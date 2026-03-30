package douban

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServiceGet(t *testing.T) {
	bookHTML := `<html><body>
		<h1 property="v:itemreviewed">Test Get Book</h1>
		<strong property="v:average"> 7.5 </strong>
		<span class="pl">ISBN:</span> 9780000000001<br/>
	</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/subject/42/" {
			fmt.Fprint(w, bookHTML)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)
	svc := NewService(loader, "", srv.URL)

	result := svc.Get("42", srv.URL+"/subject/{id}/")

	if !result.Success {
		t.Fatal("Expected success=true")
	}
	if result.Count != 1 {
		t.Errorf("Expected count=1, got %d", result.Count)
	}
	if result.Books[0].Title != "Test Get Book" {
		t.Errorf("Expected title 'Test Get Book', got %q", result.Books[0].Title)
	}
}

func TestServiceGetEmptyPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a page with no book data
		fmt.Fprint(w, "<html><body></body></html>")
	}))
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)
	svc := NewService(loader, "", srv.URL)

	result := svc.Get("999", srv.URL+"/subject/{id}/")
	// Parser returns an empty Book struct (body exists but no content)
	// The service wraps it as success=true with count=1
	if result.Count != 1 {
		t.Errorf("Expected count=1, got %d", result.Count)
	}
	// But the book should have no title
	if result.Books[0].Title != "" {
		t.Errorf("Expected empty title, got %q", result.Books[0].Title)
	}
}

func TestServiceGetByIsbn(t *testing.T) {
	bookHTML := `<html><body>
		<h1 property="v:itemreviewed">ISBN Book</h1>
		<span class="pl">ISBN:</span> 9781234567890<br/>
	</body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/isbn/9781234567890/" {
			fmt.Fprint(w, bookHTML)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)
	svc := NewService(loader, "", srv.URL)

	result := svc.Get("9781234567890", srv.URL+"/isbn/{isbn}/")

	if !result.Success {
		t.Fatal("Expected success=true")
	}
	if result.Books[0].Title != "ISBN Book" {
		t.Errorf("Expected title 'ISBN Book', got %q", result.Books[0].Title)
	}
}

func TestServiceSearchEmptyQuery(t *testing.T) {
	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, "http://localhost")
	svc := NewService(loader, "http://localhost/search", "http://localhost")

	result := svc.Search("", 5)
	if result.Success {
		t.Error("Expected success=false for empty query")
	}

	result = svc.Search("   ", 5)
	if result.Success {
		t.Error("Expected success=false for whitespace query")
	}
}

func TestServiceSearch(t *testing.T) {
	// Search page returns links to book pages
	searchHTML := `<html><body>
		<a class="nbg" href="BOOK_URL/subject/101/">Book 1</a>
		<a class="nbg" href="BOOK_URL/subject/102/">Book 2</a>
	</body></html>`

	book1HTML := `<html><body><h1 property="v:itemreviewed">Book One</h1></body></html>`
	book2HTML := `<html><body><h1 property="v:itemreviewed">Book Two</h1></body></html>`

	var bookSrv *httptest.Server
	bookSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			// Replace BOOK_URL placeholder with actual server URL
			html := fmt.Sprintf(`<html><body>
				<a class="nbg" href="%s/subject/101/">Book 1</a>
				<a class="nbg" href="%s/subject/102/">Book 2</a>
			</body></html>`, bookSrv.URL, bookSrv.URL)
			fmt.Fprint(w, html)
		case "/subject/101/":
			fmt.Fprint(w, book1HTML)
		case "/subject/102/":
			fmt.Fprint(w, book2HTML)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer bookSrv.Close()

	_ = searchHTML // using inline HTML above instead

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, bookSrv.URL)
	svc := NewService(loader, bookSrv.URL+"/search?cat={searchType}&q={searchText}", bookSrv.URL)

	result := svc.Search("test", 5)

	if !result.Success {
		t.Fatal("Expected success=true")
	}
	if result.Count != 2 {
		t.Errorf("Expected 2 books, got %d", result.Count)
	}

	titles := map[string]bool{}
	for _, b := range result.Books {
		titles[b.Title] = true
	}
	if !titles["Book One"] || !titles["Book Two"] {
		t.Errorf("Expected 'Book One' and 'Book Two', got %v", titles)
	}
}
