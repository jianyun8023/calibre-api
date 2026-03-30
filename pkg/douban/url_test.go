package douban

import (
	"testing"
)

func TestIsBookUrl(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"valid book url", "https://book.douban.com/subject/1234567/", true},
		{"valid book url without trailing slash", "https://book.douban.com/subject/1234567", true},
		{"valid book url with path", "https://book.douban.com/subject/99999999/", true},
		{"invalid url - no subject", "https://book.douban.com/author/1234567/", false},
		{"invalid url - empty", "", false},
		{"invalid url - random", "https://example.com/foo", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBookUrl(tt.url)
			if got != tt.want {
				t.Errorf("IsBookUrl(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestIdPattern(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		wantID string
		wantOk bool
	}{
		{"extract id", "https://book.douban.com/subject/1234567/", "1234567", true},
		{"extract id no slash", "https://book.douban.com/subject/99/", "99", true},
		{"no match", "https://book.douban.com/author/123/", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := IdPattern.FindStringSubmatch(tt.url)
			if tt.wantOk {
				if len(matches) < 2 || matches[1] != tt.wantID {
					t.Errorf("IdPattern on %q: got %v, want id=%q", tt.url, matches, tt.wantID)
				}
			} else {
				if len(matches) >= 2 {
					t.Errorf("IdPattern on %q: expected no match, got %v", tt.url, matches)
				}
			}
		})
	}
}

func TestSeriesPattern(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		wantID string
		wantOk bool
	}{
		{"extract series id", "https://book.douban.com/series/12345/", "12345", true},
		{"no match", "https://book.douban.com/subject/12345/", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := SeriesPattern.FindStringSubmatch(tt.url)
			if tt.wantOk {
				if len(matches) < 2 || matches[1] != tt.wantID {
					t.Errorf("SeriesPattern on %q: got %v, want id=%q", tt.url, matches, tt.wantID)
				}
			} else {
				if len(matches) >= 2 {
					t.Errorf("SeriesPattern on %q: expected no match, got %v", tt.url, matches)
				}
			}
		})
	}
}

func TestTagsPattern(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   string
		wantOk bool
	}{
		{"extract tags", "var criteria = '7:Go|7:Programming';", "7:Go|7:Programming", true},
		{"no match", "var x = 'hello';", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := TagsPattern.FindStringSubmatch(tt.input)
			if tt.wantOk {
				if len(matches) < 2 || matches[1] != tt.want {
					t.Errorf("TagsPattern on %q: got %v, want %q", tt.input, matches, tt.want)
				}
			} else {
				if len(matches) >= 2 {
					t.Errorf("TagsPattern on %q: expected no match, got %v", tt.input, matches)
				}
			}
		})
	}
}

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  map[string]string
	}{
		{"simple query", "a=1&b=2", map[string]string{"a": "1", "b": "2"}},
		{"single param", "key=value", map[string]string{"key": "value"}},
		{"empty", "", map[string]string{}},
		{"invalid", "%%%", map[string]string{}},
		{"duplicate keys takes first", "a=1&a=2", map[string]string{"a": "1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQuery(tt.query)
			if len(got) != len(tt.want) {
				t.Errorf("ParseQuery(%q) returned %d entries, want %d", tt.query, len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("ParseQuery(%q)[%q] = %q, want %q", tt.query, k, got[k], v)
				}
			}
		})
	}
}
