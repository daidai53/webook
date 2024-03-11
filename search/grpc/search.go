// Copyright@daidai53 2024
package grpc

import (
	"context"
	searchv1 "github.com/daidai53/webook/api/proto/gen/search/v1"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/service"
	"github.com/ecodeclub/ekit/slice"
)

type SearchServiceServer struct {
	searchv1.UnimplementedSearchServiceServer
	svc service.SearchService
}

func (s *SearchServiceServer) Search(ctx context.Context, request *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	resp, err := s.svc.Search(ctx, request.GetUid(), request.GetExpression())
	if err != nil {
		return nil, err
	}
	return &searchv1.SearchResponse{
		User: &searchv1.UserResult{
			Users: slice.Map(resp.Users, func(idx int, src domain.User) *searchv1.User {
				return &searchv1.User{
					Id:       src.Id,
					Nickname: src.Nickname,
					Email:    src.Email,
					Phone:    src.Phone,
				}
			}),
		},
		Article: &searchv1.ArticleResult{
			Articles: slice.Map(resp.Articles, func(idx int, src domain.Article) *searchv1.Article {
				return &searchv1.Article{
					Id:      src.Id,
					Title:   src.Title,
					Content: src.Content,
					Status:  src.Status,
				}
			}),
		},
	}, nil
}
