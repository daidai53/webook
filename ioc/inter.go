// Copyright@daidai53 2024
package ioc

import (
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	resolver2 "go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitInterClient(client *etcdv3.Client) interv1.InteractiveServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.inter", &cfg)
	if err != nil {
		panic(err)
	}

	resolver, err := resolver2.NewBuilder(client)
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{
		grpc.WithResolvers(resolver),
	}
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		panic(err)
	}
	return interv1.NewInteractiveServiceClient(cc)
}

//func InitInterClientOld(svc service.InteractiveService) interv1.InteractiveServiceClient {
//	type Config struct {
//		Addr      string `yaml:"addr"`
//		Secure    bool   `yaml:"secure"`
//		Threshold int32  `yaml:"threshold"`
//	}
//	var cfg Config
//	err := viper.UnmarshalKey("grpc.client.inter", &cfg)
//	if err != nil {
//		panic(err)
//	}
//	var opts []grpc.DialOption
//	if !cfg.Secure {
//		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	}
//	cc, err := grpc.Dial(cfg.Addr, opts...)
//	if err != nil {
//		panic(err)
//	}
//	remote := interv1.NewInteractiveServiceClient(cc)
//	local := client.NewLocalInteractiveServiceAdaptor(svc)
//	res := client.NewInteractiveClient(remote, local)
//	res.UpdateThreshold(cfg.Threshold)
//	viper.OnConfigChange(func(in fsnotify.Event) {
//		cfg = Config{}
//		err := viper.UnmarshalKey("grpc.client.inter", &cfg)
//		if err != nil {
//			panic(err)
//		}
//		res.UpdateThreshold(cfg.Threshold)
//	})
//	return res
//}
