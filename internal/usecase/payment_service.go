package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/infra/payment"
	"github.com/studio/platform/internal/pkg/apperr"
)

// PaymentService handles the payment lifecycle.
type PaymentService struct {
	orderRepo      order.Repository
	alipayGateway  payment.PaymentGateway
	wechatGateway  payment.PaymentGateway
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(
	orderRepo order.Repository,
	alipayGateway payment.PaymentGateway,
	wechatGateway payment.PaymentGateway,
) *PaymentService {
	return &PaymentService{
		orderRepo:     orderRepo,
		alipayGateway: alipayGateway,
		wechatGateway: wechatGateway,
	}
}

// InitiateAlipayPayment creates an Alipay page-pay URL for the given order.
func (s *PaymentService) InitiateAlipayPayment(ctx context.Context, orderID uuid.UUID, returnURL string) (*payment.PaymentResult, error) {
	o, err := s.getPayableOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	req := payment.CreatePaymentRequest{
		OutTradeNo:  o.ID.String(),
		Subject:     fmt.Sprintf("Studio 订单 %s", o.OrderNo),
		AmountCents: int64(o.TotalCents),
		ReturnURL:   returnURL,
	}
	return s.alipayGateway.CreatePayment(ctx, req)
}

// InitiateWechatPayment creates a WeChat Pay native QR code for the given order.
func (s *PaymentService) InitiateWechatPayment(ctx context.Context, orderID uuid.UUID) (*payment.PaymentResult, error) {
	o, err := s.getPayableOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	req := payment.CreatePaymentRequest{
		OutTradeNo:  o.ID.String(),
		Subject:     fmt.Sprintf("Studio 订单 %s", o.OrderNo),
		AmountCents: int64(o.TotalCents),
	}
	return s.wechatGateway.CreatePayment(ctx, req)
}

// HandleAlipayCallback verifies and processes an Alipay async notification.
func (s *PaymentService) HandleAlipayCallback(ctx context.Context, params map[string]string) error {
	ok, err := s.alipayGateway.VerifyCallback(ctx, params)
	if err != nil || !ok {
		return apperr.New(apperr.CodeForbidden, "无效的支付宝回调签名")
	}
	tradeStatus := params["trade_status"]
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		return nil // not a completion notification
	}
	outTradeNo := params["out_trade_no"]
	orderID, err := uuid.Parse(outTradeNo)
	if err != nil {
		return apperr.New(apperr.CodeInvalidParam, "无效的订单号")
	}
	return s.orderRepo.UpdateStatus(ctx, orderID, order.OrderStatusPaid)
}

// HandleWechatCallback verifies and processes a WeChat Pay async notification.
func (s *PaymentService) HandleWechatCallback(ctx context.Context, params map[string]string) error {
	ok, err := s.wechatGateway.VerifyCallback(ctx, params)
	if err != nil || !ok {
		return apperr.New(apperr.CodeForbidden, "无效的微信支付回调签名")
	}
	if params["result_code"] != "SUCCESS" {
		return nil
	}
	outTradeNo := params["out_trade_no"]
	orderID, err := uuid.Parse(outTradeNo)
	if err != nil {
		return apperr.New(apperr.CodeInvalidParam, "无效的订单号")
	}
	return s.orderRepo.UpdateStatus(ctx, orderID, order.OrderStatusPaid)
}

func (s *PaymentService) getPayableOrder(ctx context.Context, orderID uuid.UUID) (*order.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.OrderStatusPendingPayment {
		return nil, apperr.BadRequest("订单不在待支付状态")
	}
	return o, nil
}
