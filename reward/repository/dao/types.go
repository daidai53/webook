// Copyright@daidai53 2024
package dao

import "context"

type RewardDAO interface {
	Insert(ctx context.Context, r Reward) (int64, error)
	UpdateStatus(ctx context.Context, status uint8, rid int64) error
	GetReward(ctx context.Context, rid int64) (Reward, error)
}

type Reward struct {
	Id        int64  `gorm:"primaryKey,autoIncrement"`
	Biz       string `gorm:"index:biz_biz_id"`
	BizId     int64  `gorm:"index:biz_biz_id"`
	BizName   string
	TargetUid int64 `gorm:"index"`
	Status    uint8
	Uid       int64
	Amount    int64
	CTime     int64
	UTime     int64
}
