// Copyright@daidai53 2023
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/bwmarrin/snowflake"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/integration/startup"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ArticleMongoDBHandlerSuite struct {
	suite.Suite
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
	server  *gin.Engine
}

func (a *ArticleMongoDBHandlerSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := a.col.DeleteMany(ctx, bson.D{})
	assert.NoError(a.T(), err)
	_, err = a.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(a.T(), err)
}

func (a *ArticleMongoDBHandlerSuite) SetupSuite() {
	a.mdb = startup.InitMongoDB()
	a.col = a.mdb.Collection("articles")
	a.liveCol = a.mdb.Collection("published_articles")
	node, err := snowflake.NewNode(1)
	assert.NoError(a.T(), err)
	hdl := startup.InitArticleHandler(dao.NewMongoDBArticleDAO(a.mdb, node), nil, nil)
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user-id", int64(123))
	})
	hdl.RegisterRoutes(server)
	a.server = server
}

func (a *ArticleMongoDBHandlerSuite) TestEdit() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		art Article

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建帖子",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"author_id", 123}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, domain.ArticleStatusUnpublished.ToUint8(), art.Status)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望文章返回的Id是1
				Data: 1,
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := a.col.InsertOne(ctx, dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"id", 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望文章返回的Id是1
				Data: 2,
			},
		},
		{
			name: "修改帖子 - 修改别人的帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := a.col.InsertOne(ctx, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 要验证数据没有变
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"id", 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Ctime:    456,
					Utime:    789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望文章返回的Id是1
				Msg: "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			if tc.wantRes.Data > 0 {
				assert.True(t, res.Data > 0)
			}
		})
	}
}

func (a *ArticleMongoDBHandlerSuite) TestPublish() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		art Article

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建帖子并发表成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"author_id", 123}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				var pArt dao.PublishedArticle
				err = a.col.FindOne(ctx, bson.D{bson.E{"author_id", 123}}).Decode(&pArt)
				assert.NoError(t, err)
				assert.True(t, pArt.Ctime > 0)
				assert.True(t, pArt.Utime > 0)
				assert.True(t, pArt.Id > 0)
				assert.Equal(t, "我的标题", pArt.Title)
				assert.Equal(t, "我的内容", pArt.Content)
				assert.Equal(t, int64(123), pArt.AuthorId)
				assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), pArt.Status)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望文章返回的Id是1
				Data: 1,
			},
		},
		{
			name: "更新帖子并首次发表成功",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := a.col.InsertOne(ctx, dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"id", 2}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
				var pArt dao.PublishedArticle
				err = a.liveCol.FindOne(ctx, bson.D{bson.E{"id", 2}}).Decode(&pArt)
				assert.NoError(t, err)
				assert.True(t, pArt.Utime > 0)
				assert.True(t, pArt.Ctime > 0)
				pArt.Utime = 0
				pArt.Ctime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, pArt)
			},
			art: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望文章返回的Id是1
				Data: 2,
			},
		},
		{
			name: "更新帖子并重新发表成功",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := a.col.InsertOne(ctx, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
				_, err = a.liveCol.InsertOne(ctx, dao.PublishedArticle{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"id", 3}}).Decode(&art)
				assert.NoError(t, err)
				assert.True(t, art.Utime > 789)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
				var pArt dao.PublishedArticle
				err = a.col.FindOne(ctx, bson.D{bson.E{"id", 3}}).Decode(&pArt)
				assert.NoError(t, err)
				assert.True(t, pArt.Utime > 789)
				pArt.Utime = 0
				assert.Equal(t, dao.PublishedArticle{
					Id:       3,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorId: 123,
					Ctime:    456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, pArt)
			},
			art: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				// 希望文章返回的Id是1
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子，并发表失败",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := a.col.InsertOne(ctx, dao.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
				_, err = a.liveCol.InsertOne(ctx, dao.PublishedArticle{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				})
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.col.FindOne(ctx, bson.D{bson.E{"id", 4}}).Decode(&art)
				assert.NoError(t, err)

				assert.Equal(t, dao.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Ctime:    456,
					Utime:    789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, art)
				var pArt dao.PublishedArticle
				err = a.liveCol.FindOne(ctx, bson.D{bson.E{"id", 4}}).Decode(&pArt)
				assert.NoError(t, err)
				assert.Equal(t, dao.PublishedArticle{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Ctime:    456,
					Utime:    789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}, pArt)
			},
			art: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			if tc.wantRes.Data > 0 {
				assert.True(t, res.Data > 0)
			}
		})
	}
}

func TestArticleMongoDBHandler(t *testing.T) {
	suite.Run(t, &ArticleMongoDBHandlerSuite{})
}
