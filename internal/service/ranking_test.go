// Copyright@daidai53 2024
package service

import (
	"context"
	"github.com/daidai53/webook/internal/domain"
	svcmocks "github.com/daidai53/webook/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestBatchRankingService_TopN(t *testing.T) {
	batchSize := 2
	now := time.Now()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (InteractiveService, ArticleService)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "成功获取",
			mock: func(ctrl *gomock.Controller) (InteractiveService, ArticleService) {
				interSvc := svcmocks.NewMockInteractiveService(ctrl)
				artSvc := svcmocks.NewMockArticleService(ctrl)
				// 批量获取数据
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, 2).
					Return([]domain.Article{
						{Id: 1, UTime: now},
						{Id: 2, UTime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, 2).
					Return([]domain.Article{
						{Id: 3, UTime: now},
						{Id: 4, UTime: now},
					}, nil)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, 2).
					Return([]domain.Article{}, nil)

				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{1, 2}).
					Return(map[int64]domain.Interactive{
						1: {LikeCnt: 1},
						2: {LikeCnt: 2},
					}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{3, 4}).
					Return(map[int64]domain.Interactive{
						3: {LikeCnt: 3},
						4: {LikeCnt: 4},
					}, nil)

				return interSvc, artSvc
			},

			wantErr: nil,
			wantArts: []domain.Article{
				{Id: 4, UTime: now},
				{Id: 3, UTime: now},
				{Id: 2, UTime: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			interSvc, artSvc := tc.mock(ctrl)
			svc := NewBatchRankingService(interSvc, artSvc)
			svc.batchSize = batchSize
			svc.scoreFunc = func(likeCnt int64, utime time.Time) float64 {
				return float64(likeCnt)
			}
			svc.n = 3
			arts, err := svc.topN(context.Background())
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantArts, arts)
		})
	}
}
