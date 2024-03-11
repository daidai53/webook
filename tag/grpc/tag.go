// Copyright@daidai53 2024
package grpc

import (
	"context"
	tagv1 "github.com/daidai53/webook/api/proto/gen/tag/v1"
	"github.com/daidai53/webook/tag/domain"
	"github.com/daidai53/webook/tag/service"
	"github.com/ecodeclub/ekit/slice"
)

type TagServiceServer struct {
	tagv1.UnimplementedTagServiceServer
	svc service.TagService
}

func (t *TagServiceServer) CreateTag(ctx context.Context, request *tagv1.CreateTagRequest) (*tagv1.CreateTagResponse, error) {
	tid, err := t.svc.CreateTag(ctx, request.GetName(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &tagv1.CreateTagResponse{
		Tag: &tagv1.Tag{
			Id:   tid,
			Name: request.GetName(),
			Uid:  request.GetUid(),
		},
	}, nil
}

func (t *TagServiceServer) AttachTags(ctx context.Context, request *tagv1.AttachTagsRequest) (*tagv1.AttachTagsResponse, error) {
	err := t.svc.AttachTags(ctx, request.GetBiz(), request.GetBizId(), request.GetUid(), request.GetTids())
	if err != nil {
		return nil, err
	}
	return &tagv1.AttachTagsResponse{}, nil
}

func (t *TagServiceServer) GetTags(ctx context.Context, request *tagv1.GetTagsRequest) (*tagv1.GetTagsResponse, error) {
	tags, err := t.svc.GetTags(ctx, request.GetUid())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetTagsResponse{
		Tags: slice.Map(tags, func(idx int, src domain.Tag) *tagv1.Tag {
			return &tagv1.Tag{
				Id:   src.Id,
				Name: src.Name,
				Uid:  src.Uid,
			}
		}),
	}, nil
}

func (t *TagServiceServer) GetBizTags(ctx context.Context, request *tagv1.GetBizTagsRequest) (*tagv1.GetBizTagsResponse, error) {
	tags, err := t.svc.GetBizTags(ctx, request.GetBiz(), request.GetUid(), request.GetBizId())
	if err != nil {
		return nil, err
	}
	return &tagv1.GetBizTagsResponse{Tags: slice.Map(tags, func(idx int, src domain.Tag) *tagv1.Tag {
		return &tagv1.Tag{
			Id:   src.Id,
			Name: src.Name,
			Uid:  src.Uid,
		}
	})}, nil
}
