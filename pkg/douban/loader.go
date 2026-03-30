package douban

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jellydator/ttlcache/v3"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3573.0 Safari/537.36"

type DoubanBookLoader struct {
	client    *http.Client
	bookCache *ttlcache.Cache[string, *Book]
	imgCache  *ttlcache.Cache[string, []byte]
	parser    *DoubanBookParser
	BaseUrl   string
}

func NewDoubanBookLoader(parser *DoubanBookParser, baseUrl string) *DoubanBookLoader {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			return nil
		},
	}

	bookCache := ttlcache.New[string, *Book](
		ttlcache.WithTTL[string, *Book](24*time.Hour),
		ttlcache.WithCapacity[string, *Book](1000),
	)
	go bookCache.Start()

	imgCache := ttlcache.New[string, []byte](
		ttlcache.WithTTL[string, []byte](24*time.Hour),
		ttlcache.WithCapacity[string, []byte](500),
	)
	go imgCache.Start()

	return &DoubanBookLoader{
		client:    client,
		bookCache: bookCache,
		imgCache:  imgCache,
		parser:    parser,
		BaseUrl:   baseUrl,
	}
}

func (l *DoubanBookLoader) LoadBook(url string) *Book {
	// Check cache
	if item := l.bookCache.Get("book:" + url); item != nil {
		return item.Value()
	}

	// Request
	body, err := l.doGet(url)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", url, err)
		return nil
	}

	// Parse
	book, err := l.parser.Parse(url, string(body))
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", url, err)
		return nil
	}

	if book != nil {
		l.bookCache.Set("book:"+url, book, ttlcache.DefaultTTL)
	}

	return book
}

func (l *DoubanBookLoader) LoadImage(imageUrl string) []byte {
	// Check cache
	if item := l.imgCache.Get("img:" + imageUrl); item != nil {
		return item.Value()
	}

	body, err := l.doGet(imageUrl)
	if err != nil {
		fmt.Printf("Error fetching image %s: %v\n", imageUrl, err)
		return nil
	}

	if len(body) > 0 {
		l.imgCache.Set("img:"+imageUrl, body, ttlcache.DefaultTTL)
		fmt.Printf("获取%s图片成功\n", imageUrl)
		return body
	}

	fmt.Printf("获取%s图片失败\n", imageUrl)
	return nil
}

func (l *DoubanBookLoader) doGet(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Referer", l.BaseUrl)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
