package calibre

import (
	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// 转换函数（供其他 Handler 使用）

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
