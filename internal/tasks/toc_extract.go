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
	id           string
	mode         TaskMode
	contentApi   *content.Api
	searcher     *qdrant.Searcher
	cacheManager *cache.Manager
	status       TaskStatus
	mu           sync.RWMutex
	cancel       context.CancelFunc
	progressFile string
	progress     *TocExtractProgress
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
	if t.mode == TaskModeIncremental && t.searcher != nil {
		// For incremental mode, we could check Qdrant for books without TOC
		// For simplicity, we'll process all books and skip those already in progress
		t.mu.Lock()
		t.status.Message = "Incremental mode: checking existing TOC data..."
		t.mu.Unlock()
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
	for _, id := range bookIDs {
		if !processedMap[id] {
			booksToProcess = append(booksToProcess, id)
		}
	}

	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Processing %d books (skipped %d already processed)",
		len(booksToProcess), len(bookIDs)-len(booksToProcess))
	t.mu.Unlock()

	// Process books
	processed := len(t.progress.ProcessedIDs)
	saveInterval := 10 // Save progress every 10 books

	for i, bookID := range booksToProcess {
		select {
		case <-ctx.Done():
			t.saveProgress()
			return ctx.Err()
		default:
		}

		t.mu.Lock()
		t.status.Progress = float64(processed+i) / float64(t.progress.TotalBooks) * 100
		t.status.Message = fmt.Sprintf("Processing book %d/%d (ID: %d)",
			processed+i+1, t.progress.TotalBooks, bookID)
		t.mu.Unlock()

		// Extract TOC
		if err := t.extractAndUpdateToc(bookID); err != nil {
			log.Printf("Failed to extract TOC for book %d: %v", bookID, err)
			// Continue with next book instead of stopping
			continue
		}

		// Mark as processed
		t.mu.Lock()
		t.progress.ProcessedIDs = append(t.progress.ProcessedIDs, bookID)
		t.progress.Processed = len(t.progress.ProcessedIDs)
		t.progress.LastUpdated = time.Now()
		t.mu.Unlock()

		// Save progress periodically
		if (i+1)%saveInterval == 0 {
			if err := t.saveProgress(); err != nil {
				log.Printf("Failed to save progress: %v", err)
			}
		}
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

	// Update Qdrant with TOC data
	if err := t.searcher.UpdateToc(bookID, toc); err != nil {
		return fmt.Errorf("failed to update TOC in Qdrant: %w", err)
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
