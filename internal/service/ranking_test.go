// Copyright@daidai53 2024
package service

import (
	"context"
	interv1 "github.com/daidai53/webook/api/proto/gen/inter/v1"
	svcmocks2 "github.com/daidai53/webook/api/proto/gen/inter/v1/mocks"
	domain2 "github.com/daidai53/webook/interactive/domain"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/repository"
	repomocks "github.com/daidai53/webook/internal/repository/mocks"
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

		mock func(ctrl *gomock.Controller) (interv1.InteractiveServiceClient, ArticleService, repository.RankingRepository)

		wantErr  error
		wantArts []domain.Article
	}{
		{
			name: "成功获取",
			mock: func(ctrl *gomock.Controller) (interv1.InteractiveServiceClient, ArticleService, repository.RankingRepository) {
				interSvc := svcmocks2.NewMockInteractiveServiceClient(ctrl)
				artSvc := svcmocks.NewMockArticleService(ctrl)
				repo := repomocks.NewMockRankingRepository(ctrl)
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
					Return(map[int64]domain2.Interactive{
						1: {LikeCnt: 1},
						2: {LikeCnt: 2},
					}, nil)
				interSvc.EXPECT().GetByIds(gomock.Any(), "article", []int64{3, 4}).
					Return(map[int64]domain2.Interactive{
						3: {LikeCnt: 3},
						4: {LikeCnt: 4},
					}, nil)

				return interSvc, artSvc, repo
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
			interSvc, artSvc, repo := tc.mock(ctrl)
			svc := NewBatchRankingService1(interSvc, artSvc, repo)
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
