package handler

import (
	"encoding/xml"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/transport/http/middleware"
	"github.com/studio/platform/internal/usecase"
)

// PaymentHandler handles payment-related HTTP endpoints.
type PaymentHandler struct {
	paymentService *usecase.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(paymentService *usecase.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// PayWithAlipay initiates an Alipay page-pay for the given order.
// POST /api/v1/orders/:id/pay/alipay
func (h *PaymentHandler) PayWithAlipay(c *gin.Context) {
	_ = middleware.GetUserID(c)
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的订单ID"))
		return
	}

	var req struct {
		ReturnURL string `json:"return_url"`
	}
	_ = c.ShouldBindJSON(&req)

	result, err := h.paymentService.InitiateAlipayPayment(c.Request.Context(), orderID, req.ReturnURL)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"pay_url": result.PayURL})
}

// PayWithWechat initiates a WeChat Pay native QR code for the given order.
// POST /api/v1/orders/:id/pay/wechat
func (h *PaymentHandler) PayWithWechat(c *gin.Context) {
	_ = middleware.GetUserID(c)
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的订单ID"))
		return
	}

	result, err := h.paymentService.InitiateWechatPayment(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"qr_code": result.QRCode})
}

// AlipayCallback handles Alipay async payment notifications.
// POST /api/v1/payment/callback/alipay  (no auth required)
func (h *PaymentHandler) AlipayCallback(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(400, "fail")
		return
	}
	params := make(map[string]string, len(c.Request.PostForm))
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	if err := h.paymentService.HandleAlipayCallback(c.Request.Context(), params); err != nil {
		c.String(400, "fail")
		return
	}
	c.String(200, "success")
}

// WechatCallback handles WeChat Pay async payment notifications.
// POST /api/v1/payment/callback/wechat  (no auth required)
func (h *PaymentHandler) WechatCallback(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.XML(400, wechatCallbackReply{ReturnCode: "FAIL", ReturnMsg: "read body failed"})
		return
	}

	// Parse WeChat Pay XML notification: <xml><key><![CDATA[value]]></key>...</xml>
	var envelope struct {
		Items []xmlItem `xml:",any"`
	}
	params := make(map[string]string)
	if xmlErr := xml.Unmarshal(body, &envelope); xmlErr == nil {
		for _, item := range envelope.Items {
			params[item.XMLName.Local] = item.Value
		}
	}

	if err := h.paymentService.HandleWechatCallback(c.Request.Context(), params); err != nil {
		c.XML(400, wechatCallbackReply{ReturnCode: "FAIL", ReturnMsg: err.Error()})
		return
	}
	c.XML(200, wechatCallbackReply{ReturnCode: "SUCCESS", ReturnMsg: "OK"})
}

type wechatCallbackReply struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
}

type xmlItem struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}
