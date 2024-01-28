// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
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
	"github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/ioc"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	articleDAO := dao.NewArticleGormDAO(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, articleCache, userRepository)
	client := InitSaramaClient()
	syncProducer := InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDAO := dao2.NewGORMInteractiveDAO(db)
	interactiveCache := cache2.NewInteractiveRedisCache(cmdable)
	topLikesArticleCache := cache2.NewTopLikesCache(cmdable, loggerV1)
	interactiveRepository := repository2.NewCachedInteractiveRepository(interactiveDAO, interactiveCache, topLikesArticleCache, loggerV1)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	interactiveServiceClient := ioc.InitInterClient(interactiveService)
	rankingCache := cache.NewRankingRedisCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache)
	rankingService := service.NewBatchRankingService(interactiveServiceClient, articleService, rankingRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveServiceClient, rankingService)
	wechatService := ioc.InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	engine := ioc.InitWebServer(v, userHandler, articleHandler, oAuth2WechatHandler)
	return engine
}

func InitArticleHandler(artDao dao.ArticleDAO, interDao dao2.InteractiveDAO, userDao dao.UserDAO) *web.ArticleHandler {
	loggerV1 := ioc.InitLogger()
	cmdable := InitRedis()
	articleCache := cache.NewArticleRedisCache(cmdable)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDao, userCache)
	articleRepository := repository.NewCachedArticleRepository(artDao, articleCache, userRepository)
	client := InitSaramaClient()
	syncProducer := InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveCache := cache2.NewInteractiveRedisCache(cmdable)
	topLikesArticleCache := cache2.NewTopLikesCache(cmdable, loggerV1)
	interactiveRepository := repository2.NewCachedInteractiveRepository(interDao, interactiveCache, topLikesArticleCache, loggerV1)
	interactiveService := service2.NewInteractiveService(interactiveRepository)
	interactiveServiceClient := ioc.InitInterClient(interactiveService)
	rankingCache := cache.NewRankingRedisCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache)
	rankingService := service.NewBatchRankingService(interactiveServiceClient, articleService, rankingRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveServiceClient, rankingService)
	return articleHandler
}

func InitJobScheduler() *job.Scheduler {
	db := InitDB()
	jobDAO := dao.NewGormJobDAO(db)
	jobRepository := repository.NewPreemptJobRepository(jobDAO)
	loggerV1 := ioc.InitLogger()
	cronJobService := service.NewCronJobService(jobRepository, loggerV1)
	scheduler := job.NewScheduler(cronJobService, loggerV1)
	return scheduler
}

func InitCodeService() *service.CodeServiceImpl {
	cmdable := InitRedis()
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeServiceImpl := service.NewCodeServiceImpl(codeRepository, smsService)
	return codeServiceImpl
}

// wire.go:

var thirdPartySet = wire.NewSet(
	InitDB,
	InitRedis,
	InitSaramaClient,
	InitSyncProducer, ioc.InitLogger, ioc.InitInterClient,
)

var jobProviderSet = wire.NewSet(service.NewCronJobService, repository.NewPreemptJobRepository, dao.NewGormJobDAO)

var interactiveSvcSet = wire.NewSet(dao2.NewGORMInteractiveDAO, cache2.NewInteractiveRedisCache, repository2.NewCachedInteractiveRepository, service2.NewInteractiveService)
