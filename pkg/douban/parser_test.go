package douban

import (
	"testing"
)

func TestParseBook(t *testing.T) {
	// Simple HTML Mock
	html := `
<html>
<head><title>Test Book</title></head>
<body>
    <div id="wrapper">
        <h1 property="v:itemreviewed">Test Book Title</h1>
        <div id="content">
            <div id="info">
                <span><span class="pl">作者:</span> <a href="/author/1">Author 1</a> / <a href="/author/2">Author 2</a></span><br/>
                <span class="pl">出版社:</span> Test Publisher<br/>
                <span class="pl">副标题:</span> Subtitle<br/>
                <span class="pl">出版年:</span> 2023-01<br/>
                <span class="pl">页数:</span> 100<br/>
                <span class="pl">定价:</span> 50.00<br/>
                <span class="pl">装帧:</span> Hardcover<br/>
                <span class="pl">ISBN:</span> 9781234567890<br/>
            </div>
            <div id="link-report">
                <span class="short">
                    <div class="intro">
                        <p>Short summary.</p>
                    </div>
                </span>
                <span class="all hidden">
                    <div class="intro">
                        <p>This is a summary.</p>
                    </div>
                </span>
            </div>
             <div id="mainpic">
                <a class="nbg" href="https://img.doubanio.com/view/subject/l/public/s1234567.jpg" title="Test Book Title">
                    <img src="https://img.doubanio.com/view/subject/s/public/s1234567.jpg" title="Test Book Title" alt="Test Book Title" rel="v:photo">
                </a>
            </div>
            <div id="interest_sectl">
                <strong class="ll rating_num " property="v:average"> 9.0 </strong>
            </div>
        </div>
    </div>
    <script>
        var criteria = '7:Methodology|7:Programming';
    </script>
</body>
</html>
`
	parser := NewDoubanBookParser()
	book, err := parser.Parse("https://book.douban.com/subject/1234567/", html)

	if err != nil {
		t.Errorf("Parse failed: %v", err)
	}

	if book.Title != "Test Book Title" {
		t.Errorf("Expected title 'Test Book Title', got '%s'", book.Title)
	}
	if len(book.Author) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(book.Author))
	}
	if book.Author[0] != "Author 1" {
		t.Errorf("Expected first author 'Author 1', got '%s'", book.Author[0])
	}
	if book.Publisher != "Test Publisher" {
		t.Errorf("Expected publisher 'Test Publisher', got '%s'", book.Publisher)
	}
	if book.Rating["average"] != "9.0" {
		t.Errorf("Expected rating '9.0', got '%s'", book.Rating["average"])
	}
	if book.Summary != "<p>This is a summary.</p>" {
		t.Errorf("Expected summary '<p>This is a summary.</p>', got '%s'", book.Summary)
	}
	if book.Isbn13 != "9781234567890" {
		t.Errorf("Expected ISBN '9781234567890', got '%s'", book.Isbn13)
	}
}
