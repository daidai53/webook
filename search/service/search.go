// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/repository"
	"golang.org/x/sync/errgroup"
	"strings"
)

type searchService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
}

func (s *searchService) Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error) {
	keywords := strings.Split(expression, " ")
	var eg errgroup.Group
	var res domain.SearchResult
	eg.Go(func() error {
		users, err := s.userRepo.SearchUser(ctx, keywords)
		res.Users = users
		return err
	})
	eg.Go(func() error {
		artis, err := s.articleRepo.SearchArticle(ctx, uid, keywords)
		res.Articles = artis
		return err
	})
	return res, eg.Wait()
}
