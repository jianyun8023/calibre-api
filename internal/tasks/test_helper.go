package tasks

import (
	"fmt"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/pkg/content"
)

// RunTestExtraction is a helper for cmd/test_copyright/main.go to test the private logic.
// It instantiates a task and calls extractAndUpdateISBN on a specific book ID.
func RunTestExtraction(api *content.Api, cm *cache.Manager, bookID int64) error {
	task := NewCopyrightExtractTask("test-id", "test-mode", api, cm)

	metadata, err := task.extractAndUpdateISBN(bookID)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully extracted metadata for Book %d:\n", bookID)
	if metadata != nil {
		fmt.Printf("  Title: %s\n", metadata.BookTitle)
		fmt.Printf("  Author: %s\n", metadata.Author)
		fmt.Printf("  Translator: %s\n", metadata.Translator)
		fmt.Printf("  Publisher: %s\n", metadata.Publisher)
		fmt.Printf("  Date: %s\n", metadata.PublishDate)
		fmt.Printf("  ISBN: %s\n", metadata.ISBN)
	}
	return nil
}
