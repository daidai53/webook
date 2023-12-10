// Copyright@daidai53 2023
package service

import (
	"context"
	"errors"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository

	// V1
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
	l          logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{
		repo: repo,
	}
}

func NewArticleServiceV1(reader repository.ArticleReaderRepository, author repository.ArticleAuthorRepository) *articleService {
	return &articleService{
		readerRepo: reader,
		authorRepo: author,
		l:          logger.NewNopLogger(),
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	return a.repo.Sync(ctx, art)
}

func (a *articleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	var (
		id  = art.Id
		err error
	)

	if art.Id > 0 {
		err = a.authorRepo.Update(ctx, art)
	} else {
		id, err = a.authorRepo.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, art)
		if err != nil {
			a.l.Error("保存到制作库成功但是到线上库失败",
				logger.Error(err), logger.Int64("aid", art.Id))
		} else {
			return id, nil
		}
	}
	a.l.Error("保存到制作库成功但是到线上库失败，重试耗尽",
		logger.Error(err), logger.Int64("aid", art.Id))
	return id, errors.New("保存到线上库失败，重试次数耗尽")
}
