//go:build wireinject

// Copyright@daidai53 2024
package main

import (
	"github.com/daidai53/webook/interactive/grpc"
	"github.com/daidai53/webook/interactive/ioc"
	"github.com/daidai53/webook/interactive/repository"
	"github.com/daidai53/webook/interactive/repository/cache"
	"github.com/daidai53/webook/interactive/repository/dao"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedisClient,
	ioc.InitLogger,
	ioc.InitSaramaClient,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	cache.NewTopLikesCache,
	repository.NewCachedInteractiveRepository,
	//service.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		//grpc.NewInteractiveServiceServer,
		grpc.NewInteractiveRepoServiceServer,
		//events.NewInteractiveReadEventConsumer,
		//ioc.InitConsumers,
		//ioc.NewGrpcxServer,
		ioc.NewGrpcxServerV1,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
