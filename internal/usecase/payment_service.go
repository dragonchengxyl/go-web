package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/infra/payment"
	"github.com/studio/platform/internal/infra/streams"
	"github.com/studio/platform/internal/pkg/apperr"
)

// PaymentService handles the payment lifecycle.
type PaymentService struct {
	orderRepo     order.Repository
	alipayGateway payment.PaymentGateway
	wechatGateway payment.PaymentGateway
	publisher     *streams.Publisher
}

type PaymentServiceOption func(*PaymentService)

func WithPaymentPublisher(p *streams.Publisher) PaymentServiceOption {
	return func(s *PaymentService) { s.publisher = p }
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(
	orderRepo order.Repository,
	alipayGateway payment.PaymentGateway,
	wechatGateway payment.PaymentGateway,
) *PaymentService {
	return NewPaymentServiceWithOptions(orderRepo, alipayGateway, wechatGateway)
}

func NewPaymentServiceWithOptions(
	orderRepo order.Repository,
	alipayGateway payment.PaymentGateway,
	wechatGateway payment.PaymentGateway,
	opts ...PaymentServiceOption,
) *PaymentService {
	svc := &PaymentService{
		orderRepo:     orderRepo,
		alipayGateway: alipayGateway,
		wechatGateway: wechatGateway,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
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
	return s.markPaid(ctx, orderID, order.PaymentMethodAlipay)
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
	return s.markPaid(ctx, orderID, order.PaymentMethodWechat)
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

func (s *PaymentService) markPaid(ctx context.Context, orderID uuid.UUID, method order.PaymentMethod) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o.Status == order.OrderStatusPaid || o.Status == order.OrderStatusFulfilled {
		return nil
	}
	if o.Status != order.OrderStatusPendingPayment {
		return apperr.BadRequest("订单不在待支付状态")
	}

	now := time.Now()
	o.Status = order.OrderStatusPaid
	o.PaymentMethod = method
	o.PaidAt = &now
	o.UpdatedAt = now
	if err := s.orderRepo.Update(ctx, o); err != nil {
		return err
	}

	if s.publisher != nil {
		if orderType, _ := o.Metadata["type"].(string); orderType == "tip" {
			receiverID, _ := o.Metadata["to_user_id"].(string)
			go func() {
				_ = s.publisher.Publish(context.Background(), streams.EventTipSent, streams.TipSentPayload{
					TipID:       o.ID.String(),
					SenderID:    o.UserID.String(),
					ReceiverID:  receiverID,
					AmountCents: o.TotalCents,
				})
			}()
		}
	}

	return nil
}
