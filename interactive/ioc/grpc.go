// Copyright@daidai53 2024
package ioc

import (
	grpc2 "github.com/daidai53/webook/interactive/grpc"
	"github.com/daidai53/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewGrpcxServer(interSvc *grpc2.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		EtcdAddr string `yaml:"etcdAddr"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
	}
	s := grpc.NewServer()
	interSvc.Register(s)
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	return &grpcx.Server{
		Server:  s,
		EtcdUrl: cfg.EtcdAddr,
		Port:    cfg.Port,
		Name:    cfg.Name,
	}
}
