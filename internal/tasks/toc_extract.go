package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/kapmahc/epub"
)

type TocExtractTask struct {
	id            string
	mode          TaskMode
	contentApi    *content.Api
	searcher      *qdrant.Searcher
	cacheManager  *cache.Manager
	status        TaskStatus
	mu            sync.RWMutex
	cancel        context.CancelFunc
	progressFile  string
	progress      *TocExtractProgress
	numWorkers    int             // Number of parallel workers
	batchSize     int             // Batch size for Qdrant updates
	updateBatch   []tocUpdateItem // Batch of TOC updates for Qdrant
	updateBatchMu sync.Mutex      // Mutex for update batch
	qdrantBatch   int             // Qdrant batch update size
}

type tocUpdateItem struct {
	bookID int64
	toc    map[string]interface{}
}

type TocExtractProgress struct {
	ProcessedIDs []int64   `json:"processed_ids"`
	LastUpdated  time.Time `json:"last_updated"`
	TotalBooks   int       `json:"total_books"`
	Processed    int       `json:"processed"`
}

func NewTocExtractTask(
	id string,
	mode TaskMode,
	contentApi *content.Api,
	searcher *qdrant.Searcher,
	cacheManager *cache.Manager,
) *TocExtractTask {
	progressFile := fmt.Sprintf(".cache/toc_progress_%s.json", id)
	return &TocExtractTask{
		id:           id,
		mode:         mode,
		contentApi:   contentApi,
		searcher:     searcher,
		cacheManager: cacheManager,
		progressFile: progressFile,
		numWorkers:   10, // Increased parallel workers
		batchSize:    50, // Batch size for progress saves
		qdrantBatch:  20, // Batch size for Qdrant updates
		updateBatch:  make([]tocUpdateItem, 0, 20),
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeTocExtract,
			Mode:      mode,
			State:     "idle",
			StartTime: time.Now(),
			Message:   "Initializing TOC extraction...",
		},
	}
}

func (t *TocExtractTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *TocExtractTask) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
	if t.status.State == "running" {
		t.status.State = "stopped"
		t.status.Message = "Stopped by user"
		// Save progress before stopping
		t.saveProgress()
	}
}

func (t *TocExtractTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Loading progress..."
	t.mu.Unlock()

	// Load progress if exists
	if err := t.loadProgress(); err != nil {
		log.Printf("Failed to load progress: %v. Starting fresh.", err)
		t.progress = &TocExtractProgress{
			ProcessedIDs: []int64{},
			LastUpdated:  time.Now(),
		}
	}

	// Get all book IDs
	query := ""

	// Check for existing TOCs if in incremental mode
	if t.mode == TaskModeIncremental && t.searcher != nil {
		t.mu.Lock()
		t.status.Message = "Incremental mode: checking existing TOC data..."
		t.mu.Unlock()

		// We'll check existence during the filtering phase to avoid fetching all TOCs at once
		// or we can iterate and check. For now, let's do it in the filtering loop.
	}

	bookIDs, err := t.contentApi.GetAllBooksIds(query)
	if err != nil {
		return fmt.Errorf("failed to get book IDs: %w", err)
	}

	t.mu.Lock()
	t.progress.TotalBooks = len(bookIDs)
	t.status.Message = fmt.Sprintf("Found %d books to process", len(bookIDs))
	t.mu.Unlock()

	if len(bookIDs) == 0 {
		t.mu.Lock()
		t.status.State = "completed"
		t.status.Progress = 100
		t.status.Message = "No books to process"
		t.mu.Unlock()
		return nil
	}

	// Filter out already processed books
	processedMap := make(map[int64]bool)
	for _, id := range t.progress.ProcessedIDs {
		processedMap[id] = true
	}

	var booksToProcess []int64

	// Pre-check for incremental mode to avoid processing existing items
	if t.mode == TaskModeIncremental && t.searcher != nil {
		log.Printf("Checking %d books for existing TOC data...", len(bookIDs))
		checkCount := 0
		for _, id := range bookIDs {
			if processedMap[id] {
				continue
			}

			// Check if TOC exists in Qdrant
			// Note: This might be slow for many books, but it's safe.
			// Ideally we should have a bulk check or a way to list IDs with TOC.
			toc, err := t.searcher.GetBookToc(id)
			if err == nil && toc != nil {
				// TOC exists, skip
				processedMap[id] = true
				continue
			}

			booksToProcess = append(booksToProcess, id)
			checkCount++
			if checkCount%100 == 0 {
				t.mu.Lock()
				t.status.Message = fmt.Sprintf("Checking existing data: %d/%d checked...", checkCount, len(bookIDs))
				t.mu.Unlock()
			}
		}
	} else {
		for _, id := range bookIDs {
			if !processedMap[id] {
				booksToProcess = append(booksToProcess, id)
			}
		}
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Processing %d books (skipped %d already processed)",
		len(booksToProcess), len(bookIDs)-len(booksToProcess))
	t.mu.Unlock()

	// Process books with parallel workers
	if err := t.processBooksConcurrently(ctx, booksToProcess); err != nil {
		return err
	}

	// Final save
	t.saveProgress()

	t.mu.Lock()
	t.status.Progress = 100
	t.status.State = "completed"
	t.status.Message = fmt.Sprintf("Completed: processed %d books", len(t.progress.ProcessedIDs))
	t.mu.Unlock()

	// Clean up progress file on successful completion
	os.Remove(t.progressFile)

	return nil
}

// processBooksConcurrently processes books using a worker pool for parallel execution
func (t *TocExtractTask) processBooksConcurrently(ctx context.Context, bookIDs []int64) error {
	// Create channels with buffer to prevent blocking
	jobsChan := make(chan int64, t.numWorkers*2)
	resultsChan := make(chan processingResult, t.numWorkers*2)

	// Start worker pool
	var wg sync.WaitGroup
	for w := 0; w < t.numWorkers; w++ {
		wg.Add(1)
		go t.worker(ctx, w, jobsChan, resultsChan, &wg)
	}

	// Send jobs in a separate goroutine
	go func() {
		defer close(jobsChan)
		for _, bookID := range bookIDs {
			select {
			case <-ctx.Done():
				return
			case jobsChan <- bookID:
			}
		}
	}()

	// Collect results in a separate goroutine
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Process results and track progress
	total := t.progress.TotalBooks
	successCount := 0
	failCount := 0
	saveCounter := 0

	log.Printf("Starting parallel TOC extraction with %d workers for %d books", t.numWorkers, len(bookIDs))

	for result := range resultsChan {
		if result.err != nil {
			log.Printf("Failed to extract TOC for book %d: %v", result.bookID, result.err)
			failCount++
		} else {
			// Mark as processed
			t.mu.Lock()
			t.progress.ProcessedIDs = append(t.progress.ProcessedIDs, result.bookID)
			t.progress.Processed = len(t.progress.ProcessedIDs)
			t.progress.LastUpdated = time.Now()
			currentProcessed := len(t.progress.ProcessedIDs)
			t.mu.Unlock()

			successCount++
			saveCounter++

			// Update progress display
			t.mu.Lock()
			t.status.Progress = float64(currentProcessed) / float64(total) * 100
			t.status.Message = fmt.Sprintf("Processing: %d/%d completed (✓ %d, ✗ %d) - %d workers",
				currentProcessed, total, successCount, failCount, t.numWorkers)
			t.mu.Unlock()

			// Save progress periodically based on batch size
			if saveCounter >= t.batchSize {
				if err := t.saveProgress(); err != nil {
					log.Printf("Failed to save progress: %v", err)
				} else {
					log.Printf("Progress saved: %d/%d books processed", currentProcessed, total)
				}
				saveCounter = 0
			}
		}
	}

	// Check for context cancellation
	if ctx.Err() != nil {
		log.Printf("Task cancelled: saving progress before exit")
		t.saveProgress()
		return ctx.Err()
	}

	// Flush any remaining updates
	if err := t.flushUpdateBatch(); err != nil {
		log.Printf("Failed to flush final batch: %v", err)
	}

	log.Printf("TOC extraction completed: %d successful, %d failed out of %d total",
		successCount, failCount, len(bookIDs))

	return nil
}

// processingResult holds the result of processing a single book
type processingResult struct {
	bookID int64
	err    error
}

// worker processes books from the jobs channel
func (t *TocExtractTask) worker(ctx context.Context, id int, jobs <-chan int64, results chan<- processingResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for bookID := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			err := t.extractAndUpdateToc(bookID)
			results <- processingResult{
				bookID: bookID,
				err:    err,
			}
		}
	}
}

func (t *TocExtractTask) extractAndUpdateToc(bookID int64) error {
	bookIDStr := fmt.Sprintf("%d", bookID)

	// Get EPUB file from cache
	epubPath, err := t.cacheManager.GetOrExtractEpub(bookIDStr)
	if err != nil {
		return fmt.Errorf("failed to get EPUB file: %w", err)
	}

	// Open EPUB and extract TOC
	book, err := epub.Open(epubPath)
	if err != nil {
		return fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer book.Close()

	// Extract TOC structure
	toc := extractTocStructure(book)

	// Add to batch instead of immediate update
	t.updateBatchMu.Lock()
	t.updateBatch = append(t.updateBatch, tocUpdateItem{
		bookID: bookID,
		toc:    toc,
	})
	shouldFlush := len(t.updateBatch) >= t.qdrantBatch
	t.updateBatchMu.Unlock()

	// Flush batch if full
	if shouldFlush {
		if err := t.flushUpdateBatch(); err != nil {
			return fmt.Errorf("failed to flush batch: %w", err)
		}
	}

	return nil
}

// flushUpdateBatch flushes pending TOC updates to Qdrant in batch
func (t *TocExtractTask) flushUpdateBatch() error {
	t.updateBatchMu.Lock()
	if len(t.updateBatch) == 0 {
		t.updateBatchMu.Unlock()
		return nil
	}

	// Take current batch and reset
	batch := t.updateBatch
	t.updateBatch = make([]tocUpdateItem, 0, t.qdrantBatch)
	t.updateBatchMu.Unlock()

	// Update all items in batch
	for _, item := range batch {
		if err := t.searcher.UpdateToc(item.bookID, item.toc); err != nil {
			log.Printf("Failed to update TOC for book %d in batch: %v", item.bookID, err)
			// Continue with other items
		}
	}

	return nil
}

// extractTocStructure extracts the complete TOC structure from EPUB
func extractTocStructure(book *epub.Book) map[string]interface{} {
	result := map[string]interface{}{
		"points": convertNavPoints(book.Ncx.Points),
	}

	// Add metadata if available
	result["metadata"] = book.Opf.Metadata

	// Add base directory
	if book.Container.Rootfile.Path != "" {
		result["baseDir"] = filepath.Dir(book.Container.Rootfile.Path)
	}

	return result
}

// convertNavPoints converts epub.NavPoint to a serializable structure
func convertNavPoints(points []epub.NavPoint) []map[string]interface{} {
	var result []map[string]interface{}

	for _, point := range points {
		p := map[string]interface{}{
			"text": point.Text,
			"content": map[string]interface{}{
				"src": point.Content.Src,
			},
		}

		// Recursively convert nested points
		if len(point.Points) > 0 {
			p["points"] = convertNavPoints(point.Points)
		}

		result = append(result, p)
	}

	return result
}

func (t *TocExtractTask) loadProgress() error {
	data, err := os.ReadFile(t.progressFile)
	if err != nil {
		return err
	}

	var progress TocExtractProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		return err
	}

	t.progress = &progress
	return nil
}

func (t *TocExtractTask) saveProgress() error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.progress == nil {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(t.progressFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(t.progress, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(t.progressFile, data, 0644)
}
