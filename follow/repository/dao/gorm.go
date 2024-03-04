// Copyright@daidai53 2024
package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type gormFollowDAO struct {
	db *gorm.DB
}

func (g *gormFollowDAO) FollowRelationList(ctx context.Context, follower, offset, limit int64) ([]FollowRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (g *gormFollowDAO) FollowRelationDetail(ctx context.Context, follower int64, followee int64) (FollowRelation, error) {
	//TODO implement me
	panic("implement me")
}

func (g *gormFollowDAO) CreateFollowRelation(ctx context.Context, c FollowRelation) error {
	now := time.Now().UnixMilli()
	c.CTime = now
	c.UTime = now
	c.Status = FollowRelationStatusActive
	return g.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status": FollowRelationStatusActive,
		}),
	}).Create(&c).Error
}

func (g *gormFollowDAO) UpdateStatus(ctx context.Context, followee int64, follower int64, status uint8) error {
	return g.db.WithContext(ctx).Where("follower=? AND followee=?", follower, followee).
		Updates(map[string]any{
			"status": status,
			"u_time": time.Now().UnixMilli(),
		}).Error
}

func (g *gormFollowDAO) CntFollower(ctx context.Context, uid int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (g *gormFollowDAO) CntFollowee(ctx context.Context, uid int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}
