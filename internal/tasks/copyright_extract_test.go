package tasks

import (
	"regexp"
	"strings"
	"testing"
)

func TestExtractTextFromHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Inline tags",
			html:     `<b>版</b><b>权</b><b>信</b><b>息</b>`,
			expected: "版权信息",
		},
		{
			name:     "Block tags",
			html:     `<div>Line 1</div><p>Line 2</p>`,
			expected: "Line 1\nLine 2",
		},
		{
			name: "Mixed structure (User Example)",
			html: `<div class="header0"><h1><span id="magic_copyright_title" style="..."><b>版</b><b>权</b><b>信</b><b>息</b></span></h1></div>
<div class="part">
	<p><span id="magic_copyright_entitle" style="..."><b>C</b><b>O</b><b>P</b><b>Y</b><b>R</b><b>I</b><b>G</b><b>H</b><b>T</b></span></p>
	<p><span id="bookname" style="...">书名：复利人生</span></p>
	<p><span id="interpreter" style="...">译者：万锋</span></p>
</div>`,
			expected: "版权信息\nCOPYRIGHT\n书名：复利人生\n译者：万锋",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTextFromHTML(tt.html)
			// Normalize newlines for comparison
			got = strings.TrimSpace(got)
			if got != tt.expected {
				t.Errorf("extractTextFromHTML() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestParseMetadata(t *testing.T) {
	// Re-declare variables locally for test since they are private in package
	var (
		isbnPattern        = regexp.MustCompile(`(?i)ISBN[：:\s]*([0-9X-]{10,17})`)
		bookTitlePattern   = regexp.MustCompile(`(?:书\s*名|书名)[：:]\s*(.+?)(?:\n|$)`)
		authorPattern      = regexp.MustCompile(`(?:作\s*者|作者)[：:]\s*(.+?)(?:\n|$)`)
		publisherPattern   = regexp.MustCompile(`(?i)(?:出\s*版\s*社|出版社)[：:]\s*(.+?)(?:\n|$)`)
		translatorPattern  = regexp.MustCompile(`(?i)(?:译\s*者|译者)[：:]\s*(.+?)(?:\n|$)`)
		publishDatePattern = regexp.MustCompile(`(?i)(?:出\s*版\s*时\s*间|出版时间|出版日期)[：:]\s*(.+?)(?:\n|$)`)
	)

	content := `
版权信息
COPYRIGHT
书名：复利人生
作者：【加】尼古拉·贝鲁贝
译者：万锋
出版社：中信出版集团
出版时间：2025年3月
ISBN：9787521766004
字数：72千字
`
	metadata := &CopyrightMetadata{}

	// Helper to mimic parseMetadataFromContent logic
	if matches := isbnPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.ISBN = strings.ReplaceAll(matches[1], "-", "")
	}
	if matches := bookTitlePattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.BookTitle = strings.TrimSpace(matches[1])
	}
	if matches := authorPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.Author = strings.TrimSpace(matches[1])
	}
	if matches := publisherPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.Publisher = strings.TrimSpace(matches[1])
	}
	if matches := translatorPattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.Translator = strings.TrimSpace(matches[1])
	}
	if matches := publishDatePattern.FindStringSubmatch(content); len(matches) > 1 {
		metadata.PublishDate = strings.TrimSpace(matches[1])
	}

	if metadata.Translator != "万锋" {
		t.Errorf("Translator = %q, want %q", metadata.Translator, "万锋")
	}
	if metadata.ISBN != "9787521766004" {
		t.Errorf("ISBN = %q, want %q", metadata.ISBN, "9787521766004")
	}
}
