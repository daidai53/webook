// Copyright@daidai53 2023
//go:build wireinject

package main

import (
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
	dao.NewGORMInteractiveDAO,
	repository.NewCachedInteractiveRepository,
	cache.NewInteractiveRedisCache,
	service.NewInteractiveService,
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

		interactiveSvcSet,

		article.NewSaramaSyncProducer,
		article.NewInteractiveReadEventConsumer,
		ioc.InitConsumers,

		// cache部分
		cache.NewRedisCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,
		cache.NewArticleRedisCache,
		cache.NewTopLikesCache,

		// repository部分
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,

		// service部分
		ioc.InitSmsService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewTopArticlesService,

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
