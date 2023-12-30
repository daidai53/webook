// Copyright@daidai53 2023
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/integration/startup"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/daidai53/webook/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type ArticleInteractiveHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	cmd    redis.Cmdable
	server *gin.Engine
}

func (a *ArticleInteractiveHandlerSuite) TearDownTest() {
	err := a.db.Exec("truncate table `users`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("truncate table `articles`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("truncate table `published_articles`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("truncate table `interactives`").Error
	assert.NoError(a.T(), err)
	err = a.db.Exec("truncate table `user_like_bizs`").Error
	assert.NoError(a.T(), err)
}

func (a *ArticleInteractiveHandlerSuite) SetupSuite() {
	a.db = startup.InitDB()
	a.cmd = startup.InitRedis()
	hdl := startup.InitArticleHandler(dao.NewArticleGormDAO(a.db), dao.NewGORMInteractiveDAO(a.db),
		dao.NewUserDAO(a.db))
	server := gin.Default()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user-id", int64(123))
	})
	hdl.RegisterRoutes(server)
	a.server = server
}

func (a *ArticleInteractiveHandlerSuite) TestIncrReadCnt() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		artId string

		wantCode int
		wantRes  Result[web.ArticleVo]
	}{
		{
			name: "增加成功，原本都有记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.Create(&dao.User{
					Id:       234,
					Nickname: "作者",
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.PublishedArticle{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 234,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.Interactive{
					Id:      1,
					BizId:   1,
					Biz:     "article",
					ReadCnt: 2,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
				a.cmd.HIncrBy(ctx, "interactive:article:article:1", "read_cnt", 2)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var inter dao.Interactive
				err := a.db.WithContext(ctx).Where("id=?", 1).First(&inter).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(3), inter.ReadCnt)
				assert.True(t, inter.UTime > 100)
				rc, err := a.cmd.HGet(ctx, "interactive:article:article:1", "read_cnt").Int64()
				assert.NoError(t, err)
				assert.Equal(t, int64(3), rc)
				err = a.db.WithContext(ctx).Where("id=?", 1).Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:1", "read_cnt").Err()
				assert.NoError(t, err)
			},
			artId:    "1",
			wantCode: http.StatusOK,
			wantRes: Result[web.ArticleVo]{
				Data: web.ArticleVo{
					Id:         1,
					Title:      "我的标题",
					Content:    "我的内容",
					AuthorId:   234,
					AuthorName: "作者",
					Status:     domain.ArticleStatusPublished.ToUint8(),
					CTime:      time.UnixMilli(100).Format(time.DateTime),
					UTime:      time.UnixMilli(100).Format(time.DateTime),
				},
			},
		},
		{
			name: "增加成功，只有DB有记录",
			before: func(t *testing.T) {
				err := a.db.Create(&dao.User{
					Id:       345,
					Nickname: "作者",
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 345,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.PublishedArticle{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 345,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.Interactive{
					Id:      2,
					BizId:   2,
					Biz:     "article",
					ReadCnt: 2,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var inter dao.Interactive
				err := a.db.WithContext(ctx).Where("id=?", 2).First(&inter).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(3), inter.ReadCnt)
				assert.True(t, inter.UTime > 100)
				rc, err := a.cmd.HGet(ctx, "interactive:article:article:2", "read_cnt").Int64()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), rc)
				err = a.db.WithContext(ctx).Where("id=?", 2).Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:2", "read_cnt").Err()
				assert.NoError(t, err)
			},
			artId:    "2",
			wantCode: http.StatusOK,
			wantRes: Result[web.ArticleVo]{
				Data: web.ArticleVo{
					Id:    2,
					Title: "我的标题",

					Content:    "我的内容",
					AuthorId:   345,
					AuthorName: "作者",
					Status:     domain.ArticleStatusPublished.ToUint8(),
					CTime:      time.UnixMilli(100).Format(time.DateTime),
					UTime:      time.UnixMilli(100).Format(time.DateTime),
				},
			},
		},
		{
			name: "增加成功，db和缓存都无记录",
			before: func(t *testing.T) {
				err := a.db.Create(&dao.User{
					Id:       456,
					Nickname: "作者",
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}).Error
				assert.NoError(t, err)
				err = a.db.Create(&dao.PublishedArticle{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					Ctime:    100,
					Utime:    100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var inter dao.Interactive
				err := a.db.WithContext(ctx).Where("biz_id=? and biz=?", 3, "article").First(&inter).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(1), inter.ReadCnt)
				assert.True(t, inter.CTime > 0)
				assert.True(t, inter.UTime > 0)
				rc, err := a.cmd.HGet(ctx, "interactive:article:article:3", "read_cnt").Int64()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), rc)
				err = a.db.WithContext(ctx).Where("biz_id=? and biz=?", 3, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:3", "read_cnt").Err()
				assert.NoError(t, err)
			},
			artId:    "3",
			wantCode: http.StatusOK,
			wantRes: Result[web.ArticleVo]{
				Data: web.ArticleVo{
					Id:         3,
					Title:      "我的标题",
					Content:    "我的内容",
					AuthorId:   456,
					AuthorName: "作者",
					Status:     domain.ArticleStatusPublished.ToUint8(),
					CTime:      time.UnixMilli(100).Format(time.DateTime),
					UTime:      time.UnixMilli(100).Format(time.DateTime),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/articles/pub/%s", tc.artId), nil)
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[web.ArticleVo]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestLike() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		req LikeReq

		wantCode int
		wantRes  Result[any]
	}{
		{
			name: "点赞成功，DB无记录，缓存无记录",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var likeBiz dao.UserLikeBiz
				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 1, "article").First(&likeBiz).Error
				assert.NoError(t, err)
				assert.Equal(t, 1, likeBiz.Status)
				assert.True(t, likeBiz.CTime > 0)
				assert.True(t, likeBiz.UTime > 0)
				var interactive dao.Interactive
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 1, "article").First(&interactive).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(1), interactive.LikeCnt)
				assert.True(t, interactive.CTime > 0)
				assert.True(t, interactive.UTime > 0)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:1", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 1, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 1, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:1", "like_cnt").Err()
				assert.NoError(t, err)
			},
			req: LikeReq{
				Id:   1,
				Like: true,
			},
			wantCode: http.StatusOK,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
		{
			name: "点赞成功，DB有LikeBiz记录，无interactive，缓存无记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:    1,
					Uid:   123,
					Biz:   "article",
					BizId: 2,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var likeBiz dao.UserLikeBiz
				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 2, "article").First(&likeBiz).Error
				assert.NoError(t, err)
				assert.Equal(t, 1, likeBiz.Status)
				assert.True(t, likeBiz.CTime > 0)
				assert.True(t, likeBiz.UTime > 0)
				var interactive dao.Interactive
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 2, "article").First(&interactive).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(1), interactive.LikeCnt)
				assert.True(t, interactive.CTime > 0)
				assert.True(t, interactive.UTime > 0)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 2, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 2, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:2", "like_cnt").Err()
				assert.NoError(t, err)
			},
			req: LikeReq{
				Id:   2,
				Like: true,
			},
			wantCode: http.StatusOK,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
		{
			name: "点赞成功，DB有LikeBiz和interactive记录，缓存无记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:    1,
					Uid:   123,
					Biz:   "article",
					BizId: 3,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:      1,
					Biz:     "article",
					BizId:   3,
					LikeCnt: 10,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var likeBiz dao.UserLikeBiz
				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 3, "article").First(&likeBiz).Error
				assert.NoError(t, err)
				assert.Equal(t, 1, likeBiz.Status)
				assert.True(t, likeBiz.CTime > 0)
				assert.True(t, likeBiz.UTime > 0)
				var interactive dao.Interactive
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 3, "article").First(&interactive).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(11), interactive.LikeCnt)
				assert.True(t, interactive.CTime > 0)
				assert.True(t, interactive.UTime > 0)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:3", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 3, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 3, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:3", "like_cnt").Err()
				assert.NoError(t, err)
			},
			req: LikeReq{
				Id:   3,
				Like: true,
			},
			wantCode: http.StatusOK,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
		{
			name: "点赞成功，DB有LikeBiz和interactive记录，缓存有记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:    1,
					Uid:   123,
					Biz:   "article",
					BizId: 4,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:      1,
					Biz:     "article",
					BizId:   4,
					LikeCnt: 10,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
				err = a.cmd.HSet(ctx, "interactive:article:article:4", "like_cnt", 15).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var likeBiz dao.UserLikeBiz
				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 4, "article").First(&likeBiz).Error
				assert.NoError(t, err)
				assert.Equal(t, 1, likeBiz.Status)
				assert.True(t, likeBiz.CTime > 0)
				assert.True(t, likeBiz.UTime > 0)
				var interactive dao.Interactive
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 4, "article").First(&interactive).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(11), interactive.LikeCnt)
				assert.True(t, interactive.CTime > 0)
				assert.True(t, interactive.UTime > 0)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:4", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 16, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 4, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 4, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:4", "like_cnt").Err()
				assert.NoError(t, err)
			},
			req: LikeReq{
				Id:   4,
				Like: true,
			},
			wantCode: http.StatusOK,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/articles/pub/like"), bytes.NewReader(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[any]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestDisLike() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		req LikeReq

		wantCode int
		wantRes  Result[any]
	}{
		{
			name: "取消点赞成功",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:     1,
					Uid:    123,
					Biz:    "article",
					BizId:  1,
					Status: 1,
					CTime:  100,
					UTime:  100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:      1,
					Biz:     "article",
					BizId:   1,
					LikeCnt: 5,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
				err = a.cmd.HSet(ctx, "interactive:article:article:1", "like_cnt", 5).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var likeBiz dao.UserLikeBiz
				affected := a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 1, "article").First(&likeBiz).RowsAffected
				assert.Equal(t, int64(0), affected)
				var interactive dao.Interactive
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 1, "article").First(&interactive).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(4), interactive.LikeCnt)
				assert.True(t, interactive.UTime > 100)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:1", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 1, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 1, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:1", "like_cnt").Err()
				assert.NoError(t, err)
			},
			req: LikeReq{
				Id:   1,
				Like: false,
			},
			wantCode: http.StatusOK,
			wantRes: Result[any]{
				Msg: "OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			reqBody, err := json.Marshal(tc.req)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("/articles/pub/like"), bytes.NewReader(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			a.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}

			var res Result[any]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestArticleInteractiveHandler(t *testing.T) {
	suite.Run(t, &ArticleInteractiveHandlerSuite{})
}

type LikeReq struct {
	Id   int64 `json:"id"`
	Like bool  `json:"like"`
}
