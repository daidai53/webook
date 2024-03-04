// Copyright@daidai53 2024
package dao

import (
	"context"
	"gorm.io/gorm"
)

type gormCommentDAO struct {
	db *gorm.DB
}

func (d *gormCommentDAO) Insert(ctx context.Context, comment Comment) error {
	//TODO implement me
	panic("implement me")
}

func (d *gormCommentDAO) FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error) {
	var comments []Comment
	err := d.db.WithContext(ctx).
		Where("biz=? AND biz_id=? AND id<? AND parent_id IS NULL", biz, bizId, minId).
		Limit(int(limit)).
		Find(&comments).Error
	return comments, err
}

func (d *gormCommentDAO) FindCommentList(ctx context.Context, u Comment) ([]Comment, error) {
	//TODO implement me
	panic("implement me")
}

func (d *gormCommentDAO) FindRepliesByPId(ctx context.Context, pid int64, offset, limit int) ([]Comment, error) {
	//TODO implement me
	panic("implement me")
}

func (d *gormCommentDAO) Delete(ctx context.Context, comment Comment) error {
	//TODO implement me
	panic("implement me")
}

func (d *gormCommentDAO) FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error) {
	//TODO implement me
	panic("implement me")
}

func (d *gormCommentDAO) FindRepliesByRid(ctx context.Context, rid int64, offset, limit int64) ([]Comment, error) {
	var res []Comment
	err := d.db.WithContext(ctx).Where("root_id=? AND id >?", rid, offset).
		Order("id ASC").Limit(int(limit)).Find(&res).Error
	return res, err
}
