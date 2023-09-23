import redis
import hashlib
import csv
import sys
from dotenv import dotenv_values
from urllib.parse import urljoin, urlparse

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
GO_CRAWLER_REQUEST_HOSTNAME_STATS = "go-crawler:request:hostname:stats"


def del_redis_key(key):
    r.delete(key)


def scan_redis(key):
    return r.scan_iter(key)


def gen_go_crawler_task(url):
    r.rpush(GO_CRAWLER_TASK_QUEUE, url)


def get_domain_stats():
    return r.hgetall(OUT_LINK_HOST_COUNTER)


def get_stats(key):
    return r.hgetall(key)


def get_queue(queue):
    return r.lrange(queue, 0, -1)


def cleanup_redis():
    del_redis_key(CRAWLER_BLOOM_KEY)  # delete bloom
    del_redis_key(OUT_LINK_HOST_COUNTER)  # delete host counter
    del_redis_key(OUT_LINK_QUEUE)
    for i in scan_redis("crawler:*"):
        del_redis_key(i.decode("utf-8"))

    for i in scan_redis("go-crawler:*"):
        del_redis_key(i.decode("utf-8"))


def ext_filter(url):
    path = urlparse(url).path
    if len(path) == 0:
        return True
    if len(path.split(".")) <= 1:
        return True
    end = "." + path.split(".")[-1]
    valid_ext_list = [
        ".html",
        ".htm",
        ".aspx",
        ".php",
        ".stm",
        ".cms",
        ".app",
        ".asp",
        ".shtml",
        ".cfm",
    ]
    ext_list = [
        ".pdf",
        ".png",
        ".xml",
        ".doc",
        ".docx",
        ".jpg",
        ".jpeg",
        ".gif",
        ".cfg",
        ".zip",
        ".xls",
        ".xlsx",
        ".rss",
    ]
    if end.lower() in valid_ext_list:
        return False
    if end.lower() in ext_list:
        return True
    print(url, end)
    return False


def filter_url():
    data = get_queue(OUT_LINK_QUEUE)
    print(len(list(filter(lambda i: not ext_filter(i.decode("utf-8")), data))))
    print(len(data))


if __name__ == "__main__":
    # gen_go_crawler_task("https://www.bbc.co.uk/")
    # cleanup_redis()
    print(get_stats(GO_CRAWLER_REQUEST_HOSTNAME_STATS))
    # print(get_stats(OUT_LINK_HOST_COUNTER))
