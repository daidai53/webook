// Copyright@daidai53 2024
package dao

import (
	"context"
	"database/sql"
)

type CommentDAO interface {
	Insert(ctx context.Context, comment Comment) error
	FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]Comment, error)
	FindCommentList(ctx context.Context, u Comment) ([]Comment, error)
	FindRepliesByPId(ctx context.Context, pid int64, offset, limit int) ([]Comment, error)
	Delete(ctx context.Context, comment Comment) error
	FindOneByIds(ctx context.Context, ids []int64) ([]Comment, error)
	FindRepliesByRid(ctx context.Context, rid int64, offset, limit int64) ([]Comment, error)
}

// 评论
type Comment struct {
	Id  int64 `gorm:"autoIncrement,primaryKey"`
	Uid int64 `gorm:"index"`

	Biz   string `gorm:"index:biz_biz_id"`
	BizId int64  `gorm:"index:biz_biz_id"`

	ParentId sql.NullInt64 `gorm:"index"`
	RootId   sql.NullInt64 `gorm:"index"`
	Content  string

	ParentComment *Comment `gorm:"ForeignKey:PID;AssociationForeignKey:ID;constraint:OnDelete:CASCADE"`

	CTime int64
	UTime int64
}
