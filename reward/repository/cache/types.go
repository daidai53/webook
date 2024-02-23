// Copyright@daidai53 2024
package cache

import (
	"context"
	"github.com/daidai53/webook/reward/domain"
)

type RewardCache interface {
	GetCachedURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CachedCodeURL(ctx context.Context, url domain.CodeURL, r domain.Reward) error
}
