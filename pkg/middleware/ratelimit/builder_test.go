// Copyright@daidai53 2023
package ratelimit

import (
	"github.com/daidai53/webook/pkg/limiter"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
)

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:13394",
		Password: "123456",
	})
	return redisClient
}

func Test_RedisSlideWindowLimiter(t *testing.T) {
	limiterMiddleWare := NewRedisSliceWindowLimiter("ip-limiter",
		limiter.NewRedisSlidingWindowLimiter(initRedis(), time.Second, 1000))
	engine := gin.Default()
	engine.GET("/limit", limiterMiddleWare.BuildLua())
	err := engine.Run()
	if err != nil {
		return
	}
}
