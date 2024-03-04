// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/comment/domain"
	"github.com/daidai53/webook/comment/repository/dao"
	"golang.org/x/sync/errgroup"
	"time"
)

type commentRepository struct {
	dao dao.CommentDAO
}

func (c *commentRepository) FindByBiz(ctx context.Context, biz string, bizId, minId, limit int64) ([]domain.Comment, error) {
	// 这里只找出来了根评论
	daoComments, err := c.dao.FindByBiz(ctx, biz, bizId, minId, limit)
	if err != nil {
		return nil, err
	}

	res := make([]domain.Comment, 0, len(daoComments))
	// 下面要开始找子评论
	var eg errgroup.Group
	downgrade := ctx.Value("downgrade") == "true"
	for _, dc := range daoComments {
		newDc := dc
		cm := c.toDomain(newDc)
		if downgrade {
			continue
		}
		eg.Go(func() error {
			subComments, err := c.dao.FindRepliesByPId(ctx, newDc.Id, 0, 3)
			if err != nil {
				return err
			}

			cm.Children = make([]domain.Comment, 0, len(subComments))
			for _, subComment := range subComments {
				cm.Children = append(cm.Children, c.toDomain(subComment))
			}
			return nil
		})
	}
	return res, eg.Wait()
}

func (c *commentRepository) DeleteComment(ctx context.Context, cmt domain.Comment) error {
	//TODO implement me
	panic("implement me")
}

func (c *commentRepository) CreateComment(ctx context.Context, cmt domain.Comment) error {
	//TODO implement me
	panic("implement me")
}

func (c *commentRepository) GetCommentByIds(ctx context.Context, id []int64) ([]domain.Comment, error) {
	//TODO implement me
	panic("implement me")
}

func (c *commentRepository) GetMoreReplies(ctx context.Context, rid, maxId, limit int64) ([]domain.Comment, error) {
	comments, err := c.dao.FindRepliesByRid(ctx, rid, maxId, limit)
	if err != nil {
		return nil, err
	}
	res := make([]domain.Comment, 0, len(comments))
	for _, cmt := range comments {
		res = append(res, c.toDomain(cmt))
	}
	return res, nil
}

func (c *commentRepository) toDomain(daoComment dao.Comment) domain.Comment {
	val := domain.Comment{
		Id: daoComment.Id,
		Commentator: domain.User{
			Id: daoComment.Uid,
		},
		Biz:     daoComment.Biz,
		BizId:   daoComment.BizId,
		Content: daoComment.Content,
		CTime:   time.UnixMilli(daoComment.CTime),
		UTime:   time.UnixMilli(daoComment.UTime),
	}
	if daoComment.ParentId.Valid {
		val.ParentComment = &domain.Comment{
			Id: daoComment.ParentId.Int64,
		}
	}
	if daoComment.RootId.Valid {
		val.RootComment = &domain.Comment{
			Id: daoComment.RootId.Int64,
		}
	}
	return val
}
