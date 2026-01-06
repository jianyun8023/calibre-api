package cache

import (
	"fmt"
	"strconv"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// CachedSearcher 带缓存的搜索器
// 包装 qdrant.Searcher，为搜索操作添加缓存层
type CachedSearcher struct {
	searcher *qdrant.Searcher
	cache    *SearchCache
	metrics  *CacheMetrics
}

// NewCachedSearcher 创建带缓存的搜索器
func NewCachedSearcher(searcher *qdrant.Searcher, maxSize int, ttl int) *CachedSearcher {
	return &CachedSearcher{
		searcher: searcher,
		cache:    NewSearchCache(maxSize, 0), // ttl 将从参数设置
		metrics:  NewCacheMetrics(),
	}
}

// Search 语义搜索（带缓存）
func (cs *CachedSearcher) Search(query string, limit int) ([]semantic.SearchResult, error) {
	// 生成缓存键
	cacheKey := HashKey("semantic", query, strconv.Itoa(limit))

	// 尝试从缓存获取
	if cached, found := cs.cache.Get(cacheKey); found {
		cs.metrics.RecordHit()
		log.Debugf("Cache hit for semantic search: query=%s, limit=%d", query, limit)
		return cached.([]semantic.SearchResult), nil
	}

	// 缓存未命中，执行实际搜索
	cs.metrics.RecordMiss()
	results, err := cs.searcher.Search(query, limit)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	cs.cache.Set(cacheKey, results)
	cs.metrics.RecordSet()

	return results, nil
}

// HybridSearchCombined 混合搜索（带缓存）
func (cs *CachedSearcher) HybridSearchCombined(query string, limit int) ([]semantic.Book, error) {
	// 生成缓存键
	cacheKey := HashKey("hybrid", query, strconv.Itoa(limit))

	// 尝试从缓存获取
	if cached, found := cs.cache.Get(cacheKey); found {
		cs.metrics.RecordHit()
		log.Debugf("Cache hit for hybrid search: query=%s, limit=%d", query, limit)
		return cached.([]semantic.Book), nil
	}

	// 缓存未命中，执行实际搜索
	cs.metrics.RecordMiss()
	books, err := cs.searcher.HybridSearchCombined(query, limit)
	if err != nil {
		return nil, err
	}

	// 缓存结果
	cs.cache.Set(cacheKey, books)
	cs.metrics.RecordSet()

	return books, nil
}

// SearchByKeyword 关键词搜索（带缓存）
func (cs *CachedSearcher) SearchByKeyword(keyword, filterType string, limit, offset int) ([]semantic.Book, int64, error) {
	// 生成缓存键
	cacheKey := HashKey("keyword", keyword, filterType, strconv.Itoa(limit), strconv.Itoa(offset))

	// 尝试从缓存获取
	if cached, found := cs.cache.Get(cacheKey); found {
		cs.metrics.RecordHit()
		log.Debugf("Cache hit for keyword search: keyword=%s, filter=%s", keyword, filterType)
		result := cached.(KeywordSearchResult)
		return result.Books, result.Total, nil
	}

	// 缓存未命中，执行实际搜索
	cs.metrics.RecordMiss()
	books, total, err := cs.searcher.SearchByKeyword(keyword, filterType, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// 缓存结果
	result := KeywordSearchResult{
		Books: books,
		Total: total,
	}
	cs.cache.Set(cacheKey, result)
	cs.metrics.RecordSet()

	return books, total, nil
}

// KeywordSearchResult 关键词搜索结果（用于缓存）
type KeywordSearchResult struct {
	Books []semantic.Book
	Total int64
}

// InvalidateCache 清除缓存
// 在书籍数据更新时调用
func (cs *CachedSearcher) InvalidateCache() {
	cs.cache.Clear()
	log.Info("Search cache cleared due to data update")
}

// InvalidateBookCache 清除特定书籍相关的缓存
func (cs *CachedSearcher) InvalidateBookCache(bookID int64) {
	// 简化实现：清除所有缓存
	// TODO: 实现更精细的缓存失效策略
	cs.cache.Clear()
	log.Infof("Search cache cleared due to book update: bookID=%d", bookID)
}

// GetCacheStats 获取缓存统计信息
func (cs *CachedSearcher) GetCacheStats() map[string]interface{} {
	stats := cs.cache.Stats()
	metrics := cs.metrics.Stats()

	return map[string]interface{}{
		"cache":   stats,
		"metrics": metrics,
	}
}

// CleanExpiredCache 清理过期缓存（定期调用）
func (cs *CachedSearcher) CleanExpiredCache() int {
	count := cs.cache.CleanExpired()
	if count > 0 {
		log.Infof("Cleaned %d expired cache entries", count)
	}
	return count
}

// GetUnderlyingSearcher 获取底层搜索器（用于直接访问）
func (cs *CachedSearcher) GetUnderlyingSearcher() *qdrant.Searcher {
	return cs.searcher
}

// WrapSearcher 包装现有搜索器添加缓存功能
func WrapSearcher(searcher *qdrant.Searcher, maxSize, ttlSeconds int) *CachedSearcher {
	if maxSize <= 0 {
		maxSize = 1000 // 默认缓存 1000 个查询
	}
	if ttlSeconds <= 0 {
		ttlSeconds = 300 // 默认 5 分钟
	}

	log.Infof("Initializing search cache: maxSize=%d, ttl=%ds", maxSize, ttlSeconds)
	return NewCachedSearcher(searcher, maxSize, ttlSeconds)
}

// SearchCacheConfig 搜索缓存配置
type SearchCacheConfig struct {
	Enabled    bool `yaml:"enabled" json:"enabled"`       // 是否启用缓存
	MaxSize    int  `yaml:"max_size" json:"max_size"`     // 最大缓存条目数
	TTLSeconds int  `yaml:"ttl_seconds" json:"ttl_seconds"` // 缓存过期时间（秒）
}

// DefaultSearchCacheConfig 默认缓存配置
func DefaultSearchCacheConfig() SearchCacheConfig {
	return SearchCacheConfig{
		Enabled:    true,
		MaxSize:    1000,
		TTLSeconds: 300, // 5 分钟
	}
}

// Validate 验证配置
func (c SearchCacheConfig) Validate() error {
	if c.MaxSize < 0 {
		return fmt.Errorf("max_size must be non-negative, got %d", c.MaxSize)
	}
	if c.TTLSeconds < 0 {
		return fmt.Errorf("ttl_seconds must be non-negative, got %d", c.TTLSeconds)
	}
	return nil
}

