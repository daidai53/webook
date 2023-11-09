// Copyright@daidai53 2023
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

type UserCache interface {
	Set(ctx context.Context, du domain.User) error
	Get(ctx context.Context, uid int64) (domain.User, error)
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expireTime time.Duration
}

func NewUserCache(m redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        m,
		expireTime: 15 * time.Minute,
	}
}

func (u *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("user:info:%d", uid)
}

func (u *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := u.key(uid)
	data, err := u.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var du domain.User
	err = json.Unmarshal([]byte(data), &du)
	return du, err
}

func (u *RedisUserCache) Set(ctx context.Context, du domain.User) error {
	key := u.key(du.Id)
	data, err := json.Marshal(du)
	if err != nil {
		return nil
	}
	return u.cmd.Set(ctx, key, data, u.expireTime).Err()
}
