// Copyright@daidai53 2024
package grpc

import (
	"context"
	followv1 "github.com/daidai53/webook/api/proto/gen/follow/v1"
	"github.com/daidai53/webook/follow/service"
)

type FollowServiceServer struct {
	followv1.UnimplementedFollowServiceServer
	svc service.FollowService
}

func (f *FollowServiceServer) Follow(ctx context.Context, request *followv1.FollowRequest) (*followv1.FollowResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FollowServiceServer) CancelFollow(ctx context.Context, request *followv1.CancelFollowRequest) (*followv1.CancelFollowResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FollowServiceServer) GetFollowee(ctx context.Context, request *followv1.GetFolloweeRequest) (*followv1.GetFolloweeResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FollowServiceServer) FollowInfo(ctx context.Context, request *followv1.FollowInfoRequest) (*followv1.FollowInfoResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FollowServiceServer) GetFollower(ctx context.Context, request *followv1.GetFollowerRequest) (*followv1.GetFollowerResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (f *FollowServiceServer) GetFollowStatic(ctx context.Context, request *followv1.GetFollowStaticRequest) (*followv1.GetFollowStaticResponse, error) {
	//TODO implement me
	panic("implement me")
}
