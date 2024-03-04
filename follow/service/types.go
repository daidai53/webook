// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/follow/domain"
)

type FollowService interface {
	GetFollowee(ctx context.Context, follower, offset, limit int64) ([]domain.FollowRelation, error)
	FollowInfo(ctx context.Context,
		follower, followee int64) (domain.FollowRelation, error)
	Follow(ctx context.Context, follower, followee int64) error
	CancelFollow(ctx context.Context, follower, followee int64) error
}
