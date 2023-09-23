import redis
import hashlib
import csv
import sys
from dotenv import dotenv_values

config = dotenv_values("../.env")
r = redis.Redis(
    host=config["REDIS_HOST"],
    port=config["REDIS_PORT"],
    password=config["REDIS_PASS"],
    db=0,
)

GO_CRAWLER_TASK_QUEUE = "go-crawler:task:queue"
OUT_LINK_HOST_COUNTER = "outlink:host:counter"
OUT_LINK_QUEUE = "crawler:outlink:queue"
CRAWLER_BLOOM_KEY = "crawler:bloom"
GO_CRAWLER_RESULT_QUEUE = "go-crawler:result:queue"
GO_CRAWLER_REQUEST_STATS = "go-crawler:request:stats"


def del_redis_key(key):
    r.delete(key)


def scan_redis(key):
    return r.scan_iter(key)


def gen_go_crawler_task(url):
    r.rpush(GO_CRAWLER_TASK_QUEUE, url)


def get_domain_stats():
    return r.hgetall(OUT_LINK_HOST_COUNTER)


def cleanup_redis():
    del_redis_key(CRAWLER_BLOOM_KEY)  # delete bloom
    del_redis_key(OUT_LINK_HOST_COUNTER)  # delete host counter
    del_redis_key(OUT_LINK_QUEUE)
    for i in scan_redis("crawler:*"):
        del_redis_key(i.decode("utf-8"))

    for i in scan_redis("go-crawler:*"):
        del_redis_key(i.decode("utf-8"))


if __name__ == "__main__":
    gen_go_crawler_task("https://www.bbc.co.uk/")
    # cleanup_redis()
    # print(get_domain_stats())
