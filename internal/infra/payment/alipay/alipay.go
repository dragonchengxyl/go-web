package alipay

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/studio/platform/configs"
	payment "github.com/studio/platform/internal/infra/payment"
)

const (
	gatewayURL        = "https://openapi.alipay.com/gateway.do"
	sandboxGatewayURL = "https://openapi-sandbox.dl.alipaydev.com/gateway.do"
)

// Client implements PaymentGateway for Alipay (支付宝).
type Client struct {
	cfg        configs.AlipayConfig
	httpClient *http.Client
	privateKey *rsa.PrivateKey
}

// New creates a new Alipay client.
func New(cfg configs.AlipayConfig) (*Client, error) {
	privKey, err := parseRSAPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("alipay: parse private key: %w", err)
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		privateKey: privKey,
	}, nil
}

// CreatePayment generates an Alipay PC page-pay URL (alipay.trade.page.pay).
func (c *Client) CreatePayment(_ context.Context, req payment.CreatePaymentRequest) (*payment.PaymentResult, error) {
	bizContent, _ := json.Marshal(map[string]string{
		"out_trade_no": req.OutTradeNo,
		"product_code": "FAST_INSTANT_TRADE_PAY",
		"total_amount": fmt.Sprintf("%.2f", float64(req.AmountCents)/100),
		"subject":      req.Subject,
	})

	params := c.buildCommonParams("alipay.trade.page.pay")
	params.Set("biz_content", string(bizContent))
	params.Set("return_url", req.ReturnURL)
	params.Set("notify_url", c.cfg.NotifyURL)

	sign, err := c.sign(params)
	if err != nil {
		return nil, fmt.Errorf("alipay: sign params: %w", err)
	}
	params.Set("sign", sign)

	payURL := c.gatewayURL() + "?" + params.Encode()
	return &payment.PaymentResult{PayURL: payURL}, nil
}

// QueryPayment queries trade status via alipay.trade.query.
func (c *Client) QueryPayment(ctx context.Context, tradeNo string) (*payment.PaymentStatus, error) {
	bizContent, _ := json.Marshal(map[string]string{"trade_no": tradeNo})
	params := c.buildCommonParams("alipay.trade.query")
	params.Set("biz_content", string(bizContent))
	sign, err := c.sign(params)
	if err != nil {
		return nil, err
	}
	params.Set("sign", sign)

	body, err := c.post(ctx, params)
	if err != nil {
		return nil, err
	}

	var resp map[string]json.RawMessage
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("alipay: unmarshal query response: %w", err)
	}
	var queryResp struct {
		TradeNo    string `json:"trade_no"`
		OutTradeNo string `json:"out_trade_no"`
		TradeStatus string `json:"trade_status"`
	}
	if err := json.Unmarshal(resp["alipay_trade_query_response"], &queryResp); err != nil {
		return nil, fmt.Errorf("alipay: unmarshal trade query: %w", err)
	}

	status := tradeStatusToInternal(queryResp.TradeStatus)
	return &payment.PaymentStatus{
		TradeNo:    queryResp.TradeNo,
		OutTradeNo: queryResp.OutTradeNo,
		Status:     status,
	}, nil
}

// VerifyCallback verifies the RSA signature in an Alipay async notification.
func (c *Client) VerifyCallback(_ context.Context, params map[string]string) (bool, error) {
	sign, ok := params["sign"]
	if !ok {
		return false, nil
	}
	signType := params["sign_type"]
	if signType != "RSA2" {
		return false, fmt.Errorf("alipay: unsupported sign_type %q", signType)
	}

	// Build sorted string to verify
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "sign_type" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	message := strings.Join(parts, "&")

	pubKey, err := parseRSAPublicKey(c.cfg.PublicKey)
	if err != nil {
		return false, fmt.Errorf("alipay: parse public key: %w", err)
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return false, fmt.Errorf("alipay: decode signature: %w", err)
	}

	h := sha256.New()
	h.Write([]byte(message))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, h.Sum(nil), sigBytes); err != nil {
		return false, nil //nolint:nilerr // invalid signature is not a system error
	}
	return true, nil
}

// Refund issues a refund via alipay.trade.refund.
func (c *Client) Refund(ctx context.Context, req payment.RefundRequest) error {
	bizContent, _ := json.Marshal(map[string]any{
		"trade_no":      req.TradeNo,
		"out_trade_no":  req.OutTradeNo,
		"refund_amount": fmt.Sprintf("%.2f", float64(req.RefundAmount)/100),
		"refund_reason": req.Reason,
		"out_request_no": req.OutTradeNo + "_refund",
	})
	params := c.buildCommonParams("alipay.trade.refund")
	params.Set("biz_content", string(bizContent))
	sign, err := c.sign(params)
	if err != nil {
		return err
	}
	params.Set("sign", sign)
	_, err = c.post(ctx, params)
	return err
}

// --- helpers ---

func (c *Client) gatewayURL() string {
	if c.cfg.Sandbox {
		return sandboxGatewayURL
	}
	return gatewayURL
}

func (c *Client) buildCommonParams(method string) url.Values {
	p := url.Values{}
	p.Set("app_id", c.cfg.AppID)
	p.Set("method", method)
	p.Set("format", "JSON")
	p.Set("charset", "utf-8")
	p.Set("sign_type", "RSA2")
	p.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	p.Set("version", "1.0")
	return p
}

func (c *Client) sign(params url.Values) (string, error) {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params.Get(k))
	}
	message := strings.Join(parts, "&")

	h := sha256.New()
	h.Write([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

func (c *Client) post(ctx context.Context, params url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.gatewayURL(), strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("alipay: http request: %w", err)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func tradeStatusToInternal(s string) string {
	switch s {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		return "paid"
	case "WAIT_BUYER_PAY":
		return "pending"
	case "TRADE_CLOSED":
		return "closed"
	default:
		return "pending"
	}
}

func parseRSAPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		// Try wrapping with PKCS8 header
		pemStr = "-----BEGIN PRIVATE KEY-----\n" + pemStr + "\n-----END PRIVATE KEY-----"
		block, _ = pem.Decode([]byte(pemStr))
	}
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS1
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		pemStr = "-----BEGIN PUBLIC KEY-----\n" + pemStr + "\n-----END PUBLIC KEY-----"
		block, _ = pem.Decode([]byte(pemStr))
	}
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaKey, nil
}
