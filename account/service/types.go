// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/account/domain"
)

type AccountService interface {
	Credit(ctx context.Context, credit domain.Credit) error
}
