// Copyright@daidai53 2024
package client

import (
	"context"
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	"github.com/daidai53/webook/interactive/domain"
	"github.com/daidai53/webook/interactive/service"
	"google.golang.org/grpc"
)

type LocalInteractiveServiceAdaptor struct {
	svc service.InteractiveService
}

func NewLocalInteractiveServiceAdaptor(svc service.InteractiveService) *LocalInteractiveServiceAdaptor {
	return &LocalInteractiveServiceAdaptor{svc: svc}
}

func (l *LocalInteractiveServiceAdaptor) IncrReadCnt(ctx context.Context, in *interv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interv1.IncrReadCntResponse, error) {
	err := l.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &interv1.IncrReadCntResponse{}, err
}

func (l *LocalInteractiveServiceAdaptor) Like(ctx context.Context, in *interv1.LikeRequest, opts ...grpc.CallOption) (*interv1.LikeResponse, error) {
	err := l.svc.Like(ctx, in.GetBiz(), in.GetId(), in.GetUid())
	return &interv1.LikeResponse{}, err
}

func (l *LocalInteractiveServiceAdaptor) CancelLike(ctx context.Context, in *interv1.CancelLikeRequest, opts ...grpc.CallOption) (*interv1.CancelLikeResponse, error) {
	err := l.svc.CancelLike(ctx, in.GetBiz(), in.GetId(), in.GetUid())
	return &interv1.CancelLikeResponse{}, err
}

func (l *LocalInteractiveServiceAdaptor) Collect(ctx context.Context, in *interv1.CollectRequest, opts ...grpc.CallOption) (*interv1.CollectResponse, error) {
	err := l.svc.Collect(ctx, in.GetBiz(), in.GetId(), in.GetUid(), in.GetCid())
	return &interv1.CollectResponse{}, err
}

func (l *LocalInteractiveServiceAdaptor) Get(ctx context.Context, in *interv1.GetRequest, opts ...grpc.CallOption) (*interv1.GetResponse, error) {
	resp, err := l.svc.Get(ctx, in.GetBiz(), in.GetId(), in.GetUid())
	return &interv1.GetResponse{Inter: l.toDTO(resp)}, err
}

func (l *LocalInteractiveServiceAdaptor) GetByIds(ctx context.Context, in *interv1.GetByIdsRequest, opts ...grpc.CallOption) (*interv1.GetByIdsResponse, error) {
	resp, err := l.svc.GetByIds(ctx, in.GetBiz(), in.GetIds())
	inters := make(map[int64]*interv1.Interactive, len(resp))
	for k, v := range resp {
		inters[k] = l.toDTO(v)
	}
	return &interv1.GetByIdsResponse{
		Inters: inters,
	}, err
}

func (l *LocalInteractiveServiceAdaptor) toDTO(inter domain.Interactive) *interv1.Interactive {
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
