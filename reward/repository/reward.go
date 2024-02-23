// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/reward/domain"
	"github.com/daidai53/webook/reward/repository/cache"
	"github.com/daidai53/webook/reward/repository/dao"
)

type RewardRepository interface {
	CreateReward(ctx context.Context, r domain.Reward) (int64, error)
	GetCachedCodeURL(ctx context.Context, r domain.Reward) (domain.CodeURL, error)
	CacheCodeURL(ctx context.Context, url domain.CodeURL, r domain.Reward) error
	UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error
	GetReward(ctx context.Context, rid int64) (domain.Reward, error)
}

type rewardRepository struct {
	cache cache.RewardCache
	dao   dao.RewardDAO
}

func (r *rewardRepository) GetReward(ctx context.Context, rid int64) (domain.Reward, error) {
	rw, err := r.dao.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}
	return r.toDomain(rw), nil
}

func (r *rewardRepository) CreateReward(ctx context.Context, re domain.Reward) (int64, error) {
	rid, err := r.dao.Insert(ctx, r.toEntity(re))
	if err != nil {
		return 0, err
	}
	return rid, nil
}

func (r *rewardRepository) GetCachedCodeURL(ctx context.Context, re domain.Reward) (domain.CodeURL, error) {
	url, err := r.cache.GetCachedURL(ctx, re)
	if err != nil {
		return domain.CodeURL{}, err
	}
	return url, nil
}

func (r *rewardRepository) CacheCodeURL(ctx context.Context, url domain.CodeURL, re domain.Reward) error {
	return r.cache.CachedCodeURL(ctx, url, re)
}

func (r *rewardRepository) UpdateStatus(ctx context.Context, rid int64, status domain.RewardStatus) error {
	return r.dao.UpdateStatus(ctx, uint8(status), rid)
}

func (r *rewardRepository) toDomain(e dao.Reward) domain.Reward {
	return domain.Reward{
		Id:  e.Id,
		Uid: e.Uid,
		Target: domain.Target{
			Biz:     e.Biz,
			BizId:   e.BizId,
			BizName: e.BizName,
			Uid:     e.Uid,
		},
		Amt:    e.Amount,
		Status: domain.RewardStatus(e.Status),
	}
}

func (r *rewardRepository) toEntity(re domain.Reward) dao.Reward {
	return dao.Reward{
		Id:        re.Id,
		Biz:       re.Target.Biz,
		BizId:     re.Target.BizId,
		BizName:   re.Target.BizName,
		TargetUid: re.Target.Uid,
		Status:    uint8(re.Status),
		Uid:       re.Uid,
		Amount:    re.Amt,
	}
}
