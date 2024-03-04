// Copyright@daidai53 2024
package grpc

import (
	"context"
	commentv1 "github.com/daidai53/webook/api/proto/gen/comment/v1"
	"github.com/daidai53/webook/comment/domain"
	"github.com/daidai53/webook/comment/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CommentServiceServer struct {
	commentv1.UnimplementedCommentServiceServer
	svc service.CommentService
}

func (c *CommentServiceServer) GetCommentList(ctx context.Context, request *commentv1.CommentListRequest) (*commentv1.CommentListResponse, error) {
	comments, err := c.svc.GetCommentList(ctx, request.GetBiz(), request.GetBizid(), request.GetMinId(), request.GetLimit())
	if err != nil {
		return &commentv1.CommentListResponse{}, err
	}
	return &commentv1.CommentListResponse{
		Comments: c.toDTO(comments),
	}, nil
}

func (c *CommentServiceServer) DeleteComment(ctx context.Context, request *commentv1.DeleteCommentRequest) (*commentv1.DeleteCommentResponse, error) {
	err := c.svc.DeleteComment(ctx, request.GetId())
	return &commentv1.DeleteCommentResponse{}, err
}

func (c *CommentServiceServer) CreateComment(ctx context.Context, request *commentv1.CreateCommentRequest) (*commentv1.CreateCommentResponse, error) {
	err := c.svc.CreateComment(ctx, c.ToDomain(request.GetComment()))
	return &commentv1.CreateCommentResponse{}, err
}

func (c *CommentServiceServer) GetMoreReplies(ctx context.Context, request *commentv1.GetMoreRepliesRequest) (*commentv1.GetMoreRepliesResponse, error) {
	replies, err := c.svc.GetMoreReplies(ctx, request.GetRid(), request.GetMaxId(), request.GetLimit())
	if err != nil {
		return &commentv1.GetMoreRepliesResponse{}, err
	}
	return &commentv1.GetMoreRepliesResponse{
		Replies: c.toDTO(replies),
	}, nil
}

func (c *CommentServiceServer) toDTO(domainComments []domain.Comment) []*commentv1.Comment {
	rpcComments := make([]*commentv1.Comment, 0, len(domainComments))
	for _, domainComment := range domainComments {
		rpcComment := &commentv1.Comment{
			Id:      domainComment.Id,
			Uid:     domainComment.Commentator.Id,
			Biz:     domainComment.Biz,
			Bizid:   domainComment.BizId,
			Content: domainComment.Content,
			Ctime:   timestamppb.New(domainComment.CTime),
			Utime:   timestamppb.New(domainComment.UTime),
		}
		if domainComment.RootComment != nil {
			rpcComment.RootComment = &commentv1.Comment{
				Id: domainComment.RootComment.Id,
			}
		}
		if domainComment.ParentComment != nil {
			rpcComment.ParentComment = &commentv1.Comment{
				Id: domainComment.ParentComment.Id,
			}
		}
		rpcComments = append(rpcComments, rpcComment)
	}
	rpcCommentMap := make(map[int64]*commentv1.Comment, len(rpcComments))
	for _, rpcComment := range rpcComments {
		rpcCommentMap[rpcComment.Id] = rpcComment
	}
	for _, domainComment := range domainComments {
		rpcComment := rpcCommentMap[domainComment.Id]
		if rpcComment.RootComment != nil {
			val, ok := rpcCommentMap[rpcComment.RootComment.Id]
			if ok {
				rpcComment.RootComment = val
			}
		}
		if rpcComment.ParentComment != nil {
			val, ok := rpcCommentMap[rpcComment.ParentComment.Id]
			if ok {
				rpcComment.ParentComment = val
			}
		}
	}
	return rpcComments
}

func (c *CommentServiceServer) ToDomain(comment *commentv1.Comment) domain.Comment {
	domainComment := domain.Comment{
		Id: comment.GetId(),
		Commentator: domain.User{
			Id: comment.GetUid(),
		},
		Biz:     comment.GetBiz(),
		BizId:   comment.GetBizid(),
		Content: comment.GetContent(),
	}
	if comment.GetRootComment() != nil {
		domainComment.RootComment = &domain.Comment{
			Id: comment.GetRootComment().GetId(),
		}
	}
	if comment.GetParentComment() != nil {
		domainComment.ParentComment = &domain.Comment{
			Id: comment.GetParentComment().GetId(),
		}
	}
	return domainComment
}
