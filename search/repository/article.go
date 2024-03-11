// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/search/domain"
	"github.com/daidai53/webook/search/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type articleRepository struct {
	dao    dao.ArticleSearchDAO
	tagDao dao.TagSearchDAO
}

func (a *articleRepository) SyncArticle(ctx context.Context, arti domain.Article) error {
	return a.dao.InputArticle(ctx, a.toEntity(arti))
}

func (a *articleRepository) SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error) {
	tids, err := a.tagDao.SearchBizIds(ctx, uid, "article", keywords)
	if err != nil {
		return nil, err
	}
	arts, err := a.dao.Search(ctx, tids, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(arts, func(idx int, src dao.Article) domain.Article {
		return domain.Article{
			Id:      src.Id,
			Title:   src.Title,
			Content: src.Content,
			Status:  src.Status,
		}
	}), nil
}

func (a *articleRepository) toEntity(arti domain.Article) dao.Article {
	return dao.Article{
		Id:      arti.Id,
		Title:   arti.Title,
		Content: arti.Content,
		Status:  arti.Status,
	}
}
