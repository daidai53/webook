// Copyright@daidai53 2024
package service

import (
	"context"
	"errors"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"log"
	"math"
	"time"
)

type RankingService interface {
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	interSvc InteractiveService

	artSvc ArticleService

	batchSize int
	scoreFunc func(likeCnt int64, utime time.Time) float64
	n         int
	repo      repository.RankingRepository
}

func NewBatchRankingService(interSvc InteractiveService, artSvc ArticleService,
	rankRep repository.RankingRepository) RankingService {
	return &BatchRankingService{
		interSvc:  interSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			dura := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(dura+2, 1.5)
		},
		repo: rankRep,
	}
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	log.Println(arts)
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	ddl := start.Add(-7 * 24 * time.Hour)

	type Score struct {
		score float64
		art   domain.Article
	}
	topN := queue.NewPriorityQueue(b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		}
		if src.score < dst.score {
			return -1
		}
		return 0
	})

	for {
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		ids := slice.Map(arts, func(idx int, src domain.Article) int64 {
			return src.Id
		})
		if len(arts) == 0 {
			break
		}
		interMap, err := b.interSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		for _, art := range arts {
			inter := interMap[art.Id]
			score := b.scoreFunc(inter.LikeCnt, art.UTime)
			ele := Score{
				score: score,
				art:   art,
			}
			err1 := topN.Enqueue(ele)
			if errors.Is(err1, queue.ErrOutOfCapacity) {
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					_ = topN.Enqueue(minEle)
				}
			}
		}

		offset += len(arts)
		if len(arts) < b.batchSize || arts[len(arts)-1].UTime.Before(ddl) {
			break
		}
	}

	res := make([]domain.Article, topN.Len(), topN.Len())
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
