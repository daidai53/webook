// Copyright@daidai53 2024
package ioc

import (
	interrepov1 "github.com/daidai53/webook/api/proto/gen/inter/interrepo/v1"
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	"github.com/daidai53/webook/interactive/service"
	"github.com/daidai53/webook/internal/client"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitInterClient(svc service.InteractiveService) interv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.inter", &cfg)
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
	remote := interv1.NewInteractiveServiceClient(cc)
	local := client.NewLocalInteractiveServiceAdaptor(svc)
	res := client.NewInteractiveClient(remote, local)
	res.UpdateThreshold(cfg.Threshold)
	viper.OnConfigChange(func(in fsnotify.Event) {
		cfg = Config{}
		err := viper.UnmarshalKey("grpc.client.inter", &cfg)
		if err != nil {
			panic(err)
		}
		res.UpdateThreshold(cfg.Threshold)
	})
	return res
}

func InitInterRepoClient() interrepov1.InteractiveRepositoryClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.inter", &cfg)
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
	server := interrepov1.NewInteractiveRepositoryClient(cc)
	return server
}
