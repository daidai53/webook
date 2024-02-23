// Copyright@daidai53 2024
package domain

type AccountType uint8

type CreditItem struct {
	Uid         int64
	Account     int64
	AccountType AccountType
	Amt         int64
	Currency    string
}

type Credit struct {
	Biz   string
	BizId int64
	Items []CreditItem
}

const (
	AccountTypeUnknown AccountType = iota
	AccountTypeReward
	AccountTypeSystem
)
