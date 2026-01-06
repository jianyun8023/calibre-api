package governance

import (
	"regexp"
	"strings"
)

var (
	strongCollectionKeywords = []string{
		"套装共", "套装全", "(套装", "（套装",
		"[套装", "【套装", "合集共", "全集共",
	}

	weakCollectionKeywords = []string{
		"套装", "合集", "合辑", "全集",
	}

	seriesKeywords = []string{
		"丛书", "系列", "文集",
	}

	copyrightKeywords = []string{
		"版权", "版权信息", "版权页",
		"COPYRIGHT", "Copyright", "copyright",
	}
)

type CollectionType int

const (
	CollectionTypeNone CollectionType = iota
	CollectionTypeStrong
	CollectionTypeWeak
	CollectionTypeSeries
	CollectionTypeMagazine
)

func DetectCollectionType(title string) CollectionType {
	for _, kw := range strongCollectionKeywords {
		if strings.Contains(title, kw) {
			return CollectionTypeStrong
		}
	}

	for _, kw := range weakCollectionKeywords {
		if strings.Contains(title, kw) {
			return CollectionTypeWeak
		}
	}

	for _, kw := range seriesKeywords {
		if strings.Contains(title, kw) {
			return CollectionTypeSeries
		}
	}

	if regexp.MustCompile(`\d{4}年|\d+期|第\d+期`).MatchString(title) {
		return CollectionTypeMagazine
	}

	return CollectionTypeNone
}

func IsCollectionBook(title string) bool {
	return DetectCollectionType(title) != CollectionTypeNone
}

func IsStrongCollection(title string) bool {
	return DetectCollectionType(title) == CollectionTypeStrong
}

func ShouldSkipBook(title string) bool {
	collType := DetectCollectionType(title)
	return collType == CollectionTypeStrong || collType == CollectionTypeMagazine
}

var (
	multiBookPatterns = []*regexp.Regexp{
		regexp.MustCompile(`第[一二三四五六七八九十\d]+本`),
		regexp.MustCompile(`第[一二三四五六七八九十\d]+部`),
		regexp.MustCompile(`《[^》]+》`),
		regexp.MustCompile(`(?i)book\s*\d+`),
		regexp.MustCompile(`(?i)volume\s*\d+`),
	}
)

type TOCEntry struct {
	Title string
	Level int
}

func HasMultipleBookStructure(toc []TOCEntry) bool {
	topLevel := filterTopLevel(toc)
	bookLikeCount := 0

	for _, entry := range topLevel {
		for _, pattern := range multiBookPatterns {
			if pattern.MatchString(entry.Title) {
				bookLikeCount++
				break
			}
		}
	}

	return bookLikeCount >= 2
}

func filterTopLevel(toc []TOCEntry) []TOCEntry {
	var result []TOCEntry
	for _, entry := range toc {
		if entry.Level <= 1 {
			result = append(result, entry)
		}
	}
	return result
}

func CountBookTitlesInTOC(toc []TOCEntry) int {
	bookTitlePattern := regexp.MustCompile(`《[^》]+》`)
	count := 0
	seen := make(map[string]bool)

	for _, entry := range toc {
		matches := bookTitlePattern.FindAllString(entry.Title, -1)
		for _, match := range matches {
			if !seen[match] {
				seen[match] = true
				count++
			}
		}
	}

	return count
}

func IsCopyrightPage(title string) bool {
	for _, kw := range copyrightKeywords {
		if strings.Contains(title, kw) {
			return true
		}
	}
	return false
}
