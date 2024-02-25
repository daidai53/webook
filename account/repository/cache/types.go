// Copyright@daidai53 2024
package cache

import (
	"context"
)

type AccountCache interface {
	AddReward(ctx context.Context, biz string, bizId int64) error
}
