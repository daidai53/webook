// Copyright@daidai53 2024
package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMAccountDAO struct {
	db *gorm.DB
}

func (g *GORMAccountDAO) AddActivities(ctx context.Context, activities ...AccountActivity) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, act := range activities {
			err := tx.Where("account=? AND account_type=?", act.Account, act.AccountType).
				Clauses(clause.OnConflict{
					DoUpdates: clause.Assignments(map[string]interface{}{
						"balance": gorm.Expr("balance+?", act.Amount),
						"u_time":  now,
					}),
				}).Create(&Account{
				Id:       act.Id,
				Uid:      act.Uid,
				Account:  act.Account,
				Type:     act.AccountType,
				Balance:  act.Amount,
				Currency: act.Currency,
				CTime:    now,
				UTime:    now,
			}).Error
			if err != nil {
				return err
			}
		}
		return tx.Create(&activities).Error
	})
}
