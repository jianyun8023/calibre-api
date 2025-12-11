package calibre

import (
	"encoding/json"
	"testing"
)

// TestTokenOptimization 验证 token 优化效果
func TestTokenOptimization(t *testing.T) {
	// 创建一个完整的 Book 对象
	fullBook := Book{
		ID:           12345,
		Title:        "Complete Test Book with a Very Long Title",
		Authors:      []string{"Author One", "Author Two", "Author Three"},
		Publisher:    "Test Publisher Inc.",
		Isbn:         "9781234567890",
		Comments:     "This is a very detailed comment about the book. It contains multiple sentences and provides comprehensive information about the content, themes, and style of the book. The book is highly recommended for readers interested in the subject matter.",
		Rating:       4.5,
		SeriesIndex:  2.5,
		Tags:         []string{"fiction", "adventure", "bestseller"},
		Languages:    []string{"en", "zh"},
		Cover:        "/covers/12345.jpg",
		FilePath:     "/books/12345.epub",
		Identifiers:  map[string]string{"isbn": "9781234567890", "amazon": "B001234567"},
		Size:         1024000,
	}

	// 测试 CompactBook 的 token 消耗
	compactBook := toCompactBook(fullBook, 0.95)
	compactJSON, _ := json.Marshal(compactBook)
	compactTokens := len(compactJSON) / 4

	// 测试完整 Book 的 token 消耗（模拟优化前）
	fullJSON, _ := json.Marshal(fullBook)
	fullTokens := len(fullJSON) / 4

	// 验证 CompactBook 的 token 消耗显著减少
	reduction := float64(fullTokens-compactTokens) / float64(fullTokens) * 100

	t.Logf("Full Book JSON length: %d bytes, estimated tokens: %d", len(fullJSON), fullTokens)
	t.Logf("Compact Book JSON length: %d bytes, estimated tokens: %d", len(compactJSON), compactTokens)
	t.Logf("Token reduction: %.2f%%", reduction)

	// 验证至少减少 40% tokens
	if reduction < 40 {
		t.Errorf("Expected at least 40%% token reduction, got %.2f%%", reduction)
	}

	// 验证 CompactBook 包含必要字段
	if compactBook.ID == "" {
		t.Error("CompactBook missing ID")
	}
	if compactBook.Title == "" {
		t.Error("CompactBook missing Title")
	}
	if len(compactBook.Authors) == 0 {
		t.Error("CompactBook missing Authors")
	}
}

// TestDetailedBookOptimization 验证 DetailedBook 的优化效果
func TestDetailedBookOptimization(t *testing.T) {
	// 创建一个带有长评论的 Book
	longComment := ""
	for i := 0; i < 200; i++ {
		longComment += "这是一段很长的评论。"
	}

	book := Book{
		ID:        999,
		Title:     "Book with Long Comment",
		Authors:   []string{"Test Author"},
		Publisher: "Test Publisher",
		Isbn:      "1234567890",
		Comments:  longComment,
		Rating:    4.0,
	}

	// 创建复杂的 TOC
	toc := map[string]interface{}{
		"chapters": []interface{}{
			map[string]interface{}{"title": "Chapter 1: Introduction"},
			map[string]interface{}{"title": "Chapter 2: Background"},
			map[string]interface{}{"title": "Chapter 3: Methodology"},
			map[string]interface{}{"title": "Chapter 4: Results"},
			map[string]interface{}{"title": "Chapter 5: Discussion"},
			map[string]interface{}{"title": "Chapter 6: Conclusion"},
		},
	}

	// 测试 DetailedBook
	detailedBook := toDetailedBook(book, toc)

	// 验证 comments 被截断
	if len([]rune(detailedBook.Comments)) > 503 {
		t.Errorf("Comments not truncated properly, length: %d", len([]rune(detailedBook.Comments)))
	}

	// 验证 TOC 被转换为摘要
	if detailedBook.TocSummary == "" {
		t.Error("TOC summary is empty")
	}
	if len(detailedBook.TocSummary) > 200 {
		t.Logf("TOC summary: %s", detailedBook.TocSummary)
	}

	// 计算 token 优化
	detailedJSON, _ := json.Marshal(detailedBook)
	detailedTokens := len(detailedJSON) / 4

	// 模拟优化前的完整数据（包含完整 TOC）
	fullData := map[string]interface{}{
		"id":        book.ID,
		"title":     book.Title,
		"authors":   book.Authors,
		"publisher": book.Publisher,
		"isbn":      book.Isbn,
		"comments":  book.Comments,
		"rating":    book.Rating,
		"toc":       toc,
	}
	fullJSON, _ := json.Marshal(fullData)
	fullTokens := len(fullJSON) / 4

	reduction := float64(fullTokens-detailedTokens) / float64(fullTokens) * 100

	t.Logf("Full data tokens: %d", fullTokens)
	t.Logf("Detailed book tokens: %d", detailedTokens)
	t.Logf("Token reduction: %.2f%%", reduction)

	// 验证至少减少 40% tokens
	if reduction < 40 {
		t.Errorf("Expected at least 40%% token reduction, got %.2f%%", reduction)
	}
}

// TestBatchOptimization 验证批量数据的优化效果
func TestBatchOptimization(t *testing.T) {
	// 创建 20 本书的列表（模拟搜索结果）
	books := make([]Book, 20)
	for i := 0; i < 20; i++ {
		books[i] = Book{
			ID:           int64(i + 1),
			Title:        "Test Book " + string(rune(i+65)),
			Authors:      []string{"Author " + string(rune(i+65))},
			Publisher:    "Publisher " + string(rune(i+65)),
			Isbn:         "123456789" + string(rune(i+48)),
			Comments:     "This is a comment for book " + string(rune(i+65)),
			Rating:       4.0 + float64(i%5)*0.2,
			Tags:         []string{"tag1", "tag2"},
			Languages:    []string{"en"},
			Cover:        "/covers/" + string(rune(i+65)) + ".jpg",
			FilePath:     "/books/" + string(rune(i+65)) + ".epub",
			Size:         1024000,
		}
	}

	// 转换为 CompactBook
	compactBooks := make([]CompactBook, 20)
	for i, book := range books {
		compactBooks[i] = toCompactBook(book, 0.9)
	}

	// 计算 token 消耗
	fullJSON, _ := json.Marshal(books)
	fullTokens := len(fullJSON) / 4

	compactJSON, _ := json.Marshal(compactBooks)
	compactTokens := len(compactJSON) / 4

	reduction := float64(fullTokens-compactTokens) / float64(fullTokens) * 100

	t.Logf("20 books - Full data: %d bytes, %d tokens", len(fullJSON), fullTokens)
	t.Logf("20 books - Compact data: %d bytes, %d tokens", len(compactJSON), compactTokens)
	t.Logf("Token reduction: %.2f%%", reduction)

	// 验证至少减少 50% tokens（批量数据优化效果更明显）
	if reduction < 50 {
		t.Errorf("Expected at least 50%% token reduction for batch data, got %.2f%%", reduction)
	}

	// 验证每本书的平均 token 数
	avgTokensPerBook := compactTokens / 20
	t.Logf("Average tokens per compact book: %d", avgTokensPerBook)

	if avgTokensPerBook > 30 {
		t.Errorf("Expected average tokens per book < 30, got %d", avgTokensPerBook)
	}
}

// TestFieldPresence 验证必要字段的存在性
func TestFieldPresence(t *testing.T) {
	book := Book{
		ID:        123,
		Title:     "Test",
		Authors:   []string{"Author"},
		Publisher: "Publisher",
		Isbn:      "123",
		Comments:  "Comment",
		Rating:    4.5,
	}

	// CompactBook 应该只有 5 个字段
	compact := toCompactBook(book, 0.9)
	compactJSON, _ := json.Marshal(compact)
	var compactMap map[string]interface{}
	json.Unmarshal(compactJSON, &compactMap)

	expectedFields := []string{"id", "title", "authors", "score", "rating"}
	for _, field := range expectedFields {
		if _, exists := compactMap[field]; !exists {
			t.Errorf("CompactBook missing expected field: %s", field)
		}
	}

	// DetailedBook 应该只有 8 个字段
	detailed := toDetailedBook(book, nil)
	detailedJSON, _ := json.Marshal(detailed)
	var detailedMap map[string]interface{}
	json.Unmarshal(detailedJSON, &detailedMap)

	expectedDetailedFields := []string{"id", "title", "authors", "publisher", "isbn", "comments", "rating"}
	for _, field := range expectedDetailedFields {
		if _, exists := detailedMap[field]; !exists {
			t.Errorf("DetailedBook missing expected field: %s", field)
		}
	}

	t.Logf("CompactBook fields: %d", len(compactMap))
	t.Logf("DetailedBook fields: %d", len(detailedMap))
}
