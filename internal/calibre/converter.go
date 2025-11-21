package calibre

import (
	"strconv"

	"github.com/jianyun8023/calibre-api/pkg/content"
)

// convertContentBooks 转换 content.Book 到 calibre.Book
func convertContentBooks(books []content.Book) ([]Book, error) {
	// Use the centralized EnrichBooks method from content package
	enrichedBooks := content.EnrichBooks(books)

	// Convert content.Book to calibre.Book
	result := make([]Book, len(enrichedBooks))
	for i, b := range enrichedBooks {
		result[i] = Book{
			AuthorSort:   b.AuthorSort,
			Authors:      b.Authors,
			Comments:     b.Comments,
			ID:           b.ID,
			Isbn:         b.Isbn,
			Languages:    b.Languages,
			LastModified: b.LastModified,
			PubDate:      b.PubDate,
			Publisher:    b.Publisher,
			SeriesIndex:  b.SeriesIndex,
			Size:         b.Size,
			Title:        b.Title,
			Tags:         b.Tags,
			Rating:       b.Rating,
			Identifiers:  b.Identifiers,
			Cover:        b.Cover,
			FilePath:     b.FilePath,
		}
	}
	return result, nil
}

// convertContentToBooks 转换 content.Content 到 calibre.Book
func convertContentToBooks(content map[string]content.Content) ([]Book, error) {
	var books []Book
	for id, c := range content {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		book := Book{
			// Map fields from Content to Book
			AuthorSort:   c.AuthorSort,
			Authors:      c.Authors,
			Comments:     c.Comments,
			ID:           i,
			Isbn:         c.Isbn,
			Languages:    c.Languages,
			LastModified: c.LastModified,
			PubDate:      c.PubDate,
			Publisher:    c.Publisher,
			SeriesIndex:  c.SeriesIndex,
			Size:         c.Size,
			Title:        c.Title,
			Tags:         c.Tags,
			Rating:       c.Rating,
			Identifiers:  c.Identifiers,
			Cover:        "/api/get/cover/" + strconv.FormatInt(i, 10) + ".jpg",
			FilePath:     "/api/download/book/" + strconv.FormatInt(i, 10) + ".epub",
		}
		books = append(books, book)
	}
	return books, nil
}
