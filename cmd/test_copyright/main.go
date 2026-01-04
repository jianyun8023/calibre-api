package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/content"
)

func main() {
	// 1. Configuration
	baseURL := "http://192.168.2.236:8283"
	tempDir := "./temp_test_downloads"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir) // Cleanup

	// 2. Initialize Content API
	api, err := content.NewClient(baseURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	fmt.Printf("Connected to %s\n", baseURL)

	// Debug: Check connectivity
	// Manually check what the server is
	// Try to get /
	// We can use the exposed client if valid, but let's just use http.Get for simplicity or assume api.Client is usable.
	// Since api.Client is embedded, we can't easily access the underlying resty client if not exported, but it is.
	// Let's just try to call a method that we know simple GET
	// Or we can assume the user meant to use a different port or path?
	// But let's assume the user is right and maybe we just need to fix the path in `api.go`?
	// No, I shouldn't change `api.go` unless I know for sure.

	// Let's try to proceed, but if it fails, we catch it better.

	// 3. Initialize Cache Manager
	cacheManager, err := cache.NewManager(cache.Config{
		Dir:              tempDir,
		MaxSizeGB:        1.0,
		CleanupThreshold: 0.8,
	}, &api)
	if err != nil {
		log.Fatalf("Failed to create cache manager: %v", err)
	}
	// We need to access private methods or use public interface.
	// Since extractAndUpdateISBN is private and belongs to CopyrightExtractTask,
	// we should probably instantiate the task and run it, or expose a public helper if needed.
	// However, we want to see the OUTPUT, not just have it update the server.
	// Modification: I will temporarily expose a public method or just run the task and check logs.
	// Better yet, I will use NewCopyrightExtractTask and modify it slightly to log more verbosely,
	// OR just trust the logs it prints.

	// 5. Fetch a few books to test
	// Since listing is failing (404), we will try to blindly fetch IDs 1 to 10.
	fmt.Println("Attempting to fetch specific books (skipping list)...")

	// Create a simpler loop to test
	var booksFound []int64
	for id := int64(260000); id < 270000; id++ { // Try a range that likely has books (user logs showed 264728)
		// Try to peek if book exists or just try to extract
		// We can use api.GetBook directly to see if it downloads
		// Or assume task handles it.
		// Let's rely on GetBookMetaDatas to check existence first
		_, err := api.GetBookMetaDatas([]int64{id}, "library")
		if err == nil {
			fmt.Printf("Found book ID: %d\n", id)
			booksFound = append(booksFound, id)
			if len(booksFound) >= 5 {
				break
			}
		} else {
			// Checking if error is 404 or something else could be useful,
			// but api.go logs warn and returns error on non-success usually.
		}

		if id%100 == 0 {
			fmt.Printf("Checked up to ID %d...\r", id)
		}
	}

	if len(booksFound) == 0 {
		// Fallback to recent known IDs from logs if available, or just try 1-50
		for id := int64(1); id < 20; id++ {
			_, err := api.GetBookMetaDatas([]int64{id}, "library")
			if err == nil {
				fmt.Printf("Found book ID: %d\n", id)
				booksFound = append(booksFound, id)
				if len(booksFound) >= 1 {
					break
				}
			}
		}
	}

	if len(booksFound) == 0 {
		log.Fatal("No books found to test.")
	}

	// Now run extraction on these books
	// We can't use task.Run() because it calls GetAllBooksIds.
	// We have to invoke the worker logic manually or through reflection (hard).
	// Or we modify CopyrightExtractTask to allow passing IDs.
	// Modification: I will use a private method via export trick or just copy the logic here.
	// Copying logic is safest for a test script.

	fmt.Printf("Testing extraction on %d books...\n", len(booksFound))
	for _, id := range booksFound {
		fmt.Printf("--- Processing Book %d ---\n", id)
		err := tasks.RunTestExtraction(&api, cacheManager, id)
		if err != nil {
			log.Printf("Extraction failed: %v", err)
		}
	}
}
