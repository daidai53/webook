// Copyright@daidai53 2024
package web

import (
	rewardv1 "github.com/daidai53/webook/api/proto/gen/reward/v1"
	"github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/pkg/ginx"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/gin-gonic/gin"
)

type RewardHandler struct {
	client rewardv1.RewardServiceClient
	l      logger.LoggerV1
}

func NewRewardHandler(client rewardv1.RewardServiceClient, l logger.LoggerV1) *RewardHandler {
	return &RewardHandler{client: client, l: l}
}

func (r *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("reward")
	rg.POST("/detail", ginx.WrapBodyAndClaims[GetRewardReq, jwt.UserClaim](r.GetReward))
}

type GetRewardReq struct {
	Rid int64
}

func (r *RewardHandler) GetReward(context *gin.Context, req GetRewardReq, claims jwt.UserClaim) (ginx.Result, error) {
	resp, err := r.client.GetReward(context, &rewardv1.GetRewardRequest{
		Rid: req.Rid,
		Uid: claims.Uid,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: resp,
	}, nil
}
