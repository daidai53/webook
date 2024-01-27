// Copyright@daidai53 2023
//go:build wireinject

package main

import (
	"github.com/daidai53/webook/interactive/events"
	cache2 "github.com/daidai53/webook/interactive/repository/cache"
	dao2 "github.com/daidai53/webook/interactive/repository/dao"
	service2 "github.com/daidai53/webook/interactive/service"
	"github.com/daidai53/webook/internal/events/article"
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/internal/repository/cache"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/web"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/ioc"
	"github.com/google/wire"
)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	service2.NewInteractiveService,
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB,
		ioc.InitRedisClient,
		ioc.InitLogger,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		dao.NewUserDAO,
		dao.NewArticleGormDAO,
		//ioc.NewLocalCacheDefault,

		ioc.InitRlockClient,
		ioc.InitJobs,
		ioc.InitRankingJob,
		ioc.InitInterClient,
		ioc.InitInterRepoClient,
		article.NewSaramaSyncProducer,

		events.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,
		// cache部分
		cache.NewRedisCodeCache,

		//cache.NewLocalCodeCache,
		cache.NewUserCache,
		cache.NewArticleRedisCache,
		//cache2.NewTopLikesCache,

		// repository部分
		repository.NewCachedCodeRepository,

		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,
		//repository2.NewCachedInteractiveRepository,
		ioc.InitWechatService,

		// service部分
		ioc.InitSmsService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		interactiveSvcSet,
		rankingSvcSet,

		// handler部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitWebServer,
		ioc.InitGinMiddlewares,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
