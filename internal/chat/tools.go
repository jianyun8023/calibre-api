package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/tmc/langchaingo/tools"
)

// SemanticSearchTool 语义搜索工具
type SemanticSearchTool struct {
	searcher *qdrant.Searcher
}

// NewSemanticSearchTool 创建语义搜索工具
func NewSemanticSearchTool(searcher *qdrant.Searcher) *SemanticSearchTool {
	return &SemanticSearchTool{searcher: searcher}
}

func (t *SemanticSearchTool) Name() string {
	return "semantic_search"
}

func (t *SemanticSearchTool) Description() string {
	return "搜索书库中的书籍。输入：自然语言查询（如'机器学习相关的书'）。输出：相关书籍列表（JSON格式）。"
}

func (t *SemanticSearchTool) Call(ctx context.Context, input string) (string, error) {
	results, err := t.searcher.Search(input, 5)
	if err != nil {
		return "", fmt.Errorf("search failed: %w", err)
	}

	// 格式化结果
	books := make([]map[string]interface{}, 0)
	for _, result := range results {
		books = append(books, map[string]interface{}{
			"id":      result.Book.ID,
			"title":   result.Book.Title,
			"authors": result.Book.Authors,
			"score":   result.Score,
		})
	}

	output, _ := json.Marshal(books)
	return string(output), nil
}

// BookRecommendationTool 书籍推荐工具
type BookRecommendationTool struct {
	searcher *qdrant.Searcher
}

// NewBookRecommendationTool 创建书籍推荐工具
func NewBookRecommendationTool(searcher *qdrant.Searcher) *BookRecommendationTool {
	return &BookRecommendationTool{searcher: searcher}
}

func (t *BookRecommendationTool) Name() string {
	return "recommend_books"
}

func (t *BookRecommendationTool) Description() string {
	return "根据用户兴趣推荐书籍。输入：用户兴趣描述（如'我喜欢科幻小说'）。输出：推荐书籍列表（JSON格式）。"
}

func (t *BookRecommendationTool) Call(ctx context.Context, input string) (string, error) {
	// 使用语义搜索推荐
	results, err := t.searcher.Search(input, 3)
	if err != nil {
		return "", fmt.Errorf("recommendation failed: %w", err)
	}

	books := make([]map[string]interface{}, 0)
	for _, result := range results {
		books = append(books, map[string]interface{}{
			"id":      result.Book.ID,
			"title":   result.Book.Title,
			"authors": result.Book.Authors,
			"reason":  fmt.Sprintf("匹配度 %.2f", result.Score),
		})
	}

	output, _ := json.Marshal(books)
	return string(output), nil
}

// TocFetcher 定义获取 TOC 的函数类型
type TocFetcher func(ctx context.Context, bookID int64) (string, error)

// BookTocTool 书籍目录工具
type BookTocTool struct {
	fetcher TocFetcher
}

// NewBookTocTool 创建书籍目录工具
func NewBookTocTool(fetcher TocFetcher) *BookTocTool {
	return &BookTocTool{fetcher: fetcher}
}

func (t *BookTocTool) Name() string {
	return "get_book_toc"
}

func (t *BookTocTool) Description() string {
	return "获取书籍的目录结构。输入：书籍ID（整数）。输出：书籍目录摘要。"
}

func (t *BookTocTool) Call(ctx context.Context, input string) (string, error) {
	var bookID int64
	_, err := fmt.Sscanf(input, "%d", &bookID)
	if err != nil {
		// 尝试解析 JSON 格式的输入（有时 LLM 会传入 JSON）
		var params map[string]interface{}
		if jsonErr := json.Unmarshal([]byte(input), &params); jsonErr == nil {
			if id, ok := params["book_id"].(float64); ok {
				bookID = int64(id)
			} else if idStr, ok := params["book_id"].(string); ok {
				fmt.Sscanf(idStr, "%d", &bookID)
			}
		}
	}

	if bookID == 0 {
		return "", fmt.Errorf("invalid book ID: %s", input)
	}

	return t.fetcher(ctx, bookID)
}

// GetToolsList 获取所有可用工具
func GetToolsList(searcher *qdrant.Searcher, tocFetcher TocFetcher) []tools.Tool {
	return []tools.Tool{
		NewSemanticSearchTool(searcher),
		NewBookRecommendationTool(searcher),
		NewBookTocTool(tocFetcher),
	}
}
