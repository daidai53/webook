// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/search/domain"
)

type UserRepository interface {
	SyncUser(ctx context.Context, user domain.User) error
	SearchUser(ctx context.Context, keywords []string) ([]domain.User, error)
}

type ArticleRepository interface {
	SyncArticle(ctx context.Context, arti domain.Article) error
	SearchArticle(ctx context.Context, uid int64, keywords []string) ([]domain.Article, error)
}

type AnyRepository interface {
	Input(ctx context.Context, index, docId, data string) error
}
