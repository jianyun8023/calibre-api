package content

import (
	"testing"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
)

// TestEnrichBooks 测试 EnrichBooks 函数
func TestEnrichBooks(t *testing.T) {
	tests := []struct {
		name  string
		books []Book
		want  int
	}{
		{
			name:  "空书籍列表",
			books: []Book{},
			want:  0,
		},
		{
			name: "单本书籍",
			books: []Book{
				{
					ID:    123,
					Title: "Test Book",
				},
			},
			want: 1,
		},
		{
			name: "多本书籍",
			books: []Book{
				{ID: 1, Title: "Book 1"},
				{ID: 2, Title: "Book 2"},
				{ID: 3, Title: "Book 3"},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EnrichBooks(tt.books)
			if len(result) != tt.want {
				t.Errorf("EnrichBooks() length = %d, want %d", len(result), tt.want)
			}
		})
	}
}

// TestCombineBookText 测试 CombineBookText 函数
func TestCombineBookText(t *testing.T) {
	tests := []struct {
		name string
		book semantic.Book
		want string
	}{
		{
			name: "完整书籍信息",
			book: semantic.Book{
				ID:        123,
				Title:     "Test Book",
				Authors:   []string{"Author A", "Author B"},
				Publisher: "Test Publisher",
				Isbn:      "1234567890",
				Comments:  "Test comments",
				Tags:      []string{"fiction", "scifi"},
			},
			want: "Test Book | Test Book | Author A, Author B | Test comments | fiction, scifi",
		},
		{
			name: "最小信息",
			book: semantic.Book{
				ID:    456,
				Title: "Min Book",
			},
			want: "Min Book | Min Book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := embedding.CombineBookText(tt.book)
			if got != tt.want {
				t.Errorf("CombineBookText() = %q, want %q", got, tt.want)
			}
		})
	}
}

// BenchmarkEnrichBooks 基准测试 EnrichBooks
func BenchmarkEnrichBooks(b *testing.B) {
	books := make([]Book, 100)
	for i := 0; i < 100; i++ {
		books[i] = Book{
			ID:        int64(i),
			Title:     "Book " + string(rune(i)),
			Authors:   []string{"Author " + string(rune(i))},
			Publisher: "Publisher",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EnrichBooks(books)
	}
}

// BenchmarkCombineBookText 基准测试 CombineBookText
func BenchmarkCombineBookText(b *testing.B) {
	book := semantic.Book{
		ID:        123,
		Title:     "Test Book with a Long Title",
		Authors:   []string{"Author A", "Author B", "Author C"},
		Publisher: "Test Publisher Name",
		Isbn:      "1234567890123",
		Comments:  "This is a long comment with many details about the book",
		Tags:      []string{"fiction", "scifi", "adventure", "bestseller"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = embedding.CombineBookText(book)
	}
}
