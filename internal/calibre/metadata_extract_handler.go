package calibre

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/response"
	"github.com/kapmahc/epub"
)

// parseChineseDate 解析中文日期格式，转换为 ISO 格式 (YYYY-MM-DD)
// 支持格式: "2025年2月", "2025年02月", "2025年2月1日", "2025-02", "2025.2" 等
func parseChineseDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// 尝试多种正则匹配
	patterns := []struct {
		regex  *regexp.Regexp
		format func([]string) string
	}{
		// "2025年2月1日" 或 "2025年02月01日"
		{
			regexp.MustCompile(`(\d{4})\s*年\s*(\d{1,2})\s*月\s*(\d{1,2})\s*日?`),
			func(m []string) string {
				year := m[1]
				month := m[2]
				day := m[3]
				if len(month) == 1 {
					month = "0" + month
				}
				if len(day) == 1 {
					day = "0" + day
				}
				return year + "-" + month + "-" + day
			},
		},
		// "2025年2月" 或 "2025年02月"
		{
			regexp.MustCompile(`(\d{4})\s*年\s*(\d{1,2})\s*月`),
			func(m []string) string {
				year := m[1]
				month := m[2]
				if len(month) == 1 {
					month = "0" + month
				}
				return year + "-" + month + "-01"
			},
		},
		// "2025-02" 或 "2025-2"
		{
			regexp.MustCompile(`(\d{4})[-/.](\d{1,2})$`),
			func(m []string) string {
				year := m[1]
				month := m[2]
				if len(month) == 1 {
					month = "0" + month
				}
				return year + "-" + month + "-01"
			},
		},
		// "2025-02-01"
		{
			regexp.MustCompile(`(\d{4})[-/.](\d{1,2})[-/.](\d{1,2})`),
			func(m []string) string {
				year := m[1]
				month := m[2]
				day := m[3]
				if len(month) == 1 {
					month = "0" + month
				}
				if len(day) == 1 {
					day = "0" + day
				}
				return year + "-" + month + "-" + day
			},
		},
		// "2025年"
		{
			regexp.MustCompile(`(\d{4})\s*年`),
			func(m []string) string {
				return m[1] + "-01-01"
			},
		},
	}

	// 清理空白字符
	dateStr = strings.TrimSpace(dateStr)

	for _, p := range patterns {
		if matches := p.regex.FindStringSubmatch(dateStr); len(matches) > 1 {
			return p.format(matches)
		}
	}

	// 无法解析，返回原始字符串
	return dateStr
}

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

	// 7. 格式化日期并返回抽取的元数据
	bookIDInt, _ := strconv.ParseInt(id, 10, 64)
	formattedDate := parseChineseDate(metadata.PublishDate)

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
			"publish_date": formattedDate,
		},
	})
}
