// Copyright@daidai53 2024
package repository

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository/dao"
	"time"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, jId int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, jid int64, time time.Time) error
}

type PreemptJobRepository struct {
	dao dao.JobDAO
}

func NewPreemptJobRepository(dao dao.JobDAO) JobRepository {
	return &PreemptJobRepository{dao: dao}
}

func (p *PreemptJobRepository) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	return domain.Job{
		Id:         j.Id,
		Expression: j.Expression,
		Executor:   j.Executor,
		Name:       j.Name,
	}, err
}

func (p *PreemptJobRepository) Release(ctx context.Context, jId int64) error {
	return p.dao.Release(ctx, jId)
}

func (p *PreemptJobRepository) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptJobRepository) UpdateNextTime(ctx context.Context, jid int64, time time.Time) error {
	return p.dao.UpdateNextTime(ctx, jid, time)
}
