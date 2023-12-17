// Copyright@daidai53 2023
package dao

import (
	"context"
	"errors"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBArticleDAO struct {
	node    *snowflake.Node
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	art.Id = m.node.Generate().Int64()
	_, err := m.col.InsertOne(ctx, &art)
	return art.Id, err
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	filter := bson.D{bson.E{"id", art.Id}, bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", bson.M{
		"title":   art.Title,
		"content": art.Content,
		"status":  art.Status,
		"utime":   now,
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("更新失败，ID不对或作者不对")
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)
	if art.Id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	now := time.Now().UnixMilli()
	art.Utime = now
	filter := bson.D{
		bson.E{"id", art.Id},
		bson.E{"author_id", art.AuthorId},
	}
	set := bson.D{
		bson.E{"$set", PublishedArticle(art)},
		bson.E{"$setOnInsert",
			bson.D{
				bson.E{"ctime", now},
			},
		},
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
	return id, err
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, uid int64, id int64, status uint8) error {
	filter := bson.D{
		bson.E{"id", id},
		bson.E{"author_id", uid},
	}
	set := bson.D{
		bson.E{"$set", bson.D{
			bson.E{"status", status},
		}},
	}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount != 1 {
		return errors.New("更新失败，ID不对或作者不对")
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, res)
	return err
}

func NewMongoDBArticleDAO(db *mongo.Database, node *snowflake.Node) ArticleDAO {
	return &MongoDBArticleDAO{
		node:    node,
		liveCol: db.Collection("published_articles"),
		col:     db.Collection("articles"),
	}
}
