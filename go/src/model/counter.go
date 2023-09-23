package model

import (
	"fmt"
	"time"

	"github.com/tonywangcn/distributed-web-crawler/pkg/crypto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Counter struct {
	Id        string    `bson:"_id,omitempty"`
	Hostname  string    `bson:"hostname"`
	Count     int64     `bson:"count"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (m *Counter) Upsert() error {
	m.Id = crypto.Md5(m.Hostname)
	filter := bson.D{{Key: "_id", Value: m.Id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "updated_at", Value: time.Now()}}},
		{Key: "$inc", Value: bson.D{{Key: "count", Value: m.Count}}},
		{Key: "$setOnInsert", Value: bson.D{{Key: "created_at", Value: time.Now()}, {Key: "hostname", Value: m.Hostname}}},
	}
	opts := options.Update().SetUpsert(true)

	_, err := counter.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update doc %+v", m)
	}
	return nil

}
