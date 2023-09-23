package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredisSync "github.com/go-redsync/redsync/v4/redis"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
)

var Redis *redis.Client
var pool goredisSync.Pool
var rs *redsync.Redsync
var lockKey = "lock:"
var ctx = context.TODO()

func init() {
	log.Info("Initializing Redis: %s", os.Getenv("REDIS_HOST"))
	option := &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0,
	}

	Redis = redis.NewClient(option)
	pong, err := Redis.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Errorf("failed to inititialize Redis %s", err.Error()))
	}
	pool = goredis.NewPool(Redis)
	rs = redsync.New(pool)
	log.Info("Redis Initialized %s", pong)
}

func GetLock(key string) (*redsync.Mutex, error) {
	mutex := rs.NewMutex(lockKey + key)
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	return mutex, nil
}

func ReleaseLock(key string) error {
	mutex := rs.NewMutex(lockKey + key)
	if ok, err := mutex.Unlock(); !ok || err != nil {
		return err
	}
	return nil
}

func Get(key string) (string, error) {
	return Redis.Get(ctx, key).Result()
}

func Del(key string) error {
	return Redis.Del(ctx, key).Err()
}

func Set(key string, value string) error {

	return Redis.Set(ctx, key, value, 0).Err()
}

func SetEx(key string, value string, ex time.Duration) error {
	return Redis.SetEx(ctx, key, value, ex).Err()
}

func LPush(key string, members ...interface{}) error {
	return Redis.LPush(ctx, key, members...).Err()
}

func LRange(key string) []string {

	return Redis.LRange(ctx, key, 0, -1).Val()
}

func LLen(key string) int64 {
	return Redis.LLen(ctx, key).Val()
}

func LPop(key string) string {
	val, err := Redis.LPop(ctx, key).Result()
	if err != nil {
		return ""
	}
	return val
}

func Incr(key string) error {
	return Redis.Incr(ctx, key).Err()
}

func Decr(key string) error {
	return Redis.Decr(ctx, key).Err()
}

func HIncryBy(key string, field string, incr int64) error {
	return Redis.HIncrBy(ctx, key, field, incr).Err()
}

func HGetAll(key string) map[string]string {
	return Redis.HGetAll(ctx, key).Val()
}

func HDel(key, val string) error {
	return Redis.HDel(ctx, key, val).Err()
}

func SIsMember(key string, value string) bool {
	return Redis.SIsMember(ctx, key, value).Val()
}

func SAdd(key string, members ...interface{}) error {
	return Redis.SAdd(ctx, key, members...).Err()
}

func Exists(key string) bool {
	res, err := Redis.Exists(ctx, key).Result()
	if err != nil {
		log.Error("failed to run redis cmd exists on key %s, err:%s", key, err.Error())
		return false
	}
	if res == 0 {
		return false
	}
	return true
}

func Rename(oldKey, newKey string) error {
	return Redis.Rename(ctx, oldKey, newKey).Err()
}

func HGetAndDel(key, val string) (string, error) {
	var get *redis.StringCmd

	_, err := Redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		get = pipe.HGet(ctx, key, val)
		pipe.HDel(ctx, key, val)
		return nil
	})
	if err != nil {
		return "", err
	}
	if get.Err() == nil {
		return get.Val(), nil
	}
	return "", get.Err()

}
