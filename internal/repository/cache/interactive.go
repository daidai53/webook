// Copyright@daidai53 2023
package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/daidai53/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var (
	//go:embed lua/incr_cnt.lua
	luaIncrCnt string
)

const fieldReadCnt = "read_cnt"
const fieldLikeCnt = "like_cnt"
const fieldCollectCnt = "collect_cnt"

type InteractiveCache interface {
	IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error
	Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error)
	Set(ctx context.Context, res domain.Interactive, biz string, bizId int64) error
}

type InteractiveRedisCache struct {
	client redis.Cmdable
}

func NewInteractiveRedisCache(cmd redis.Cmdable) InteractiveCache {
	return &InteractiveRedisCache{
		client: cmd,
	}
}

func (i *InteractiveRedisCache) Set(ctx context.Context, res domain.Interactive,
	biz string, bizId int64) error {
	key := i.key(biz, bizId)
	err := i.client.HSet(ctx, key,
		fieldLikeCnt, res.LikeCnt,
		fieldReadCnt, res.ReadCnt,
		fieldCollectCnt, res.CollectCnt).Err()
	if err != nil {
		return err
	}
	return i.client.Expire(ctx, key, time.Minute*15).Err()
}

func (i *InteractiveRedisCache) Get(ctx context.Context, biz string, bizId int64) (domain.Interactive, error) {
	key := i.key(biz, bizId)
	res, err := i.client.HGetAll(ctx, key).Result()
	if err != nil {
		return domain.Interactive{}, err
	}
	if len(res) == 0 {
		return domain.Interactive{}, ErrKeyNotFound
	}
	var intr domain.Interactive
	intr.ReadCnt, _ = strconv.ParseInt(res[fieldReadCnt], 10, 64)
	intr.LikeCnt, _ = strconv.ParseInt(res[fieldLikeCnt], 10, 64)
	intr.CollectCnt, _ = strconv.ParseInt(res[fieldCollectCnt], 10, 64)
	return intr, nil
}

func (i *InteractiveRedisCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizId int64) error {
	key := i.key(biz, bizId)
	_, err := i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldCollectCnt, 1).Int()
	return err
}

func (i *InteractiveRedisCache) IncrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	_, err := i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, 1).Int()
	return err
}

func (i *InteractiveRedisCache) DecrLikeCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	_, err := i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldLikeCnt, -1).Int()
	return err
}

func (i *InteractiveRedisCache) IncrReadCntIfPresent(ctx context.Context, biz string, id int64) error {
	key := i.key(biz, id)
	_, err := i.client.Eval(ctx, luaIncrCnt, []string{key}, fieldReadCnt, 1).Int()
	return err
}
func (i *InteractiveRedisCache) key(biz string, bizId int64) string {
	return fmt.Sprintf("interactive:article:%s:%d", biz, bizId)
}
