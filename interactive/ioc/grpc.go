// Copyright@daidai53 2024
package ioc

import (
	grpc2 "github.com/daidai53/webook/interactive/grpc"
	"github.com/daidai53/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewGrpcxServer(interSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	s := grpc.NewServer()
	interSvc.Register(s)
	return &grpcx.Server{
		Server: s,
		Addr:   viper.GetString("grpc.server.addr"),
	}
}

func NewGrpcxServerV1(interRepo *grpc2.InteractiveRepoServiceServer) *grpcx.Server {
	s := grpc.NewServer()
	interRepo.Register(s)
	return &grpcx.Server{
		Server: s,
		Addr:   viper.GetString("grpc.server.addr"),
	}
}
