// Copyright@daidai53 2024
package grpc

import (
	"context"
	"github.com/daidai53/webook/account/domain"
	"github.com/daidai53/webook/account/service"
	accountv1 "github.com/daidai53/webook/api/proto/gen/account/v1"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func (a AccountServiceServer) Credit(ctx context.Context, request *accountv1.CreditRequest) (*accountv1.CreditResponse, error) {
	items := make([]domain.CreditItem, 0, len(request.GetItems()))
	for _, item := range request.GetItems() {
		items = append(items, domain.CreditItem{
			Uid:         item.GetUid(),
			Account:     item.GetAccount(),
			AccountType: domain.AccountType(item.GetAccountType()),
			Amt:         item.GetAmt(),
			Currency:    item.GetCurrency(),
		})
	}
	err := a.svc.Credit(
		ctx,
		domain.Credit{
			Biz:   request.GetBiz(),
			BizId: request.GetBizId(),
			Items: items,
		},
	)
	if err != nil {
		return &accountv1.CreditResponse{}, err
	}
	return &accountv1.CreditResponse{}, nil
}
