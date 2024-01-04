// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
)

type TopArticlesService interface {
	GetTopArticles(ctx context.Context, n int) ([]domain.Article, error)
}

type topArticlesService struct {
	artiRepo  repository.ArticleRepository
	interRepo repository.InteractiveRepository
}

func NewTopArticlesService(artiRepo repository.ArticleRepository, interRepo repository.InteractiveRepository) TopArticlesService {
	return &topArticlesService{artiRepo: artiRepo, interRepo: interRepo}
}

func (t *topArticlesService) GetTopArticles(ctx context.Context, n int) ([]domain.Article, error) {
	topIds, err := t.interRepo.TopIds(ctx, n)
	if err != nil {
		return []domain.Article{}, err
	}
	res := make([]domain.Article, 0, n)
	for _, id := range topIds {
		arti, err := t.artiRepo.GetPubById(ctx, id)
		if err != nil {
			return []domain.Article{}, err
		}
		res = append(res, arti)
	}
	return res, nil
}
