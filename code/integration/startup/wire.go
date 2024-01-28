//go:build !wireinject

// Copyright@daidai53 2024
package startup

import (
	repository3 "github.com/daidai53/webook/code/repository"
	cache3 "github.com/daidai53/webook/code/repository/cache"
	service3 "github.com/daidai53/webook/code/service"
	"github.com/daidai53/webook/ioc"
	"github.com/google/wire"
)

func InitCodeService() *service3.CodeServiceImpl {
	wire.Build(
		InitRedis,
		ioc.InitSmsService,
		cache3.NewRedisCodeCache,
		repository3.NewCachedCodeRepository,
		service3.NewCodeServiceImpl,
	)
	return new(service3.CodeServiceImpl)
}
