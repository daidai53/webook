// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/account/domain"
	"github.com/daidai53/webook/account/repository"
	"github.com/daidai53/webook/pkg/logger"
)

type accountService struct {
	repo repository.AccountRepository
	l    logger.LoggerV1
}

func (a *accountService) Credit(ctx context.Context, credit domain.Credit) error {
	err := a.repo.AddReward(ctx, credit)
	if err != nil {
		a.l.Info(
			"重复记账",
			logger.Error(err),
			logger.String("biz", credit.Biz),
			logger.Int64("biz_id", credit.BizId),
		)
		return err
	}
	return a.repo.AddActivities(ctx, credit)
}
