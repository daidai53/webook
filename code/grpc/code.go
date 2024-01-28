// Copyright@daidai53 2024
package grpc

import (
	"context"
	codev1 "github.com/daidai53/webook/api/proto/gen/code/v1"
	"github.com/daidai53/webook/code/service"
	"google.golang.org/grpc"
)

type CodeServiceServer struct {
	codev1.UnimplementedCodeServiceServer
	svc service.CodeService
}

func NewCodeServiceServer(svc service.CodeService) *CodeServiceServer {
	return &CodeServiceServer{svc: svc}
}

func (c *CodeServiceServer) Register(s *grpc.Server) {
	codev1.RegisterCodeServiceServer(s, c)
}

func (c *CodeServiceServer) Send(ctx context.Context, request *codev1.SendRequest) (*codev1.SendResponse, error) {
	err := c.svc.Send(ctx, request.GetBiz(), request.GetPhone())
	return &codev1.SendResponse{}, err
}

func (c *CodeServiceServer) Verify(ctx context.Context, request *codev1.VerifyRequest) (*codev1.VerifyResponse, error) {
	verified, err := c.svc.Verify(ctx, request.GetBiz(), request.GetPhone(), request.GetCode())
	if err != nil {
		return &codev1.VerifyResponse{}, err
	}
	return &codev1.VerifyResponse{Verified: verified}, nil
}
