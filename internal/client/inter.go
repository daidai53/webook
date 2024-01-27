// Copyright@daidai53 2024
package client

import (
	"context"
	"fmt"
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
)

type InteractiveClient struct {
	remote interv1.InteractiveServiceClient
	local  *LocalInteractiveServiceAdaptor

	threshold *atomicx.Value[int32]
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *interv1.IncrReadCntRequest, opts ...grpc.CallOption) (*interv1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveClient) Like(ctx context.Context, in *interv1.LikeRequest, opts ...grpc.CallOption) (*interv1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *InteractiveClient) CancelLike(ctx context.Context, in *interv1.CancelLikeRequest, opts ...grpc.CallOption) (*interv1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in, opts...)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *interv1.CollectRequest, opts ...grpc.CallOption) (*interv1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *InteractiveClient) Get(ctx context.Context, in *interv1.GetRequest, opts ...grpc.CallOption) (*interv1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *interv1.GetByIdsRequest, opts ...grpc.CallOption) (*interv1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}

func (i *InteractiveClient) selectClient() interv1.InteractiveServiceClient {
	zap.L().Info(fmt.Sprintf("selectClient:%v", i.threshold.Load()))
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveClient) UpdateThreshold(val int32) {
	i.threshold.Store(val)
}

func NewInteractiveClient(remote interv1.InteractiveServiceClient, local *LocalInteractiveServiceAdaptor) *InteractiveClient {
	return &InteractiveClient{
		remote:    remote,
		local:     local,
		threshold: atomicx.NewValue[int32](),
	}
}
