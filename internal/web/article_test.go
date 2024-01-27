// Copyright@daidai53 2023
package web

import (
	"bytes"
	"encoding/json"
	"errors"
	service2 "github.com/daidai53/webook/interactive/service"
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/service"
	svcmocks "github.com/daidai53/webook/internal/service/mocks"
	"github.com/daidai53/webook/pkg/ginx"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name      string
		mockArti  func(ctrl *gomock.Controller) service.ArticleService
		mockInter func(ctrl *gomock.Controller) service2.InteractiveService
		mockRank  func(ctrl *gomock.Controller) service.RankingService
		reqBody   string

		wantCode int
		wantRes  ginx.Result
	}{
		{
			name: "新建并发表成功",
			mockArti: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: int64(123),
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"title":"我的标题",
	"content":"我的内容"
}`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Data: float64(1),
			},
		},
		{
			name: "已有帖子发表失败",
			mockArti: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      int64(123),
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: int64(123),
					},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody: `
{
	"id":123,
	"title":"我的标题",
	"content":"我的内容"
}`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Data: float64(1),
			},
		},
		{
			name: "发表失败",
			mockArti: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: int64(123),
					},
				}).Return(int64(0), errors.New("mockArti error"))
				return svc
			},
			reqBody: `
{
	"title":"我的标题",
	"content":"我的内容"
}`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "Bind错误",
			mockArti: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				return svc
			},
			reqBody: `
{
	"title":"我的标题",aaa
	"content":"我的内容"
}`,
			wantCode: http.StatusBadRequest,
			wantRes: ginx.Result{
				Data: float64(1),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			articleSvc := tc.mockArti(ctrl)
			interSvc := tc.mockInter(ctrl)
			rankSvc := tc.mockRank(ctrl)
			hdl := NewArticleHandler(logger.NewNopLogger(), articleSvc, interSvc, rankSvc)

			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user-id", int64(123))
			})
			hdl.RegisterRoutes(server)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			record := httptest.NewRecorder()
			server.ServeHTTP(record, req)
			assert.Equal(t, tc.wantCode, record.Code)
			if record.Code != http.StatusOK {
				return
			}

			var res ginx.Result
			err = json.NewDecoder(record.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
