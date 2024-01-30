//go:build wireinject

// Copyright@daidai53 2024
package main

import (
	"github.com/daidai53/webook/interactive/events"
	"github.com/daidai53/webook/interactive/grpc"
	"github.com/daidai53/webook/interactive/ioc"
	"github.com/daidai53/webook/interactive/repository"
	"github.com/daidai53/webook/interactive/repository/cache"
	"github.com/daidai53/webook/interactive/repository/dao"
	"github.com/daidai53/webook/interactive/service"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	ioc.InitDstDB,
	ioc.InitSrcDB,
	ioc.InitDoubleWritePool,
	ioc.InitBizDB,
	ioc.InitSaramaSyncProducer,
	ioc.InitRedisClient,
	ioc.InitLogger,
	ioc.InitSaramaClient,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	cache.NewTopLikesCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitApp() *App {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		grpc.NewInteractiveServiceServer,
		events.NewInteractiveReadEventConsumer,
		ioc.InitInteractiveProducer,
		ioc.InitFixerConsumer,
		ioc.InitConsumers,
		ioc.NewGrpcxServer,
		ioc.InitGinxServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
