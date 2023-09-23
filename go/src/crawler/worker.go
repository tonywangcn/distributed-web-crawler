package crawler

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tonywangcn/distributed-web-crawler/model"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
	"github.com/tonywangcn/distributed-web-crawler/pkg/redis"
)

var DB_INSERT_BATCH_COUNT int64

func init() {
	count := os.Getenv("DB_INSERT_BATCH_COUNT")
	DB_INSERT_BATCH_COUNT, err = strconv.ParseInt(count, 10, 0)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if DB_INSERT_BATCH_COUNT < 100 {
		DB_INSERT_BATCH_COUNT = 100
	}

}

type Worker struct {
	wg   *sync.WaitGroup
	jobs chan *model.Content
}

func NewWorker() *Worker {
	return &Worker{
		wg:   &sync.WaitGroup{},
		jobs: make(chan *model.Content, 1000000),
	}
}

func RunWorker(n int) {
	worker := NewWorker()
	go func() {
		for {
			if len(worker.jobs) > 10000 {
				time.Sleep(time.Millisecond * 200)
				continue
			}
			val := redis.LPop(GO_CRAWLER_RESULT_QUEUE)
			if val != "" {
				var content model.Content
				if err := json.Unmarshal([]byte(val), &content); err != nil {
					log.Error("failed to unmarshal data %s, err:%s", val, err.Error())
					continue
				}
				if len(content.URL) > 0 && len(content.Domain) > 0 {
					worker.AddJob(&content)
				} else {
					log.Error("invalid content %s, %+v", val, content)
				}

			}
			time.Sleep(time.Millisecond * 20)
		}
	}()
	worker.run(n)
}

func (w *Worker) AddJob(content *model.Content) {
	w.jobs <- content
}

func (w *Worker) run(count int) {
	w.wg.Add(1)
	if count < 2 {
		count = 2
	}
	for i := 0; i < count; i++ {
		go func() {
			ticker := time.NewTicker(time.Second * 50)
			data := make([]interface{}, 0)
			for {
				select {
				case <-ticker.C:
					log.Debug("time is ticking ! syncing data %d to MongoDB", len(data))
					if len(data) > 0 {
						w.worker(data)
						data = make([]interface{}, 0)
					}
				case job := <-w.jobs:
					data = append(data, *job)
					if len(data) >= int(DB_INSERT_BATCH_COUNT) {
						w.worker(data)
						data = make([]interface{}, 0)
					}
				}
			}
		}()
	}
	w.wg.Wait()
}

func (w *Worker) worker(jobs []interface{}) {
	var retries = 3
	for {
		if retries <= 0 {
			return
		}
		if err := model.InsertManyContents(jobs); err != nil {
			log.Error(err.Error())
		} else {
			return
		}
		retries -= 1
	}

}
