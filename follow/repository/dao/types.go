// Copyright@daidai53 2024
package dao

import "context"

type FollowRelation struct {
	Id int64 `gorm:"autoIncrement;primaryKey;"`

	// 不能重复关注，创建联合唯一索引
	Followee int64 `gorm:"uniqueIndex:follower_followee"`
	Follower int64 `gorm:"uniqueIndex:follower_followee"`

	// 软删除
	Status uint8

	CTime int64
	UTime int64
}

const (
	FollowRelationStatusUnknown uint8 = iota
	FollowRelationStatusActive
	FollowRelationStatusInactive
)

type FollowRelationDao interface {
	// FollowRelationList 获取某人的关注列表
	FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error)
	FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error)
	// CreateFollowRelation 创建联系人
	CreateFollowRelation(ctx context.Context, c FollowRelation) error
	// UpdateStatus 更新状态
	UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error
	// CntFollower 统计计算关注自己的人有多少
	CntFollower(ctx context.Context, uid int64) (int64, error)
	// CntFollowee 统计自己关注了多少人
	CntFollowee(ctx context.Context, uid int64) (int64, error)
}
