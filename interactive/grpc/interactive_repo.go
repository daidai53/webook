// Copyright@daidai53 2024
package grpc

import (
	"context"
	interrepov1 "github.com/daidai53/webook/api/proto/gen/inter/interrepo/v1"
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	"github.com/daidai53/webook/interactive/domain"
	"github.com/daidai53/webook/interactive/repository"
	"google.golang.org/grpc"
)

type InteractiveRepoServiceServer struct {
	interrepov1.UnimplementedInteractiveRepositoryServer
	repo repository.InteractiveRepository
}

func NewInteractiveRepoServiceServer(repo repository.InteractiveRepository) *InteractiveRepoServiceServer {
	return &InteractiveRepoServiceServer{repo: repo}
}

func (i *InteractiveRepoServiceServer) Register(server *grpc.Server) {
	interrepov1.RegisterInteractiveRepositoryServer(server, i)
}

func (i *InteractiveRepoServiceServer) IncrReadCnt(ctx context.Context, request *interrepov1.IncrReadCntRequest) (*interrepov1.IncrReadCntResponse, error) {
	err := i.repo.IncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &interrepov1.IncrReadCntResponse{}, err
}

func (i *InteractiveRepoServiceServer) BatchIncrReadCnt(ctx context.Context, request *interrepov1.BatchIncrReadCntRequest) (*interrepov1.BatchIncrReadCntResponse, error) {
	err := i.repo.BatchIncrReadCnt(ctx, request.GetBiz(), request.GetBizId())
	return &interrepov1.BatchIncrReadCntResponse{}, err
}

func (i *InteractiveRepoServiceServer) IncrLike(ctx context.Context, request *interrepov1.IncrLikeRequest) (*interrepov1.IncrLikeResponse, error) {
	err := i.repo.IncrLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &interrepov1.IncrLikeResponse{}, err
}

func (i *InteractiveRepoServiceServer) DecrLike(ctx context.Context, request *interrepov1.DecrLikeRequest) (*interrepov1.DecrLikeResponse, error) {
	err := i.repo.DecrLike(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	return &interrepov1.DecrLikeResponse{}, err
}

func (i *InteractiveRepoServiceServer) AddCollectionItem(ctx context.Context, request *interrepov1.AddCollectionItemRequest) (*interrepov1.AddCollectionItemResponse, error) {
	err := i.repo.AddCollectionItem(ctx, request.GetBiz(), request.GetBizId(), request.GetUid(), request.GetCid())
	return &interrepov1.AddCollectionItemResponse{}, err
}

func (i *InteractiveRepoServiceServer) Get(ctx context.Context, request *interrepov1.GetRequest) (*interrepov1.GetResponse, error) {
	inter, err := i.repo.Get(ctx, request.GetBiz(), request.GetBizId())
	if err != nil {
		return &interrepov1.GetResponse{}, err
	}
	return &interrepov1.GetResponse{Inter: i.toDTO(inter)}, nil
}

func (i *InteractiveRepoServiceServer) Liked(ctx context.Context, request *interrepov1.LikedRequest) (*interrepov1.LikedResponse, error) {
	liked, err := i.repo.Liked(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return &interrepov1.LikedResponse{}, err
	}
	return &interrepov1.LikedResponse{Liked: liked}, nil
}

func (i *InteractiveRepoServiceServer) Collected(ctx context.Context, request *interrepov1.CollectedRequest) (*interrepov1.CollectedResponse, error) {
	collected, err := i.repo.Collected(ctx, request.GetBiz(), request.GetBizId(), request.GetUid())
	if err != nil {
		return &interrepov1.CollectedResponse{}, err
	}
	return &interrepov1.CollectedResponse{Collected: collected}, nil
}

func (i *InteractiveRepoServiceServer) TopIds(ctx context.Context, request *interrepov1.TopIdsRequest) (*interrepov1.TopIdsResponse, error) {
	ids, err := i.repo.TopIds(ctx, int(request.GetN()))
	if err != nil {
		return &interrepov1.TopIdsResponse{}, err
	}
	return &interrepov1.TopIdsResponse{TopIds: ids}, nil
}

func (i *InteractiveRepoServiceServer) GetByIds(ctx context.Context, request *interrepov1.GetByIdsRequest) (*interrepov1.GetByIdsResponse, error) {
	res, err := i.repo.GetByIds(ctx, request.GetBiz(), request.GetBizIds())
	if err != nil {
		return &interrepov1.GetByIdsResponse{}, err
	}
	inters := make([]*interv1.Interactive, len(res))
	for idx := range res {
		inters[idx] = i.toDTO(res[idx])
	}
	return &interrepov1.GetByIdsResponse{Inters: inters}, err
}

func (i *InteractiveRepoServiceServer) toDTO(inter domain.Interactive) *interv1.Interactive {
	return &interv1.Interactive{
		Biz:        inter.Biz,
		BizId:      inter.BizId,
		ReadCnt:    inter.ReadCnt,
		CollectCnt: inter.CollectCnt,
		LikeCnt:    inter.LikeCnt,
		Liked:      inter.Liked,
		Collected:  inter.Collected,
	}
}
