// Copyright@daidai53 2023
//go:build wireinject

package startup

import (
	repository3 "github.com/daidai53/webook/code/repository"
	cache3 "github.com/daidai53/webook/code/repository/cache"
	service3 "github.com/daidai53/webook/code/service"
	repository2 "github.com/daidai53/webook/interactive/repository"
	cache2 "github.com/daidai53/webook/interactive/repository/cache"
	dao2 "github.com/daidai53/webook/interactive/repository/dao"
	service2 "github.com/daidai53/webook/interactive/service"
	"github.com/daidai53/webook/internal/events/article"
	"github.com/daidai53/webook/internal/job"
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
	InitSaramaClient,
	InitSyncProducer,
	ioc.InitLogger,
	ioc.InitInterClient,
)

var jobProviderSet = wire.NewSet(
	service.NewCronJobService,
	repository.NewPreemptJobRepository,
	dao.NewGormJobDAO)

var interactiveSvcSet = wire.NewSet(
	dao2.NewGORMInteractiveDAO,
	cache2.NewInteractiveRedisCache,
	repository2.NewCachedInteractiveRepository,
	service2.NewInteractiveService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		// 第三方依赖
		thirdPartySet,

		dao.NewUserDAO,
		dao.NewArticleGormDAO,
		ijwt.NewRedisJWTHandler,
		ioc.InitWechatService,
		dao2.NewGORMInteractiveDAO,

		//ioc.NewLocalCacheDefault,

		// cache部分
		cache3.NewRedisCodeCache,
		//cache.NewLocalCodeCache,
		cache.NewArticleRedisCache,
		cache2.NewInteractiveRedisCache,
		cache.NewUserCache,
		cache2.NewTopLikesCache,
		cache.NewRankingRedisCache,

		// repository部分
		repository3.NewCachedCodeRepository,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,
		repository2.NewCachedInteractiveRepository,
		repository.NewCachedRankingRepository,

		article.NewSaramaSyncProducer,

		// service部分
		ioc.InitSmsService,
		service.NewUserService,
		service3.NewCodeService,
		service.NewArticleService,
		service2.NewInteractiveService,
		service.NewBatchRankingService,

		// handler部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ioc.InitWebServer,
		ioc.InitGinMiddlewares,
	)
	return gin.Default()
}

func InitArticleHandler(artDao dao.ArticleDAO, interDao dao2.InteractiveDAO, userDao dao.UserDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		article.NewSaramaSyncProducer,
		repository.NewCachedUserRepository,
		repository.NewCachedArticleRepository,
		repository2.NewCachedInteractiveRepository,
		repository.NewCachedRankingRepository,
		service.NewArticleService,
		service2.NewInteractiveService,
		service.NewBatchRankingService,
		cache2.NewInteractiveRedisCache,
		cache.NewArticleRedisCache,
		cache.NewUserCache,
		cache2.NewTopLikesCache,
		cache.NewRankingRedisCache,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

func InitJobScheduler() *job.Scheduler {
	wire.Build(jobProviderSet, thirdPartySet, job.NewScheduler)
	return &job.Scheduler{}
}
