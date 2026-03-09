package wechat

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // WeChat Pay v2 requires MD5 for some operations
	"crypto/rand"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/studio/platform/configs"
	payment "github.com/studio/platform/internal/infra/payment"
)

const (
	unifiedOrderURL = "https://api.mch.weixin.qq.com/pay/unifiedorder"
	orderQueryURL   = "https://api.mch.weixin.qq.com/pay/orderquery"
	refundURL       = "https://api.mch.weixin.qq.com/secapi/pay/refund"
)

// Client implements PaymentGateway for WeChat Pay (微信支付 v2 API).
type Client struct {
	cfg        configs.WechatConfig
	httpClient *http.Client
}

// New creates a new WeChat Pay client.
func New(cfg configs.WechatConfig) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// CreatePayment creates a WeChat Pay native QR code payment.
func (c *Client) CreatePayment(ctx context.Context, req payment.CreatePaymentRequest) (*payment.PaymentResult, error) {
	nonceStr := randomString(32)
	params := map[string]string{
		"appid":            c.cfg.AppID,
		"mch_id":           c.cfg.MchID,
		"nonce_str":        nonceStr,
		"body":             req.Subject,
		"out_trade_no":     req.OutTradeNo,
		"total_fee":        fmt.Sprintf("%d", req.AmountCents),
		"spbill_create_ip": "127.0.0.1",
		"notify_url":       c.cfg.NotifyURL,
		"trade_type":       "NATIVE",
	}
	params["sign"] = c.sign(params)

	xmlBody := mapToXML(params)
	respBytes, err := c.postXML(ctx, unifiedOrderURL, xmlBody)
	if err != nil {
		return nil, err
	}

	var resp struct {
		ReturnCode string `xml:"return_code"`
		ResultCode string `xml:"result_code"`
		CodeURL    string `xml:"code_url"`
		ErrCodeDes string `xml:"err_code_des"`
	}
	if err := xml.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("wechat: unmarshal response: %w", err)
	}
	if resp.ReturnCode != "SUCCESS" || resp.ResultCode != "SUCCESS" {
		return nil, fmt.Errorf("wechat: payment failed: %s", resp.ErrCodeDes)
	}

	return &payment.PaymentResult{QRCode: resp.CodeURL}, nil
}

// QueryPayment queries the payment status.
func (c *Client) QueryPayment(ctx context.Context, tradeNo string) (*payment.PaymentStatus, error) {
	params := map[string]string{
		"appid":     c.cfg.AppID,
		"mch_id":    c.cfg.MchID,
		"nonce_str": randomString(32),
		"trade_no":  tradeNo,
	}
	params["sign"] = c.sign(params)

	respBytes, err := c.postXML(ctx, orderQueryURL, mapToXML(params))
	if err != nil {
		return nil, err
	}

	var resp struct {
		TradeState string `xml:"trade_state"`
		TradeNo    string `xml:"transaction_id"`
		OutTradeNo string `xml:"out_trade_no"`
	}
	if err := xml.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("wechat: unmarshal query response: %w", err)
	}

	return &payment.PaymentStatus{
		TradeNo:    resp.TradeNo,
		OutTradeNo: resp.OutTradeNo,
		Status:     tradeStateToInternal(resp.TradeState),
	}, nil
}

// VerifyCallback verifies the signature of a WeChat Pay callback notification.
func (c *Client) VerifyCallback(_ context.Context, params map[string]string) (bool, error) {
	sign, ok := params["sign"]
	if !ok {
		return false, nil
	}
	// Remove sign from params for verification
	paramsCopy := make(map[string]string, len(params))
	for k, v := range params {
		if k != "sign" {
			paramsCopy[k] = v
		}
	}
	expected := c.sign(paramsCopy)
	return hmac.Equal([]byte(sign), []byte(expected)), nil
}

// Refund initiates a refund via WeChat Pay.
func (c *Client) Refund(ctx context.Context, req payment.RefundRequest) error {
	params := map[string]string{
		"appid":         c.cfg.AppID,
		"mch_id":        c.cfg.MchID,
		"nonce_str":     randomString(32),
		"out_trade_no":  req.OutTradeNo,
		"transaction_id": req.TradeNo,
		"out_refund_no": req.OutTradeNo + "_refund",
		"total_fee":     fmt.Sprintf("%d", req.RefundAmount),
		"refund_fee":    fmt.Sprintf("%d", req.RefundAmount),
	}
	params["sign"] = c.sign(params)

	_, err := c.postXML(ctx, refundURL, mapToXML(params))
	return err
}

// --- helpers ---

func (c *Client) sign(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	stringA := strings.Join(parts, "&")
	stringSignTemp := stringA + "&key=" + c.cfg.APIKey
	h := md5.New() //nolint:gosec
	h.Write([]byte(stringSignTemp))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

func (c *Client) postXML(ctx context.Context, apiURL string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wechat: http request: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func mapToXML(params map[string]string) []byte {
	var b strings.Builder
	b.WriteString("<xml>")
	for k, v := range params {
		b.WriteString(fmt.Sprintf("<%s><![CDATA[%s]]></%s>", k, v, k))
	}
	b.WriteString("</xml>")
	return []byte(b.String())
}

func randomString(n int) string {
	b := make([]byte, n/2)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

func tradeStateToInternal(s string) string {
	switch s {
	case "SUCCESS":
		return "paid"
	case "REFUND":
		return "refunded"
	case "CLOSED", "REVOKED":
		return "closed"
	default:
		return "pending"
	}
}
