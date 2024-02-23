// Copyright@daidai53 2024
package ioc

import (
	wechat_grpc "github.com/daidai53/webook/payment/grpc"
	"github.com/daidai53/webook/pkg/grpcx"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func InitGRPCServer(server *wechat_grpc.WechatNativeServiceServer,
) *grpcx.Server {
	type Config struct {
		Port    int    `yaml:"port"`
		EtcdUrl string `yaml:"etcd_url"`
	}

	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}
	grpcSvr := grpc.NewServer()
	server.Register(grpcSvr)
	return &grpcx.Server{
		Server:  grpcSvr,
		EtcdUrl: cfg.EtcdUrl,
		Port:    cfg.Port,
		Name:    "wechat_native_service_server",
	}
}
