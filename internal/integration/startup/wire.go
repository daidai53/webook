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
		dao.NewGORMInteractiveDAO,
		//ioc.NewLocalCacheDefault,

		// cache部分
		cache.NewRedisCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewArticleRedisCache,
		cache.NewInteractiveRedisCache,
		cache.NewUserCache,

		// repository部分
		repository.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,

		// service部分
		ioc.InitSmsService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,

		// handler部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ioc.InitWebServer,
		ioc.InitGinMiddlewares,
	)
	return gin.Default()
}

func InitArticleHandler(artDao dao.ArticleDAO, interDao dao.InteractiveDAO, userDao dao.UserDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,
		service.NewArticleService,
		service.NewInteractiveService,
		cache.NewInteractiveRedisCache,
		cache.NewArticleRedisCache,
		cache.NewUserCache,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}
