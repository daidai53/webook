// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/account/domain"
)

type AccountRepository interface {
	AddActivities(ctx context.Context, credit domain.Credit) error
	AddReward(ctx context.Context, credit domain.Credit) error
}
