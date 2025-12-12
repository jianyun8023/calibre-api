package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode string

// 错误码定义
const (
	// 通用错误码 (1xxx)
	CodeInternalError      ErrorCode = "INTERNAL_ERROR"      // 内部错误
	CodeInvalidRequest     ErrorCode = "INVALID_REQUEST"     // 请求参数错误
	CodeUnauthorized       ErrorCode = "UNAUTHORIZED"        // 未授权
	CodeForbidden          ErrorCode = "FORBIDDEN"           // 禁止访问
	CodeNotFound           ErrorCode = "NOT_FOUND"           // 资源未找到
	CodeConflict           ErrorCode = "CONFLICT"            // 资源冲突
	CodeTooManyRequests    ErrorCode = "TOO_MANY_REQUESTS"   // 请求过多
	CodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE" // 服务不可用

	// 业务错误码 (2xxx)
	CodeBookNotFound     ErrorCode = "BOOK_NOT_FOUND"     // 书籍未找到
	CodeInvalidBookID    ErrorCode = "INVALID_BOOK_ID"    // 无效的书籍 ID
	CodeMetadataNotFound ErrorCode = "METADATA_NOT_FOUND" // 元数据未找到
	CodeFileNotFound     ErrorCode = "FILE_NOT_FOUND"     // 文件未找到
	CodeInvalidFormat    ErrorCode = "INVALID_FORMAT"     // 无效的格式
	CodeConversionFailed ErrorCode = "CONVERSION_FAILED"  // 转换失败
	CodeUploadFailed     ErrorCode = "UPLOAD_FAILED"      // 上传失败
	CodeDeleteFailed     ErrorCode = "DELETE_FAILED"      // 删除失败

	// 搜索错误码 (3xxx)
	CodeSearchServiceNotAvailable ErrorCode = "SEARCH_SERVICE_NOT_AVAILABLE" // 搜索服务不可用
	CodeInvalidQuery              ErrorCode = "INVALID_QUERY"                // 无效的查询
	CodeSearchFailed              ErrorCode = "SEARCH_FAILED"                // 搜索失败

	// 任务错误码 (4xxx)
	CodeTaskNotFound       ErrorCode = "TASK_NOT_FOUND"       // 任务未找到
	CodeTaskAlreadyRunning ErrorCode = "TASK_ALREADY_RUNNING" // 任务已在运行
	CodeTaskFailed         ErrorCode = "TASK_FAILED"          // 任务失败

	// 聊天错误码 (5xxx)
	CodeChatServiceNotAvailable ErrorCode = "CHAT_SERVICE_NOT_AVAILABLE" // 聊天服务不可用
	CodeConversationNotFound    ErrorCode = "CONVERSATION_NOT_FOUND"     // 会话未找到
	CodeMessageNotFound         ErrorCode = "MESSAGE_NOT_FOUND"          // 消息未找到
	CodeChatFailed              ErrorCode = "CHAT_FAILED"                // 聊天失败

	// 验证错误码 (6xxx)
	CodeValidationFailed ErrorCode = "VALIDATION_FAILED" // 验证失败
	CodeInvalidParameter ErrorCode = "INVALID_PARAMETER" // 无效参数
	CodeMissingParameter ErrorCode = "MISSING_PARAMETER" // 缺少参数
)

// AppError 应用错误类型
type AppError struct {
	Code       ErrorCode              // 错误码
	Message    string                 // 错误消息（用户可见）
	Details    string                 // 错误详情（开发者可见）
	StatusCode int                    // HTTP 状态码
	Err        error                  // 原始错误
	Context    map[string]interface{} // 错误上下文
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithContext 添加错误上下文
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// New 创建新的应用错误
func New(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap 包装已有错误
func Wrap(err error, code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// WrapWithDetails 包装错误并添加详情
func WrapWithDetails(err error, code ErrorCode, message, details string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
		Err:        err,
	}
}

// IsAppError 判断是否为应用错误
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError 将错误转换为应用错误
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// 预定义的错误构造函数

// NewInternalError 创建内部错误
func NewInternalError(message string, err error) *AppError {
	return Wrap(err, CodeInternalError, message, http.StatusInternalServerError)
}

// NewInvalidRequestError 创建请求参数错误
func NewInvalidRequestError(message string) *AppError {
	return New(CodeInvalidRequest, message, http.StatusBadRequest)
}

// NewNotFoundError 创建资源未找到错误
func NewNotFoundError(resource string) *AppError {
	return New(CodeNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

// NewServiceUnavailableError 创建服务不可用错误
func NewServiceUnavailableError(service string) *AppError {
	return New(CodeServiceUnavailable, fmt.Sprintf("%s service not available", service), http.StatusServiceUnavailable)
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string) *AppError {
	return New(CodeValidationFailed, fmt.Sprintf("validation failed for field '%s': %s", field, message), http.StatusBadRequest).
		WithContext("field", field)
}

// 业务错误构造函数

// NewBookNotFoundError 创建书籍未找到错误
func NewBookNotFoundError(bookID string) *AppError {
	return New(CodeBookNotFound, "book not found", http.StatusNotFound).
		WithContext("book_id", bookID)
}

// NewInvalidBookIDError 创建无效书籍 ID 错误
func NewInvalidBookIDError(bookID string) *AppError {
	return New(CodeInvalidBookID, "invalid book ID", http.StatusBadRequest).
		WithContext("book_id", bookID)
}

// NewMetadataNotFoundError 创建元数据未找到错误
func NewMetadataNotFoundError(bookID string) *AppError {
	return New(CodeMetadataNotFound, "metadata not found", http.StatusNotFound).
		WithContext("book_id", bookID)
}

// NewFileNotFoundError 创建文件未找到错误
func NewFileNotFoundError(filename string) *AppError {
	return New(CodeFileNotFound, "file not found", http.StatusNotFound).
		WithContext("filename", filename)
}

// NewSearchServiceNotAvailableError 创建搜索服务不可用错误
func NewSearchServiceNotAvailableError() *AppError {
	return New(CodeSearchServiceNotAvailable, "search service not available", http.StatusServiceUnavailable)
}

// NewInvalidQueryError 创建无效查询错误
func NewInvalidQueryError(query string) *AppError {
	return New(CodeInvalidQuery, "invalid query", http.StatusBadRequest).
		WithContext("query", query)
}

// NewTaskNotFoundError 创建任务未找到错误
func NewTaskNotFoundError(taskID string) *AppError {
	return New(CodeTaskNotFound, "task not found", http.StatusNotFound).
		WithContext("task_id", taskID)
}

// NewTaskAlreadyRunningError 创建任务已在运行错误
func NewTaskAlreadyRunningError(taskType string) *AppError {
	return New(CodeTaskAlreadyRunning, "task already running", http.StatusConflict).
		WithContext("task_type", taskType)
}

// NewChatServiceNotAvailableError 创建聊天服务不可用错误
func NewChatServiceNotAvailableError() *AppError {
	return New(CodeChatServiceNotAvailable, "chat service not available", http.StatusServiceUnavailable)
}

// NewConversationNotFoundError 创建会话未找到错误
func NewConversationNotFoundError(conversationID string) *AppError {
	return New(CodeConversationNotFound, "conversation not found", http.StatusNotFound).
		WithContext("conversation_id", conversationID)
}

// 错误处理辅助函数

// GetHTTPStatus 获取错误的 HTTP 状态码
func GetHTTPStatus(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) ErrorCode {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return CodeInternalError
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(err error) string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Message
	}
	return err.Error()
}
