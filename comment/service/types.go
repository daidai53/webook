// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/comment/domain"
)

type CommentService interface {
	GetCommentList(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	CreateComment(ctx context.Context, cmt domain.Comment) error
	GetMoreReplies(ctx context.Context, rid, maxId, limit int64) ([]domain.Comment, error)
}
