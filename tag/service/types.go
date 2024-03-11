// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/tag/domain"
)

type TagService interface {
	CreateTag(ctx context.Context, name string, uid int64) (int64, error)
	AttachTags(ctx context.Context, biz string, bizId int64, uid int64, tagIds []int64) error
	GetTags(ctx context.Context, uid int64) ([]domain.Tag, error)
	GetBizTags(ctx context.Context, biz string, uid int64, bizId int64) ([]domain.Tag, error)
}
