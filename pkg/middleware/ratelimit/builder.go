// Copyright@daidai53 2023
package ratelimit

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

//go:embed ratelimit.lua
var luaScript string

type RedisSliceWindowLimiter struct {
	client     redis.Cmdable
	keyGenFunc func(ctx *gin.Context) string
	// 滑动窗口的大小
	interval time.Duration

	threshold int
}

func NewRedisSliceWindowLimiter(client redis.Cmdable, interval time.Duration,
	threshold int) *RedisSliceWindowLimiter {
	return &RedisSliceWindowLimiter{
		client: client,
		keyGenFunc: func(ctx *gin.Context) string {
			return "ip-limit-" + ctx.ClientIP()
		},
		interval:  interval,
		threshold: threshold,
	}
}

func (r *RedisSliceWindowLimiter) WithTimeWindow(dur time.Duration) *RedisSliceWindowLimiter {
	r.interval = dur
	return r
}

func (r *RedisSliceWindowLimiter) WithRedisClient(cli redis.Cmdable) *RedisSliceWindowLimiter {
	r.client = cli
	return r
}

func (r *RedisSliceWindowLimiter) WithKeyGenFunc(fn func(ctx *gin.Context) string) *RedisSliceWindowLimiter {
	r.keyGenFunc = fn
	return r
}

func (r *RedisSliceWindowLimiter) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 当前时间戳
		now := time.Now().UnixNano()
		start := fmt.Sprintf("%d", now-int64(r.interval))
		key := r.keyGenFunc(ctx)
		// 清理时间窗外的
		err := r.client.ZRemRangeByScore(ctx, key,
			"0", start).Err()
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		reqs, err := r.client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: start,
			Max: fmt.Sprintf("%d", now),
		}).Result()

		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if len(reqs) >= r.threshold {
			// 限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// 这边正常执行
		err = r.client.ZAddNX(ctx, key, redis.Z{
			Score:  float64(now),
			Member: now,
		}).Err()
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		err = r.client.Expire(ctx, key, r.interval).Err()
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"text": "success",
		})
	}
}

func (r *RedisSliceWindowLimiter) BuildLua() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limit, err := r.limit(ctx)
		if err != nil {
			// redis崩溃了要不要限流
			// 保守策略是返回错误码
			// 尽可能保证可用，请求会被处理，不限流
			// 更加高级的做法是启用单机限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limit {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (r *RedisSliceWindowLimiter) limit(ctx *gin.Context) (bool, error) {
	key := r.keyGenFunc(ctx)
	return r.client.Eval(ctx, luaScript, []string{key},
		r.interval.Milliseconds(), r.threshold, time.Now().UnixMilli()).Bool()
}
