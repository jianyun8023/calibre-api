package content

import "time"

type Content struct {
	Formats             []string          `json:"formats"`
	FormatSizes         FormatSizes       `json:"format_sizes"`
	Authors             []string          `json:"authors"`
	Languages           []string          `json:"languages"`
	Publisher           string            `json:"publisher"`
	Identifiers         map[string]string `json:"identifiers"`
	AuthorSort          string            `json:"author_sort"`
	Comments            string            `json:"comments"`
	LastModified        time.Time         `json:"last_modified"`
	PubDate             time.Time         `json:"pubdate"`
	SeriesIndex         float64           `json:"series_index"`
	Sort                string            `json:"sort"`
	Size                int64             `json:"size"`
	Timestamp           string            `json:"timestamp"`
	Title               string            `json:"title"`
	UUID                string            `json:"uuid"`
	ID                  string            `json:"#id"`
	Isbn                string            `json:"#isbn"`
	UrlsFromIdentifiers [][]string        `json:"urls_from_identifiers"`
	LangNames           LangNames         `json:"lang_names"`
}

type Book struct {
	AuthorSort   string            `json:"author_sort"`
	Authors      []string          `json:"authors"`
	Comments     string            `json:"comments"`
	ID           int64             `json:"id"`
	Isbn         string            `json:"isbn"`
	Languages    []string          `json:"languages"`
	LastModified time.Time         `json:"last_modified"`
	PubDate      time.Time         `json:"pubdate"`
	Publisher    string            `json:"publisher"`
	SeriesIndex  float64           `json:"series_index"`
	Size         int64             `json:"size"`
	Tags         []string          `json:"tags"`
	Rating       int               `json:"rating"`
	Title        string            `json:"title"`
	Identifiers  map[string]string `json:"identifiers"`
}

type FormatSizes struct {
	Epub int64 `json:"EPUB"`
}

type Identifiers struct {
	Isbn string `json:"isbn"`
}

type LangNames struct {
	Zho string `json:"zho"`
}
