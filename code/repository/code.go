// Copyright@daidai53 2023
package repository

import (
	"context"
	"github.com/daidai53/webook/code/repository/cache"
)

var (
	ErrCodeSendTooMany   = cache.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = cache.ErrVerifySendTooMany
)

type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCachedCodeRepository(cache cache.CodeCache) CodeRepository {
	return &CachedCodeRepository{
		cache: cache,
	}
}

func (c *CachedCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return c.cache.Set(ctx, biz, phone, code)
}

func (c *CachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return c.cache.Verify(ctx, biz, phone, code)
}
