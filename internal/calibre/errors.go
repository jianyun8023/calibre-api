package calibre

import (
	"errors"
	"fmt"
)

var (
	// ErrSearchServiceNotAvailable 搜索服务不可用
	ErrSearchServiceNotAvailable = errors.New("search service not available")

	// ErrBookNotFound 书籍未找到
	ErrBookNotFound = errors.New("book not found")

	// ErrInvalidParameters 无效参数
	ErrInvalidParameters = errors.New("invalid parameters")

	// ErrInvalidBookID 无效的书籍 ID
	ErrInvalidBookID = errors.New("invalid book ID")

	// ErrTaskAlreadyRunning 任务已在运行
	ErrTaskAlreadyRunning = errors.New("task already running")

	// ErrTaskNotFound 任务未找到
	ErrTaskNotFound = errors.New("task not found")

	// ErrInvalidQuery 无效的查询
	ErrInvalidQuery = errors.New("invalid query")

	// ErrInvalidLimit 无效的限制参数
	ErrInvalidLimit = errors.New("invalid limit parameter")

	// ErrInvalidOffset 无效的偏移参数
	ErrInvalidOffset = errors.New("invalid offset parameter")

	// ErrMetadataNotFound 元数据未找到
	ErrMetadataNotFound = errors.New("metadata not found")

	// ErrFileNotFound 文件未找到
	ErrFileNotFound = errors.New("file not found")

	// ErrCacheError 缓存错误
	ErrCacheError = errors.New("cache error")
)

// WrapError wraps an error with additional context
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// NewValidationError creates a new validation error with details
func NewValidationError(field, message string) error {
	return fmt.Errorf("validation error for field '%s': %s", field, message)
}
