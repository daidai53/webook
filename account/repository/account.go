// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/account/domain"
	"github.com/daidai53/webook/account/repository/cache"
	"github.com/daidai53/webook/account/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type accountRepository struct {
	dao   dao.AccountDAO
	cache cache.AccountCache
}

func (a *accountRepository) AddReward(ctx context.Context, credit domain.Credit) error {
	return a.cache.AddReward(ctx, credit.Biz, credit.BizId)
}

func (a *accountRepository) AddActivities(ctx context.Context, credit domain.Credit) error {
	activities := slice.Map[domain.CreditItem, dao.AccountActivity](credit.Items, func(idx int, src domain.CreditItem) dao.AccountActivity {
		return dao.AccountActivity{
			Biz:         credit.Biz,
			BizId:       credit.BizId,
			Account:     src.Account,
			AccountType: uint8(src.AccountType),
			Amount:      src.Amt,
			Currency:    src.Currency,
		}
	})
	return a.dao.AddActivities(ctx, activities...)
}
