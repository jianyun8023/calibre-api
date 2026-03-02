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

type Metadata struct {
	DoubanUrl string `json:"doubanurl"`
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
