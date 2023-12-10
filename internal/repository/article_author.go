// Copyright@daidai53 2023
package repository

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, arti domain.Article) (int64, error)
	Update(ctx context.Context, arti domain.Article) error
}
