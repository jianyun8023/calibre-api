package douban

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Service struct {
	loader    *DoubanBookLoader
	client    *http.Client
	SearchUrl string
	BaseUrl   string
}

func NewService(l *DoubanBookLoader, searchUrl string, baseUrl string) *Service {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			return nil
		},
	}
	return &Service{
		loader:    l,
		client:    client,
		SearchUrl: searchUrl,
		BaseUrl:   baseUrl,
	}
}

func (ds *Service) Search(query string, count int) *SearchResult {
	if strings.TrimSpace(query) == "" {
		return &SearchResult{Success: false}
	}

	// 1. Search Page
	// urlTemplate: https://www.douban.com/search?cat={searchType}&q={searchText}
	// searchType = 1001 (book)
	targetUrl := strings.Replace(ds.SearchUrl, "{searchType}", "1001", -1)
	targetUrl = strings.Replace(targetUrl, "{searchText}", url.QueryEscape(query), -1)

	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		return &SearchResult{Success: false}
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Referer", ds.BaseUrl)

	resp, err := ds.client.Do(req)
	if err != nil {
		fmt.Printf("Search error: %v\n", err)
		return &SearchResult{Success: false}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Search read error: %v\n", err)
		return &SearchResult{Success: false}
	}

	// 2. Parse Search Results (List<Element>)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return &SearchResult{Success: false}
	}

	links := doc.Find("a.nbg")

	// 3. Concurrency Load
	var wg sync.WaitGroup
	resultChan := make(chan *Book, count)

	links.Each(func(i int, s *goquery.Selection) {
		if i >= count {
			return
		}
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		// Parse "url" param from href if it's a redirect link?
		targetBookUrl := href
		u, err := url.Parse(href)
		if err == nil {
			q := u.Query()
			if val := q.Get("url"); val != "" {
				targetBookUrl = val
			}
		}

		if IsBookUrl(targetBookUrl) {
			wg.Add(1)
			go func(u string) {
				defer wg.Done()
				book := ds.loader.LoadBook(u)
				if book != nil {
					resultChan <- book
				}
			}(targetBookUrl)
		}
	})

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var books []Book
	for book := range resultChan {
		books = append(books, *book)
	}

	success := len(books) > 0
	return &SearchResult{
		Books:   books,
		Success: success,
		Count:   len(books),
	}
}

func (ds *Service) Get(id string, urlTemplate string) *SearchResult {
	// Resolve template
	// https://book.douban.com/subject/{id}/
	targetUrl := strings.Replace(urlTemplate, "{id}", id, -1)
	targetUrl = strings.Replace(targetUrl, "{isbn}", id, -1)

	book := ds.loader.LoadBook(targetUrl)
	if book != nil {
		return &SearchResult{
			Books:   []Book{*book},
			Success: true,
			Count:   1,
		}
	}
	return &SearchResult{Success: false}
}
