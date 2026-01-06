package calibre

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// search 通用搜索接口（支持关键词、语义、混合搜索）
func (c *Api) search(r *gin.Context) {
	// Get search parameters
	q := r.Query("q")
	if q == "" {
		q = r.PostForm("q")
	}

	filterType := r.Query("filter")
	if filterType == "" {
		filterType = "title" // default to title search
	}

	limit := 20
	if limitStr := r.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	offset := 0
	if offsetStr := r.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Parse POST body for filters
	type SearchRequest struct {
		Filter []string `json:"Filter"`
		Limit  int      `json:"Limit"`
		Offset int      `json:"Offset"`
		Sort   []string `json:"Sort"`
	}

	var searchReq SearchRequest
	hasFilter := false
	if r.Request.Method == "POST" {
		if err := r.ShouldBindJSON(&searchReq); err == nil {
			// Use POST body parameters if available
			if searchReq.Limit > 0 {
				limit = searchReq.Limit
			}
			if searchReq.Offset >= 0 {
				offset = searchReq.Offset
			}
			// Parse filter from Filter array
			if len(searchReq.Filter) > 0 {
				hasFilter = true
				// Parse filter like: 'publisher = "xxx"' or 'authors = "xxx"'
				for _, filter := range searchReq.Filter {
					parts := strings.SplitN(filter, "=", 2)
					if len(parts) == 2 {
						fieldName := strings.TrimSpace(parts[0])
						fieldValue := strings.Trim(strings.TrimSpace(parts[1]), "\"")

						// Override query with filter value
						q = fieldValue
						// Set filterType based on field name
						if fieldName == "publisher" {
							filterType = "publisher"
						} else if fieldName == "authors" {
							filterType = "author"
						} else if fieldName == "tags" {
							filterType = "tags"
						} else if fieldName == "isbn" {
							filterType = "isbn"
						}
						break // Use first filter
					}
				}
			}
		}
	}

	log.Infof("Qdrant search query: %s, filter: %s, limit: %d, offset: %d, hasFilter: %v", q, filterType, limit, offset, hasFilter)

	// 优先使用带缓存的搜索器
	var searcher interface {
		Search(query string, limit int) ([]semantic.SearchResult, error)
		HybridSearchCombined(query string, limit int) ([]semantic.Book, error)
		SearchByKeyword(keyword, filterType string, limit, offset int) ([]semantic.Book, int64, error)
	}

	if c.cachedSearcher != nil {
		searcher = c.cachedSearcher
		log.Debugf("Using cached searcher for query: %s", q)
	} else {
		// 回退到原始 Qdrant 搜索器
		qdrantSearcher, ok := c.semanticSearcher.(*qdrant.Searcher)
		if !ok || qdrantSearcher == nil {
			log.Warnf("Qdrant searcher not available")
			r.JSON(http.StatusInternalServerError, gin.H{
				"error": "Search service not available",
				"code":  500,
			})
			return
		}
		searcher = qdrantSearcher
	}

	mode := r.Query("mode")
	if mode == "" {
		mode = "hybrid" // default to hybrid
	}

	// Force keyword mode for filtered searches (publisher, author, tags, isbn)
	if hasFilter {
		mode = "keyword"
	}

	var books []semantic.Book
	var total int64
	var err error

	switch mode {
	case "semantic":
		// Semantic search
		var results []semantic.SearchResult
		results, err = searcher.Search(q, limit)
		if err == nil {
			books = make([]semantic.Book, len(results))
			for i, res := range results {
				books[i] = res.Book
			}
			total = int64(len(books)) // Approximate total for semantic
		}
	case "hybrid":
		// Hybrid search
		books, err = searcher.HybridSearchCombined(q, limit)
		if err == nil {
			total = int64(len(books)) // Approximate total for hybrid
		}
	default:
		// Keyword search (default fallback or explicit 'keyword')
		books, total, err = searcher.SearchByKeyword(q, filterType, limit, offset)
	}

	if err != nil {
		log.Warnf("Search failed (mode=%s): %v", mode, err)
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := convertSemanticToBooks(books)

	r.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"records": calibreBooks,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
		"code": 200,
	})
}

// semanticSearch 语义搜索接口
func (c *Api) semanticSearch(r *gin.Context) {
	if c.semanticSearcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Semantic search is not initialized",
		})
		return
	}

	q := r.Query("q")
	if q == "" {
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Query parameter 'q' is required",
		})
		return
	}

	limit := 10
	if l := r.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	// Use Qdrant for semantic search
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Qdrant semantic search service not available",
		})
		return
	}

	// Perform semantic search with Qdrant
	results, err := searcher.Search(q, limit)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Search failed: " + err.Error(),
		})
		return
	}

	// Convert results to response format
	books := make([]map[string]interface{}, len(results))
	for i, result := range results {
		books[i] = map[string]interface{}{
			"id":            result.Book.ID,
			"title":         result.Book.Title,
			"authors":       result.Book.Authors,
			"publisher":     result.Book.Publisher,
			"pubdate":       result.Book.PubDate,
			"isbn":          result.Book.Isbn,
			"tags":          result.Book.Tags,
			"rating":        result.Book.Rating,
			"series_index":  result.Book.SeriesIndex,
			"comments":      result.Book.Comments,
			"languages":     result.Book.Languages,
			"last_modified": result.Book.LastModified,
			"score":         result.Score,
			"rank":          result.Rank,
		}
	}

	r.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": books,
	})
}
