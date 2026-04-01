package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// BookService 书籍业务逻辑接口
type BookService interface {
	// GetBookByID 根据 ID 获取书籍
	GetBookByID(id string) (*Book, error)

	// DeleteBook 删除书籍
	DeleteBook(id string) error

	// UpdateMetadata 更新书籍元数据
	UpdateMetadata(id string, updates *BookUpdate) error

	// GetRecentBooks 获取最近更新的书籍
	GetRecentBooks(limit, offset int) ([]Book, int64, error)

	// GetRandomBooks 获取随机书籍
	GetRandomBooks(limit int) ([]Book, error)

	// GetAllBooks 获取所有书籍（游标分页）
	GetAllBooks(limit int, cursor string) ([]Book, int64, string, error)

	// ListPublishers 获取所有出版社列表
	ListPublishers() ([]string, error)
}

// Book 书籍模型（简化版，不包含 API 层特有字段）
type Book struct {
	ID           int64             `json:"id"`
	Title        string            `json:"title"`
	Authors      []string          `json:"authors"`
	Publisher    string            `json:"publisher"`
	PubDate      time.Time         `json:"pubdate"`
	Isbn         string            `json:"isbn"`
	Tags         []string          `json:"tags"`
	Rating       float64           `json:"rating"`
	SeriesIndex  float64           `json:"series_index"`
	Comments     string            `json:"comments"`
	Languages    []string          `json:"languages"`
	LastModified time.Time         `json:"last_modified"`
	Cover        string            `json:"cover"`
	FilePath     string            `json:"file_path"`
	Identifiers  map[string]string `json:"identifiers"`
	Size         int64             `json:"size"`
}

// BookUpdate 书籍更新请求
type BookUpdate struct {
	Title     *string    `json:"title,omitempty"`
	Authors   *[]string  `json:"authors,omitempty"` // 使用指针类型：nil=不更新, 空数组会被拒绝
	Publisher *string    `json:"publisher,omitempty"`
	PubDate   *time.Time `json:"pubdate,omitempty"` // 使用指针类型，未提供时为 nil
	Isbn      *string    `json:"isbn,omitempty"`
	Tags      *[]string  `json:"tags,omitempty"` // 使用指针类型：nil=不更新, &[]=清空, &[...]=更新
	Rating    float64    `json:"rating,omitempty"`
	Comments  *string    `json:"comments,omitempty"`
}

// UnmarshalJSON 自定义 JSON 解析，支持灵活的日期格式和 rating 类型
func (bu *BookUpdate) UnmarshalJSON(data []byte) error {
	type Alias BookUpdate // 避免递归调用
	
	// 先尝试使用辅助结构解析，支持字符串类型的 rating
	aux := &struct {
		PubDate interface{} `json:"pubdate,omitempty"` // 支持字符串格式的日期
		Rating  interface{} `json:"rating,omitempty"`  // 支持字符串或数字
		*Alias
	}{
		Alias: (*Alias)(bu),
	}
	
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	
	// 处理 PubDate 字段
	if aux.PubDate != nil {
		switch v := aux.PubDate.(type) {
		case string:
			if v != "" {
				parsedDate, err := parseFlexibleDate(v)
				if err != nil {
					return fmt.Errorf("invalid pubdate format '%s': %v", v, err)
				}
				bu.PubDate = &parsedDate
			}
		}
	}
	
	// 处理 Rating 字段
	if aux.Rating != nil {
		switch v := aux.Rating.(type) {
		case string:
			if v != "" {
				rating, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return fmt.Errorf("invalid rating format '%s': %v", v, err)
				}
				bu.Rating = rating
			}
		case float64:
			bu.Rating = v
		}
	}
	
	return nil
}

// parseFlexibleDate 解析多种日期格式
func parseFlexibleDate(dateStr string) (time.Time, error) {
	// 支持的日期格式列表
	formats := []string{
		"2006-01-02",           // 标准格式: 2023-09-01
		"2006-1-2",             // 短格式: 2023-9-1
		"2006-01",              // 年月: 2023-09
		"2006-1",               // 年月短格式: 2023-9
		"2006",                 // 仅年份: 2023
		time.RFC3339,           // RFC3339: 2023-09-01T00:00:00Z
		"2006-01-02T15:04:05",  // 无时区
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// 如果只有年月或年份，设置为该月的第一天
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// ContentAPI 内容服务接口（用于与 Calibre Content Server 交互）
type ContentAPI interface {
	// DeleteBooks 删除书籍
	DeleteBooks(ids []string, library string) error

	// UpdateMetaData 更新元数据
	UpdateMetaData(id string, metadata map[string]interface{}, library string) (bool, error)

	// GetBookDetail 从 Calibre 获取书籍详情（最新数据）
	GetBookDetail(id int64) (*Book, error)

	// GetAllPublisher 获取所有出版社
	GetAllPublisher() ([]string, error)
}

// 辅助函数

// convertSemanticToBook 将 semantic.Book 转换为 service.Book
func convertSemanticToBook(book semantic.Book) Book {
	return Book{
		ID:           book.ID,
		Title:        book.Title,
		Authors:      book.Authors,
		Publisher:    book.Publisher,
		PubDate:      book.PubDate,
		Isbn:         book.Isbn,
		Tags:         book.Tags,
		Rating:       book.Rating,
		SeriesIndex:  book.SeriesIndex,
		Comments:     book.Comments,
		Languages:    book.Languages,
		LastModified: book.LastModified,
		Cover:        book.Cover,
		FilePath:     book.FilePath,
		Identifiers:  book.Identifiers,
		Size:         book.Size,
	}
}

// convertSemanticToBooks 批量转换 semantic.Book 到 service.Book
func convertSemanticToBooks(books []semantic.Book) []Book {
	serviceBooks := make([]Book, len(books))
	for i, book := range books {
		serviceBooks[i] = convertSemanticToBook(book)
	}
	return serviceBooks
}

// convertBookToSemantic 将 service.Book 转换为 semantic.Book
func convertBookToSemantic(book *Book) semantic.Book {
	return semantic.Book{
		ID:           book.ID,
		Title:        book.Title,
		Authors:      book.Authors,
		Publisher:    book.Publisher,
		PubDate:      book.PubDate,
		Isbn:         book.Isbn,
		Tags:         book.Tags,
		Rating:       book.Rating,
		SeriesIndex:  book.SeriesIndex,
		Comments:     book.Comments,
		Languages:    book.Languages,
		LastModified: book.LastModified,
		Cover:        book.Cover,
		FilePath:     book.FilePath,
		Identifiers:  book.Identifiers,
		Size:         book.Size,
	}
}

// buildUpdateParams 构建更新参数
func buildUpdateParams(updates *BookUpdate, oldBook *Book) map[string]interface{} {
	metadata := map[string]interface{}{}

	// Comments: 只允许补全信息（非空值），不允许清空
	if updates.Comments != nil && *updates.Comments != "" {
		metadata["comments"] = *updates.Comments
	}

	// Isbn: 只允许补全信息（非空值）
	if updates.Isbn != nil && *updates.Isbn != "" {
		identifiers := oldBook.Identifiers
		if identifiers == nil {
			identifiers = make(map[string]string)
		}
		identifiers["isbn"] = *updates.Isbn
		metadata["identifiers"] = identifiers
	}

	// Title: 只允许补全信息（非空值），不允许清空
	if updates.Title != nil && *updates.Title != "" {
		metadata["title"] = *updates.Title
	}

	// Publisher: 只允许补全信息（非空值），不允许清空
	if updates.Publisher != nil && *updates.Publisher != "" {
		metadata["publisher"] = *updates.Publisher
	}

	// PubDate: 保持原有逻辑
	if updates.PubDate != nil && !updates.PubDate.IsZero() {
		metadata["pubdate"] = updates.PubDate.Format("2006-01-02T15:04:05+00:00")
	}

	// Authors: 只允许非空数组，不允许清空
	if updates.Authors != nil && len(*updates.Authors) > 0 {
		metadata["authors"] = *updates.Authors
	}

	// Tags: 允许清空（nil = 不更新，&[] = 清空）
	if updates.Tags != nil {
		metadata["tags"] = *updates.Tags
	}

	// Rating: 保持原有逻辑（0 表示未设置或移除评分）
	if updates.Rating > 0 {
		metadata["rating"] = updates.Rating
	}

	return metadata
}

// mergeBookUpdates 合并书籍更新
func mergeBookUpdates(oldBook *Book, updates *BookUpdate) *Book {
	merged := *oldBook // 复制旧书籍

	// 字符串字段：只允许非空值更新
	if updates.Title != nil && *updates.Title != "" {
		merged.Title = *updates.Title
	}
	if updates.Publisher != nil && *updates.Publisher != "" {
		merged.Publisher = *updates.Publisher
	}
	if updates.Comments != nil && *updates.Comments != "" {
		merged.Comments = *updates.Comments
	}
	if updates.Isbn != nil && *updates.Isbn != "" {
		merged.Isbn = *updates.Isbn
		if merged.Identifiers == nil {
			merged.Identifiers = make(map[string]string)
		}
		merged.Identifiers["isbn"] = *updates.Isbn
	}

	// Authors: 只允许非空数组，不允许清空
	if updates.Authors != nil && len(*updates.Authors) > 0 {
		merged.Authors = *updates.Authors
	}
	
	// Tags: 允许清空（nil = 不更新，[] = 清空）
	if updates.Tags != nil {
		merged.Tags = *updates.Tags
	}

	// 其他字段
	if updates.Rating > 0 {
		merged.Rating = updates.Rating
	}
	if updates.PubDate != nil && !updates.PubDate.IsZero() {
		merged.PubDate = *updates.PubDate
	}

	return &merged
}
