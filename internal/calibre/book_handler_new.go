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

// GetBookByIDInternal 内部方法，供其他 Handler 使用
func (h *BookHandlerV2) GetBookByIDInternal(id string) (*service.Book, error) {
	return h.bookService.GetBookByID(id)
}

// API 路由适配方法

// getBookV2 获取书籍信息
func (c *Api) getBookV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.GetBook(r)
}

// deleteBookV2 删除书籍
func (c *Api) deleteBookV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.DeleteBook(r)
}

// updateMetadataV2 更新书籍元数据
func (c *Api) updateMetadataV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.UpdateMetadata(r)
}

// recentlyV2 获取最近更新的书籍
func (c *Api) recentlyV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.GetRecentBooks(r)
}

// randomV2 获取随机书籍
func (c *Api) randomV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.GetRandomBooks(r)
}

// getAllBooksV2 获取所有书籍（游标分页）
func (c *Api) getAllBooksV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.GetAllBooks(r)
}

// listPublisherV2 获取所有出版社列表
func (c *Api) listPublisherV2(r *gin.Context) {
	if c.bookHandler == nil {
		response.Error(r, ErrSearchServiceNotAvailable)
		return
	}
	c.bookHandler.ListPublishers(r)
}

// getBookByIDV2 替代旧的私有方法，供内部使用
func (c *Api) getBookByIDV2(id string) (*Book, error) {
	if c.bookHandler == nil {
		return nil, ErrSearchServiceNotAvailable
	}

	serviceBook, err := c.bookHandler.GetBookByIDInternal(id)
	if err != nil {
		return nil, err
	}

	// 转换 service.Book 到 calibre.Book
	book := &Book{
		ID:           serviceBook.ID,
		Title:        serviceBook.Title,
		Authors:      serviceBook.Authors,
		Publisher:    serviceBook.Publisher,
		PubDate:      serviceBook.PubDate,
		Isbn:         serviceBook.Isbn,
		Tags:         serviceBook.Tags,
		Rating:       serviceBook.Rating,
		SeriesIndex:  serviceBook.SeriesIndex,
		Comments:     serviceBook.Comments,
		Languages:    serviceBook.Languages,
		LastModified: serviceBook.LastModified,
		Cover:        serviceBook.Cover,
		FilePath:     serviceBook.FilePath,
		Identifiers:  serviceBook.Identifiers,
		Size:         serviceBook.Size,
		AuthorSort:   "", // service.Book 不包含 AuthorSort
	}
	return book, nil
}
