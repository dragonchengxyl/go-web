package payment

import (
	"context"
	"time"
)

// PaymentGateway defines the interface for payment providers.
type PaymentGateway interface {
	// CreatePayment initiates a payment and returns redirect URL or QR code content.
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResult, error)
	// QueryPayment queries the payment status from the upstream gateway.
	QueryPayment(ctx context.Context, tradeNo string) (*PaymentStatus, error)
	// VerifyCallback verifies the authenticity of an async payment callback.
	VerifyCallback(ctx context.Context, params map[string]string) (bool, error)
	// Refund initiates a refund for a completed payment.
	Refund(ctx context.Context, req RefundRequest) error
}

// CreatePaymentRequest contains the data needed to initiate a payment.
type CreatePaymentRequest struct {
	OutTradeNo  string // Our order ID used as the trade number
	Subject     string // Order subject / description
	AmountCents int64  // Amount in cents (¥ × 100)
	ReturnURL   string // Synchronous redirect URL (for Alipay page pay)
}

// PaymentResult holds the result of a CreatePayment call.
type PaymentResult struct {
	// PayURL is set for Alipay redirect-based payment (web page).
	PayURL string
	// QRCode is set for WeChat Pay native QR code payment.
	QRCode string
	// TradeNo is the gateway-assigned trade number (empty until payment completes).
	TradeNo string
	// RawResponse is the raw response body from the gateway for auditing.
	RawResponse string
}

// PaymentStatus describes the current status of a payment as reported by the gateway.
type PaymentStatus struct {
	TradeNo    string
	OutTradeNo string
	Status     string // "pending" | "paid" | "closed" | "refunded"
	PaidAt     *time.Time
}

// RefundRequest contains the data needed to initiate a refund.
type RefundRequest struct {
	OutTradeNo   string // Our original order ID
	TradeNo      string // Gateway trade number
	RefundAmount int64  // Amount to refund in cents
	Reason       string
}
