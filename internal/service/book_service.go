package service

import (
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
	Title     string    `json:"title"`
	Authors   []string  `json:"authors"`
	Publisher string    `json:"publisher"`
	PubDate   time.Time `json:"pubdate"`
	Isbn      string    `json:"isbn"`
	Tags      []string  `json:"tags"`
	Rating    float64   `json:"rating"`
	Comments  string    `json:"comments"`
}

// ContentAPI 内容服务接口（用于与 Calibre Content Server 交互）
type ContentAPI interface {
	// DeleteBooks 删除书籍
	DeleteBooks(ids []string, library string) error

	// UpdateMetaData 更新元数据
	UpdateMetaData(id string, metadata map[string]interface{}, library string) (bool, error)

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

	if updates.Comments != "" {
		metadata["comments"] = updates.Comments
	}
	if updates.Isbn != "" {
		identifiers := oldBook.Identifiers
		if identifiers == nil {
			identifiers = make(map[string]string)
		}
		identifiers["isbn"] = updates.Isbn
		metadata["identifiers"] = identifiers
	}
	if updates.Title != "" {
		metadata["title"] = updates.Title
	}
	if updates.Publisher != "" {
		metadata["publisher"] = updates.Publisher
	}
	if !updates.PubDate.IsZero() {
		metadata["pubdate"] = updates.PubDate.Format("2006-01-02T15:04:05+00:00")
	}
	if updates.Authors != nil {
		metadata["authors"] = updates.Authors
	}
	if updates.Tags != nil {
		metadata["tags"] = updates.Tags
	}
	if updates.Rating > 0 {
		metadata["rating"] = updates.Rating
	}

	return metadata
}

// mergeBookUpdates 合并书籍更新
func mergeBookUpdates(oldBook *Book, updates *BookUpdate) *Book {
	merged := *oldBook // 复制旧书籍

	if updates.Title != "" {
		merged.Title = updates.Title
	}
	if updates.Authors != nil {
		merged.Authors = updates.Authors
	}
	if updates.Publisher != "" {
		merged.Publisher = updates.Publisher
	}
	if updates.Comments != "" {
		merged.Comments = updates.Comments
	}
	if updates.Tags != nil {
		merged.Tags = updates.Tags
	}
	if updates.Rating > 0 {
		merged.Rating = updates.Rating
	}
	if updates.Isbn != "" {
		merged.Isbn = updates.Isbn
		if merged.Identifiers == nil {
			merged.Identifiers = make(map[string]string)
		}
		merged.Identifiers["isbn"] = updates.Isbn
	}
	if !updates.PubDate.IsZero() {
		merged.PubDate = updates.PubDate
	}

	return &merged
}
