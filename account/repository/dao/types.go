// Copyright@daidai53 2024
package dao

import "context"

type AccountDAO interface {
	AddActivities(ctx context.Context, activities ...AccountActivity) error
}

type Account struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	Uid int64

	// 唯一标识一个账号
	Account int64 `gorm:"uniqueIndex:account_type"`
	Type    uint8 `gorm:"uniqueIndex:account_type""`

	Balance  int64
	Currency string

	CTime int64
	UTime int64
}

type AccountActivity struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`

	Biz   string `gorm:"uniqueIndex:biz_type_id"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id"`

	// 唯一标识一个账号
	Account     int64 `gorm:"uniqueIndex:account_type"`
	AccountType uint8 `gorm:"uniqueIndex:account_type""`

	Amount   int64
	Currency string

	Uid int64

	CTime int64
	UTime int64
}
