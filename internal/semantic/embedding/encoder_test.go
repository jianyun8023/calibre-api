package embedding

import (
	"testing"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/stretchr/testify/assert"
)

func TestCombineBookText(t *testing.T) {
	tests := []struct {
		name     string
		book     semantic.Book
		contains []string
	}{
		{
			name: "Full book info",
			book: semantic.Book{
				Title:    "The Go Programming Language",
				Authors:  []string{"Alan A. A. Donovan", "Brian W. Kernighan"},
				Comments: "The Go Programming Language is the authoritative resource for any programmer who wants to learn Go.",
				Tags:     []string{"programming", "golang"},
			},
			contains: []string{
				"The Go Programming Language",
				"Alan A. A. Donovan",
				"Brian W. Kernighan",
				"The Go Programming Language is the authoritative resource",
				"programming",
				"golang",
			},
		},
		{
			name: "Empty fields",
			book: semantic.Book{
				Title: "Minimal Book",
			},
			contains: []string{
				"Minimal Book",
			},
		},
		{
			name: "HTML in comments",
			book: semantic.Book{
				Title:    "HTML Book",
				Comments: "<p>This is <b>bold</b> text.</p>",
			},
			contains: []string{
				"This is bold text.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := CombineBookText(tt.book)
			for _, c := range tt.contains {
				assert.Contains(t, text, c)
			}
		})
	}
}
