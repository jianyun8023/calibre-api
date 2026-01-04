package tasks

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/kapmahc/epub"
)

// CopyrightMetadata 版权页抽取的元数据
type CopyrightMetadata struct {
	ISBN        string `json:"isbn"`         // 核心目标
	BookTitle   string `json:"book_title"`   // 书名
	Author      string `json:"author"`       // 作者
	Translator  string `json:"translator"`   // 译者
	Publisher   string `json:"publisher"`    // 出版社
	PublishDate string `json:"publish_date"` // 出版时间
}

// CopyrightExtractTask 版权页元数据抽取任务
type CopyrightExtractTask struct {
	id           string
	mode         TaskMode
	contentApi   *content.Api
	cacheManager *cache.Manager
	status       TaskStatus
	mu           sync.RWMutex
	cancel       context.CancelFunc
	numWorkers   int // 并行工作者数量
}

// 正则表达式模式
var (
	// ISBN 匹配 - 最重要
	isbnPattern = regexp.MustCompile(`(?i)ISBN[：:\s]*([0-9X-]{10,17})`)

	// 书名匹配
	bookTitlePattern = regexp.MustCompile(`(?:书\s*名|书名)[：:]\s*(.+?)(?:\n|$)`)

	// 作者匹配
	authorPattern = regexp.MustCompile(`(?:作\s*者|作者)[：:]\s*(.+?)(?:\n|$)`)

	// 出版社匹配
	publisherPattern = regexp.MustCompile(`(?i)(?:出\s*版\s*社|出版社)[：:]\s*(.+?)(?:\n|$)`)

	// 译者匹配
	translatorPattern = regexp.MustCompile(`(?i)(?:译\s*者|译者)[：:]\s*(.+?)(?:\n|$)`)

	// 出版时间匹配
	publishDatePattern = regexp.MustCompile(`(?i)(?:出\s*版\s*时\s*间|出版时间|出版日期)[：:]\s*(.+?)(?:\n|$)`)
)

// 合集/套装检测关键词（从书籍 Title 检测）
var collectionKeywords = []string{
	"合集", "合辑", "套装", "全集",
	"丛书", "系列", "文集", "全本",
	"(套装", "（套装", "[套装", "【套装",
}

// 版权页识别关键词
var copyrightKeywords = []string{
	"版权", "版权信息", "版权页",
	"COPYRIGHT", "Copyright", "copyright",
}

// NewCopyrightExtractTask 创建新的版权页元数据抽取任务
func NewCopyrightExtractTask(
	id string,
	mode TaskMode,
	contentApi *content.Api,
	cacheManager *cache.Manager,
) *CopyrightExtractTask {
	return &CopyrightExtractTask{
		id:           id,
		mode:         mode,
		contentApi:   contentApi,
		cacheManager: cacheManager,
		numWorkers:   5, // 并行工作者数量
		status: TaskStatus{
			ID:        id,
			Type:      TaskTypeCopyrightExtract,
			Mode:      mode,
			State:     "idle",
			StartTime: time.Now(),
			Message:   "Initializing copyright extraction...",
		},
	}
}

func (t *CopyrightExtractTask) GetStatus() TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *CopyrightExtractTask) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
	if t.status.State == "running" {
		t.status.State = "stopped"
		t.status.Message = "Stopped by user"
	}
}

func (t *CopyrightExtractTask) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	t.mu.Lock()
	t.cancel = cancel
	t.status.State = "running"
	t.status.Message = "Fetching books without ISBN..."
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	// 1. 获取缺少 ISBN 的书籍 ID（倒序）
	bookIDs, err := t.getBooksWithoutISBN()
	if err != nil {
		return fmt.Errorf("failed to get books without ISBN: %w", err)
	}

	totalBooks := len(bookIDs)
	t.mu.Lock()
	t.status.Message = fmt.Sprintf("Found %d books without ISBN", totalBooks)
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	if totalBooks == 0 {
		t.mu.Lock()
		t.status.State = "completed"
		t.status.Progress = 100
		t.status.Message = "No books without ISBN found"
		t.mu.Unlock()
		return nil
	}

	// 2. 分片处理书籍（每片100本）
	const batchSize = 100
	totalBatches := (totalBooks + batchSize - 1) / batchSize
	processedTotal := 0
	successTotal := 0
	failTotal := 0
	skippedTotal := 0

	for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
		// 检查是否取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 计算当前批次的范围
		start := batchIndex * batchSize
		end := start + batchSize
		if end > totalBooks {
			end = totalBooks
		}

		currentBatch := bookIDs[start:end]
		batchNum := batchIndex + 1

		t.mu.Lock()
		t.status.Message = fmt.Sprintf("Processing batch %d/%d (%d books in this batch)...",
			batchNum, totalBatches, len(currentBatch))
		t.mu.Unlock()
		GetManager().BroadcastTaskProgress(t.id)

		log.Printf("Starting batch %d/%d with %d books", batchNum, totalBatches, len(currentBatch))

		// 处理当前批次
		success, fail, skipped, err := t.processBatch(ctx, currentBatch)
		if err != nil {
			return err
		}

		processedTotal += len(currentBatch)
		successTotal += success
		failTotal += fail
		skippedTotal += skipped

		// 更新整体进度
		t.mu.Lock()
		t.status.Progress = float64(processedTotal) / float64(totalBooks) * 100
		t.status.Message = fmt.Sprintf("Batch %d/%d completed. Total: %d/%d (✓ %d ISBN, ✗ %d failed, ⊘ %d skipped)",
			batchNum, totalBatches, processedTotal, totalBooks, successTotal, failTotal, skippedTotal)
		t.mu.Unlock()
		GetManager().BroadcastTaskProgress(t.id)

		log.Printf("Batch %d/%d completed: %d success, %d fail, %d skipped",
			batchNum, totalBatches, success, fail, skipped)

		// 批次间短暂休息，避免过度占用资源
		if batchIndex < totalBatches-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	t.mu.Lock()
	t.status.Progress = 100
	t.status.State = "completed"
	t.status.EndTime = time.Now()
	t.status.Message = fmt.Sprintf("Completed: %d/%d books processed (✓ %d ISBN extracted, ✗ %d failed, ⊘ %d skipped)",
		processedTotal, totalBooks, successTotal, failTotal, skippedTotal)
	t.mu.Unlock()
	GetManager().BroadcastTaskProgress(t.id)

	return nil
}

// getBooksWithoutISBN 获取缺少 ISBN 的书籍 ID（倒序排列）
func (t *CopyrightExtractTask) getBooksWithoutISBN() ([]int64, error) {
	// 获取所有书籍 ID
	allIDs, err := t.contentApi.GetAllBooksIds("")
	if err != nil {
		return nil, err
	}

	if len(allIDs) == 0 {
		return nil, nil
	}

	// 批量获取元数据（分批处理）
	var missingISBN []int64
	batchSize := 500

	for i := 0; i < len(allIDs); i += batchSize {
		end := i + batchSize
		if end > len(allIDs) {
			end = len(allIDs)
		}

		batch := allIDs[i:end]
		books, err := t.contentApi.GetBookMetaDatas(batch, "library")
		if err != nil {
			log.Printf("Warning: failed to get metadata for batch: %v", err)
			continue
		}

		// 过滤出缺少 ISBN 且不是合集的书籍
		for _, book := range books {
			if book.Isbn == "" && !isCollectionBook(book.Title) {
				missingISBN = append(missingISBN, book.ID)
			}
		}

		t.mu.Lock()
		t.status.Message = fmt.Sprintf("Scanning books: %d/%d, found %d without ISBN",
			end, len(allIDs), len(missingISBN))
		t.mu.Unlock()
		GetManager().BroadcastTaskProgress(t.id)
	}

	// 倒序排列（新书优先）
	sort.Slice(missingISBN, func(i, j int) bool {
		return missingISBN[i] > missingISBN[j]
	})

	return missingISBN, nil
}

// isCollectionBook 检测书籍是否为合集/套装（从书籍 Title 判断）
func isCollectionBook(title string) bool {
	for _, keyword := range collectionKeywords {
		if strings.Contains(title, keyword) {
			return true
		}
	}
	return false
}

// copyrightResult 处理结果
type copyrightResult struct {
	bookID   int64
	metadata *CopyrightMetadata
	err      error
}

// processBatch 处理一个批次的书籍，返回 (success, fail, skipped, error)
func (t *CopyrightExtractTask) processBatch(ctx context.Context, bookIDs []int64) (int, int, int, error) {
	jobsChan := make(chan int64, t.numWorkers*2)
	resultsChan := make(chan copyrightResult, t.numWorkers*2)

	// 启动工作者池
	var wg sync.WaitGroup
	for w := 0; w < t.numWorkers; w++ {
		wg.Add(1)
		go t.worker(ctx, jobsChan, resultsChan, &wg)
	}

	// 发送任务
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

	// 收集结果
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 处理结果
	total := len(bookIDs)
	processed := 0
	successCount := 0
	failCount := 0
	skippedCount := 0

	for result := range resultsChan {
		processed++

		if result.err != nil {
			if strings.Contains(result.err.Error(), "copyright page not found") ||
				strings.Contains(result.err.Error(), "ISBN not found") {
				skippedCount++
				log.Printf("Skipped book %d: %v", result.bookID, result.err)
			} else {
				failCount++
				log.Printf("Failed to extract from book %d: %v", result.bookID, result.err)
			}
		} else if result.metadata != nil && result.metadata.ISBN != "" {
			successCount++
			log.Printf("Extracted ISBN %s from book %d", result.metadata.ISBN, result.bookID)
		}

		// 更新批次内进度
		if processed%10 == 0 {
			t.mu.Lock()
			t.status.Message = fmt.Sprintf("Batch progress: %d/%d (✓ %d, ✗ %d, ⊘ %d)",
				processed, total, successCount, failCount, skippedCount)
			t.mu.Unlock()
			GetManager().BroadcastTaskProgress(t.id)
		}
	}

	// 检查上下文取消
	if ctx.Err() != nil {
		return successCount, failCount, skippedCount, ctx.Err()
	}

	return successCount, failCount, skippedCount, nil
}

// worker 工作者处理书籍
func (t *CopyrightExtractTask) worker(ctx context.Context, jobs <-chan int64, results chan<- copyrightResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for bookID := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			metadata, err := t.extractAndUpdateISBN(bookID)
			results <- copyrightResult{
				bookID:   bookID,
				metadata: metadata,
				err:      err,
			}
		}
	}
}

// extractAndUpdateISBN 抽取并更新 ISBN
func (t *CopyrightExtractTask) extractAndUpdateISBN(bookID int64) (*CopyrightMetadata, error) {
	bookIDStr := strconv.FormatInt(bookID, 10)

	// 获取 EPUB 文件
	epubPath, err := t.cacheManager.GetOrExtractEpub(bookIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get EPUB file: %w", err)
	}

	// 打开 EPUB
	book, err := epub.Open(epubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB: %w", err)
	}
	defer book.Close()

	// 查找版权页
	copyrightPath, err := findCopyrightPage(book)
	if err != nil {
		return nil, err
	}

	// 读取页面内容
	content, err := readPageContent(book, copyrightPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read copyright page: %w", err)
	}

	// 解析元数据
	metadata, err := parseMetadataFromContent(content)
	if err != nil {
		return nil, err
	}

	// 更新 ISBN 到 Calibre
	if metadata.ISBN != "" {
		if err := t.updateBookISBN(bookID, metadata.ISBN); err != nil {
			return metadata, fmt.Errorf("failed to update ISBN: %w", err)
		}
	}

	return metadata, nil
}

// findCopyrightPage 查找版权页
func findCopyrightPage(book *epub.Book) (string, error) {
	// 遍历所有导航点查找版权页
	for _, point := range flattenNavPoints(book.Ncx.Points) {
		for _, keyword := range copyrightKeywords {
			if strings.Contains(point.Text, keyword) {
				return point.Content.Src, nil
			}
		}
	}
	return "", fmt.Errorf("copyright page not found in TOC")
}

// flattenNavPoints 展平导航点
func flattenNavPoints(points []epub.NavPoint) []epub.NavPoint {
	var result []epub.NavPoint
	for _, point := range points {
		result = append(result, point)
		if len(point.Points) > 0 {
			result = append(result, flattenNavPoints(point.Points)...)
		}
	}
	return result
}

// readPageContent 读取页面内容
func readPageContent(book *epub.Book, pagePath string) (string, error) {
	// 处理路径中的锚点
	if idx := strings.Index(pagePath, "#"); idx != -1 {
		pagePath = pagePath[:idx]
	}

	// 解析相对路径
	resolvedPath := resolveContentPath(book, pagePath)

	// 打开文件
	reader, err := book.Open(resolvedPath)
	if err != nil {
		// 尝试原始路径
		reader, err = book.Open(pagePath)
		if err != nil {
			return "", fmt.Errorf("failed to open page %s: %w", pagePath, err)
		}
	}
	defer reader.Close()

	// 读取内容
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read page content: %w", err)
	}

	// 提取纯文本
	return extractTextFromHTML(string(data)), nil
}

// resolveContentPath 解析内容路径
func resolveContentPath(book *epub.Book, src string) string {
	baseDir := filepath.Dir(book.Container.Rootfile.Path)
	if baseDir == "." || baseDir == "" {
		return src
	}
	return filepath.Join(baseDir, src)
}

// extractTextFromHTML 从 HTML 中提取纯文本
func extractTextFromHTML(htmlContent string) string {
	// 1. 移除 style 和 script 标签及其内容
	reScript := regexp.MustCompile(`(?i)<script[^>]*>[\s\S]*?</script>|<style[^>]*>[\s\S]*?</style>`)
	content := reScript.ReplaceAllString(htmlContent, "")

	// 2. 将块级元素替换为换行符
	// div, p, br, h1-h6, li, ul, ol, tr, td, table, blockquote, pre, form, header, footer, nav, section, article
	reBlock := regexp.MustCompile(`(?i)</?(?:div|p|br|h[1-6]|li|ul|ol|tr|td|table|blockquote|pre|form|header|footer|nav|section|article)[^>]*>`)
	content = reBlock.ReplaceAllString(content, "\n")

	// 3. 移除所有剩余的标签（内联元素，如 span, b, i, a 等），保留其内容
	reTag := regexp.MustCompile(`<[^>]+>`)
	content = reTag.ReplaceAllString(content, "")

	// 4. 处理 HTML 实体
	content = strings.ReplaceAll(content, "&nbsp;", " ")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	// 可以添加更多常见的实体解码，或者使用 html.UnescapeString

	// 5. 清理多余空白
	// 将连续的换行符替换为单个换行符，并去除行首行尾空白
	reNewlines := regexp.MustCompile(`\n\s*\n+`)
	content = reNewlines.ReplaceAllString(content, "\n")

	return strings.TrimSpace(content)
}

// parseMetadataFromContent 从内容中解析元数据
func parseMetadataFromContent(content string) (*CopyrightMetadata, error) {
	metadata := &CopyrightMetadata{}

	// 解析 ISBN（最重要）
	if matches := isbnPattern.FindStringSubmatch(content); len(matches) > 1 {
		// 清理 ISBN 中的短横线
		isbn := strings.ReplaceAll(matches[1], "-", "")
		metadata.ISBN = isbn
	}

	// 如果没有找到 ISBN，返回错误
	if metadata.ISBN == "" {
		return nil, fmt.Errorf("ISBN not found in copyright page")
	}

	// 解析书名
	if matches := bookTitlePattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.BookTitle = strings.TrimSpace(matches[1])
	}

	// 解析作者
	if matches := authorPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.Author = strings.TrimSpace(matches[1])
	}

	// 解析出版社
	if matches := publisherPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.Publisher = strings.TrimSpace(matches[1])
	}

	// 解析译者
	if matches := translatorPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.Translator = strings.TrimSpace(matches[1])
	}

	// 解析出版时间
	if matches := publishDatePattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.PublishDate = strings.TrimSpace(matches[1])
	}

	return metadata, nil
}

// updateBookISBN 更新书籍的 ISBN
func (t *CopyrightExtractTask) updateBookISBN(bookID int64, isbn string) error {
	metadata := map[string]interface{}{
		"identifiers": map[string]string{
			"isbn": isbn,
		},
	}
	_, err := t.contentApi.UpdateMetaData(
		strconv.FormatInt(bookID, 10),
		metadata,
		"library",
	)
	return err
}
