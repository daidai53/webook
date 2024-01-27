// Copyright@daidai53 2023
package repository

import (
	"context"
	"errors"
	"github.com/daidai53/webook/interactive/domain"
	cache2 "github.com/daidai53/webook/interactive/repository/cache"
	dao2 "github.com/daidai53/webook/interactive/repository/dao"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"time"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, uid int64, cid int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error)
	TopIds(ctx context.Context, n int) ([]int64, error)
	GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error)
}

type CachedInteractiveRepository struct {
	dao   dao2.InteractiveDAO
	cache cache2.InteractiveCache
	top   cache2.TopLikesArticleCache
	l     logger.LoggerV1
}

func NewCachedInteractiveRepository(dao dao2.InteractiveDAO, cache cache2.InteractiveCache, top cache2.TopLikesArticleCache,
	l logger.LoggerV1) InteractiveRepository {
	ret := &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
		top:   top,
		l:     l,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	allLikes, err := ret.dao.GetAllArticleLikes()
	if err != nil {
		ret.l.Error("GetAllArticleLikes失败", logger.Error(err))
	} else {
		ret.top.Init(ctx, allLikes)
	}
	return ret
}

func (c *CachedInteractiveRepository) GetByIds(ctx context.Context, biz string, ids []int64) ([]domain.Interactive, error) {
	intrs, err := c.dao.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(intrs, func(idx int, src dao2.Interactive) domain.Interactive {
		return c.toDomain(src)
	}), nil
}

func (c *CachedInteractiveRepository) TopIds(ctx context.Context, n int) ([]int64, error) {
	return c.top.GetTopLikesIds(ctx, n)
}

func (c *CachedInteractiveRepository) BatchIncrReadCnt(ctx context.Context, biz []string, bizId []int64) error {
	err := c.dao.BatchIncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 部分失败问题-数据不一致
	go func() {
		ctx1, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		for i := 0; i < len(biz); i++ {
			er := c.cache.IncrReadCntIfPresent(ctx1, biz[i], bizId[i])
			if er != nil {
				// log
			}
		}
	}()
	return nil
}

func (c *CachedInteractiveRepository) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	intr, err := c.cache.Get(ctx, biz, bizId)
	if err == nil {
		return intr, err
	}
	ie, err := c.dao.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}

	res := c.toDomain(ie)
	err = c.cache.Set(ctx, res, biz, bizId)
	if err != nil {
		c.l.Error("回写缓存失败。",
			logger.Error(err),
			logger.String("biz", biz),
			logger.Int64("bizId", bizId))
	}
	return res, nil
}

func (c *CachedInteractiveRepository) Liked(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetLikeInfo(ctx, biz, bizId, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) Collected(ctx context.Context, biz string, bizId int64, uid int64) (bool, error) {
	_, err := c.dao.GetCollectInfo(ctx, biz, bizId, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrRecordNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizId int64, uid int64, cid int64) error {
	now := time.Now().UnixMilli()
	err := c.dao.InsertCollectionBiz(ctx, dao2.UserCollectionBiz{
		Biz:   biz,
		BizId: bizId,
		Uid:   uid,
		Cid:   cid,
		CTime: now,
		UTime: now,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	err = c.top.IncrLike(ctx, id, 1)
	if err != nil {
		c.l.Error("增加点赞数失败", logger.Error(err))
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	err = c.top.IncrLike(ctx, id, -1)
	if err != nil {
		c.l.Error("减少点赞数失败", logger.Error(err))
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	err := c.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}
	// 部分失败问题-数据不一致
	return c.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) toDomain(ie dao2.Interactive) domain.Interactive {
	return domain.Interactive{
		BizId:      ie.BizId,
		ReadCnt:    ie.ReadCnt,
		LikeCnt:    ie.LikeCnt,
		CollectCnt: ie.CollectCnt,
	}
}
