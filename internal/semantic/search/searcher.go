package search

import (
	"fmt"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/milvus"
)

type Searcher struct {
	provider embedding.Provider
	client   *milvus.Client
}

func NewSearcher(provider embedding.Provider, client *milvus.Client) *Searcher {
	return &Searcher{
		provider: provider,
		client:   client,
	}
}

func (s *Searcher) Search(query string, topK int) ([]semantic.SearchResult, error) {
	// 1. Vectorize the query
	embeddings, err := s.provider.Embed([]string{query})
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated for query")
	}

	// 2. Search in Milvus
	results, err := s.client.Search(embeddings[0], topK)
	if err != nil {
		return nil, fmt.Errorf("milvus search failed: %w", err)
	}

	return results, nil
}
