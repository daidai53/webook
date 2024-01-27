// Copyright@daidai53 2023
package integration

import (
	"context"
	"errors"
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	"github.com/daidai53/webook/interactive/integration/startup"
	"github.com/daidai53/webook/interactive/repository/dao"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
)

type ArticleInteractiveHandlerSuite struct {
	suite.Suite
	db  *gorm.DB
	cmd redis.Cmdable
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
	err = a.cmd.FlushDB(context.Background()).Err()
	assert.NoError(a.T(), err)
}

func (a *ArticleInteractiveHandlerSuite) SetupSuite() {
	a.db = startup.InitDB()
	a.cmd = startup.InitRedis()
}

func (a *ArticleInteractiveHandlerSuite) TestIncrReadCnt() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64

		wantErr  error
		wantResp *interv1.IncrReadCntResponse
	}{
		{
			name: "增加成功，原本都有记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.Create(&dao.Interactive{
					Id:      5,
					BizId:   5,
					Biz:     "article",
					ReadCnt: 2,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
				a.cmd.HSet(ctx, "interactive:article:article:5", "read_cnt", 2)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var inter dao.Interactive
				err := a.db.WithContext(ctx).Where("id=?", 5).First(&inter).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(3), inter.ReadCnt)
				assert.True(t, inter.UTime > 100)
				rc, err := a.cmd.HGet(ctx, "interactive:article:article:5", "read_cnt").Int64()
				assert.NoError(t, err)
				assert.Equal(t, int64(3), rc)
				err = a.db.WithContext(ctx).Where("id=?", 5).Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:5", "read_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    5,
			wantResp: &interv1.IncrReadCntResponse{},
		},
		{
			name: "增加成功，只有DB有记录",
			before: func(t *testing.T) {
				err := a.db.Create(&dao.Interactive{
					Id:      6,
					BizId:   6,
					Biz:     "article",
					ReadCnt: 3,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var inter dao.Interactive
				err := a.db.WithContext(ctx).Where("id=?", 6).First(&inter).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(4), inter.ReadCnt)
				assert.True(t, inter.UTime > 100)
				rc, err := a.cmd.HGet(ctx, "interactive:article:article:6", "read_cnt").Int64()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), rc)
				err = a.db.WithContext(ctx).Where("id=?", 6).Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:6", "read_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    6,
			wantResp: &interv1.IncrReadCntResponse{},
		},
		{
			name:   "增加成功，db和缓存都无记录",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var inter dao.Interactive
				err := a.db.WithContext(ctx).Where("biz_id=?", 7).First(&inter).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(1), inter.ReadCnt)
				assert.True(t, inter.UTime > 100)
				rc, err := a.cmd.HGet(ctx, "interactive:article:article:7", "read_cnt").Int64()
				assert.NoError(t, err)
				assert.Equal(t, int64(1), rc)
				err = a.db.WithContext(ctx).Where("id=?", 7).Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:7", "read_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    7,
			wantResp: &interv1.IncrReadCntResponse{},
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			resp, err := svc.IncrReadCnt(context.Background(), &interv1.IncrReadCntRequest{
				Biz:   tc.biz,
				BizId: tc.bizId,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
			tc.after(t)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestLike() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr  error
		wantResp *interv1.LikeResponse
	}{
		{
			name: "点赞成功，DB无记录，缓存无记录",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var likeBiz dao.UserLikeBiz
				err := a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 1, "article").First(&likeBiz).Error
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
				likeTop := a.cmd.ZScore(ctx, "article:likes:top", "1").Val()
				assert.Equal(t, float64(1), likeTop)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:1", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 1, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 1, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.ZRem(ctx, "article:likes:top", "1").Err()
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:1", "like_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    1,
			uid:      123,
			wantResp: &interv1.LikeResponse{},
		},
		{
			name: "点赞成功，DB有LikeBiz记录，无interactive，缓存无记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:    2,
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

				var likeBiz dao.UserLikeBiz
				err := a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 2, "article").First(&likeBiz).Error
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
				likeTop := a.cmd.ZScore(ctx, "article:likes:top", "2").Val()
				assert.Equal(t, float64(1), likeTop)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:2", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 2, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 2, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.ZRem(ctx, "article:likes:top", "2").Err()
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:2", "like_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    2,
			uid:      123,
			wantResp: &interv1.LikeResponse{},
		},
		{
			name: "点赞成功，DB有LikeBiz和interactive记录，缓存无记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:    3,
					Uid:   123,
					Biz:   "article",
					BizId: 3,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:      3,
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

				var likeBiz dao.UserLikeBiz
				err := a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 3, "article").First(&likeBiz).Error
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
				likeTop := a.cmd.ZScore(ctx, "article:likes:top", "3").Val()
				assert.Equal(t, float64(1), likeTop)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:3", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 3, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 3, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.ZRem(ctx, "article:likes:top", "3").Err()
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:3", "like_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    3,
			uid:      123,
			wantResp: &interv1.LikeResponse{},
		},
		{
			name: "点赞成功，DB有LikeBiz和interactive记录，缓存有记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:    4,
					Uid:   123,
					Biz:   "article",
					BizId: 4,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:      4,
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

				var likeBiz dao.UserLikeBiz
				err := a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 4, "article").First(&likeBiz).Error
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
				likeTop := a.cmd.ZScore(ctx, "article:likes:top", "4").Val()
				assert.Equal(t, float64(1), likeTop)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:4", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 16, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 4, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 4, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.ZRem(ctx, "article:likes:top", "4").Err()
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:4", "like_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    4,
			uid:      123,
			wantResp: &interv1.LikeResponse{},
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			resp, err := svc.Like(context.Background(), &interv1.LikeRequest{
				Biz: tc.biz,
				Id:  tc.bizId,
				Uid: tc.uid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestDisLike() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr  error
		wantResp *interv1.CancelLikeResponse
	}{
		{
			name: "取消点赞成功",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:     7,
					Uid:    123,
					Biz:    "article",
					BizId:  7,
					Status: 1,
					CTime:  100,
					UTime:  100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:      7,
					Biz:     "article",
					BizId:   7,
					LikeCnt: 5,
					CTime:   100,
					UTime:   100,
				}).Error
				assert.NoError(t, err)
				err = a.cmd.ZIncrBy(ctx, "article:likes:top", 5, "7").Err()
				assert.NoError(t, err)
				err = a.cmd.HSet(ctx, "interactive:article:article:7", "like_cnt", 5).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				var err error
				var likeBiz dao.UserLikeBiz
				affected := a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=? AND status=1", 123, 7, "article").First(&likeBiz).RowsAffected
				assert.Equal(t, int64(0), affected)
				var interactive dao.Interactive
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 7, "article").First(&interactive).Error
				assert.NoError(t, err)
				assert.Equal(t, int64(4), interactive.LikeCnt)
				assert.True(t, interactive.UTime > 100)
				artiLikesTop := a.cmd.ZScore(ctx, "article:likes:top", "7").Val()
				assert.Equal(t, float64(4), artiLikesTop)
				redisLikeCnt, err := a.cmd.HGet(ctx, "interactive:article:article:7", "like_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 4, redisLikeCnt)

				err = a.db.WithContext(ctx).Where("uid=? AND biz_id=? AND biz=?", 123, 7, "article").Delete(dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz_id=? AND biz=?", 7, "article").Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.ZRem(ctx, "article:likes:top", "7").Err()
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:7", "like_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    7,
			uid:      123,
			wantResp: &interv1.CancelLikeResponse{},
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			resp, err := svc.CancelLike(context.Background(), &interv1.CancelLikeRequest{
				Biz: tc.biz,
				Id:  tc.bizId,
				Uid: tc.uid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestCollect() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64
		cid   int64

		wantErr  error
		wantResp *interv1.CollectResponse
	}{
		{
			name: "添加收藏成功,DB和Redis都无记录",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var collect dao.UserCollectionBiz
				err := a.db.WithContext(ctx).Model(&dao.UserCollectionBiz{}).
					Where("biz=? AND biz_id=? AND uid=?", "article", 8, 123).First(&collect).Error
				assert.NoError(t, err)
				assert.True(t, collect.UTime > 0)
				var inter dao.Interactive
				err = a.db.WithContext(ctx).Model(&dao.Interactive{}).
					Where("biz=? AND biz_id=?", "article", 8).First(&inter).Error
				assert.NoError(t, err)
				assert.True(t, inter.CTime > 0 && inter.UTime > 0)
				assert.Equal(t, int64(1), inter.CollectCnt)
				colCnt, err := a.cmd.HGet(ctx, "interactive:article:article:8", "collect_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, colCnt)

				err = a.db.WithContext(ctx).Where("biz=? AND biz_id=? AND uid=?", "article", 8, 123).
					Delete(dao.UserCollectionBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz=? AND biz_id=?", "article", 8).
					Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:8", "collect_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    8,
			uid:      123,
			cid:      8,
			wantResp: &interv1.CollectResponse{},
		},
		{
			name: "添加收藏成功,DB有interactive记录，Redis无记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         9,
					Biz:        "article",
					BizId:      9,
					CollectCnt: 5,
					CTime:      100,
					UTime:      100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var collect dao.UserCollectionBiz
				err := a.db.WithContext(ctx).Model(&dao.UserCollectionBiz{}).
					Where("biz=? AND biz_id=? AND uid=?", "article", 9, 123).First(&collect).Error
				assert.NoError(t, err)
				assert.True(t, collect.UTime > 0)
				var inter dao.Interactive
				err = a.db.WithContext(ctx).Model(&dao.Interactive{}).
					Where("biz=? AND biz_id=?", "article", 9).First(&inter).Error
				assert.NoError(t, err)
				assert.True(t, inter.CTime == 100 && inter.UTime > 100)
				assert.Equal(t, int64(6), inter.CollectCnt)
				colCnt, err := a.cmd.HGet(ctx, "interactive:article:article:9", "collect_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 1, colCnt)

				err = a.db.WithContext(ctx).Where("biz=? AND biz_id=? AND uid=?", "article", 9, 123).
					Delete(dao.UserCollectionBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz=? AND biz_id=?", "article", 9).
					Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:9", "collect_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    9,
			uid:      123,
			cid:      9,
			wantResp: &interv1.CollectResponse{},
		},
		{
			name: "添加收藏成功,DB有interactive记录，Redis有记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				err := a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         10,
					Biz:        "article",
					BizId:      10,
					CollectCnt: 5,
					CTime:      100,
					UTime:      100,
				}).Error
				assert.NoError(t, err)
				err = a.cmd.HIncrBy(ctx, "interactive:article:article:10", "collect_cnt", 10).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				var collect dao.UserCollectionBiz
				err := a.db.WithContext(ctx).Model(&dao.UserCollectionBiz{}).
					Where("biz=? AND biz_id=? AND uid=?", "article", 10, 123).First(&collect).Error
				assert.NoError(t, err)
				assert.True(t, collect.UTime > 0)
				var inter dao.Interactive
				err = a.db.WithContext(ctx).Model(&dao.Interactive{}).
					Where("biz=? AND biz_id=?", "article", 10).First(&inter).Error
				assert.NoError(t, err)
				assert.True(t, inter.CTime == 100 && inter.UTime > 100)
				assert.Equal(t, int64(6), inter.CollectCnt)
				colCnt, err := a.cmd.HGet(ctx, "interactive:article:article:10", "collect_cnt").Int()
				assert.NoError(t, err)
				assert.Equal(t, 11, colCnt)

				err = a.db.WithContext(ctx).Where("biz=? AND biz_id=? AND uid=?", "article", 10, 123).
					Delete(dao.UserCollectionBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("biz=? AND biz_id=?", "article", 10).
					Delete(dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.cmd.HDel(ctx, "interactive:article:article:10", "collect_cnt").Err()
				assert.NoError(t, err)
			},
			biz:      "article",
			bizId:    10,
			uid:      123,
			cid:      10,
			wantResp: &interv1.CollectResponse{},
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			resp, err := svc.Collect(context.Background(), &interv1.CollectRequest{
				Biz: tc.biz,
				Id:  tc.bizId,
				Uid: tc.uid,
				Cid: tc.cid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestGet() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz   string
		bizId int64
		uid   int64

		wantErr  error
		wantResp *interv1.GetResponse
	}{
		{
			name: "Get失败，缓存和DB都无Interactive记录",
			before: func(t *testing.T) {
			},
			after: func(t *testing.T) {
			},
			biz:     "article",
			bizId:   11,
			uid:     123,
			wantErr: errors.New("record not found"),
		},
		{
			name: "Get成功，DB有Interactive记录，DB无Like和Collect记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         12,
					Biz:        "article",
					BizId:      12,
					LikeCnt:    1,
					CollectCnt: 2,
					ReadCnt:    3,
					CTime:      100,
					UTime:      100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).
					Where("id=?", 12).Delete(&dao.Interactive{}).Error
				assert.NoError(t, err)
				interCache := a.cmd.HMGet(ctx, "interactive:article:article:12", "like_cnt", "collect_cnt", "read_cnt").Val()
				assert.Equal(t, []interface{}{"1", "2", "3"}, interCache)
				err = a.cmd.HDel(ctx, "interactive:article:article:12", "like_cnt", "collect_cnt", "read_cnt").Err()
				assert.NoError(t, err)
			},
			biz:     "article",
			bizId:   12,
			uid:     123,
			wantErr: nil,
			wantResp: &interv1.GetResponse{
				Inter: &interv1.Interactive{
					BizId:      12,
					LikeCnt:    1,
					CollectCnt: 2,
					ReadCnt:    3,
				},
			},
		},
		{
			name: "Get成功，DB有Interactive记录，DB有Like和Collect记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         13,
					Biz:        "article",
					BizId:      13,
					LikeCnt:    1,
					CollectCnt: 2,
					ReadCnt:    3,
					CTime:      100,
					UTime:      100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:     13,
					Biz:    "article",
					BizId:  13,
					Uid:    123,
					Status: 1,
					CTime:  100,
					UTime:  100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.UserCollectionBiz{
					Id:    13,
					Biz:   "article",
					BizId: 13,
					Uid:   123,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).
					Where("id=?", 13).Delete(&dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).
					Where("id=?", 13).Delete(&dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).
					Where("id=?", 13).Delete(&dao.UserCollectionBiz{}).Error
				assert.NoError(t, err)
				interCache := a.cmd.HMGet(ctx, "interactive:article:article:13", "like_cnt", "collect_cnt", "read_cnt").Val()
				assert.Equal(t, []interface{}{"1", "2", "3"}, interCache)
				err = a.cmd.HDel(ctx, "interactive:article:article:13", "like_cnt", "collect_cnt", "read_cnt").Err()
				assert.NoError(t, err)
			},
			biz:   "article",
			bizId: 13,
			uid:   123,
			wantResp: &interv1.GetResponse{
				Inter: &interv1.Interactive{
					BizId:      13,
					LikeCnt:    1,
					CollectCnt: 2,
					ReadCnt:    3,
					Liked:      true,
					Collected:  true,
				},
			},
		},
		{
			name: "Get成功，Redis有Interactive记录，DB有Like和Collect记录",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.cmd.HMSet(ctx, "interactive:article:article:14", "like_cnt", "1", "collect_cnt", "2",
					"read_cnt", "3").Err()
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.UserLikeBiz{
					Id:     14,
					Biz:    "article",
					BizId:  14,
					Uid:    123,
					Status: 1,
					CTime:  100,
					UTime:  100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.UserCollectionBiz{
					Id:    14,
					Biz:   "article",
					BizId: 14,
					Uid:   123,
					CTime: 100,
					UTime: 100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).
					Where("id=?", 14).Delete(&dao.UserLikeBiz{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).
					Where("id=?", 14).Delete(&dao.UserCollectionBiz{}).Error
				assert.NoError(t, err)
				interCache := a.cmd.HMGet(ctx, "interactive:article:article:14", "like_cnt", "collect_cnt", "read_cnt").Val()
				assert.Equal(t, []interface{}{"1", "2", "3"}, interCache)
				err = a.cmd.HDel(ctx, "interactive:article:article:14", "like_cnt", "collect_cnt", "read_cnt").Err()
				assert.NoError(t, err)
			},
			biz:   "article",
			bizId: 14,
			uid:   123,
			wantResp: &interv1.GetResponse{
				Inter: &interv1.Interactive{
					BizId:      14,
					LikeCnt:    1,
					CollectCnt: 2,
					ReadCnt:    3,
					Liked:      true,
					Collected:  true,
				},
			},
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			resp, err := svc.Get(context.Background(), &interv1.GetRequest{
				Biz: tc.biz,
				Id:  tc.bizId,
				Uid: tc.uid,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func (a *ArticleInteractiveHandlerSuite) TestGetByIds() {
	t := a.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		biz    string
		bizIds []int64

		wantResp *interv1.GetByIdsResponse
		wantErr  error
	}{
		{
			name: "Get成功",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         15,
					Biz:        "article",
					BizId:      15,
					LikeCnt:    1,
					CollectCnt: 2,
					ReadCnt:    3,
					CTime:      100,
					UTime:      100,
				}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Create(&dao.Interactive{
					Id:         16,
					Biz:        "article",
					BizId:      16,
					LikeCnt:    4,
					CollectCnt: 5,
					ReadCnt:    6,
					CTime:      100,
					UTime:      100,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()

				err := a.db.WithContext(ctx).Where("id=?", 15).Delete(&dao.Interactive{}).Error
				assert.NoError(t, err)
				err = a.db.WithContext(ctx).Where("id=?", 16).Delete(&dao.Interactive{}).Error
				assert.NoError(t, err)
			},
			biz:    "article",
			bizIds: []int64{15, 16},
			wantResp: &interv1.GetByIdsResponse{
				Inters: map[int64]*interv1.Interactive{
					15: {
						BizId:      15,
						ReadCnt:    3,
						LikeCnt:    1,
						CollectCnt: 2,
					},
					16: {
						BizId:      16,
						ReadCnt:    6,
						LikeCnt:    4,
						CollectCnt: 5,
					},
				},
			},
		},
		{
			name:   "Get失败，DB中无interactive",
			before: func(t *testing.T) {},
			after:  func(t *testing.T) {},
			biz:    "article",
			bizIds: []int64{15, 16},
			wantResp: &interv1.GetByIdsResponse{
				Inters: map[int64]*interv1.Interactive{},
			},
		},
	}

	svc := startup.InitInteractiveService()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			resp, err := svc.GetByIds(context.Background(), &interv1.GetByIdsRequest{
				Biz: tc.biz,
				Ids: tc.bizIds,
			})
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantResp, resp)
		})
	}
}

func TestArticleInteractiveHandler(t *testing.T) {
	suite.Run(t, &ArticleInteractiveHandlerSuite{})
}
