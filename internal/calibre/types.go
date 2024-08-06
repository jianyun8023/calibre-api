package calibre

import (
	"database/sql"
	"time"
)

type Book struct {
	AuthorSort   string            `json:"author_sort"`
	Authors      []string          `json:"authors"`
	Comments     string            `json:"comments"`
	Cover        string            `json:"cover"`
	FilePath     string            `json:"file_path"`
	ID           int64             `json:"id"`
	Isbn         string            `json:"isbn"`
	Languages    []string          `json:"languages"`
	LastModified time.Time         `json:"last_modified"`
	PubDate      time.Time         `json:"pubdate"`
	Publisher    string            `json:"publisher"`
	SeriesIndex  float64           `json:"series_index"`
	Size         int64             `json:"size"`
	Tags         []string          `json:"tags"`
	Title        string            `json:"title"`
	Rating       float64           `json:"rating"`
	Identifiers  map[string]string `json:"identifiers"`
}

type BookRaw struct {
	AuthorSort   string         `json:"author_sort"`
	Authors      string         `json:"authors"`
	Comments     sql.NullString `json:"comments"`
	Cover        string         `json:"cover"`
	FilePath     string         `json:"file_path"`
	ID           int64          `json:"id"`
	Isbn         sql.NullString `json:"isbn"`
	Languages    []string       `json:"languages"`
	LastModified time.Time      `json:"last_modified"`
	Pubdate      time.Time      `json:"pubdate"`
	Publisher    sql.NullString `json:"publisher"`
	SeriesIndex  float64        `json:"series_index"`
	Size         int64          `json:"size"`
	Tags         []string       `json:"tags"`
	Timestamp    time.Time      `json:"timestamp"`
	Title        string         `json:"title"`
	UUID         string         `json:"uuid"`
}

type Config struct {
	Address   string  `mapstructure:"address"`
	Debug     bool    `mapstructure:"debug"`
	StaticDir string  `mapstructure:"staticDir"`
	TmpDir    string  `mapstructure:"tmpdir"`
	Content   Content `mapstructure:"content"`
	Search    Search  `mapstructure:"search"`
}

type Content struct {
	Server string `mapstructure:"server"`
}

type Search struct {
	Host   string `mapstructure:"host"`
	APIKey string `mapstructure:"apikey"`
	Index  string `mapstructure:"index"`
}
