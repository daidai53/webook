// Copyright@daidai53 2023
package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

type MongoDBTestSuite struct {
	suite.Suite
	col *mongo.Collection
}

func (s *MongoDBTestSuite) SetupSuite() {
	t := s.T()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:example@localhost:27017/").SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)
	cols := client.Database("webook").Collection("articles")
	s.col = cols

	manyRes, err := cols.InsertMany(ctx, []any{Article{
		Id:       123,
		AuthorId: 11,
	}, Article{Id: 234, AuthorId: 22}})
	assert.NoError(t, err)
	t.Log("插入数量", len(manyRes.InsertedIDs))
}

func (s *MongoDBTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *MongoDBTestSuite) TestOr() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.A{bson.D{bson.E{"id", 123}}, bson.D{bson.E{"id", 234}}}
	res, err := s.col.Find(ctx, bson.D{bson.E{"$or", filter}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestAnd() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.A{bson.D{bson.E{"id", 123}}, bson.D{bson.E{"author_id", 11}}}
	res, err := s.col.Find(ctx, bson.D{bson.E{"$and", filter}})
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func (s *MongoDBTestSuite) TestIn() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	filter := bson.D{bson.E{"id",
		bson.D{bson.E{"$in", []int{123, 234}}}}}
	res, err := s.col.Find(ctx, filter)
	assert.NoError(s.T(), err)
	var arts []Article
	err = res.All(ctx, &arts)
	assert.NoError(s.T(), err)
	s.T().Log("查询结果", arts)
}

func TestMongoDBQueries(t *testing.T) {
	suite.Run(t, &MongoDBTestSuite{})
}
