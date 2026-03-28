package service

import (
	"context"
	"strconv"

	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	apperrors "github.com/jianyun8023/calibre-api/pkg/errors"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// bookServiceV2 使用 Repository 层的新实现
type bookServiceV2 struct {
	bookRepo    repository.BookRepository
	contentAPI  ContentAPI
	taskManager *tasks.Manager
	searcher    semantic.Searcher // 保留用于任务调度
}

// NewBookServiceWithRepository 创建使用 Repository 的书籍服务
func NewBookServiceWithRepository(
	bookRepo repository.BookRepository,
	contentAPI ContentAPI,
	taskManager *tasks.Manager,
	searcher semantic.Searcher,
) BookService {
	return &bookServiceV2{
		bookRepo:    bookRepo,
		contentAPI:  contentAPI,
		taskManager: taskManager,
		searcher:    searcher,
	}
}

// GetBookByID 根据 ID 获取书籍
func (s *bookServiceV2) GetBookByID(id string) (*Book, error) {
	ctx := context.Background()
	repoBook, err := s.bookRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 转换 repository.Book 到 service.Book
	book := convertRepoBookToServiceBook(repoBook)
	return &book, nil
}

// DeleteBook 删除书籍
func (s *bookServiceV2) DeleteBook(id string) error {
	// 删除 Calibre 中的书籍
	err := s.contentAPI.DeleteBooks([]string{id}, "")
	if err != nil {
		log.Infof("Failed to delete book %s: %v", id, err)
		return apperrors.WrapWithDetails(err, apperrors.CodeDeleteFailed, "failed to delete book", err.Error(), 500)
	}

	// 异步删除 Repository 中的数据
	if s.searcher != nil {
		bookID, _ := strconv.ParseInt(id, 10, 64)
		_, err := s.taskManager.StartTask(tasks.TaskTypeDeleteBook, tasks.TaskModeFull, func(taskID string) tasks.Task {
			return tasks.NewDeleteBookTask(taskID, bookID, s.searcher)
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
func (s *bookServiceV2) UpdateMetadata(id string, updates *BookUpdate) error {
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

	// 从 Calibre Content Server 重新查询最新书籍数据（而不是从 MeiliSearch 读取）
	idInt, _ := strconv.ParseInt(id, 10, 64)
	updatedBook, err := s.contentAPI.GetBookDetail(idInt)
	if err != nil {
		log.Warnf("Failed to fetch updated book %s from Calibre: %v, using merge fallback", id, err)
		// 降级方案：使用本地 merge
		updatedBook = mergeBookUpdates(oldBook, updates)
	}

	// 异步更新 Repository 中的数据
	if s.searcher != nil {
		semanticBook := convertBookToSemantic(updatedBook)
		_, err := s.taskManager.StartTask(tasks.TaskTypeUpdateMetadata, tasks.TaskModeFull, func(taskID string) tasks.Task {
			return tasks.NewUpdateMetadataTask(taskID, semanticBook, s.searcher)
		})
		if err != nil {
			log.Warnf("Failed to start update task for book %d: %v", updatedBook.ID, err)
		} else {
			log.Infof("Started update task for book %d", updatedBook.ID)
		}
	}

	return nil
}

// GetRecentBooks 获取最近更新的书籍
func (s *bookServiceV2) GetRecentBooks(limit, offset int) ([]Book, int64, error) {
	// 验证参数
	if limit <= 0 {
		return nil, 0, apperrors.NewValidationError("limit", "must be positive")
	}
	if offset < 0 {
		return nil, 0, apperrors.NewValidationError("offset", "must be non-negative")
	}

	ctx := context.Background()
	repoBooks, total, err := s.bookRepo.FindRecent(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 转换 repository.Book 到 service.Book
	books := convertRepoBooksToServiceBooks(repoBooks)
	return books, total, nil
}

// GetRandomBooks 获取随机书籍
func (s *bookServiceV2) GetRandomBooks(limit int) ([]Book, error) {
	// 验证参数
	if limit <= 0 {
		return nil, apperrors.NewValidationError("limit", "must be positive")
	}

	ctx := context.Background()
	repoBooks, err := s.bookRepo.FindRandom(ctx, limit)
	if err != nil {
		return nil, err
	}

	// 转换 repository.Book 到 service.Book
	books := convertRepoBooksToServiceBooks(repoBooks)
	return books, nil
}

// GetAllBooks 获取所有书籍（游标分页）
func (s *bookServiceV2) GetAllBooks(limit int, cursor string) ([]Book, int64, string, error) {
	// 验证参数
	if limit <= 0 {
		return nil, 0, "", apperrors.NewValidationError("limit", "must be positive")
	}

	ctx := context.Background()
	repoBooks, total, nextCursor, err := s.bookRepo.FindAllWithCursor(ctx, limit, cursor)
	if err != nil {
		return nil, 0, "", err
	}

	// 转换 repository.Book 到 service.Book
	books := convertRepoBooksToServiceBooks(repoBooks)
	return books, total, nextCursor, nil
}

// ListPublishers 获取所有出版社列表
func (s *bookServiceV2) ListPublishers() ([]string, error) {
	publishers, err := s.contentAPI.GetAllPublisher()
	if err != nil {
		return nil, apperrors.WrapWithDetails(err, apperrors.CodeInternalError, "failed to list publishers", err.Error(), 500)
	}
	return publishers, nil
}

// 转换函数

// convertRepoBookToServiceBook 将 repository.Book 转换为 service.Book
func convertRepoBookToServiceBook(book *repository.Book) Book {
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

// convertRepoBooksToServiceBooks 批量转换 repository.Book 到 service.Book
func convertRepoBooksToServiceBooks(books []repository.Book) []Book {
	serviceBooks := make([]Book, len(books))
	for i, book := range books {
		serviceBooks[i] = Book{
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
	return serviceBooks
}

// convertBookToSemantic is already defined in book_service.go
