// Copyright@daidai53 2023
package integration

import (
	"bytes"
	"encoding/json"
	dao2 "github.com/daidai53/webook/interactive/repository/dao"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/integration/startup"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ArticleHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

func (a *ArticleHandlerSuite) TearDownTest() {
	err := a.db.Exec("truncate table `articles`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("truncate table `published_articles`").Error
	assert.NoError(a.T(), err)
}

func (a *ArticleHandlerSuite) SetupSuite() {
	a.db = startup.InitDB()
	hdl := startup.InitArticleHandler(dao.NewArticleGormDAO(a.db), dao2.NewGORMInteractiveDAO(a.db),
		dao.NewUserDAO(a.db))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user-id", int64(123))
	})
	hdl.RegisterRoutes(server)
	a.server = server
}

func (a *ArticleHandlerSuite) TestEdit() {
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
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.db.Where("author_id = ?", 123).First(&art).Error
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
				err := a.db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.db.Where("id = ?", 2).First(&art).Error
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
				err := a.db.Create(dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 要验证数据没有变
				var art dao.Article
				err := a.db.Where("id = ?", 3).First(&art).Error
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
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (a *ArticleHandlerSuite) TestPublish() {
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
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				assert.True(t, art.Id > 0)
				assert.Equal(t, "我的标题", art.Title)
				assert.Equal(t, "我的内容", art.Content)
				assert.Equal(t, int64(123), art.AuthorId)
				assert.Equal(t, domain.ArticleStatusPublished.ToUint8(), art.Status)
				var pArt dao.PublishedArticle
				err = a.db.Where("author_id=?", 123).First(&pArt).Error
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
				err := a.db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.db.Where("id = ?", 2).First(&art).Error
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
				err = a.db.Where("id = ?", 2).First(&pArt).Error
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
				err := a.db.Create(dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(dao.PublishedArticle{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.db.Where("id = ?", 3).First(&art).Error
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
				err = a.db.Where("id = ?", 3).First(&pArt).Error
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
				err := a.db.Create(dao.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(dao.PublishedArticle{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    456,
					Utime:    789,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 要验证保存到了数据库里
				var art dao.Article
				err := a.db.Where("id = ?", 4).First(&art).Error
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
				err = a.db.Where("id = ?", 4).First(&pArt).Error
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
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestArticleHandler(t *testing.T) {
	suite.Run(t, &ArticleHandlerSuite{})
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
