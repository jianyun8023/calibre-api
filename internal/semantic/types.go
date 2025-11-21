package semantic

import (
	"time"
)

// Book represents a book with its metadata (mirrors internal/calibre/types.go Book)
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
	Toc          interface{}       `json:"toc,omitempty"` // Table of Contents data
}

// BookEmbedding represents a book with its vector embedding
type BookEmbedding struct {
	Book      Book
	Embedding []float32
}

// SearchResult represents a semantic search result
type SearchResult struct {
	Book  Book
	Score float32
	Rank  int
}

type Embedding struct {
	Provider    string            `mapstructure:"provider"`
	Ollama      OllamaConfig      `mapstructure:"ollama"`
	SiliconFlow SiliconFlowConfig `mapstructure:"siliconflow"`
}

type OllamaConfig struct {
	APIURL string `mapstructure:"api_url"`
	Model  string `mapstructure:"model"`
}

type SiliconFlowConfig struct {
	APIURL   string `mapstructure:"api_url"`
	Model    string `mapstructure:"model"`
	APIToken string `mapstructure:"api_token"`
}
