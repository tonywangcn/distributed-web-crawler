package crawler

import (
	"bytes"
	"encoding/json"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/gocolly/redisstorage"
	"github.com/tonywangcn/distributed-web-crawler/model"
	"github.com/tonywangcn/distributed-web-crawler/pkg/crypto"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
	"github.com/tonywangcn/distributed-web-crawler/pkg/redis"
	"github.com/tonywangcn/distributed-web-crawler/pkg/robots"
	"github.com/tonywangcn/distributed-web-crawler/pkg/utils"
)

const GO_CRAWLER_TASK_QUEUE = "go-crawler:task:queue"
const GO_CRAWLER_RESULT_QUEUE = "go-crawler:result:queue"
const GO_CRAWLER_REQUEST_STATS = "go-crawler:request:stats"
const GO_CRAWLER_REQUEST_HOSTNAME_STATS = "go-crawler:request:hostname:stats"
const CRAWLER_BLOOM_KEY = "crawler:bloom"
const OUT_LINK_QUEUE = "crawler:outlink:queue"
const OUT_LINK_HOST_COUNTER = "outlink:host:counter"

var CRAWLER_MAX_RETRIES int64 = 3

var DOMAIN_STATS = make(map[string]int64)
var err error

var Sigchan = make(chan os.Signal, 1)

func Exit() {
	Sigchan <- syscall.SIGTERM
}

func init() {
	if err != nil {
		panic(err)
	}
	reserveBloomKey()
	go func() {
		for {
			time.Sleep(time.Minute * 10)
			SyncCounterToMongo()

		}
	}()
	retries, exists := os.LookupEnv("CRAWLER_MAX_RETRIES")
	var maxRetries int64
	if exists {
		maxRetries, err = strconv.ParseInt(retries, 10, 0)
		if err != nil {
			log.Error("invalid CRAWLER_MAX_RETRIES %s", retries)
			return
		}
		CRAWLER_MAX_RETRIES = maxRetries
	}
}

func SyncCounterToMongo() {
	stat := redis.HGetAll(OUT_LINK_HOST_COUNTER)
	log.Info("total %d domains' stats is going be synced!", len(stat))
	if len(stat) == 0 {
		log.Info("counter is empty!")
		return
	}
	for k, v := range stat {
		count, err := redis.HGetAndDel(OUT_LINK_HOST_COUNTER, k)
		if err != nil {
			log.Error("failed to get and delete hash key %s by field %s", OUT_LINK_HOST_COUNTER, k)
			continue
		}
		c, err := strconv.ParseInt(count, 10, 0)
		if err != nil {
			log.Error("failed to parse string %s to int", v)
			redis.HDel(OUT_LINK_HOST_COUNTER, k)
			continue
		}

		counter := &model.Counter{}
		counter.Hostname = k
		counter.Count = c
		var retries = 3

		// should retry n times if upsert operation failed to make sure no data loss when sync stats to MongoDB

		for {
			if retries < 0 {
				break
			}
			if err := counter.Upsert(); err != nil {
				log.Error("syncCounterToMongo: failed to upsert, hostname %s, count %d, err:%s", k, c, err.Error())
			} else {
				break
			}
			time.Sleep(time.Second * 1)
			retries -= 1
		}

		time.Sleep(time.Millisecond * 20)
	}

}

// Reserve bloom key if not exist. With error rate  0.0000001 and total 1000000000 items, the memory usage is 3.9 GB.
func reserveBloomKey() {
	if ok := redis.Exists(CRAWLER_BLOOM_KEY); ok {
		return
	}
	if err := redis.BloomReserve(CRAWLER_BLOOM_KEY, 0.0000001, 1000000000); err != nil {
		log.Error("failed to reserve redis bloom key %s, err:%s", CRAWLER_BLOOM_KEY, err.Error())
		return
	}
	log.Info("redis bloom key %s is reserved successfully!", CRAWLER_BLOOM_KEY)
}

func Scrape(n int) {
	for {
		scrape()
		time.Sleep(time.Second * 10)
	}
}

// the main function of scraping.
func scrape() {

	// recieve task from redis queue
	u := redis.LPop(GO_CRAWLER_TASK_QUEUE)
	if len(u) == 0 {
		log.Error("invalid url %s", u)
		return
	}

	// parse hostname from url
	var hostname = utils.GetHostname(u)

	// check hostname if valid
	if !utils.IsValidHostname(hostname) {
		log.Error("illegal url %s, hostname %s", u, hostname)
		return
	}
	// read robots.txt file and check url is allowed
	robo := robots.New("http://" + hostname)
	if !robo.AgentAllowed("GoogleBot", u) {
		log.Error("URL is not allowed to visit in robots.txt. URL: %s", u)
		return
	}

	var c = colly.NewCollector(
		colly.UserAgent(os.Getenv("GO_BOT_UA")),
		colly.AllowURLRevisit(),
	)
	c.Limit(&colly.LimitRule{
		RandomDelay: 2 * time.Second,
	})

	// create the redis storage
	storage := &redisstorage.Storage{
		Address:  os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
		Prefix:   "crawler:" + hostname,
	}

	// add storage to the collector
	err = c.SetStorage(storage)
	if err != nil {
		log.Error(err.Error())
		panic(err)
	}

	// close redis client
	defer storage.Client.Close()

	// create a new request queue with redis storage backend
	q, _ := queue.New(50, storage)

	// Process any errors caused by timeout, status code >= 400
	// Exponential Backoff, use E as the base.
	c.OnError(func(r *colly.Response, err error) {
		retriesLeft := CRAWLER_MAX_RETRIES
		if x, ok := r.Ctx.GetAny("retriesLeft").(int64); ok {
			retriesLeft = x
		}

		log.Error("error %s |  retriesLeft %d", err.Error(), retriesLeft)

		if retriesLeft > 0 {
			r.Ctx.Put("retriesLeft", retriesLeft-1)
			time.Sleep(time.Duration(math.Exp(float64(CRAWLER_MAX_RETRIES-retriesLeft+1))) * time.Second)
			r.Request.Retry()
		}

	})
	c.OnResponse(func(r *colly.Response) {
		if err = redis.HIncryBy(GO_CRAWLER_REQUEST_HOSTNAME_STATS, hostname, 1); err != nil {
			log.Error(err.Error())
		}
		content := parse(r)
		content.Domain = hostname
		content.URL = r.Request.URL.String()
		b, err := json.Marshal(content)
		if err != nil {
			log.Error("failed to marshal struct to json, url: %s", content.URL)
			return
		}
		if err := redis.LPush(GO_CRAWLER_RESULT_QUEUE, b); err != nil {
			log.Error(err.Error())
		}

	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		// get the url of the element
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if len(link) == 0 {
			return
		}

		// read robots.txt file and check url is allowed
		if !robo.AgentAllowed("GoogleBot", link) {
			log.Error("URL is not allowed to visit in robots.txt. URL: %s", link)
			return
		}

		u, err := url.ParseRequestURI(link)
		if err != nil {
			log.Error("illegal url %s, err:%s", link, err.Error())
			return
		}
		if !utils.IsValidHostname(u.Hostname()) {
			log.Error("illegal url %s, hostname %s", link, u.Hostname())
			return
		}
		if !utils.IsValidPath(u.Path) {
			log.Error("illegal path %s, hostname %s", u.Path, u.Hostname())
			return
		}
		link = utils.CleanUpUrlParam(u)
		ok, err := redis.BloomAdd(CRAWLER_BLOOM_KEY, crypto.Md5(strings.Replace(strings.TrimSuffix(link, "/"), "http://", "https://", 1)))
		if err != nil {
			log.Error(err.Error())
			return
		}
		if !ok {
			log.Debug("URL has been crawled already, %s", u)
			return
		}
		log.Info("Found new Url %s", u)

		// prioritize links in the same domain
		if u.Hostname() == hostname {
			q.AddURL(link)
		} else {
			if err = redis.LPush(OUT_LINK_QUEUE, link); err != nil {
				log.Error(err.Error())
			}
			if err := redis.HIncryBy(OUT_LINK_HOST_COUNTER, u.Hostname(), 1); err != nil {
				log.Error(err.Error())
			}
		}

	})

	c.OnRequest(func(r *colly.Request) {
		log.Info("Visiting %s", r.URL.String())
	})

	// add URLs to the queue
	q.AddURL(u)
	// consume requests
	q.Run(c)
}

func parse(r *colly.Response) *model.Content {
	content := &model.Content{}
	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(r.Body))
	title := doc.Find("h1").Text()
	if len(title) > 0 {
		content.Title = title
	}
	var desc string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if val, _ := s.Attr("name"); strings.Contains(val, "description") {
			desc, _ = s.Attr("content")
		}
		if len(desc) == 0 {
			if val, _ := s.Attr("property"); strings.Contains(val, "description") {
				desc, _ = s.Attr("content")
			}
		}
	})
	if len(desc) > 0 {
		content.Desc = desc
	}
	return content
}
