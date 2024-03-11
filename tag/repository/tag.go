// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/tag/domain"
	"github.com/daidai53/webook/tag/repository/cache"
	"github.com/daidai53/webook/tag/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type CachedTagRepository struct {
	dao   dao.TagDAO
	cache cache.TagCache
	l     logger.LoggerV1
}

func (c *CachedTagRepository) PreloadUserTags(ctx context.Context) error {
	offset := 0
	const batch = 100
	for {
		tags, err := c.dao.GetTags(ctx, offset, batch)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			err = c.cache.Append(ctx, tag.Uid, []domain.Tag{c.toDomain(tag)})
			if err != nil {
				return err
			}
		}
		if len(tags) < batch {
			break
		}
		offset += batch
	}
	return nil
}

func (c *CachedTagRepository) CreateTag(ctx context.Context, tag domain.Tag) (int64, error) {
	tid, err := c.dao.CreateTag(ctx, c.toEntity(tag))
	if err != nil {
		return 0, err
	}
	err = c.cache.Append(ctx, tag.Uid, []domain.Tag{tag})
	if err != nil {
		return 0, err
	}
	return tid, nil
}

func (c *CachedTagRepository) BindTagToBiz(ctx context.Context, biz string, bizId int64, uid int64, tags []int64) error {
	return c.dao.CreateTagBiz(ctx, slice.Map(tags, func(idx int, src int64) dao.TagBiz {
		return dao.TagBiz{
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
			Tid:   src,
		}
	}))
}

func (c *CachedTagRepository) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	tags, err := c.cache.GetTags(ctx, uid)
	if err == nil {
		return tags, nil
	}
	daoTags, err := c.dao.GetTagsByUid(ctx, uid)
	if err != nil {
		return nil, err
	}
	tags = slice.Map(daoTags, func(idx int, src dao.Tag) domain.Tag {
		return c.toDomain(src)
	})
	err = c.cache.Append(ctx, uid, tags)
	if err != nil {
		c.l.Error("缓存失败", logger.Error(err), logger.Int64("uid", uid))
	}
	return tags, nil
}

func (c *CachedTagRepository) GetTagsById(ctx context.Context, ids []int64) ([]domain.Tag, error) {
	res, err := c.dao.GetTagsById(ctx, ids)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao.Tag) domain.Tag {
		return c.toDomain(src)
	}), nil
}

func (c *CachedTagRepository) GetBizTags(ctx context.Context, uid int64, biz string, bizId int64) ([]domain.Tag, error) {
	res, err := c.dao.GetTagsByBiz(ctx, uid, biz, bizId)
	if err != nil {
		return nil, err
	}
	return slice.Map(res, func(idx int, src dao.Tag) domain.Tag {
		return c.toDomain(src)
	}), nil
}

func (c *CachedTagRepository) toDomain(tag dao.Tag) domain.Tag {
	return domain.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}

func (c *CachedTagRepository) toEntity(tag domain.Tag) dao.Tag {
	return dao.Tag{
		Id:   tag.Id,
		Name: tag.Name,
		Uid:  tag.Uid,
	}
}
