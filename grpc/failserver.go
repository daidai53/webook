// Copyright@daidai53 2024
package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FailServer struct {
	UnimplementedUserServiceServer
	Name string
}

func (s *FailServer) GetByID(ctx context.Context, request *GetByIDRequest) (*GetByIDResponse, error) {
	return &GetByIDResponse{}, status.Errorf(codes.Unavailable, "假装我被熔断了")
}
