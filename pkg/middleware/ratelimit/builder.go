// Copyright@daidai53 2023
package ratelimit

import (
	_ "embed"
	"fmt"
	"github.com/daidai53/webook/pkg/limiter"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RedisSliceWindowLimiter struct {
	prefix  string
	limiter limiter.Limiter
}

func NewRedisSliceWindowLimiter(p string, l limiter.Limiter) *RedisSliceWindowLimiter {
	return &RedisSliceWindowLimiter{
		prefix:  p,
		limiter: l,
	}
}

func (r *RedisSliceWindowLimiter) BuildLua() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limit, err := r.limiter.Limit(ctx, fmt.Sprintf("%s%s", r.prefix, ctx.ClientIP()))
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
