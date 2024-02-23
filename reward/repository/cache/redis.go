// Copyright@daidai53 2024
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/daidai53/webook/reward/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisRewardCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

func (r2 *RedisRewardCache) GetCachedURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	url, err := r2.client.Get(ctx, r2.key(r)).Result()
	if err != nil {
		return domain.CodeURL{}, err
	}
	var res domain.CodeURL
	err = json.Unmarshal([]byte(url), &res)
	if err != nil {
		return domain.CodeURL{}, err
	}
	return res, nil
}

func (r2 *RedisRewardCache) CachedCodeURL(ctx context.Context, url domain.CodeURL, r domain.Reward) error {
	value, err := json.Marshal(url)
	if err != nil {
		return err
	}
	return r2.client.Set(ctx, r2.key(r), value, r2.expiration).Err()
}

func (r2 *RedisRewardCache) key(r domain.Reward) string {
	return fmt.Sprintf("reward-url:%s:%d:%d", r.Target.Biz, r.Target.BizId, r.Uid)
}
