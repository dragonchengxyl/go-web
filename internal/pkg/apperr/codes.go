package apperr

// Error codes
const (
	// Success
	CodeSuccess = 0

	// Client errors (4xxxx)
	CodeInvalidParam      = 40001
	CodeValidationFailed  = 40002
	CodeUnauthorized      = 40101
	CodeTokenExpired      = 40102
	CodeForbidden         = 40301
	CodeNotFound          = 40401
	CodeEmailExists       = 40901
	CodeUsernameExists    = 40902
	CodeRateLimited       = 42901

	// Server errors (5xxxx)
	CodeInternalError     = 50001
	CodeDependencyFailed  = 50002
	CodeStorageFailed     = 50003
)

// Predefined errors
var (
	ErrInvalidParam      = New(CodeInvalidParam, "参数错误")
	ErrValidationFailed  = New(CodeValidationFailed, "验证失败")
	ErrUnauthorized      = New(CodeUnauthorized, "未认证")
	ErrTokenExpired      = New(CodeTokenExpired, "令牌已过期")
	ErrForbidden         = New(CodeForbidden, "无权限")
	ErrNotFound          = New(CodeNotFound, "资源不存在")
	ErrEmailExists       = New(CodeEmailExists, "邮箱已存在")
	ErrUsernameExists    = New(CodeUsernameExists, "用户名已存在")
	ErrRateLimited       = New(CodeRateLimited, "请求过于频繁")
	ErrInternalError     = New(CodeInternalError, "内部错误")
	ErrDependencyFailed  = New(CodeDependencyFailed, "依赖服务失败")
	ErrStorageFailed     = New(CodeStorageFailed, "存储失败")
)
