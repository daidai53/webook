// Copyright@daidai53 2024
package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type GORMRewardDAO struct {
	db *gorm.DB
}

func (dao *GORMRewardDAO) Insert(ctx context.Context, r Reward) (int64, error) {
	now := time.Now().UnixMilli()
	r.CTime = now
	r.UTime = now
	err := dao.db.WithContext(ctx).Create(&r).Error
	if err != nil {
		return 0, err
	}
	return r.Id, nil
}

func (dao *GORMRewardDAO) UpdateStatus(ctx context.Context, status uint8, rid int64) error {
	return dao.db.WithContext(ctx).Model(&Reward{}).
		Where("id=?", rid).
		Updates(map[string]any{
			"status": status,
			"u_time": time.Now().UnixMilli(),
		}).Error
}

func (dao *GORMRewardDAO) GetReward(ctx context.Context, rid int64) (Reward, error) {
	var r Reward
	err := dao.db.WithContext(ctx).Where("id=?", rid).First(&r).Error
	if err != nil {
		return Reward{}, err
	}
	return r, nil
}
