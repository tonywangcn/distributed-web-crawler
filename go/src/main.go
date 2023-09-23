package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/tonywangcn/distributed-web-crawler/crawler"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
)

var workerMap map[string]func(count int) = map[string]func(count int){
	"worker":  crawler.RunWorker,
	"crawler": crawler.Scrape,
}

func main() {
	shutdown := make(chan int)
	// Catch interuption or termination signal
	signal.Notify(crawler.Sigchan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	defer func() {
		if r := recover(); r != nil {
			log.Error("Panicing %s!", string(debug.Stack()))
			shutdown <- 1
			log.Info("Shutting down... because of panic")
			// Persist any data in memory to Redis or Database
			log.Info("Shutted down... because of panic")
		}
	}()

	go func() {
		<-crawler.Sigchan
		log.Info("Shutting down...")
		// Persist any data in memory to Redis or Database
		log.Info("Shutted down...")
		shutdown <- 1
	}()

	var worker string
	var count int

	flag.StringVar(&worker, "w", "crawler", "Choose the worker to run")
	flag.IntVar(&count, "c", 2, "Number of workers to run")

	flag.Parse()

	f, ok := workerMap[worker]
	if !ok {
		panic("worker not found")
	}
	go f(count)

	<-shutdown
}
