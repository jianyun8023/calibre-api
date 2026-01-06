package governance

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type CopyrightContext struct {
	ISBN        string
	BookTitle   string
	Author      string
	Publisher   string
	PublishDate string
	PageContent string
	ISBNCount   int
}

type BookInfo struct {
	Title   string
	Authors []string
}

func CalculateConfidence(ctx *CopyrightContext, book *BookInfo) *ConfidenceBreakdown {
	breakdown := &ConfidenceBreakdown{}

	breakdown.ISBNScore = calculateISBNScore(ctx.ISBN)
	breakdown.ContextScore = calculateContextScore(ctx)
	breakdown.ComplexityPenalty = calculateComplexityPenalty(book.Title, book.Authors)

	breakdown.FinalScore = breakdown.ISBNScore + breakdown.ContextScore - breakdown.ComplexityPenalty
	if breakdown.FinalScore < 0 {
		breakdown.FinalScore = 0
	}
	if breakdown.FinalScore > 1 {
		breakdown.FinalScore = 1
	}

	breakdown.Details = fmt.Sprintf(
		"ISBN=%.2f, Context=%.2f, Penalty=%.2f",
		breakdown.ISBNScore, breakdown.ContextScore, breakdown.ComplexityPenalty,
	)

	return breakdown
}

func calculateISBNScore(isbn string) float64 {
	if isbn == "" {
		return 0
	}

	score := 0.0
	clean := strings.ReplaceAll(isbn, "-", "")

	if len(clean) == 13 {
		score += 0.05
		if validateISBN13Checksum(clean) {
			score += 0.30
		}
		if strings.HasPrefix(clean, "978") || strings.HasPrefix(clean, "979") {
			score += 0.10
		}
	} else if len(clean) == 10 {
		if validateISBN10Checksum(clean) {
			score += 0.30
		}
	}

	if !isTestISBN(clean) {
		score += 0.05
	}

	return score
}

func calculateContextScore(ctx *CopyrightContext) float64 {
	score := 0.0

	if ctx.Publisher != "" {
		score += 0.04
	}
	if ctx.Author != "" {
		score += 0.04
	}
	if ctx.PublishDate != "" {
		score += 0.04
	}
	if ctx.BookTitle != "" {
		score += 0.04
	}

	if regexp.MustCompile(`(?i)ISBN[：:\s]`).MatchString(ctx.PageContent) {
		score += 0.10
	}

	if ctx.ISBNCount > 1 {
		score -= 0.20
	} else if ctx.ISBNCount == 1 {
		score += 0.10
	}

	return score
}

func calculateComplexityPenalty(title string, authors []string) float64 {
	penalty := 0.0

	strongKeywords := []string{"套装共", "套装全", "(套装", "（套装", "[套装", "【套装", "合集共", "全集共"}
	for _, kw := range strongKeywords {
		if strings.Contains(title, kw) {
			return 0.50
		}
	}

	weakKeywords := []string{"套装", "合集", "合辑", "全集"}
	for _, kw := range weakKeywords {
		if strings.Contains(title, kw) {
			penalty += 0.20
			break
		}
	}

	seriesKeywords := []string{"丛书", "系列", "文集"}
	for _, kw := range seriesKeywords {
		if strings.Contains(title, kw) {
			penalty += 0.10
			break
		}
	}

	if regexp.MustCompile(`\d{4}年|\d+期|第\d+期`).MatchString(title) {
		penalty += 0.30
	}

	if utf8.RuneCountInString(title) > 50 {
		penalty += 0.10
	}

	if len(authors) > 3 {
		penalty += 0.05
	}

	return penalty
}

func validateISBN13Checksum(isbn string) bool {
	if len(isbn) != 13 {
		return false
	}
	sum := 0
	for i, c := range isbn {
		if c < '0' || c > '9' {
			return false
		}
		digit := int(c - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	return sum%10 == 0
}

func validateISBN10Checksum(isbn string) bool {
	if len(isbn) != 10 {
		return false
	}
	sum := 0
	for i, c := range isbn {
		var digit int
		if c == 'X' || c == 'x' {
			if i != 9 {
				return false
			}
			digit = 10
		} else if c >= '0' && c <= '9' {
			digit = int(c - '0')
		} else {
			return false
		}
		sum += digit * (10 - i)
	}
	return sum%11 == 0
}

var testISBNPrefixes = []string{
	"9780000000",
	"9790000000",
	"0000000000",
}

func isTestISBN(isbn string) bool {
	for _, prefix := range testISBNPrefixes {
		if strings.HasPrefix(isbn, prefix) {
			return true
		}
	}
	return false
}

func DetermineSuggestedAction(confidence float64, autoApplyThreshold, reviewThreshold float64) SuggestedAction {
	if confidence >= autoApplyThreshold {
		return ActionAutoApply
	}
	if confidence >= reviewThreshold {
		return ActionReview
	}
	return ActionSkip
}

func DetectFlags(ctx *CopyrightContext, book *BookInfo) []DraftFlag {
	var flags []DraftFlag

	if IsCollectionBook(book.Title) {
		flags = append(flags, FlagCollectionSuspected)
	}

	if ctx.ISBNCount > 1 {
		flags = append(flags, FlagMultipleISBN)
	}

	if ctx.ISBN != "" {
		clean := strings.ReplaceAll(ctx.ISBN, "-", "")
		if len(clean) == 13 && !validateISBN13Checksum(clean) {
			flags = append(flags, FlagISBNInvalidChecksum)
		}
		if len(clean) == 10 && !validateISBN10Checksum(clean) {
			flags = append(flags, FlagISBNInvalidChecksum)
		}
	}

	if utf8.RuneCountInString(book.Title) > 50 {
		flags = append(flags, FlagTitleTooLong)
	}

	if len(book.Authors) > 3 {
		flags = append(flags, FlagMultipleAuthors)
	}

	if regexp.MustCompile(`\d{4}年|\d+期|第\d+期`).MatchString(book.Title) {
		flags = append(flags, FlagMagazineSuspected)
	}

	return flags
}
