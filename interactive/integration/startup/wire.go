//go:build wireinject

// Copyright@daidai53 2024
package startup

import (
	"github.com/daidai53/webook/interactive/grpc"
	"github.com/daidai53/webook/interactive/repository"
	"github.com/daidai53/webook/interactive/repository/cache"
	"github.com/daidai53/webook/interactive/repository/dao"
	"github.com/daidai53/webook/interactive/service"
	"github.com/daidai53/webook/ioc"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	InitDB,
	InitRedis,
	//InitSaramaClient,
	//InitSyncProducer,
	ioc.InitLogger,
)

var interactiveSvcSet = wire.NewSet(
	dao.NewGORMInteractiveDAO,
	cache.NewInteractiveRedisCache,
	repository.NewCachedInteractiveRepository,
	service.NewInteractiveService,
)

func InitInteractiveService() *grpc.InteractiveServiceServer {
	wire.Build(
		thirdPartySet,
		interactiveSvcSet,
		cache.NewTopLikesCache,
		grpc.NewInteractiveServiceServer,
	)
	return new(grpc.InteractiveServiceServer)
}
