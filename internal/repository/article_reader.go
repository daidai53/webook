// Copyright@daidai53 2023
package repository

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
)

type ArticleReaderRepository interface {
	Save(ctx context.Context, arti domain.Article) error
}
