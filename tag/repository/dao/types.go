// Copyright@daidai53 2024
package dao

import "context"

type TagDAO interface {
	CreateTag(ctx context.Context, tag Tag) (int64, error)
	CreateTagBiz(ctx context.Context, tagBizs []TagBiz) error
	GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error)
	GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error)
	GetTags(ctx context.Context, offset, limit int) ([]Tag, error)
	GetTagsById(ctx context.Context, tIds []int64) ([]Tag, error)
}

// Tag 标签表
type Tag struct {
	Id    int64 `gorm:"primaryKey;autoIncrement;"`
	Name  string
	Uid   int64 `gorm:"index;"`
	CTime int64
	UTime int64
}

// TagBiz 领域标签表，每个Tag表中的标签可能和多个资源有绑定关系
type TagBiz struct {
	Id    int64  `gorm:"primaryKey;autoIncrement;"`
	Uid   int64  `gorm:"index:index_uid_biz_biz_id"`
	Biz   string `gorm:"index:index_uid_biz_biz_id"`
	BizId int64  `gorm:"index:index_uid_biz_biz_id"`
	Tid   int64
	CTime int64
	UTime int64
}
