// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/reward/domain"
)

type RewardService interface {
	PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error)
	UpdateReward(ctx context.Context, bizTradeNo string, status domain.RewardStatus) error
}
