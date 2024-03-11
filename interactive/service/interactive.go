// Copyright@daidai53 2023
package service

import (
	"context"
	"github.com/daidai53/webook/interactive/domain"
	events2 "github.com/daidai53/webook/interactive/events"
	"github.com/daidai53/webook/interactive/repository"
	"github.com/daidai53/webook/pkg/logger"
	"golang.org/x/sync/errgroup"
	"time"
)

//go:generate mockgen -source=./interactive.go -package=svcmocks -destination=./mocks/interactive.mock.go
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId int64, uid int64, cid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
	GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo     repository.InteractiveRepository
	producer events2.Producer
	l        logger.LoggerV1
}

func NewInteractiveService(repo repository.InteractiveRepository, producer events2.Producer) InteractiveService {
	return &interactiveService{
		repo:     repo,
		producer: producer,
	}
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error) {
	intrs, err := i.repo.GetByIds(ctx, biz, ids)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizId] = intr
	}
	return res, nil
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, bizId)
	if err != nil {
		return domain.Interactive{}, err
	}
	var eg errgroup.Group
	eg.Go(func() error {
		var er error
		intr.Liked, er = i.repo.Liked(ctx, biz, bizId, uid)
		return er
	})
	eg.Go(func() error {
		var er error
		intr.Collected, er = i.repo.Collected(ctx, biz, bizId, uid)
		return er
	})
	return intr, eg.Wait()
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId int64, uid int64, cid int64) error {
	err := i.repo.AddCollectionItem(ctx, biz, bizId, uid, cid)
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := i.producer.ProduceInteractiveEvent(ctx, events2.InteractiveEvent{
			Uid:   uid,
			Biz:   biz,
			BizId: bizId,
			Type:  events2.TypeCollect,
		})
		if err != nil {
			i.l.Error("发送收藏事件到Kafka失败",
				logger.Error(err),
				logger.Int64("uid", uid),
				logger.String("biz", biz),
				logger.Int64("biz_id", bizId))
		}
	}()
	return nil
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	err := i.repo.IncrLike(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err := i.producer.ProduceInteractiveEvent(ctx, events2.InteractiveEvent{
			Uid:   uid,
			Biz:   biz,
			BizId: id,
			Type:  events2.TypeLike,
		})
		if err != nil {
			i.l.Error("发送点赞事件到Kafka失败",
				logger.Error(err),
				logger.Int64("uid", uid),
				logger.String("biz", biz),
				logger.Int64("biz_id", id))
		}
	}()
	return nil
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, id, uid)
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}
