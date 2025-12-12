package calibre

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/service"
	"github.com/jianyun8023/calibre-api/pkg/response"
)

// BookHandlerV2 使用 Service 层的新版本 Handler
type BookHandlerV2 struct {
	bookService service.BookService
}

// NewBookHandler 创建书籍 Handler
func NewBookHandler(bookService service.BookService) *BookHandlerV2 {
	return &BookHandlerV2{
		bookService: bookService,
	}
}

// GetBook 获取书籍信息
func (h *BookHandlerV2) GetBook(c *gin.Context) {
	id := c.Param("id")

	book, err := h.bookService.GetBookByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, book)
}

// DeleteBook 删除书籍
func (h *BookHandlerV2) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	err := h.bookService.DeleteBook(id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "book deleted successfully", gin.H{"deleted": true})
}

// UpdateMetadata 更新书籍元数据
func (h *BookHandlerV2) UpdateMetadata(c *gin.Context) {
	id := c.Param("id")

	var updates service.BookUpdate
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	err := h.bookService.UpdateMetadata(id, &updates)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithMessage(c, "metadata updated successfully", gin.H{"updated": true})
}

// GetRecentBooks 获取最近更新的书籍
func (h *BookHandlerV2) GetRecentBooks(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		response.BadRequest(c, "invalid limit parameter")
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil {
		response.BadRequest(c, "invalid offset parameter")
		return
	}

	books, total, err := h.bookService.GetRecentBooks(limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 计算页码
	page := (offset / limit) + 1
	response.Paginated(c, books, total, page, limit)
}

// GetRandomBooks 获取随机书籍
func (h *BookHandlerV2) GetRandomBooks(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		response.BadRequest(c, "invalid limit parameter")
		return
	}

	books, err := h.bookService.GetRandomBooks(limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, books)
}

// GetAllBooks 获取所有书籍（游标分页）
func (h *BookHandlerV2) GetAllBooks(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "12"))
	if err != nil {
		response.BadRequest(c, "invalid limit parameter")
		return
	}

	cursor := c.DefaultQuery("cursor", "")

	books, total, nextCursor, err := h.bookService.GetAllBooks(limit, cursor)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"records":     books,
		"total":       total,
		"limit":       limit,
		"next_cursor": nextCursor,
	})
}

// ListPublishers 获取所有出版社列表
func (h *BookHandlerV2) ListPublishers(c *gin.Context) {
	publishers, err := h.bookService.ListPublishers()
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, publishers)
}

// 保持与旧 Handler 的兼容性，添加适配方法

// getBook 旧版 API 适配
func (c *Api) getBookV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.GetBook(r)
		return
	}
	// 降级到旧实现
	c.getBook(r)
}

// deleteBook 旧版 API 适配
func (c *Api) deleteBookV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.DeleteBook(r)
		return
	}
	// 降级到旧实现
	c.deleteBook(r)
}

// updateMetadata 旧版 API 适配
func (c *Api) updateMetadataV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.UpdateMetadata(r)
		return
	}
	// 降级到旧实现
	c.updateMetadata(r)
}

// recently 旧版 API 适配
func (c *Api) recentlyV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.GetRecentBooks(r)
		return
	}
	// 降级到旧实现
	c.recently(r)
}

// random 旧版 API 适配
func (c *Api) randomV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.GetRandomBooks(r)
		return
	}
	// 降级到旧实现
	c.random(r)
}

// getAllBooks 旧版 API 适配
func (c *Api) getAllBooksV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.GetAllBooks(r)
		return
	}
	// 降级到旧实现
	c.getAllBooks(r)
}

// listPublisher 旧版 API 适配
func (c *Api) listPublisherV2(r *gin.Context) {
	if c.bookHandler != nil {
		c.bookHandler.ListPublishers(r)
		return
	}
	// 降级到旧实现
	c.listPublisher(r)
}
