// Copyright@daidai53 2024
package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type tagGormDAO struct {
	db *gorm.DB
}

func (t *tagGormDAO) CreateTag(ctx context.Context, tag Tag) (int64, error) {
	now := time.Now().UnixMilli()
	tag.CTime = now
	tag.UTime = now
	err := t.db.WithContext(ctx).Create(&tag).Error
	return tag.Id, err
}

func (t *tagGormDAO) CreateTagBiz(ctx context.Context, tagBizs []TagBiz) error {
	now := time.Now().UnixMilli()
	for i := range tagBizs {
		tagBizs[i].CTime = now
		tagBizs[i].UTime = now
	}
	first := tagBizs[0]
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&TagBiz{}).Delete(
			"uid=? AND biz=? AND biz_id=?", first.Uid, first.Biz, first.BizId).Error
		if err != nil {
			return err
		}
		return tx.Create(&tagBizs).Error
	})
}

func (t *tagGormDAO) GetTagsByUid(ctx context.Context, uid int64) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Where("uid=?", uid).Find(&res).Error
	return res, err
}

func (t *tagGormDAO) GetTagsByBiz(ctx context.Context, uid int64, biz string, bizId int64) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Where("uid=? AND biz=? AND biz_id=?", uid, biz, bizId).Find(&res).Error
	return res, err
}

func (t *tagGormDAO) GetTags(ctx context.Context, offset, limit int) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Model(&TagBiz{}).Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (t *tagGormDAO) GetTagsById(ctx context.Context, tIds []int64) ([]Tag, error) {
	var res []Tag
	err := t.db.WithContext(ctx).Where("id IN ?", tIds).Find(&res).Error
	return res, err
}
