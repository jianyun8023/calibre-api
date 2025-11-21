package calibre

import "errors"

var (
	// ErrSearchServiceNotAvailable 搜索服务不可用
	ErrSearchServiceNotAvailable = errors.New("search service not available")

	// ErrBookNotFound 书籍未找到
	ErrBookNotFound = errors.New("book not found")

	// ErrInvalidParameters 无效参数
	ErrInvalidParameters = errors.New("invalid parameters")
)
