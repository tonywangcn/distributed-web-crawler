package model

import (
	"fmt"
	"time"

	"github.com/tonywangcn/distributed-web-crawler/pkg/crypto"
	"github.com/tonywangcn/distributed-web-crawler/pkg/log"
)

type Content struct {
	Id            string    `bson:"_id,omitempty" json:"_id,omitempty"`
	Domain        string    `bson:"domain" json:"domain"`
	URL           string    `bson:"url" json:"url"`
	Title         string    `bson:"title" json:"title"`
	Desc          string    `bson:"desc" json:"desc"`
	Author        string    `bson:"author" json:"author"`
	RawHtml       string    `bson:"raw_html" json:"raw_html"`
	Content       string    `bson:"content" json:"content"`
	Text          string    `bson:"Text" json:"text"`
	JsRender      bool      `bson:"js_render" json:"js_render"`
	DatePublished time.Time `bson:"date_published" json:"date_published"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
}

func (m *Content) Insert() error {
	m.Id = crypto.Md5(m.URL)
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	_, err := content.InsertOne(ctx, m)
	if err != nil {
		return fmt.Errorf("failed to insert doc %+v, err:%s", m, err.Error())
	}
	return nil
}

func InsertManyContents(c []interface{}) error {

	res, err := content.InsertMany(ctx, c)

	if err != nil {
		return fmt.Errorf("failed to insert many docs, err:%s", err.Error())
	}
	log.Debug("Inserted count %d", len(res.InsertedIDs))
	if len(res.InsertedIDs) != len(c) {
		log.Error("Total inserted count %d, total %d, missing %d", len(res.InsertedIDs), len(c), len(c)-len(res.InsertedIDs))
	}

	return nil

}
