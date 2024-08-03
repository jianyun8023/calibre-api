package content

type Content struct {
	Formats             []string    `json:"formats"`
	FormatSizes         FormatSizes `json:"format_sizes"`
	Authors             []string    `json:"authors"`
	Languages           []string    `json:"languages"`
	Publisher           string      `json:"publisher"`
	Identifiers         Identifiers `json:"identifiers"`
	AuthorSort          string      `json:"author_sort"`
	Comments            string      `json:"comments"`
	LastModified        string      `json:"last_modified"`
	Pubdate             string      `json:"pubdate"`
	SeriesIndex         float64     `json:"series_index"`
	Sort                string      `json:"sort"`
	Size                int64       `json:"size"`
	Timestamp           string      `json:"timestamp"`
	Title               string      `json:"title"`
	UUID                string      `json:"uuid"`
	ID                  string      `json:"#id"`
	Isbn                string      `json:"#isbn"`
	UrlsFromIdentifiers [][]string  `json:"urls_from_identifiers"`
	LangNames           LangNames   `json:"lang_names"`
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
