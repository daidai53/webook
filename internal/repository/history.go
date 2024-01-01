// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, record domain.HistoryRecord) error
}
