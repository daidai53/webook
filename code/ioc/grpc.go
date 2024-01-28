// Copyright@daidai53 2024
package ioc

import (
	code_grpc "github.com/daidai53/webook/code/grpc"
	"github.com/daidai53/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewGrpcxServer(codeSvc *code_grpc.CodeServiceServer) *grpcx.Server {
	s := grpc.NewServer()
	codeSvc.Register(s)
	return &grpcx.Server{
		Server: s,
		Addr:   viper.GetString("grpc.server.addr"),
	}
}
