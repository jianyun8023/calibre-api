package calibre

import (
	"time"
)

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

type Config struct {
	Address string  `mapstructure:"address"`
	Debug   bool    `mapstructure:"debug"`
	Search  Search  `mapstructure:"search"`
	Storage Storage `mapstructure:"storage"`
}
type Search struct {
	Host     string `mapstructure:"host"`
	APIKey   string `mapstructure:"apikey"`
	Index    string `mapstructure:"index"`
	TrimPath string `mapstructure:"trimPath"`
}
type Storage struct {
	Use    string `mapstructure:"use"`
	TmpDir string `mapstructure:"tmpdir"`
	Webdav Webdav `mapstructure:"webdav"`
	Minio  Minio  `mapstructure:"minio"`
	Local  Local  `mapstructure:"local"`
}

type Webdav struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Path     string `mapstructure:"path"`
}

type Minio struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	SecretAccessKey string `mapstructure:"secretAccessKey"`
	UseSSL          bool   `mapstructure:"useSSL"`
	BucketName      string `mapstructure:"bucketName"`
	Path            string `mapstructure:"path"`
}

type Local struct {
	Path string `mapstructure:"path"`
}
