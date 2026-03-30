package calibre

import (
	"database/sql"
	"time"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/semantic"
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

// Douban 默认配置常量
const (
	DefaultDoubanBaseUrl   = "https://book.douban.com/"
	DefaultDoubanSearchUrl = "https://www.douban.com/search?cat={searchType}&q={searchText}"
	DefaultDoubanIsbnUrl   = "https://book.douban.com/isbn/{isbn}/"
	DefaultDoubanDetailUrl = "https://book.douban.com/subject/{id}/"
)

type DoubanConfig struct {
	BaseUrl   string `mapstructure:"baseurl"`
	SearchUrl string `mapstructure:"searchurl"`
	IsbnUrl   string `mapstructure:"isbnurl"`
	DetailUrl string `mapstructure:"detailurl"`
}

type Metadata struct {
	DoubanUrl    string       `mapstructure:"doubanurl"`    // HTTP 模式 URL（兼容旧配置）
	DoubanMode   string       `mapstructure:"doubanmode"`   // "local" 或 "http"，默认 "http"
	DoubanConfig DoubanConfig `mapstructure:"doubanconfig"` // 本地模式配置（可选，使用默认值）
}

// ApplyDoubanDefaults 应用默认配置到 Metadata
func (m *Metadata) ApplyDoubanDefaults() {
	// 默认使用 HTTP 模式（兼容性）
	if m.DoubanMode == "" {
		m.DoubanMode = "http"
	}

	// 只有在 local 模式下才填充默认值
	if m.DoubanMode == "local" {
		if m.DoubanConfig.BaseUrl == "" {
			m.DoubanConfig.BaseUrl = DefaultDoubanBaseUrl
		}
		if m.DoubanConfig.SearchUrl == "" {
			m.DoubanConfig.SearchUrl = DefaultDoubanSearchUrl
		}
		if m.DoubanConfig.IsbnUrl == "" {
			m.DoubanConfig.IsbnUrl = DefaultDoubanIsbnUrl
		}
		if m.DoubanConfig.DetailUrl == "" {
			m.DoubanConfig.DetailUrl = DefaultDoubanDetailUrl
		}
	}
}

type DraftsConfig struct {
	DBPath     string `mapstructure:"db_path"`
	ExpireDays int    `mapstructure:"expire_days"`
}

type Config struct {
	Address     string             `mapstructure:"address"`
	Debug       bool               `mapstructure:"debug"`
	TmpDir      string             `mapstructure:"tmpdir"`
	Content     Content            `mapstructure:"content"`
	Metadata    Metadata           `mapstructure:"metadata"`
	MCP         MCPConfig          `mapstructure:"mcp"`
	Embedding   semantic.Embedding `mapstructure:"embedding"`
	Qdrant      QdrantConfig       `mapstructure:"qdrant"`
	Meilisearch MeilisearchConfig  `mapstructure:"meilisearch"`
	Cache       cache.Config       `mapstructure:"cache"`
	Drafts      DraftsConfig       `mapstructure:"drafts"`
}

type QdrantConfig struct {
	URL        string `mapstructure:"url"`
	Collection string `mapstructure:"collection"`
	Timeout    int    `mapstructure:"timeout"`
}

type MeilisearchConfig struct {
	Host      string `mapstructure:"host"`
	APIKey    string `mapstructure:"api_key"`
	IndexName string `mapstructure:"index_name"`
}

type Content struct {
	Server string `mapstructure:"server"`
}

type MCPConfig struct {
	Enabled         bool   `mapstructure:"enabled"`
	ServerName      string `mapstructure:"server_name"`
	Version         string `mapstructure:"version"`
	Transport       string `mapstructure:"transport"`        // "sse" or "http"
	SSEEndpoint     string `mapstructure:"sse_endpoint"`     // SSE endpoint path
	MessageEndpoint string `mapstructure:"message_endpoint"` // Message endpoint path
	BaseURL         string `mapstructure:"base_url"`
	Timeout         int    `mapstructure:"timeout"`
}
