// Copyright@daidai53 2024
package cache

import (
	"context"
	"github.com/daidai53/webook/follow/domain"
)

type FollowCache interface {
	StaticsInfo(ctx context.Context, uid int64) (domain.FollowStatics, error)
	SetStaticsInfo(ctx context.Context, uid int64, statics domain.FollowStatics) error
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
}
