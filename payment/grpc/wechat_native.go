// Copyright@daidai53 2024
package grpc

import (
	"context"
	pmtv1 "github.com/daidai53/webook/api/proto/gen/payment/v1"
	"github.com/daidai53/webook/payment/domain"
	"github.com/daidai53/webook/payment/service"
	"google.golang.org/grpc"
)

type WechatNativeServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc service.PaymentService
}

func (w *WechatNativeServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, w)
}

func (w *WechatNativeServiceServer) NativePrePay(ctx context.Context, request *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	url, err := w.svc.PrePay(ctx, domain.Payment{
		Amt: domain.Amount{
			Total:    request.Amt.GetTotal(),
			Currency: request.Amt.GetCurrency(),
		},
		BizTradeNo:  request.GetBizTradeNo(),
		Description: request.GetDescription(),
	})
	if err != nil {
		return &pmtv1.NativePrePayResponse{}, err
	}
	return &pmtv1.NativePrePayResponse{CodeUrl: url}, nil
}

func (w *WechatNativeServiceServer) GetPayment(ctx context.Context, request *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	pmt, err := w.svc.GetPayment(ctx, request.BizTradeNo)
	if err != nil {
		return &pmtv1.GetPaymentResponse{}, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(pmt.Status),
	}, nil
}
