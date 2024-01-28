// Copyright@daidai53 2024
package ioc

import (
	codev1 "github.com/daidai53/webook/api/proto/gen/code/v1"
	"github.com/daidai53/webook/code/service"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitCodeClient(svc service.CodeService) codev1.CodeServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.code", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	server := codev1.NewCodeServiceClient(cc)
	return server
}
