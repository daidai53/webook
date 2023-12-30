// Copyright@daidai53 2023
package service

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	"golang.org/x/sync/errgroup"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(ctx context.Context, biz string, id int64, uid int64) error
	CancelLike(ctx context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId int64, uid int64, cid int64) error
	Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	repo repository.InteractiveRepository
}

func NewInteractiveService(repo repository.InteractiveRepository) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
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
	return i.repo.AddCollectionItem(ctx, biz, bizId, uid, cid)
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.IncrLike(ctx, biz, id, uid)
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	return i.repo.DecrLike(ctx, biz, id, uid)
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}
