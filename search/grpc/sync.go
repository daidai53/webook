// Copyright@daidai53 2024
package grpc

import (
	"context"
	searchv1 "github.com/daidai53/webook/api/proto/gen/search/v1"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/service"
)

type SyncServiceServer struct {
	searchv1.UnimplementedSyncServiceServer
	svc service.SyncService
}

func (s *SyncServiceServer) InputUser(ctx context.Context, request *searchv1.InputUserRequest) (*searchv1.InputUserResponse, error) {
	err := s.svc.SyncUser(ctx, domain.User{
		Id:       request.GetUser().GetId(),
		Nickname: request.GetUser().GetNickname(),
		Email:    request.GetUser().GetEmail(),
		Phone:    request.GetUser().GetPhone(),
	})
	if err != nil {
		return nil, err
	}
	return &searchv1.InputUserResponse{}, nil
}

func (s *SyncServiceServer) InputArticle(ctx context.Context, request *searchv1.InputArticleRequest) (*searchv1.InputArticleResponse, error) {
	err := s.svc.SyncArticle(ctx, domain.Article{
		Id:      request.GetArticle().GetId(),
		Title:   request.GetArticle().GetTitle(),
		Content: request.GetArticle().GetContent(),
		Status:  request.GetArticle().GetStatus(),
	})
	if err != nil {
		return nil, err
	}
	return &searchv1.InputArticleResponse{}, nil
}

func (s *SyncServiceServer) InputAny(ctx context.Context, request *searchv1.InputAnyRequest) (*searchv1.InputAnyResponse, error) {
	err := s.svc.SyncAny(ctx, request.GetIndexName(), request.GetDocId(), request.GetData())
	if err != nil {
		return nil, err
	}
	return &searchv1.InputAnyResponse{}, nil
}
