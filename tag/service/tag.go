// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/daidai53/webook/tag/domain"
	"github.com/daidai53/webook/tag/events"
	"github.com/daidai53/webook/tag/repository"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type tagService struct {
	repo     repository.TagRepository
	l        logger.LoggerV1
	producer events.Producer
}

func (t *tagService) CreateTag(ctx context.Context, name string, uid int64) (int64, error) {
	return t.repo.CreateTag(ctx, domain.Tag{
		Name: name,
		Uid:  uid,
	})
}

func (t *tagService) AttachTags(ctx context.Context, biz string, bizId int64, uid int64, tagIds []int64) error {
	err := t.repo.BindTagToBiz(ctx, biz, bizId, uid, tagIds)
	if err != nil {
		return err
	}
	go func() {
		tags, err := t.repo.GetTagsById(ctx, tagIds)
		if err != nil {
			t.l.Error("查询Tags失败",
				logger.Error(err),
				logger.Int64("uid", uid),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId))
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = t.producer.ProduceSyncEvent(ctx, events.BizTags{
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
			Tags: slice.Map(tags, func(idx int, src domain.Tag) string {
				return src.Name
			}),
		})
		if err != nil {
			t.l.Error("发送同步Tag事件到Kafka失败",
				logger.Error(err),
				logger.Int64("uid", uid),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId))
		}
	}()
	return nil
}

func (t *tagService) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	return t.repo.GetTags(ctx, uid)
}

func (t *tagService) GetBizTags(ctx context.Context, biz string, uid int64, bizId int64) ([]domain.Tag, error) {
	return t.repo.GetBizTags(ctx, uid, biz, bizId)
}
