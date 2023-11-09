// Copyright@daidai53 2023
package ioc

import (
	"github.com/daidai53/webook/config"
	"github.com/redis/go-redis/v9"
)

func InitRedisClient() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
}
