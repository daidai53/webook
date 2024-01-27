// Copyright@daidai53 2023
package service

import (
	"context"
	interrepov1 "github.com/daidai53/webook/api/proto/gen/inter/interrepo/v1"
	"github.com/daidai53/webook/interactive/domain"
	"golang.org/x/sync/errgroup"
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
	repo interrepov1.InteractiveRepositoryClient
}

func NewInteractiveService(repo interrepov1.InteractiveRepositoryClient) InteractiveService {
	return &interactiveService{
		repo: repo,
	}
}

func (i *interactiveService) GetByIds(ctx context.Context, biz string, ids []int64) (map[int64]domain.Interactive, error) {
	resp, err := i.repo.GetByIds(ctx, &interrepov1.GetByIdsRequest{
		Biz:    biz,
		BizIds: ids,
	})
	if err != nil {
		return nil, err
	}
	inters := resp.GetInters()
	res := make(map[int64]domain.Interactive, len(inters))
	for _, intr := range inters {
		res[intr.GetBizId()] = domain.Interactive{
			Biz:        intr.GetBiz(),
			BizId:      intr.GetBizId(),
			ReadCnt:    intr.GetReadCnt(),
			LikeCnt:    intr.GetLikeCnt(),
			CollectCnt: intr.GetCollectCnt(),
			Liked:      intr.GetLiked(),
			Collected:  intr.GetCollected(),
		}
	}
	return res, nil
}

func (i *interactiveService) Get(ctx context.Context, biz string, bizId int64, uid int64) (domain.Interactive, error) {
	resp, err := i.repo.Get(ctx, &interrepov1.GetRequest{
		Biz:   biz,
		BizId: bizId,
	})
	if err != nil {
		return domain.Interactive{}, err
	}
	inter := domain.Interactive{
		Biz:        resp.GetInter().GetBiz(),
		BizId:      resp.GetInter().GetBizId(),
		ReadCnt:    resp.GetInter().GetReadCnt(),
		LikeCnt:    resp.GetInter().GetLikeCnt(),
		CollectCnt: resp.GetInter().GetCollectCnt(),
		Liked:      resp.GetInter().GetLiked(),
		Collected:  resp.GetInter().GetCollected(),
	}
	var eg errgroup.Group
	eg.Go(func() error {
		likeResp, er := i.repo.Liked(ctx, &interrepov1.LikedRequest{
			Biz:   biz,
			BizId: bizId,
			Uid:   uid,
		})
		inter.Liked = likeResp.GetLiked()
		return er
	})
	eg.Go(func() error {
		collectResp, er := i.repo.Collected(ctx, &interrepov1.CollectedRequest{
			Biz:   biz,
			BizId: bizId,
			Uid:   uid,
		})
		inter.Collected = collectResp.GetCollected()
		return er
	})
	return inter, eg.Wait()
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId int64, uid int64, cid int64) error {
	_, err := i.repo.AddCollectionItem(ctx, &interrepov1.AddCollectionItemRequest{
		Biz:   biz,
		BizId: bizId,
		Uid:   uid,
		Cid:   cid,
	})
	return err
}

func (i *interactiveService) Like(ctx context.Context, biz string, id int64, uid int64) error {
	_, err := i.repo.IncrLike(ctx, &interrepov1.IncrLikeRequest{
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	return err
}

func (i *interactiveService) CancelLike(ctx context.Context, biz string, id int64, uid int64) error {
	_, err := i.repo.DecrLike(ctx, &interrepov1.DecrLikeRequest{
		Biz:   biz,
		BizId: id,
		Uid:   uid,
	})
	return err
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	_, err := i.repo.IncrReadCnt(ctx, &interrepov1.IncrReadCntRequest{
		Biz:   biz,
		BizId: bizId,
	})
	return err
}
