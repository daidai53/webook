// Copyright@daidai53 2023
package limiter

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed ratelimit.lua
var luaScript string

type RedisSlidingWindowLimiter struct {
	client    redis.Cmdable
	interval  time.Duration
	threshold int
}

func NewRedisSlidingWindowLimiter(client redis.Cmdable, interval time.Duration,
	threshold int) *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{
		client:    client,
		interval:  interval,
		threshold: threshold,
	}
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.client.Eval(ctx, luaScript, []string{key},
		r.interval.Milliseconds(), r.threshold, time.Now().UnixMilli()).Bool()
}
