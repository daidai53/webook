// Copyright@daidai53 2024
package domain

type Target struct {
	Biz     string
	BizId   int64
	BizName string

	Uid int64
}

type RewardStatus uint8

const (
	RewardStatusUnknown RewardStatus = iota
	RewardStatusInit
	RewardStatusSuccess
	RewardStatusPayed
	RewardStatusFailed
	RewardStatusRefund
)

type Reward struct {
	Id     int64
	Uid    int64
	Target Target
	Amt    int64
	Status RewardStatus
}

func (r Reward) Completed() bool {
	return r.Status == RewardStatusPayed || r.Status == RewardStatusFailed
}

type CodeURL struct {
	Rid int64
	URL string
}
