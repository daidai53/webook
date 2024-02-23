// Copyright@daidai53 2023
//go:build wireinject

package main

import (
	repository3 "github.com/daidai53/webook/code/repository"
	cache3 "github.com/daidai53/webook/code/repository/cache"
	service3 "github.com/daidai53/webook/code/service"
	"github.com/daidai53/webook/internal/events/article"
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/internal/repository/cache"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/web"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/ioc"
	"github.com/daidai53/webook/pkg/app"
	"github.com/google/wire"
)

var rankingSvcSet = wire.NewSet(
	cache.NewRankingRedisCache,
	repository.NewCachedRankingRepository,
	service.NewBatchRankingService,
)

func InitWebServer() *app.App {
	wire.Build(
		// 第三方依赖
		ioc.InitDB,
		ioc.InitRedisClient,
		ioc.InitLogger,
		ioc.InitEtcd,
		ioc.InitSaramaClient,
		ioc.InitSyncProducer,
		dao.NewUserDAO,
		dao.NewArticleGormDAO,
		//ioc.NewLocalCacheDefault,

		rankingSvcSet,
		ioc.InitRlockClient,
		ioc.InitJobs,
		ioc.InitRankingJob,
		ioc.InitInterClient,
		ioc.InitCodeClient,

		article.NewSaramaSyncProducer,
		ioc.InitConsumers,

		// cache部分
		cache3.NewRedisCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,
		cache.NewArticleRedisCache,

		// repository部分
		repository3.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,

		// service部分
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewUserService,
		service3.NewCodeService,
		service.NewArticleService,

		// handler部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitWebServer,
		ioc.InitGinMiddlewares,

		wire.Struct(new(app.App), "*"),
	)
	return new(app.App)
}
