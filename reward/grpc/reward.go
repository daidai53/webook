// Copyright@daidai53 2024
package grpc

import (
	"context"
	rewardv1 "github.com/daidai53/webook/api/proto/gen/reward/v1"
	"github.com/daidai53/webook/reward/domain"
	"github.com/daidai53/webook/reward/service"
)

type RewardServiceServer struct {
	rewardv1.UnimplementedRewardServiceServer
	svc service.RewardService
}

func (r *RewardServiceServer) PreReward(ctx context.Context, request *rewardv1.PreRewardRequest) (*rewardv1.PreRewardResponse, error) {
	url, err := r.svc.PreReward(ctx, domain.Reward{
		Uid: request.GetUid(),
		Target: domain.Target{
			Biz:     request.GetBiz(),
			BizId:   request.GetBizId(),
			BizName: request.GetBizName(),
			Uid:     request.GetUid(),
		},
		Amt: request.GetAmt(),
	})
	if err != nil {
		return &rewardv1.PreRewardResponse{}, err
	}
	return &rewardv1.PreRewardResponse{
		Rid:     url.Rid,
		CodeUrl: url.URL,
	}, nil

}

func (r *RewardServiceServer) GetReward(ctx context.Context, request *rewardv1.GetRewardRequest) (*rewardv1.GetRewardResponse, error) {
	reward, err := r.svc.GetReward(ctx, request.GetRid(), request.GetUid())
	if err != nil {
		return &rewardv1.GetRewardResponse{}, err
	}
	return &rewardv1.GetRewardResponse{
		Status: rewardv1.RewardStatus(reward.Status),
	}, err
}
