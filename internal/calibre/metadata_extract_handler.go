package calibre

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/response"
	"github.com/kapmahc/epub"
)

// extractMetadata 从书籍文件中抽取元数据（版权页抽取）
// POST /api/book/:id/extract-metadata
func (c *Api) extractMetadata(r *gin.Context) {
	id := r.Param("id")

	// 1. 验证 cacheManager 可用
	if c.cacheManager == nil {
		response.BadRequest(r, "cache manager not available")
		return
	}

	// 2. 获取 EPUB 文件
	epubPath, err := c.cacheManager.GetOrExtractEpub(id)
	if err != nil {
		response.InternalError(r, "failed to get EPUB file", err)
		return
	}

	// 3. 打开 EPUB
	book, err := epub.Open(epubPath)
	if err != nil {
		response.InternalError(r, "failed to open EPUB", err)
		return
	}
	defer book.Close()

	// 4. 查找版权页
	copyrightPath, err := tasks.FindCopyrightPage(book)
	if err != nil {
		response.NotFound(r, "copyright page not found in book TOC")
		return
	}

	// 5. 读取页面内容
	content, err := tasks.ReadPageContent(book, copyrightPath)
	if err != nil {
		response.InternalError(r, "failed to read copyright page", err)
		return
	}

	// 6. 解析元数据
	metadata, err := tasks.ParseMetadataFromContent(content)
	if err != nil {
		// 如果解析失败（如没有找到 ISBN），返回空元数据但状态为成功
		response.Success(r, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	// 7. 返回抽取的元数据
	bookIDInt, _ := strconv.ParseInt(id, 10, 64)
	response.Success(r, gin.H{
		"success": true,
		"message": "metadata extracted successfully",
		"data": gin.H{
			"book_id":      bookIDInt,
			"isbn":         metadata.ISBN,
			"book_title":   metadata.BookTitle,
			"author":       metadata.Author,
			"translator":   metadata.Translator,
			"publisher":    metadata.Publisher,
			"publish_date": metadata.PublishDate,
		},
	})
}
