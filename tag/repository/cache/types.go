// Copyright@daidai53 2024
package cache

import (
	"context"
	"github.com/daidai53/webook/tag/domain"
)

type TagCache interface {
	Append(ctx context.Context, uid int64, tags []domain.Tag) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	DelTags(ctx context.Context, uid int64) error
}
