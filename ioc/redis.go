// Copyright@daidai53 2023
package ioc

import (
	rlock "github.com/gotomicro/redis-lock"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedisClient() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis.Addr"),
	})
}

func InitRlockClient(cmd redis.Cmdable) *rlock.Client {
	return rlock.NewClient(cmd)
}
