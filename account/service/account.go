// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/account/domain"
	"github.com/daidai53/webook/account/repository"
)

type accountService struct {
	repo repository.AccountRepository
}

func (a *accountService) Credit(ctx context.Context, credit domain.Credit) error {
	return a.repo.AddActivities(ctx, credit)
}
