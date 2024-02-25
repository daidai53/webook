// Copyright@daidai53 2024
package cache

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisAccountCache struct {
	client redis.Cmdable
}

func (r *RedisAccountCache) AddReward(ctx context.Context, biz string, bizId int64) error {
	return r.client.Set(ctx, r.key(biz, bizId), "record", time.Hour*24*3).Err()
}

func (r *RedisAccountCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("account:%s:%d", biz, bizId)
}
