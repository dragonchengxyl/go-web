package response

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/apperr"
)

// Response represents the unified API response format
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp int64       `json:"timestamp"`
}

// Success sends a success response
func Success(c *gin.Context, data interface{}) {
	requestID, _ := c.Get("request_id")
	c.JSON(200, Response{
		Code:      apperr.CodeSuccess,
		Message:   "success",
		Data:      data,
		RequestID: requestID.(string),
		Timestamp: time.Now().Unix(),
	})
}

// PaginationData represents paginated response data
type PaginationData struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// SuccessWithPagination sends a success response with pagination info
func SuccessWithPagination(c *gin.Context, items interface{}, total, page, pageSize int) {
	requestID, _ := c.Get("request_id")
	c.JSON(200, Response{
		Code:    apperr.CodeSuccess,
		Message: "success",
		Data: PaginationData{
			Items: items,
			Total: total,
			Page:  page,
			Size:  pageSize,
		},
		RequestID: requestID.(string),
		Timestamp: time.Now().Unix(),
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	requestID, _ := c.Get("request_id")

	appErr, ok := err.(*apperr.AppError)
	if !ok {
		appErr = apperr.Wrap(apperr.CodeInternalError, "内部错误", err)
	}

	httpStatus := getHTTPStatus(appErr.Code)

	c.JSON(httpStatus, Response{
		Code:      appErr.Code,
		Message:   appErr.Message,
		RequestID: requestID.(string),
		Timestamp: time.Now().Unix(),
	})
}

// getHTTPStatus maps error code to HTTP status
func getHTTPStatus(code int) int {
	switch {
	case code >= 40000 && code < 41000:
		return 400
	case code >= 40100 && code < 40200:
		return 401
	case code >= 40300 && code < 40400:
		return 403
	case code >= 40400 && code < 40500:
		return 404
	case code >= 40900 && code < 41000:
		return 409
	case code >= 42900 && code < 43000:
		return 429
	case code >= 50000:
		return 500
	default:
		return 500
	}
}
