package web

import (
	"github.com/daidai53/webook/payment/service"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"net/http"
)

type WechatHandler struct {
	handler   *notify.Handler
	l         logger.LoggerV1
	nativeSvc *service.NativePaymentService
}

func NewWechatHandler(handler *notify.Handler, l logger.LoggerV1, nativeSvc *service.NativePaymentService) *WechatHandler {
	return &WechatHandler{
		handler:   handler,
		l:         l,
		nativeSvc: nativeSvc,
	}
}

func (w *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.GET("hello", func(context *gin.Context) {
		context.String(http.StatusOK, "进来了")
	})

	server.Any("/pay/callback", w.HandleNative)
}

func (w *WechatHandler) HandleNative(ctx *gin.Context) {
	transaction := new(payments.Transaction)
	_, err := w.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		ctx.String(http.StatusBadRequest, "参数解析失败")
		w.l.Error("参数解析失败", logger.Error(err))
		// 走到这里可能是有黑客在尝试攻击系统
		return
	}
	err = w.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "处理回调失败")
		w.l.Error("处理回调失败", logger.Error(err),
			logger.String("biz_trade_no", *transaction.OutTradeNo))
		return
	}
	ctx.String(http.StatusOK, "OK")
}
