package douban

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMockBookServer() *httptest.Server {
	html := `<html><head><title>Test</title></head><body>
		<h1 property="v:itemreviewed">Mock Book</h1>
		<div id="mainpic"><a class="nbg" href="https://img.example.com/cover.jpg"></a></div>
		<strong property="v:average"> 8.5 </strong>
		<span class="pl">ISBN:</span> 9781234567890<br/>
		<span class="pl">出版社:</span> Mock Publisher<br/>
		<div id="link-report"><span class="all hidden"><div class="intro"><p>Summary</p></div></span></div>
	</body></html>`

	mux := http.NewServeMux()
	mux.HandleFunc("/subject/123/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, html)
	})
	mux.HandleFunc("/cover.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte{0xFF, 0xD8, 0xFF}) // minimal JPEG header bytes
	})
	return httptest.NewServer(mux)
}

func TestLoaderLoadBook(t *testing.T) {
	srv := newMockBookServer()
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)

	book := loader.LoadBook(srv.URL + "/subject/123/")
	if book == nil {
		t.Fatal("LoadBook returned nil")
	}

	if book.Title != "Mock Book" {
		t.Errorf("Expected title 'Mock Book', got %q", book.Title)
	}
	if book.Rating["average"] != "8.5" {
		t.Errorf("Expected rating '8.5', got %q", book.Rating["average"])
	}
	if book.Isbn13 != "9781234567890" {
		t.Errorf("Expected ISBN '9781234567890', got %q", book.Isbn13)
	}
}

func TestLoaderLoadBookCaching(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprint(w, `<html><body><h1 property="v:itemreviewed">Cached Book</h1></body></html>`)
	}))
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)

	// First call
	book1 := loader.LoadBook(srv.URL + "/subject/1/")
	// Second call - should hit cache
	book2 := loader.LoadBook(srv.URL + "/subject/1/")

	if callCount != 1 {
		t.Errorf("Expected 1 HTTP call (cached), got %d", callCount)
	}
	if book1 == nil || book2 == nil {
		t.Fatal("LoadBook returned nil")
	}
	if book1.Title != book2.Title {
		t.Errorf("Cached result mismatch: %q vs %q", book1.Title, book2.Title)
	}
}

func TestLoaderLoadBookNetworkError(t *testing.T) {
	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, "http://localhost:1")

	book := loader.LoadBook("http://localhost:1/nonexistent")
	if book != nil {
		t.Error("Expected nil for network error, got non-nil")
	}
}

func TestLoaderLoadImage(t *testing.T) {
	srv := newMockBookServer()
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)

	data := loader.LoadImage(srv.URL + "/cover.jpg")
	if len(data) == 0 {
		t.Error("Expected non-empty image data")
	}
	if data[0] != 0xFF {
		t.Errorf("Expected JPEG magic byte 0xFF, got 0x%X", data[0])
	}
}

func TestLoaderLoadImageCaching(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Write([]byte{0x89, 0x50, 0x4E, 0x47}) // PNG header
	}))
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)

	loader.LoadImage(srv.URL + "/img.png")
	loader.LoadImage(srv.URL + "/img.png")

	if callCount != 1 {
		t.Errorf("Expected 1 HTTP call (cached), got %d", callCount)
	}
}

func TestLoaderLoadImage404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	parser := NewDoubanBookParser()
	loader := NewDoubanBookLoader(parser, srv.URL)

	data := loader.LoadImage(srv.URL + "/missing.jpg")
	if data != nil {
		t.Error("Expected nil for 404, got non-nil")
	}
}
