// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/comment/domain"
)

type CommentRepository interface {
	FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, cmt domain.Comment) error
	CreateComment(ctx context.Context, cmt domain.Comment) error
	GetCommentByIds(ctx context.Context, id []int64) ([]domain.Comment, error)
	GetMoreReplies(ctx context.Context, rid, maxId, limit int64) ([]domain.Comment, error)
}
