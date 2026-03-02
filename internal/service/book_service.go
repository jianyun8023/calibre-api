package service

import (
	"strconv"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	apperrors "github.com/jianyun8023/calibre-api/pkg/errors"
	"github.com/jianyun8023/calibre-api/pkg/log"
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

// bookService 书籍业务逻辑实现
type bookService struct {
	semanticSearcher semantic.Searcher
	contentAPI       ContentAPI
	taskManager      *tasks.Manager
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

// NewBookService 创建书籍服务
func NewBookService(searcher semantic.Searcher, contentAPI ContentAPI, taskManager *tasks.Manager) BookService {
	return &bookService{
		semanticSearcher: searcher,
		contentAPI:       contentAPI,
		taskManager:      taskManager,
	}
}

// GetBookByID 根据 ID 获取书籍
func (s *bookService) GetBookByID(id string) (*Book, error) {
	if s.semanticSearcher == nil {
		return nil, apperrors.NewSearchServiceNotAvailableError()
	}

	// 搜索书籍
	books, _, err := s.semanticSearcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil {
		return nil, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to search book", err.Error(), 500)
	}
	if len(books) == 0 {
		return nil, apperrors.NewBookNotFoundError(id)
	}

	book := convertSemanticToBook(books[0])
	return &book, nil
}

// DeleteBook 删除书籍
func (s *bookService) DeleteBook(id string) error {
	// 删除 Calibre 中的书籍
	err := s.contentAPI.DeleteBooks([]string{id}, "")
	if err != nil {
		log.Infof("Failed to delete book %s: %v", id, err)
		return apperrors.WrapWithDetails(err, apperrors.CodeDeleteFailed, "failed to delete book", err.Error(), 500)
	}

	// 异步删除 Qdrant 中的向量
	if s.semanticSearcher != nil {
		bookID, _ := strconv.ParseInt(id, 10, 64)
		_, err := s.taskManager.StartTask(tasks.TaskTypeDeleteBook, tasks.TaskModeFull, func(taskID string) tasks.Task {
			return tasks.NewDeleteBookTask(taskID, bookID, s.semanticSearcher)
		})
		if err != nil {
			log.Warnf("Failed to start delete task for book %d: %v", bookID, err)
		} else {
			log.Infof("Started delete task for book %d", bookID)
		}
	}

	return nil
}

// UpdateMetadata 更新书籍元数据
func (s *bookService) UpdateMetadata(id string, updates *BookUpdate) error {
	// 获取旧书籍信息
	oldBook, err := s.GetBookByID(id)
	if err != nil {
		return err
	}

	// 构建更新参数
	metadata := buildUpdateParams(updates, oldBook)

	// 更新 Calibre 中的元数据
	_, err = s.contentAPI.UpdateMetaData(id, metadata, "")
	if err != nil {
		return apperrors.WrapWithDetails(err, apperrors.CodeInternalError, "failed to update metadata", err.Error(), 500)
	}

	// 合并更新到旧书籍
	mergedBook := mergeBookUpdates(oldBook, updates)

	// 异步更新 Qdrant 中的向量
	if s.semanticSearcher != nil {
		semanticBook := convertBookToSemantic(mergedBook)
		_, err := s.taskManager.StartTask(tasks.TaskTypeUpdateMetadata, tasks.TaskModeFull, func(taskID string) tasks.Task {
			return tasks.NewUpdateMetadataTask(taskID, semanticBook, s.semanticSearcher)
		})
		if err != nil {
			log.Warnf("Failed to start update task for book %d: %v", oldBook.ID, err)
		} else {
			log.Infof("Started update task for book %d", oldBook.ID)
		}
	}

	return nil
}

// GetRecentBooks 获取最近更新的书籍
func (s *bookService) GetRecentBooks(limit, offset int) ([]Book, int64, error) {
	// 验证参数
	if limit <= 0 {
		return nil, 0, apperrors.NewValidationError("limit", "must be positive")
	}
	if offset < 0 {
		return nil, 0, apperrors.NewValidationError("offset", "must be non-negative")
	}

	if s.semanticSearcher == nil {
		return nil, 0, apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, total, err := s.semanticSearcher.GetRecent(limit, offset)
	if err != nil {
		return nil, 0, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to get recent books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, total, nil
}

// GetRandomBooks 获取随机书籍
func (s *bookService) GetRandomBooks(limit int) ([]Book, error) {
	// 验证参数
	if limit <= 0 {
		return nil, apperrors.NewValidationError("limit", "must be positive")
	}

	if s.semanticSearcher == nil {
		return nil, apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, err := s.semanticSearcher.GetRandom(limit)
	if err != nil {
		return nil, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to get random books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, nil
}

// GetAllBooks 获取所有书籍（游标分页）
func (s *bookService) GetAllBooks(limit int, cursor string) ([]Book, int64, string, error) {
	// 验证参数
	if limit <= 0 {
		return nil, 0, "", apperrors.NewValidationError("limit", "must be positive")
	}

	if s.semanticSearcher == nil {
		return nil, 0, "", apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, total, nextCursor, err := s.semanticSearcher.GetAllWithCursor(limit, cursor)
	if err != nil {
		log.Warnf("GetAllWithCursor failed: %v", err)
		return nil, 0, "", apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to get all books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, total, nextCursor, nil
}

// ListPublishers 获取所有出版社列表
func (s *bookService) ListPublishers() ([]string, error) {
	publishers, err := s.contentAPI.GetAllPublisher()
	if err != nil {
		return nil, apperrors.WrapWithDetails(err, apperrors.CodeInternalError, "failed to list publishers", err.Error(), 500)
	}
	return publishers, nil
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
