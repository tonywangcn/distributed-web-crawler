package redis

import (
	"os"

	redisbloom "github.com/RedisBloom/redisbloom-go"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
)

var rb *redisbloom.Client

func init() {
	pass := os.Getenv("REDIS_PASS")
	rb = redisbloom.NewClient(os.Getenv("REDIS_HOST")+":"+os.Getenv("REDIS_PORT"), "bloom", &pass)
}

func BloomReserve(key string, error_rate float64, capacity uint64) error {
	return rb.Reserve(key, error_rate, capacity)
}

func BloomAdd(key string, item string) (bool, error) {
	ok, err := rb.Add(key, item)
	if err != nil {
		log.Error("failed to add item %s to bloom key %s, err: %s", item, key, err.Error())
		return false, err
	}
	return ok, nil
}
