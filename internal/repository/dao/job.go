// Copyright@daidai53 2024
package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, jid int64, t time.Time) error
}

type GormJobDAO struct {
	db *gorm.DB

	interval time.Duration
}

func NewGormJobDAO(db *gorm.DB) JobDAO {
	return &GormJobDAO{
		db:       db,
		interval: time.Minute,
	}
}

func (g *GormJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := g.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now()
		nowUm := now.UnixMilli()
		err := db.Where("(status = ? AND next_time<?) OR (status = ? AND next_time<? AND u_time<?)", JobStatusWaiting, nowUm,
			JobStatusRunning, nowUm, now.Add(-1*g.interval).UnixMilli()).
			First(&j).Error
		if err != nil {
			return j, err
		}
		res := db.
			Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).
			Updates(map[string]any{
				"status":  JobStatusRunning,
				"version": j.Version + 1,
				"u_time":  nowUm,
			})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到
			continue
		}
		return j, err
	}
}

func (g *GormJobDAO) Release(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).
		Model(&Job{}).
		Where("id=?", id).
		Updates(map[string]any{
			"status": JobStatusWaiting,
			"u_time": now,
		}).Error
}

func (g *GormJobDAO) UpdateUtime(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).
		Model(&Job{}).
		Where("id=?", id).
		Updates(map[string]any{
			"u_time": now,
		}).Error
}

func (g *GormJobDAO) UpdateNextTime(ctx context.Context, jid int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).
		Model(&Job{}).
		Where("id=?", jid).
		Updates(map[string]any{
			"u_time":    now,
			"next_time": t.UnixMilli(),
		}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string

	// 状态来表达，是不是可以抢占，有没有被人抢占
	Status int

	Version int

	NextTime int64 `gorm:"index"`

	CTime int64
	UTime int64
}

const (
	// JobStatusWaiting 没人抢
	JobStatusWaiting = iota
	// JobStatusRunning 已经被人抢了
	JobStatusRunning
	// JobStatusPaused 不再需要调度了
	JobStatusPaused
)
