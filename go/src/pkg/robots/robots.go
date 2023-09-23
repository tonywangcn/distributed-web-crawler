package robots

import (
	"io"
	"net/http"

	"github.com/jimsmart/grobotstxt"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
)

type RobotsTxt struct {
	txt string
}

func read(uri string) (string, error) {
	resp, err := http.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil

}

func New(domain string) *RobotsTxt {
	txt, err := read(domain + "/robots.txt")
	if err != nil {
		log.Error("failed to read %s/robots.txt, err:%s", domain, err.Error())
		return nil
	}
	if len(txt) == 0 {
		log.Error("robots.txt is empty, url %s, err:%s", domain, err.Error())
		return nil
	}
	return &RobotsTxt{
		txt: txt,
	}
}

func (r *RobotsTxt) AgentAllowed(ua, uri string) bool {
	return grobotstxt.AgentAllowed(r.txt, "FooBot/1.0", uri)
}
