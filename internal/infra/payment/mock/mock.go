package mock

import (
	"context"
	"fmt"

	payment "github.com/studio/platform/internal/infra/payment"
)

// Gateway is a mock PaymentGateway for testing and development.
type Gateway struct{}

// New creates a new mock gateway.
func New() *Gateway { return &Gateway{} }

func (g *Gateway) CreatePayment(_ context.Context, req payment.CreatePaymentRequest) (*payment.PaymentResult, error) {
	return &payment.PaymentResult{
		PayURL:  fmt.Sprintf("https://mock-pay.example.com/pay?order=%s", req.OutTradeNo),
		QRCode:  fmt.Sprintf("mock://pay?order=%s", req.OutTradeNo),
		TradeNo: "MOCK_" + req.OutTradeNo,
	}, nil
}

func (g *Gateway) QueryPayment(_ context.Context, tradeNo string) (*payment.PaymentStatus, error) {
	return &payment.PaymentStatus{
		TradeNo: tradeNo,
		Status:  "paid",
	}, nil
}

func (g *Gateway) VerifyCallback(_ context.Context, _ map[string]string) (bool, error) {
	return true, nil
}

func (g *Gateway) Refund(_ context.Context, _ payment.RefundRequest) error {
	return nil
}
