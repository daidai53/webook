// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	repository2 "github.com/daidai53/webook/code/repository"
	cache2 "github.com/daidai53/webook/code/repository/cache"
	service2 "github.com/daidai53/webook/code/service"
	"github.com/daidai53/webook/internal/events/article"
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/internal/repository/cache"
	"github.com/daidai53/webook/internal/repository/dao"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/web"
	"github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/ioc"
	"github.com/daidai53/webook/pkg/app"
	"github.com/google/wire"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *app.App {
	cmdable := ioc.InitRedisClient()
	handler := jwt.NewRedisJWTHandler(cmdable)
	loggerV1 := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, loggerV1)
	db := ioc.InitDB(loggerV1)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache2.NewRedisCodeCache(cmdable)
	codeRepository := repository2.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSmsService()
	codeService := service2.NewCodeService(codeRepository, smsService)
	codeServiceClient := ioc.InitCodeClient(codeService)
	userHandler := web.NewUserHandler(userService, codeServiceClient, handler)
	articleDAO := dao.NewArticleGormDAO(db)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, articleCache, userRepository)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	clientv3Client := ioc.InitEtcd()
	interactiveServiceClient := ioc.InitInterClient(clientv3Client)
	rankingCache := cache.NewRankingRedisCache(cmdable)
	rankingRepository := repository.NewCachedRankingRepository(rankingCache)
	rankingService := service.NewBatchRankingService(interactiveServiceClient, articleService, rankingRepository)
	articleHandler := web.NewArticleHandler(loggerV1, articleService, interactiveServiceClient, rankingService)
	wechatService := ioc.InitWechatService(loggerV1)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	engine := ioc.InitWebServer(v, userHandler, articleHandler, oAuth2WechatHandler)
	v2 := ioc.InitConsumers()
	rlockClient := ioc.InitRlockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingService, loggerV1, rlockClient)
	cron := ioc.InitJobs(loggerV1, rankingJob)
	app := &app.App{
		Server:    engine,
		Consumers: v2,
		Cron:      cron,
	}
	return app
}

// wire.go:

var rankingSvcSet = wire.NewSet(cache.NewRankingRedisCache, repository.NewCachedRankingRepository, service.NewBatchRankingService)
