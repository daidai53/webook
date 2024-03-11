// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/search/domain"
)

type SyncService interface {
	SyncUser(ctx context.Context, user domain.User) error
	SyncArticle(ctx context.Context, arti domain.Article) error
	SyncAny(ctx context.Context, index, docId, data string) error
}

type SearchService interface {
	Search(ctx context.Context, uid int64, expression string) (domain.SearchResult, error)
}
