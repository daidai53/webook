// Copyright@daidai53 2024
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/daidai53/webook/tag/domain"
	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisTagCache struct {
	client     redis.Cmdable
	Expiration time.Duration
}

func (r *RedisTagCache) Append(ctx context.Context, uid int64, tags []domain.Tag) error {
	data := slice.Map(tags, func(idx int, src domain.Tag) any {
		val, _ := json.Marshal(src)
		return val
	})
	key := r.key(uid)
	pip := r.client.Pipeline()
	pip.RPush(ctx, key, data)
	pip.Expire(ctx, key, r.Expiration)
	_, err := pip.Exec(ctx)
	return err
}

func (r *RedisTagCache) GetTags(ctx context.Context, uid int64) ([]domain.Tag, error) {
	key := r.key(uid)
	data, err := r.client.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	res := make([]domain.Tag, 0, len(data))
	for _, ele := range data {
		var t domain.Tag
		err = json.Unmarshal([]byte(ele), &t)
		if err != nil {
			return nil, err
		}
		res = append(res, t)
	}
	return res, nil
}

func (r *RedisTagCache) DelTags(ctx context.Context, uid int64) error {
	return r.client.Del(ctx, r.key(uid)).Err()
}

func (r *RedisTagCache) key(uid int64) string {
	return fmt.Sprintf("tag:user_tags:%d", uid)
}
