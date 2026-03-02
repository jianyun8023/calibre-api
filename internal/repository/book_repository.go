package repository

import (
	"context"
	"time"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	apperrors "github.com/jianyun8023/calibre-api/pkg/errors"
)

// BookRepository 书籍数据访问接口
// 抽象数据层操作，解耦具体实现（Qdrant, MySQL 等）
type BookRepository interface {
	// FindByID 根据 ID 查找书籍
	FindByID(ctx context.Context, id string) (*Book, error)

	// FindRecent 查找最近更新的书籍
	FindRecent(ctx context.Context, limit, offset int) ([]Book, int64, error)

	// FindRandom 查找随机书籍
	FindRandom(ctx context.Context, limit int) ([]Book, error)

	// FindAllWithCursor 使用游标分页查找所有书籍
	FindAllWithCursor(ctx context.Context, limit int, cursor string) ([]Book, int64, string, error)

	// SearchByKeyword 根据关键词搜索书籍
	SearchByKeyword(ctx context.Context, keyword, field string, limit, offset int) ([]Book, int64, error)

	// Update 更新书籍（元数据）
	Update(ctx context.Context, book *Book) error

	// Delete 删除书籍
	Delete(ctx context.Context, id int64) error
}

// Book 书籍模型（Repository 层）
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

// qdrantBookRepository Qdrant 实现的 BookRepository
type qdrantBookRepository struct {
	searcher semantic.Searcher
}

// NewQdrantBookRepository 创建 Qdrant 书籍仓储
func NewQdrantBookRepository(searcher semantic.Searcher) BookRepository {
	return &qdrantBookRepository{
		searcher: searcher,
	}
}

// FindByID 根据 ID 查找书籍
func (r *qdrantBookRepository) FindByID(ctx context.Context, id string) (*Book, error) {
	if r.searcher == nil {
		return nil, apperrors.NewSearchServiceNotAvailableError()
	}

	books, _, err := r.searcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil {
		return nil, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to find book by ID", err.Error(), 500)
	}
	if len(books) == 0 {
		return nil, apperrors.NewBookNotFoundError(id)
	}

	book := convertSemanticToBook(books[0])
	return &book, nil
}

// FindRecent 查找最近更新的书籍
func (r *qdrantBookRepository) FindRecent(ctx context.Context, limit, offset int) ([]Book, int64, error) {
	if r.searcher == nil {
		return nil, 0, apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, total, err := r.searcher.GetRecent(limit, offset)
	if err != nil {
		return nil, 0, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to get recent books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, total, nil
}

// FindRandom 查找随机书籍
func (r *qdrantBookRepository) FindRandom(ctx context.Context, limit int) ([]Book, error) {
	if r.searcher == nil {
		return nil, apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, err := r.searcher.GetRandom(limit)
	if err != nil {
		return nil, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to get random books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, nil
}

// FindAllWithCursor 使用游标分页查找所有书籍
func (r *qdrantBookRepository) FindAllWithCursor(ctx context.Context, limit int, cursor string) ([]Book, int64, string, error) {
	if r.searcher == nil {
		return nil, 0, "", apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, total, nextCursor, err := r.searcher.GetAllWithCursor(limit, cursor)
	if err != nil {
		return nil, 0, "", apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to get all books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, total, nextCursor, nil
}

// SearchByKeyword 根据关键词搜索书籍
func (r *qdrantBookRepository) SearchByKeyword(ctx context.Context, keyword, field string, limit, offset int) ([]Book, int64, error) {
	if r.searcher == nil {
		return nil, 0, apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBooks, total, err := r.searcher.SearchByKeyword(keyword, field, limit, offset)
	if err != nil {
		return nil, 0, apperrors.WrapWithDetails(err, apperrors.CodeSearchFailed, "failed to search books", err.Error(), 500)
	}

	books := convertSemanticToBooks(semanticBooks)
	return books, total, nil
}

// Update 更新书籍（元数据）
func (r *qdrantBookRepository) Update(ctx context.Context, book *Book) error {
	if r.searcher == nil {
		return apperrors.NewSearchServiceNotAvailableError()
	}

	semanticBook := convertBookToSemantic(book)

	// 使用 IndexBooks 进行 Upsert（单本书籍）
	err := r.searcher.IndexBooks(ctx, []semantic.Book{semanticBook})
	if err != nil {
		return apperrors.WrapWithDetails(err, apperrors.CodeInternalError, "failed to update book", err.Error(), 500)
	}

	return nil
}

// Delete 删除书籍
func (r *qdrantBookRepository) Delete(ctx context.Context, id int64) error {
	if r.searcher == nil {
		return apperrors.NewSearchServiceNotAvailableError()
	}

	// DeleteBook 方法内部创建 context，所以这里不传递
	err := r.searcher.DeleteBook(id)
	if err != nil {
		return apperrors.WrapWithDetails(err, apperrors.CodeDeleteFailed, "failed to delete book", err.Error(), 500)
	}

	return nil
}

// 转换函数

// convertSemanticToBook 将 semantic.Book 转换为 repository.Book
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

// convertSemanticToBooks 批量转换 semantic.Book 到 repository.Book
func convertSemanticToBooks(books []semantic.Book) []Book {
	repoBooks := make([]Book, len(books))
	for i, book := range books {
		repoBooks[i] = convertSemanticToBook(book)
	}
	return repoBooks
}

// convertBookToSemantic 将 repository.Book 转换为 semantic.Book
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
