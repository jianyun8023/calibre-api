package calibre

import (
	"database/sql"
	"time"
)

type Book struct {
	AuthorSort   string    `json:"author_sort"`
	Authors      []string  `json:"authors"`
	Comments     string    `json:"comments"`
	Cover        string    `json:"cover"`
	FilePath     string    `json:"file_path"`
	ID           int64     `json:"id"`
	Isbn         string    `json:"isbn"`
	Languages    []string  `json:"languages"`
	LastModified time.Time `json:"last_modified"`
	Pubdate      time.Time `json:"pubdate"`
	Publisher    string    `json:"publisher"`
	SeriesIndex  float64   `json:"series_index"`
	Size         int64     `json:"size"`
	Tags         []string  `json:"tags"`
	Title        string    `json:"title"`
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
	Address       string  `mapstructure:"address"`
	Debug         bool    `mapstructure:"debug"`
	StaticDir     string  `mapstructure:"static_dir"`
	TemplateDir   string  `mapstructure:"template_dir"`
	ContentServer string  `mapstructure:"content_server"`
	Search        Search  `mapstructure:"search"`
	Storage       Storage `mapstructure:"storage"`
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
