package calibre

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// getBook 获取书籍信息接口
func (c *Api) getBook(r *gin.Context) {
	id := r.Param("id")

	// Use Qdrant to search by ID
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "Search service not available",
			"code":    http.StatusServiceUnavailable,
		})
		return
	}

	// Search by ID using keyword search
	books, _, err := searcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil || len(books) == 0 {
		r.JSON(http.StatusOK, gin.H{
			"message": "book not found",
			"code":    http.StatusNotFound,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	book := convertSemanticToBook(books[0])

	r.JSON(http.StatusOK, gin.H{
		"data":    &book,
		"message": "ok",
		"code":    200,
	})
}

// deleteBook 删除书籍接口
func (c *Api) deleteBook(r *gin.Context) {
	id := r.Param("id")

	err := c.contentApi.DeleteBooks([]string{id}, "")
	if err != nil {
		log.Infof("Failed to delete book %s: %v", id, err)
		r.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("Failed to delete book: %v", err),
			"code":    http.StatusInternalServerError,
		})
		return
	}

	// Dispatch delete task
	if searcher, ok := c.semanticSearcher.(*qdrant.Searcher); ok && searcher != nil {
		manager := tasks.GetManager()
		bookID, _ := strconv.ParseInt(id, 10, 64)
		_, err := manager.StartTask(tasks.TaskTypeDeleteBook, tasks.TaskModeFull, func(taskID string) tasks.Task {
			return tasks.NewDeleteBookTask(taskID, bookID, searcher)
		})
		if err != nil {
			log.Warnf("Failed to start delete task for book %d: %v", bookID, err)
		} else {
			log.Infof("Started delete task for book %d", bookID)
		}
	}

	r.JSON(http.StatusOK, gin.H{
		"data":    true,
		"message": "ok",
		"code":    200,
	})
}

// updateMetadata 更新书籍元数据
func (c *Api) updateMetadata(r *gin.Context) {
	id := r.Param("id")
	book := &Book{}
	err := r.Bind(book)
	if err != nil {
		r.JSON(http.StatusOK, gin.H{
			"code":    400,
			"data":    false,
			"message": "请求参数错误" + err.Error(),
		})
		return
	}

	oldBook, err := c.getBookByID(id)
	if err != nil {
		r.JSON(http.StatusOK, gin.H{
			"code":    500,
			"data":    false,
			"message": "元数据更新失败",
		})
		return
	}
	_, err = c.contentApi.UpdateMetaData(id, parseParams(book, oldBook), "")
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":    500,
			"data":    false,
			"message": "元数据更新失败",
		})
		return
	}

	// Dispatch update task
	if searcher, ok := c.semanticSearcher.(*qdrant.Searcher); ok && searcher != nil {
		// Merge changes into oldBook
		if book.Title != "" {
			oldBook.Title = book.Title
		}
		if book.Authors != nil {
			oldBook.Authors = book.Authors
		}
		if book.Publisher != "" {
			oldBook.Publisher = book.Publisher
		}
		if book.Comments != "" {
			oldBook.Comments = book.Comments
		}
		if book.Tags != nil {
			oldBook.Tags = book.Tags
		}
		if book.Rating > 0 {
			oldBook.Rating = book.Rating
		}
		if book.Isbn != "" {
			oldBook.Isbn = book.Isbn
			if oldBook.Identifiers == nil {
				oldBook.Identifiers = make(map[string]string)
			}
			oldBook.Identifiers["isbn"] = book.Isbn
		}
		if !book.PubDate.IsZero() {
			oldBook.PubDate = book.PubDate
		}

		semanticBook := convertBookToSemantic(oldBook)

		manager := tasks.GetManager()
		_, err := manager.StartTask(tasks.TaskTypeUpdateMetadata, tasks.TaskModeFull, func(taskID string) tasks.Task {
			return tasks.NewUpdateMetadataTask(taskID, semanticBook, searcher)
		})
		if err != nil {
			log.Warnf("Failed to start update task for book %d: %v", oldBook.ID, err)
		} else {
			log.Infof("Started update task for book %d", oldBook.ID)
		}
	}

	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    true,
		"message": "元数据更新成功",
	})
}

// recently 获取最近更新的书籍
func (c *Api) recently(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "10"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	offset, err := strconv.Atoi(r.DefaultQuery("offset", "0"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
		return
	}

	// Use Qdrant GetRecent
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Search service not available",
			"code":  503,
		})
		return
	}

	books, total, err := searcher.GetRecent(limit, offset)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := convertSemanticToBooks(books)

	r.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"records": calibreBooks,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
		"code": 200,
	})
}

// random 获取随机书籍
func (c *Api) random(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "10"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	// Use Qdrant GetRandom
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Search service not available",
			"code":  503,
		})
		return
	}

	books, err := searcher.GetRandom(limit)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := convertSemanticToBooks(books)

	r.JSON(http.StatusOK, gin.H{
		"data": calibreBooks,
		"code": 200,
	})
}

// getAllBooks 获取所有书籍（支持游标分页）
func (c *Api) getAllBooks(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "12"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	// Use cursor instead of offset
	// cursor format: "last_modified:2024-01-01T00:00:00Z,id:123"
	cursor := r.DefaultQuery("cursor", "")

	// Use Qdrant searcher
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Search service not available",
			"code":  503,
		})
		return
	}

	// Get books with cursor-based pagination
	books, total, nextCursor, err := searcher.GetAllWithCursor(limit, cursor)
	if err != nil {
		log.Warnf("GetAllWithCursor failed: %v", err)
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := convertSemanticToBooks(books)

	r.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"records":     calibreBooks,
			"total":       total,
			"limit":       limit,
			"next_cursor": nextCursor,
		},
		"code": 200,
	})
}

// listPublisher 获取所有出版社列表
func (c *Api) listPublisher(context *gin.Context) {
	publishers, err := c.contentApi.GetAllPublisher()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": publishers,
	})
}

// getBookByID retrieves a book by ID from Qdrant
func (c *Api) getBookByID(id string) (*Book, error) {
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		return nil, ErrSearchServiceNotAvailable
	}

	books, _, err := searcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return nil, ErrBookNotFound
	}

	book := convertSemanticToBook(books[0])
	return &book, nil
}

// convertSemanticToBook 将 semantic.Book 转换为 calibre.Book
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
	}
}

// convertSemanticToBooks 批量转换 semantic.Book 到 calibre.Book
func convertSemanticToBooks(books []semantic.Book) []Book {
	calibreBooks := make([]Book, len(books))
	for i, book := range books {
		calibreBooks[i] = convertSemanticToBook(book)
	}
	return calibreBooks
}

// convertBookToSemantic converts calibre.Book to semantic.Book
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

// parseParams 解析更新参数
func parseParams(book *Book, oldBook *Book) map[string]interface{} {
	metadata := map[string]interface{}{}
	if book.Comments != "" {
		metadata["comments"] = book.Comments
	}
	if book.Isbn != "" {
		identifiers := oldBook.Identifiers
		if identifiers == nil {
			identifiers = make(map[string]string)
		}
		identifiers["isbn"] = book.Isbn
		metadata["identifiers"] = identifiers
	}
	if book.Title != "" {
		metadata["title"] = book.Title
	}
	if book.Publisher != "" {
		metadata["publisher"] = book.Publisher
	}

	if !book.PubDate.IsZero() {
		metadata["pubdate"] = book.PubDate.Format("2006-01-02T15:04:05+00:00")
	}
	if book.Authors != nil {
		metadata["authors"] = book.Authors
	}
	if book.Tags != nil {
		metadata["tags"] = book.Tags
	}
	if book.Rating > 0 {
		metadata["rating"] = book.Rating
	}
	return metadata
}
