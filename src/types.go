package calibreApi

import (
	"github.com/studio-b12/gowebdav"
	"io"
	"os"
	"time"
)

type Config struct {
	Address string `mapstructure:"address"`
	Debug   bool   `mapstructure:"debug"`
	Calibre struct {
		Library string `mapstructure:"library"`
		FixPath struct {
			From string `mapstructure:"from"`
			To   string `mapstructure:"to"`
		} `mapstructure:"fixpath"`
	} `mapstructure:"calibre"`
	Search struct {
		Host   string `mapstructure:"host"`
		APIKey string `mapstructure:"apikey"`
		Index  string `mapstructure:"index"`
	} `mapstructure:"search"`
	Storage struct {
		Use    string `mapstructure:"use"`
		TmpDir string `mapstructure:"tmpdir"`
		Webdav struct {
			Host     string `mapstructure:"host"`
			User     string `mapstructure:"user"`
			Password string `mapstructure:"password"`
			Path     string `mapstructure:"path"`
		} `mapstructure:"webdav"`
	} `mapstructure:"storage"`
}

type Book struct {
	AuthorSort  string   `json:"author_sort"`
	Authors     string   `json:"authors"`
	Cover       string   `json:"cover"`
	Formats     []string `json:"formats"`
	ID          int64    `json:"id"`
	Identifiers struct {
		MobiAsin string `json:"mobi-asin"`
	} `json:"identifiers"`
	Isbn         string    `json:"isbn"`
	Languages    []string  `json:"languages"`
	LastModified time.Time `json:"last_modified"`
	Pubdate      time.Time `json:"pubdate"`
	Publisher    string    `json:"publisher"`
	SeriesIndex  float64   `json:"series_index"`
	Size         int64     `json:"size"`
	Tags         []string  `json:"tags"`
	Timestamp    time.Time `json:"timestamp"`
	Title        string    `json:"title"`
	UUID         string    `json:"uuid"`
}

type FileClient interface {
	Stat(path string) (os.FileInfo, error)
	ReadStream(path string) (io.ReadCloser, error)
}

func NewWebDavClient(uri, user, pw string) FileClient {
	newClient := gowebdav.NewClient(uri, user, pw)
	client := interface{}(newClient).(FileClient)
	return client
}
