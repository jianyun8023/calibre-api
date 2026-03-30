package douban

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type DoubanBookParser struct{}

func NewDoubanBookParser() *DoubanBookParser {
	return &DoubanBookParser{}
}

func (p *DoubanBookParser) Parse(url string, htmlContent string) (*Book, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	book := &Book{}
	content := doc.Find("body")
	if content.Length() == 0 {
		return nil, nil
	}

	// URL and ID
	shareElement := content.Find("a.bn-sharing").First()
	if shareElement.Length() > 0 {
		if val, exists := shareElement.Attr("data-url"); exists {
			url = val
		}
	}
	book.Url = url
	matches := IdPattern.FindStringSubmatch(url)
	if len(matches) > 1 {
		book.ID = matches[1]
	}

	// Image
	aNbg := content.Find("a.nbg").First()
	if aNbg.Length() > 0 {
		if val, exists := aNbg.Attr("href"); exists {
			book.Image = val
		}
	}

	// Title
	book.Title = doc.Find("[property='v:itemreviewed']").Text()

	// Rating
	rateElement := content.Find("[property='v:average']").First()
	if rateElement.Length() > 0 {
		book.Rating = map[string]string{
			"average": strings.TrimSpace(rateElement.Text()),
		}
	}

	// Info block parsing
	content.Find("span.pl").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		isTranslator := strings.HasPrefix(text, "译者")

		if strings.HasPrefix(text, "作者") || isTranslator {
			var authors []string

			// Re-implementing strictly matching Java logic which uses nextElementSiblings
			// GoQuery NextAll() gets all following siblings. We iterate and break at br.
			s.NextAll().EachWithBreak(func(i int, sib *goquery.Selection) bool {
				if sib.Is("br") {
					return false // Break
				}
				// Only process non-br elements (mostly <a>)
				parts := regexp.MustCompile(`\s*/\s*`).Split(sib.Text(), -1)
				for _, part := range parts {
					if p := strings.TrimSpace(part); p != "" {
						authors = append(authors, p)
					}
				}
				return true
			})

			if isTranslator {
				book.Translator = authors
			} else {
				book.Author = authors
			}
		} else if strings.HasPrefix(text, "原作名") {
			book.OriginTitle = getInfo(s)
		} else if strings.HasPrefix(text, "副标题") {
			book.SubTitle = getInfo(s)
		} else if strings.HasPrefix(text, "出版社") {
			book.Publisher = getInfoOrNext(s)
		} else if strings.HasPrefix(text, "出版年") {
			book.PublishDate = getInfo(s)
		} else if strings.HasPrefix(text, "ISBN") {
			book.Isbn13 = getInfo(s)
		} else if strings.HasPrefix(text, "页数") {
			book.Pages = getInfo(s)
		} else if strings.HasPrefix(text, "定价") {
			book.Price = getInfo(s)
		} else if strings.HasPrefix(text, "装帧") {
			book.Binding = getInfo(s)
		} else if strings.HasPrefix(text, "丛书") {
			seriesElement := s.Next()
			if seriesElement.Length() > 0 {
				href, _ := seriesElement.Attr("href")
				matches := SeriesPattern.FindStringSubmatch(href)
				series := map[string]string{
					"title": seriesElement.Text(),
				}
				if len(matches) > 1 {
					series["id"] = matches[1]
				}
				book.Series = series
			}
		}
	})

	// Summary
	// Java uses: #link-report :not(.short) .intro
	summaryElement := content.Find("#link-report :not(.short) .intro").First()
	if summaryElement.Length() > 0 {
		htmlStr, _ := summaryElement.Html()
		book.Summary = strings.TrimSpace(htmlStr)
	}

	// Tags
	matchesTags := TagsPattern.FindStringSubmatch(htmlContent)
	if len(matchesTags) > 1 {
		rawTags := matchesTags[1]
		parts := strings.Split(rawTags, "|")
		for _, part := range parts {
			if strings.HasPrefix(part, "7:") {
				tagName := strings.TrimPrefix(part, "7:")
				exists := false
				for _, t := range book.Tags {
					if t.Name == tagName {
						exists = true
						break
					}
				}
				if !exists {
					book.Tags = append(book.Tags, Tag{Name: tagName, Title: tagName})
				}
			}
		}
	}

	return book, nil
}

func getInfo(s *goquery.Selection) string {
	// Get the immediate next sibling node, strictly following the element
	if len(s.Nodes) == 0 {
		return ""
	}
	node := s.Nodes[0].NextSibling
	if node != nil && node.Type == html.TextNode {
		return strings.TrimSpace(node.Data)
	}
	return ""
}

func getInfoOrNext(s *goquery.Selection) string {
	info := getInfo(s)
	if info == "" {
		next := s.Next()
		if next.Length() > 0 {
			info = strings.TrimSpace(next.Text())
		}
	}
	return info
}
