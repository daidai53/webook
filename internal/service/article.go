// Copyright@daidai53 2023
package service

import (
	"context"
	"errors"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/events/article"
	"github.com/daidai53/webook/internal/repository"
	"github.com/daidai53/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuhtor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer

	// V1
	readerRepo repository.ArticleReaderRepository
	authorRepo repository.ArticleAuthorRepository
	l          logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository, prod article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: prod,
	}
}

func (a *articleService) GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)

	go func() {
		if err == nil {
			er := a.producer.ProduceReadEvent(article.ReadEvent{
				Uid: uid,
				Aid: id,
			})
			if er != nil {
				a.l.Error("发送 ReadEvent失败",
					logger.Error(err),
					logger.Int64("uid", uid),
					logger.Int64("aid", id))
			}
		}
	}()
	return res, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuhtor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func NewArticleServiceV1(reader repository.ArticleReaderRepository, author repository.ArticleAuthorRepository) *articleService {
	return &articleService{
		readerRepo: reader,
		authorRepo: author,
		l:          logger.NewNopLogger(),
	}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
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
