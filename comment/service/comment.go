// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/comment/domain"
	"github.com/daidai53/webook/comment/repository"
)

type commentService struct {
	repo repository.CommentRepository
}

func (c *commentService) GetCommentList(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error) {
	list, err := c.repo.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (c *commentService) DeleteComment(ctx context.Context, id int64) error {
	return c.repo.DeleteComment(ctx, domain.Comment{
		Id: id,
	})
}

func (c *commentService) CreateComment(ctx context.Context, cmt domain.Comment) error {
	return c.repo.CreateComment(ctx, cmt)
}

func (c *commentService) GetMoreReplies(ctx context.Context, rid, maxId, limit int64) ([]domain.Comment, error) {
	return c.repo.GetMoreReplies(ctx, rid, maxId, limit)
}
