package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/jianyun8023/calibre-api/pkg/errors"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`               // HTTP 状态码
	Message string      `json:"message"`            // 响应消息
	Data    interface{} `json:"data,omitempty"`     // 响应数据
	Error   *ErrorInfo  `json:"error,omitempty"`    // 错误信息（仅错误响应）
	TraceID string      `json:"trace_id,omitempty"` // 追踪 ID（可选）
}

// ErrorInfo 错误详情
type ErrorInfo struct {
	Code    string                 `json:"code"`              // 错误码
	Message string                 `json:"message"`           // 错误消息
	Details string                 `json:"details,omitempty"` // 错误详情
	Context map[string]interface{} `json:"context,omitempty"` // 错误上下文
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Code       int         `json:"code"`               // HTTP 状态码
	Message    string      `json:"message"`            // 响应消息
	Data       interface{} `json:"data"`               // 响应数据
	Pagination *Pagination `json:"pagination"`         // 分页信息
	TraceID    string      `json:"trace_id,omitempty"` // 追踪 ID（可选）
}

// Pagination 分页信息
type Pagination struct {
	Total      int64 `json:"total"`       // 总数
	Page       int   `json:"page"`        // 当前页码（从 1 开始）
	PageSize   int   `json:"page_size"`   // 每页大小
	TotalPages int   `json:"total_pages"` // 总页数
}

// Builder 响应构建器
type Builder struct {
	ctx     *gin.Context
	code    int
	message string
	data    interface{}
	err     error
	traceID string
}

// NewBuilder 创建响应构建器
func NewBuilder(c *gin.Context) *Builder {
	return &Builder{
		ctx: c,
	}
}

// WithCode 设置状态码
func (b *Builder) WithCode(code int) *Builder {
	b.code = code
	return b
}

// WithMessage 设置消息
func (b *Builder) WithMessage(message string) *Builder {
	b.message = message
	return b
}

// WithData 设置数据
func (b *Builder) WithData(data interface{}) *Builder {
	b.data = data
	return b
}

// WithError 设置错误
func (b *Builder) WithError(err error) *Builder {
	b.err = err
	return b
}

// WithTraceID 设置追踪 ID
func (b *Builder) WithTraceID(traceID string) *Builder {
	b.traceID = traceID
	return b
}

// Build 构建并发送响应
func (b *Builder) Build() {
	if b.err != nil {
		b.sendError()
	} else {
		b.sendSuccess()
	}
}

// sendSuccess 发送成功响应
func (b *Builder) sendSuccess() {
	if b.code == 0 {
		b.code = http.StatusOK
	}
	if b.message == "" {
		b.message = "success"
	}

	resp := Response{
		Code:    b.code,
		Message: b.message,
		Data:    b.data,
		TraceID: b.traceID,
	}

	b.ctx.JSON(b.code, resp)
}

// sendError 发送错误响应
func (b *Builder) sendError() {
	// 如果是应用错误，提取详细信息
	if appErr, ok := apperrors.AsAppError(b.err); ok {
		if b.code == 0 {
			b.code = appErr.StatusCode
		}
		if b.message == "" {
			b.message = "error"
		}

		resp := Response{
			Code:    b.code,
			Message: b.message,
			Error: &ErrorInfo{
				Code:    string(appErr.Code),
				Message: appErr.Message,
				Details: appErr.Details,
				Context: appErr.Context,
			},
			TraceID: b.traceID,
		}

		b.ctx.JSON(b.code, resp)
		return
	}

	// 普通错误
	if b.code == 0 {
		b.code = http.StatusInternalServerError
	}
	if b.message == "" {
		b.message = "error"
	}

	resp := Response{
		Code:    b.code,
		Message: b.message,
		Error: &ErrorInfo{
			Code:    "INTERNAL_ERROR",
			Message: b.err.Error(),
		},
		TraceID: b.traceID,
	}

	b.ctx.JSON(b.code, resp)
}

// 便捷函数

// Success 发送成功响应
func Success(c *gin.Context, data interface{}) {
	NewBuilder(c).WithData(data).Build()
}

// SuccessWithMessage 发送带消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	NewBuilder(c).WithMessage(message).WithData(data).Build()
}

// Error 发送错误响应
func Error(c *gin.Context, err error) {
	NewBuilder(c).WithError(err).Build()
}

// ErrorWithCode 发送带状态码的错误响应
func ErrorWithCode(c *gin.Context, code int, err error) {
	NewBuilder(c).WithCode(code).WithError(err).Build()
}

// ErrorWithMessage 发送带消息的错误响应
func ErrorWithMessage(c *gin.Context, message string, err error) {
	NewBuilder(c).WithMessage(message).WithError(err).Build()
}

// BadRequest 发送 400 错误响应
func BadRequest(c *gin.Context, message string) {
	NewBuilder(c).
		WithCode(http.StatusBadRequest).
		WithMessage("bad request").
		WithError(apperrors.NewInvalidRequestError(message)).
		Build()
}

// NotFound 发送 404 错误响应
func NotFound(c *gin.Context, resource string) {
	NewBuilder(c).
		WithCode(http.StatusNotFound).
		WithMessage("not found").
		WithError(apperrors.NewNotFoundError(resource)).
		Build()
}

// InternalError 发送 500 错误响应
func InternalError(c *gin.Context, message string, err error) {
	NewBuilder(c).
		WithCode(http.StatusInternalServerError).
		WithMessage("internal error").
		WithError(apperrors.NewInternalError(message, err)).
		Build()
}

// ServiceUnavailable 发送 503 错误响应
func ServiceUnavailable(c *gin.Context, service string) {
	NewBuilder(c).
		WithCode(http.StatusServiceUnavailable).
		WithMessage("service unavailable").
		WithError(apperrors.NewServiceUnavailableError(service)).
		Build()
}

// Paginated 发送分页响应
func Paginated(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	resp := PaginatedResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		Pagination: &Pagination{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// PaginatedWithMessage 发送带消息的分页响应
func PaginatedWithMessage(c *gin.Context, message string, data interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	resp := PaginatedResponse{
		Code:    http.StatusOK,
		Message: message,
		Data:    data,
		Pagination: &Pagination{
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	}

	c.JSON(http.StatusOK, resp)
}

// AbortWithError 中止请求并发送错误响应
func AbortWithError(c *gin.Context, err error) {
	c.Abort()
	Error(c, err)
}

// AbortWithErrorCode 中止请求并发送带状态码的错误响应
func AbortWithErrorCode(c *gin.Context, code int, err error) {
	c.Abort()
	ErrorWithCode(c, code, err)
}
