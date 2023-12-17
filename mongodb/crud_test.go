// Copyright@daidai53 2023
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestMongoDB(t *testing.T) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)
	cols := client.Database("webook").Collection("articles")

	insertRes, err := cols.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
	})
	assert.NoError(t, err)
	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log(string(oid[:12]))

	filter := bson.D{bson.E{"id", 1}}
	findRes := cols.FindOne(ctx, filter)
	if errors.Is(findRes.Err(), mongo.ErrNoDocuments) {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	updateFilter := bson.D{bson.E{"id", 1}}
	set := bson.D{bson.E{"$set", bson.E{"title", "新的标题"}}}
	updateOneRes, err := cols.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	updateManyRes, err := cols.UpdateMany(ctx, updateFilter, bson.D{bson.E{"$set", Article{
		Id:      1,
		Content: "新的内容",
	}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)
	deleterFilter := bson.D{bson.E{"id", 1}, bson.E{"id", 0}}
	delRes, err := cols.DeleteMany(ctx, deleterFilter)
	assert.NoError(t, err)
	t.Log("删除文档数量", delRes.DeletedCount)
}

type Article struct {
	Id       int64  `bson:"id,omitempty"`
	Title    string `bson:"title,omitempty"`
	Content  string `bson:"content"`
	AuthorId int64  `bson:"author_id"`
	Status   uint8  `bson:"status"`
	Ctime    int64  `bson:"ctime"`
	Utime    int64  `bson:"utime"`
}
