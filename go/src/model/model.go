package model

import (
	"context"
	"os"

	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	Setup()
}

var DB *mongo.Database
var content *mongo.Collection
var counter *mongo.Collection
var ctx = context.TODO()

func Setup() {
	log.Info("Initializing MongoDB ")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		panic(err.Error())
	}
	if err = client.Ping(ctx, nil); err != nil {
		panic(err.Error())
	}
	DB = client.Database(os.Getenv("MONGO_DBNAME"))
	content = DB.Collection("content")
	counter = DB.Collection("counter")
	log.Info("Mongodb is initialized")
	buildIndex()
}

func buildIndex() {
	var err error
	_, err = content.Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"domain": 1,
			},
			Options: options.Index().SetName("domain_1"),
		})
	if err != nil {
		log.Error("failed to create index for domain, err:%s", err.Error())
		panic(err)
	}
	_, err = content.Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"hostname": 1,
			},
			Options: options.Index().SetName("hostname_1"),
		})
	if err != nil {
		log.Error("failed to create index for hostname, err:%s", err.Error())
		panic(err)
	}
}
