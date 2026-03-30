package douban

type Book struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	OriginTitle string            `json:"origin_title,omitempty"`
	SubTitle    string            `json:"sub_title,omitempty"`
	Author      []string          `json:"author"`
	Translator  []string          `json:"translator,omitempty"`
	Summary     string            `json:"summary,omitempty"`
	Publisher   string            `json:"publisher,omitempty"`
	PublishDate string            `json:"pubdate,omitempty"` // Default "1900-01" in Java
	Tags        []Tag             `json:"tags,omitempty"`
	Rating      map[string]string `json:"rating,omitempty"`
	Series      map[string]string `json:"series,omitempty"`
	Image       string            `json:"image,omitempty"`
	Url         string            `json:"url,omitempty"`
	Isbn13      string            `json:"isbn13,omitempty"`
	Isbn10      string            `json:"isbn10,omitempty"`
	Pages       string            `json:"pages,omitempty"`
	Binding     string            `json:"binding,omitempty"`
	Price       string            `json:"price,omitempty"`
	AuthorIntro string            `json:"author_intro,omitempty"`
	Catalog     string            `json:"catalog,omitempty"`
	EbookUrl    string            `json:"ebook_url,omitempty"`
	EbookPrice  string            `json:"ebook_price,omitempty"`
}

type Tag struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

type SearchResult struct {
	Start   int    `json:"start,omitempty"`
	Count   int    `json:"count,omitempty"`
	Total   int    `json:"total,omitempty"`
	Books   []Book `json:"books"`
	Success bool   `json:"success,omitempty"`
}
