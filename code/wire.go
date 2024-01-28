//go:build wireinject

// Copyright@daidai53 2024
package main

import (
	"github.com/daidai53/webook/code/grpc"
	"github.com/daidai53/webook/code/ioc"
	"github.com/daidai53/webook/code/repository"
	"github.com/daidai53/webook/code/repository/cache"
	"github.com/daidai53/webook/code/service"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		ioc.InitRedisClient,
		ioc.InitSmsService,
		cache.NewRedisCodeCache,
		repository.NewCachedCodeRepository,
		service.NewCodeService,
		grpc.NewCodeServiceServer,
		ioc.NewGrpcxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
