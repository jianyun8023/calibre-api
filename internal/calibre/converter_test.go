package calibre

import (
	"testing"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// BenchmarkConvertSemanticToBook 测试 convertSemanticToBook 的性能
func BenchmarkConvertSemanticToBook(b *testing.B) {
	book := semantic.Book{
		ID:        123,
		Title:     "Test Book",
		Authors:   []string{"Author A", "Author B", "Author C"},
		Publisher: "Test Publisher",
		Isbn:      "1234567890",
		Comments:  "This is a test book with some comments",
		Tags:      []string{"fiction", "scifi", "adventure"},
		Languages: []string{"en", "zh"},
		Identifiers: map[string]string{
			"isbn": "1234567890",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertSemanticToBook(book)
	}
}

// TestConvertSemanticToBook 测试 convertSemanticToBook 的正确性
func TestConvertSemanticToBook(t *testing.T) {
	tests := []struct {
		name    string
		input   semantic.Book
		wantID  int64
		wantLen int // 期望的 Authors 数量
	}{
		{
			name: "正常转换",
			input: semantic.Book{
				ID:      123,
				Title:   "Test Book",
				Authors: []string{"Author A", "Author B"},
			},
			wantID:  123,
			wantLen: 2,
		},
		{
			name: "单个作者",
			input: semantic.Book{
				ID:      456,
				Title:   "Single Author",
				Authors: []string{"Single Author"},
			},
			wantID:  456,
			wantLen: 1,
		},
		{
			name: "空作者",
			input: semantic.Book{
				ID:      789,
				Title:   "No Author",
				Authors: []string{},
			},
			wantID:  789,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSemanticToBook(tt.input)

			if result.ID != tt.wantID {
				t.Errorf("ID = %d, want %d", result.ID, tt.wantID)
			}

			if len(result.Authors) != tt.wantLen {
				t.Errorf("Authors length = %d, want %d", len(result.Authors), tt.wantLen)
			}
		})
	}
}
