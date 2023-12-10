// Copyright@daidai53 2023
//go:build wireinject

package startup

import (
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/internal/repository/cache"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/web"
	ijwt "github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

var thirdPartySet = wire.NewSet(
	InitDB,
	InitRedis,
	ioc.InitLogger,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		InitDB,
		InitRedis,
		ioc.InitLogger,

		dao.NewUserDAO,
		dao.NewArticleGormDAO,
		ijwt.NewRedisJWTHandler,
		ioc.InitWechatService,
		//ioc.NewLocalCacheDefault,

		// cache部分
		cache.NewRedisCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewUserCache,

		// repository部分
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,

		// service部分
		ioc.InitSmsService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,

		// handler部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ioc.InitWebServer,
		ioc.InitGinMiddlewares,
	)
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		dao.NewArticleGormDAO,
		repository.NewCachedArticleRepository,
		service.NewArticleService,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
