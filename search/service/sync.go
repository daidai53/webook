// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/repository"
)

type syncService struct {
	userRepo repository.UserRepository
	artiRepo repository.ArticleRepository
	anyRepo  repository.AnyRepository
}

func (s *syncService) SyncAny(ctx context.Context, index, docId, data string) error {
	return s.anyRepo.Input(ctx, index, docId, data)
}

func (s *syncService) SyncUser(ctx context.Context, user domain.User) error {
	return s.userRepo.SyncUser(ctx, user)
}

func (s *syncService) SyncArticle(ctx context.Context, arti domain.Article) error {
	return s.artiRepo.SyncArticle(ctx, arti)
}
