package calibre

import (
	"testing"
	"time"
)

func TestToCompactBook(t *testing.T) {
	book := Book{
		ID:      123,
		Title:   "Test Book",
		Authors: []string{"Author 1", "Author 2"},
		Rating:  4.5,
	}

	compact := toCompactBook(book, 0.95)

	if compact.ID != "123" {
		t.Errorf("Expected ID '123', got '%s'", compact.ID)
	}
	if compact.Title != "Test Book" {
		t.Errorf("Expected Title 'Test Book', got '%s'", compact.Title)
	}
	if len(compact.Authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(compact.Authors))
	}
	// Use approximate comparison for float values
	if compact.Score < 0.949 || compact.Score > 0.951 {
		t.Errorf("Expected Score ~0.95, got %f", compact.Score)
	}
	if compact.Rating != 4.5 {
		t.Errorf("Expected Rating 4.5, got %f", compact.Rating)
	}
}

func TestToDetailedBook(t *testing.T) {
	book := Book{
		ID:        456,
		Title:     "Detailed Test Book",
		Authors:   []string{"Author A"},
		Publisher: "Test Publisher",
		Isbn:      "1234567890",
		Comments:  "This is a test comment",
		Rating:    3.8,
	}

	toc := map[string]interface{}{
		"chapters": []interface{}{
			map[string]interface{}{"title": "Chapter 1"},
			map[string]interface{}{"title": "Chapter 2"},
		},
	}

	detailed := toDetailedBook(book, toc)

	if detailed.ID != "456" {
		t.Errorf("Expected ID '456', got '%s'", detailed.ID)
	}
	if detailed.Title != "Detailed Test Book" {
		t.Errorf("Expected Title 'Detailed Test Book', got '%s'", detailed.Title)
	}
	if detailed.Publisher != "Test Publisher" {
		t.Errorf("Expected Publisher 'Test Publisher', got '%s'", detailed.Publisher)
	}
	if detailed.ISBN != "1234567890" {
		t.Errorf("Expected ISBN '1234567890', got '%s'", detailed.ISBN)
	}
	if detailed.Comments != "This is a test comment" {
		t.Errorf("Expected Comments 'This is a test comment', got '%s'", detailed.Comments)
	}
	if detailed.TocSummary == "" {
		t.Error("Expected non-empty TocSummary")
	}
	if detailed.Rating != 3.8 {
		t.Errorf("Expected Rating 3.8, got %f", detailed.Rating)
	}
}

func TestTruncateComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			maxLen:   100,
			expected: "",
		},
		{
			name:     "short string",
			input:    "short text",
			maxLen:   100,
			expected: "short text",
		},
		{
			name:     "exact length",
			input:    "exact",
			maxLen:   5,
			expected: "exact",
		},
		{
			name:     "long string",
			input:    "this is a very long comment that needs to be truncated",
			maxLen:   10,
			expected: "this is a ...",
		},
		{
			name:     "multibyte characters",
			input:    "这是一个很长的中文评论需要被截断",
			maxLen:   5,
			expected: "这是一个很...",
		},
		{
			name:     "mixed characters",
			input:    "Hello 世界 this is a test",
			maxLen:   8,
			expected: "Hello 世界...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateComments(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGenerateTocSummary(t *testing.T) {
	tests := []struct {
		name     string
		toc      interface{}
		expected string
	}{
		{
			name:     "nil toc",
			toc:      nil,
			expected: "",
		},
		{
			name: "empty chapters",
			toc: map[string]interface{}{
				"chapters": []interface{}{},
			},
			expected: "无目录信息",
		},
		{
			name: "one chapter",
			toc: map[string]interface{}{
				"chapters": []interface{}{
					map[string]interface{}{"title": "Introduction"},
				},
			},
			expected: "共 1 章，包括： 1. Introduction",
		},
		{
			name: "three chapters",
			toc: map[string]interface{}{
				"chapters": []interface{}{
					map[string]interface{}{"title": "Chapter 1"},
					map[string]interface{}{"title": "Chapter 2"},
					map[string]interface{}{"title": "Chapter 3"},
				},
			},
			expected: "共 3 章，包括： 1. Chapter 1, 2. Chapter 2, 3. Chapter 3",
		},
		{
			name: "more than three chapters",
			toc: map[string]interface{}{
				"chapters": []interface{}{
					map[string]interface{}{"title": "Chapter 1"},
					map[string]interface{}{"title": "Chapter 2"},
					map[string]interface{}{"title": "Chapter 3"},
					map[string]interface{}{"title": "Chapter 4"},
					map[string]interface{}{"title": "Chapter 5"},
				},
			},
			expected: "共 5 章，包括： 1. Chapter 1, 2. Chapter 2, 3. Chapter 3...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateTocSummary(tt.toc)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		minToken int
		maxToken int
	}{
		{
			name:     "simple string",
			data:     "hello",
			minToken: 1,
			maxToken: 5,
		},
		{
			name: "compact book",
			data: CompactBook{
				ID:      "123",
				Title:   "Test Book",
				Authors: []string{"Author 1"},
				Score:   0.95,
				Rating:  4.5,
			},
			minToken: 15,
			maxToken: 30,
		},
		{
			name: "detailed book",
			data: DetailedBook{
				ID:         "456",
				Title:      "Detailed Test Book",
				Authors:    []string{"Author A", "Author B"},
				Publisher:  "Test Publisher",
				ISBN:       "1234567890",
				Comments:   "This is a test comment with some details",
				TocSummary: "共 3 章",
				Rating:     3.8,
			},
			minToken: 30,
			maxToken: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateTokens(tt.data)
			if result < tt.minToken || result > tt.maxToken {
				t.Errorf("Expected tokens between %d and %d, got %d", tt.minToken, tt.maxToken, result)
			}
		})
	}
}

func TestCompactBookSerialization(t *testing.T) {
	book := CompactBook{
		ID:      "789",
		Title:   "Serialization Test",
		Authors: []string{"Test Author"},
		Score:   0.88,
		Rating:  4.2,
	}

	// Test that omitempty works - fields with zero values should be omitted
	emptyBook := CompactBook{
		ID:      "000",
		Title:   "Empty Test",
		Authors: []string{},
	}

	tokens1 := estimateTokens(book)
	tokens2 := estimateTokens(emptyBook)

	if tokens1 <= 0 {
		t.Error("Expected positive token count for book")
	}
	if tokens2 <= 0 {
		t.Error("Expected positive token count for emptyBook")
	}
	if tokens2 >= tokens1 {
		t.Error("Expected emptyBook to have fewer tokens than book")
	}
}

func TestDetailedBookWithLongComments(t *testing.T) {
	longComment := ""
	for i := 0; i < 1000; i++ {
		longComment += "这是一个很长的评论。"
	}

	book := Book{
		ID:       999,
		Title:    "Long Comment Book",
		Authors:  []string{"Author"},
		Comments: longComment,
		Rating:   4.0,
	}

	detailed := toDetailedBook(book, nil)

	// Comments should be truncated to 500 characters + "..."
	if len([]rune(detailed.Comments)) > 503 {
		t.Errorf("Expected comments to be truncated to ~503 runes, got %d", len([]rune(detailed.Comments)))
	}

	// Should end with "..."
	if len(detailed.Comments) > 3 && detailed.Comments[len(detailed.Comments)-3:] != "..." {
		t.Error("Expected truncated comments to end with '...'")
	}
}

func TestToCompactBookWithZeroValues(t *testing.T) {
	book := Book{
		ID:      0,
		Title:   "",
		Authors: []string{},
		Rating:  0,
	}

	compact := toCompactBook(book, 0)

	if compact.ID != "0" {
		t.Errorf("Expected ID '0', got '%s'", compact.ID)
	}
	if compact.Title != "" {
		t.Errorf("Expected empty Title, got '%s'", compact.Title)
	}
	if len(compact.Authors) != 0 {
		t.Errorf("Expected 0 authors, got %d", len(compact.Authors))
	}
}

func TestToDetailedBookWithNilToc(t *testing.T) {
	book := Book{
		ID:        111,
		Title:     "No TOC Book",
		Authors:   []string{"Author"},
		Publisher: "Publisher",
		Isbn:      "123",
		Comments:  "Short comment",
		Rating:    3.5,
	}

	detailed := toDetailedBook(book, nil)

	if detailed.TocSummary != "" {
		t.Errorf("Expected empty TocSummary for nil toc, got '%s'", detailed.TocSummary)
	}
	if detailed.Comments != "Short comment" {
		t.Errorf("Expected 'Short comment', got '%s'", detailed.Comments)
	}
}

func TestBookFieldMapping(t *testing.T) {
	now := time.Now()
	book := Book{
		ID:           12345,
		Title:        "Complete Book",
		Authors:      []string{"Author 1", "Author 2", "Author 3"},
		Publisher:    "Great Publisher",
		PubDate:      now,
		Isbn:         "9781234567890",
		Tags:         []string{"fiction", "adventure"},
		Rating:       4.7,
		SeriesIndex:  2.5,
		Comments:     "An excellent book with great reviews",
		Languages:    []string{"en", "zh"},
		LastModified: now,
		Cover:        "/covers/12345.jpg",
		FilePath:     "/books/12345.epub",
		Identifiers:  map[string]string{"isbn": "9781234567890"},
		Size:         1024000,
	}

	// Test CompactBook conversion
	compact := toCompactBook(book, 0.92)
	if compact.ID != "12345" {
		t.Errorf("CompactBook ID mismatch: expected '12345', got '%s'", compact.ID)
	}
	if len(compact.Authors) != 3 {
		t.Errorf("CompactBook Authors count mismatch: expected 3, got %d", len(compact.Authors))
	}

	// Test DetailedBook conversion
	detailed := toDetailedBook(book, nil)
	if detailed.ID != "12345" {
		t.Errorf("DetailedBook ID mismatch: expected '12345', got '%s'", detailed.ID)
	}
	if detailed.Publisher != "Great Publisher" {
		t.Errorf("DetailedBook Publisher mismatch: expected 'Great Publisher', got '%s'", detailed.Publisher)
	}
	if detailed.ISBN != "9781234567890" {
		t.Errorf("DetailedBook ISBN mismatch: expected '9781234567890', got '%s'", detailed.ISBN)
	}
}
