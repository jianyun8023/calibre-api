package calibre

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/kapmahc/epub"

	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/internal/tasks"
)

type Api struct {
	config           *Config
	contentApi       *content.Api
	baseDir          string
	http             *client.Client
	qdrantClient     *qdrant.Client
	semanticSearcher interface{} // *qdrant.Searcher
}

func (c *Api) SetupRouter(r *gin.Engine) {

	base := r.Group("/api")
	base.GET("/get/cover/:id", c.getCover)
	base.GET("/proxy/cover/*path", c.proxyCover)
	base.GET("/download/book/:id", c.getBookFile)
	base.GET("/read/:id/toc", c.getBookToc)
	base.GET("/read/:id/file/*path", c.getBookContent)
	base.GET("/book/:id", c.getBook)
	base.GET("/book/content", c.getBookContentByQuery)
	base.POST("/book/:id/delete", c.deleteBook)
	base.POST("/book/:id/update", c.updateMetadata)
	base.GET("/search", c.search)
	base.GET("/metadata/isbn/:isbn", c.getIsbn)
	base.GET("/metadata/search", c.queryMetadata)
	base.POST("/search", c.search)
	base.GET("/publisher", c.listPublisher)
	// 最近更新Recently
	base.GET("/recently", c.recently)
	base.GET("/random", c.random)
	// 所有书籍（支持排序）
	base.GET("/books/all", c.getAllBooks)

	// Enhanced Tools MCP 端点
	base.GET("/mcp/tools/enhanced", c.getEnhancedTools)
	base.POST("/mcp/tools/enhanced/:tool", c.executeEnhancedTool)

	// Semantic Search Endpoints
	base.GET("/search/semantic", c.semanticSearch)

	// Task Management Endpoints
	base.GET("/tasks", c.listTasks)
	base.POST("/tasks/start", c.startTask)
	base.POST("/tasks/:id/stop", c.stopTask)
}

// Deprecated: This function will be removed after migration to Qdrant

// getBookByID retrieves a book by ID from Qdrant
func (c *Api) getBookByID(id string) (*Book, error) {
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		return nil, fmt.Errorf("search service not available")
	}

	books, _, err := searcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil {
		return nil, err
	}
	if len(books) == 0 {
		return nil, fmt.Errorf("book not found")
	}

	book := &Book{
		ID:           books[0].ID,
		Title:        books[0].Title,
		Authors:      books[0].Authors,
		Publisher:    books[0].Publisher,
		PubDate:      books[0].PubDate,
		Isbn:         books[0].Isbn,
		Tags:         books[0].Tags,
		Rating:       books[0].Rating,
		SeriesIndex:  books[0].SeriesIndex,
		Comments:     books[0].Comments,
		Languages:    books[0].Languages,
		LastModified: books[0].LastModified,
	}
	return book, nil
}

func NewClient(config *Config) *Api {
	baseDir := config.TmpDir
	if !Exists(baseDir) {
		os.MkdirAll(baseDir, fs.ModePerm)
	}

	newClient, err := content.NewClient(config.Content.Server)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Qdrant client and searcher
	var qdrantClient *qdrant.Client
	var qdrantSearcher *qdrant.Searcher

	if config.Qdrant.URL != "" {
		qdrantClient = qdrant.NewClient(
			config.Qdrant.URL,
			config.Qdrant.Collection,
			time.Duration(config.Qdrant.Timeout)*time.Second,
		)
		log.Info("Qdrant client initialized successfully")

		// Initialize embedding provider for Qdrant searcher
		qdrantProviderConfig := embedding.ProviderConfig{
			Provider:    config.Embedding.Provider,
			Ollama:      config.Embedding.Ollama,
			SiliconFlow: config.Embedding.SiliconFlow,
			VectorDim:   4096, // Qdrant uses 4096 dimensions
		}

		qdrantProvider, err := embedding.NewProvider(qdrantProviderConfig)
		if err != nil {
			log.Warnf("Failed to create Qdrant embedding provider: %v", err)
		} else {
			qdrantSearcher = qdrant.NewSearcher(qdrantProvider, qdrantClient)
			log.Info("Qdrant searcher initialized successfully")

			// Ensure required payload indexes exist
			ctx := context.Background()
			if err := qdrantSearcher.EnsureIndexes(ctx); err != nil {
				log.Warnf("Failed to ensure Qdrant indexes (indexes may already exist): %v", err)
			} else {
				log.Info("Qdrant payload indexes ensured successfully")
			}
		}
	}

	api := Api{
		config:           config,
		baseDir:          config.TmpDir,
		contentApi:       &newClient,
		http:             newClient.Client,
		qdrantClient:     qdrantClient,
		semanticSearcher: qdrantSearcher,
	}

	// 初始化 SSE MCP 服务器（在 HTTP 模式下默认启用）
	// 注意：这里不依赖 config.MCP.Enabled，因为那个控制的是传统 MCP 模式
	baseURL := config.MCP.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	log.Infof("SSE MCP Server initialized with base URL: %s", baseURL)

	return &api
}

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

	// Check if Qdrant searcher is available
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		log.Warnf("Qdrant searcher not available")
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search service not available",
			"code":  500,
		})
		return
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
	calibreBooks := make([]Book, len(books))
	for i, book := range books {
		calibreBooks[i] = Book{
			ID:           book.ID,
			Title:        book.Title,
			Authors:      book.Authors,
			Publisher:    book.Publisher,
			PubDate:      book.PubDate,
			Isbn:         book.Isbn,
			Tags:         book.Tags,
			Rating:       book.Rating,
			SeriesIndex:  book.SeriesIndex,
			Comments:     book.Comments,
			Languages:    book.Languages,
			LastModified: book.LastModified,
			Cover:        book.Cover,
			FilePath:     book.FilePath,
		}
	}

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

// 获取书籍信息接口
func (c *Api) getBook(r *gin.Context) {
	id := r.Param("id")

	// Use Qdrant to search by ID
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "Search service not available",
			"code":    http.StatusServiceUnavailable,
		})
		return
	}

	// Search by ID using keyword search
	books, _, err := searcher.SearchByKeyword(id, "id", 1, 0)
	if err != nil || len(books) == 0 {
		r.JSON(http.StatusOK, gin.H{
			"message": "book not found",
			"code":    http.StatusNotFound,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	book := Book{
		ID:           books[0].ID,
		Title:        books[0].Title,
		Authors:      books[0].Authors,
		Publisher:    books[0].Publisher,
		PubDate:      books[0].PubDate,
		Isbn:         books[0].Isbn,
		Tags:         books[0].Tags,
		Rating:       books[0].Rating,
		SeriesIndex:  books[0].SeriesIndex,
		Comments:     books[0].Comments,
		Languages:    books[0].Languages,
		LastModified: books[0].LastModified,
		Cover:        books[0].Cover,
		FilePath:     books[0].FilePath,
	}

	r.JSON(http.StatusOK, gin.H{
		"data":    &book,
		"message": "ok",
		"code":    200,
	})

}

// 删除书籍接口
func (c *Api) deleteBook(r *gin.Context) {
	id := r.Param("id")

	err := c.contentApi.DeleteBooks([]string{id}, "")
	if err != nil {
		r.JSON(http.StatusOK, gin.H{
			"message": "book not found" + err.Error(),
			"code":    http.StatusNotFound,
		})
		return
	}

	// TODO: Also delete from Qdrant when delete API is implemented

	r.JSON(http.StatusOK, gin.H{
		"data":    true,
		"message": "ok",
		"code":    200,
	})
}

// 获取书籍目录接口
func (c *Api) getBookToc(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")

	filepath, _ := c.getFileOrCache(id)
	book, _ := epub.Open(filepath)
	points := c.expansionTree(book.Ncx.Points)
	var p []epub.NavPoint
	for i := range points {
		point := points[i]
		p = append(p, epub.NavPoint{
			Text: point.Text,
			Content: epub.Content{
				Src: path.Join("/read/"+id+"/file", path.Dir(book.Container.Rootfile.Path), point.Content.Src),
			},
		})
	}

	defer book.Close()

	r.JSON(http.StatusOK, gin.H{
		"points":   p,
		"metadata": book.Opf.Metadata,
		"manifest": book.Opf.Manifest,
		"baseDir":  path.Dir(book.Container.Rootfile.Path),
	})

}

func (c *Api) expansionTree(ori []epub.NavPoint) []epub.NavPoint {
	var points []epub.NavPoint
	for i := range ori {
		point := ori[i]
		points = append(points, point)
		if len(point.Points) > 0 {
			points = append(points, c.expansionTree(point.Points)...)
		}
	}
	return points
}

// 获取书籍内容接口
func (c *Api) getBookContent(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".epub")

	//path1 := path.Join(c.Query("baseDir"), c.Query("path"))
	path1 := r.Param("path")
	_, err := c.getBookByID(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, err)
	} else {

		filepath, _ := c.getFileOrCache(id)

		destDir := path.Join(c.baseDir, id)

		if Exists(destDir) {
			s, _ := ioutil.ReadDir(destDir)
			if len(s) == 0 {
				fmt.Println("empty")
			}
		} else {
			os.MkdirAll(destDir, fs.ModePerm)
		}

		err := unzipSource(filepath, destDir)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": err.Error(),
			})
		}
		r.FileFromFS(path1, http.Dir(destDir))
	}
}

// getBookContentByQuery 通过 query 参数获取书籍内容
func (c *Api) getBookContentByQuery(r *gin.Context) {
	// 从 query 参数获取书籍 ID
	id := r.Query("id")
	if id == "" {
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必需的参数: id",
		})
		return
	}

	// 获取文件路径参数（可选）
	filePath := r.Query("path")
	if filePath == "" {
		filePath = "OEBPS/content.opf" // 默认返回 OPF 文件
	}

	_, err := c.getBookByID(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取书籍信息失败: " + err.Error(),
		})
		return
	}

	// 获取或缓存书籍文件
	filepath, err := c.getFileOrCache(id)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取书籍文件失败: " + err.Error(),
		})
		return
	}

	// 解压目录
	destDir := path.Join(c.baseDir, id)
	if !Exists(destDir) {
		os.MkdirAll(destDir, fs.ModePerm)
	}

	// 检查目录是否为空，如果为空则解压
	if Exists(destDir) {
		s, _ := ioutil.ReadDir(destDir)
		if len(s) == 0 {
			// 目录为空，需要解压
			err := unzipSource(filepath, destDir)
			if err != nil {
				r.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "解压书籍文件失败: " + err.Error(),
				})
				return
			}
		}
	} else {
		// 目录不存在，创建并解压
		os.MkdirAll(destDir, fs.ModePerm)
		err := unzipSource(filepath, destDir)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "解压书籍文件失败: " + err.Error(),
			})
			return
		}
	}

	// 返回指定路径的文件
	r.FileFromFS(filePath, http.Dir(destDir))
}

func (c *Api) getFile(id string) (int64, io.ReadCloser, error) {
	size, reader, err := c.contentApi.GetBook(id, "library")
	return size, reader, err
}

func (c *Api) getBookFile(r *gin.Context) {
	filesuffix := path.Ext(r.Param("id"))
	id := strings.TrimSuffix(r.Param("id"), filesuffix)

	size, reader, err := c.contentApi.GetBook(id, "library")
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	defer reader.Close()
	r.DataFromReader(http.StatusOK, size, "application/epub+zip", reader, nil)
}

func (c *Api) getCover(r *gin.Context) {
	id := strings.TrimSuffix(r.Param("id"), ".jpg")
	size, reader, err := c.contentApi.GetCover(id, "library")
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	defer reader.Close()
	r.DataFromReader(http.StatusOK, size, "image/jpeg", reader, nil)
}

func (c *Api) getFileOrCache(id string) (string, error) {
	filename := path.Join(c.baseDir, id+".epub")
	_, err := os.Stat(filename)
	if Exists(filename) {
		return filename, nil
	}
	_, closer, err := c.getFile(id)
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(closer)
	if err != nil {
		return "", err
	}
	closer.Close()

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer f.Close()
	_, err = f.Write(b)
	return filename, err
}

func (c *Api) recently(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "10"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}
	offset, err := strconv.Atoi(r.DefaultQuery("offset", "0"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
		return
	}

	// Use Qdrant GetRecent
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Search service not available",
			"code":  503,
		})
		return
	}

	books, total, err := searcher.GetRecent(limit, offset)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := make([]Book, len(books))
	for i, book := range books {
		calibreBooks[i] = Book{
			ID:           book.ID,
			Title:        book.Title,
			Authors:      book.Authors,
			Publisher:    book.Publisher,
			PubDate:      book.PubDate,
			Isbn:         book.Isbn,
			Tags:         book.Tags,
			Rating:       book.Rating,
			SeriesIndex:  book.SeriesIndex,
			Comments:     book.Comments,
			Languages:    book.Languages,
			LastModified: book.LastModified,
			Cover:        book.Cover,
			FilePath:     book.FilePath,
		}
	}

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

func (c *Api) random(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "10"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	// Use Qdrant GetRandom
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Search service not available",
			"code":  503,
		})
		return
	}

	books, err := searcher.GetRandom(limit)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := make([]Book, len(books))
	for i, book := range books {
		calibreBooks[i] = Book{
			ID:           book.ID,
			Title:        book.Title,
			Authors:      book.Authors,
			Publisher:    book.Publisher,
			PubDate:      book.PubDate,
			Isbn:         book.Isbn,
			Tags:         book.Tags,
			Rating:       book.Rating,
			SeriesIndex:  book.SeriesIndex,
			Comments:     book.Comments,
			Languages:    book.Languages,
			LastModified: book.LastModified,
			Cover:        book.Cover,
			FilePath:     book.FilePath,
		}
	}

	r.JSON(http.StatusOK, gin.H{
		"data": calibreBooks,
		"code": 200,
	})
}

// getAllBooks returns all books with cursor-based pagination (no offset needed)
func (c *Api) getAllBooks(r *gin.Context) {
	limit, err := strconv.Atoi(r.DefaultQuery("limit", "12"))
	if err != nil {
		r.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
		return
	}

	// Use cursor instead of offset
	// cursor format: "last_modified:2024-01-01T00:00:00Z,id:123"
	cursor := r.DefaultQuery("cursor", "")

	// Use Qdrant searcher
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Search service not available",
			"code":  503,
		})
		return
	}

	// Get books with cursor-based pagination
	books, total, nextCursor, err := searcher.GetAllWithCursor(limit, cursor)
	if err != nil {
		log.Warnf("GetAllWithCursor failed: %v", err)
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  500,
		})
		return
	}

	// Convert semantic.Book to calibre.Book
	calibreBooks := make([]Book, len(books))
	for i, book := range books {
		calibreBooks[i] = Book{
			ID:           book.ID,
			Title:        book.Title,
			Authors:      book.Authors,
			Publisher:    book.Publisher,
			PubDate:      book.PubDate,
			Isbn:         book.Isbn,
			Tags:         book.Tags,
			Rating:       book.Rating,
			SeriesIndex:  book.SeriesIndex,
			Comments:     book.Comments,
			Languages:    book.Languages,
			LastModified: book.LastModified,
			Cover:        book.Cover,
			FilePath:     book.FilePath,
		}
	}

	r.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"records":     calibreBooks,
			"total":       total,
			"limit":       limit,
			"next_cursor": nextCursor,
		},
		"code": 200,
	})
}

func (c *Api) updateMetadata(r *gin.Context) {
	id := r.Param("id")
	book := &Book{}
	err := r.Bind(book)
	if err != nil {
		r.JSON(http.StatusOK, gin.H{
			"code":    400,
			"data":    false,
			"message": "请求参数错误" + err.Error(),
		})
		return
	}

	oldBook, err := c.getBookByID(id)
	if err != nil {
		r.JSON(http.StatusOK, gin.H{
			"code":    500,
			"data":    false,
			"message": "元数据更新失败",
		})
		return
	}
	_, err = c.contentApi.UpdateMetaData(id, parseParams(book, oldBook), "")
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":    500,
			"data":    false,
			"message": "元数据更新失败",
		})
		return
	}

	// TODO: Update Qdrant metadata when UpdateBookMetadata is implemented
	// For now, metadata will be updated on next sync

	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"data":    true,
		"message": "元数据更新成功",
	})
}

// Task Management Handlers

func (c *Api) listTasks(r *gin.Context) {
	manager := tasks.GetManager()
	taskList := manager.GetTasks()
	r.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": taskList,
	})
}

type StartTaskRequest struct {
	Type string `json:"type"` // "qdrant_sync"
	Mode string `json:"mode"` // "full" or "incremental"
}

func (c *Api) startTask(r *gin.Context) {
	var req StartTaskRequest
	if err := r.ShouldBindJSON(&req); err != nil {
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body",
		})
		return
	}

	switch tasks.TaskType(req.Type) {
	case tasks.TaskTypeQdrantSync:
		if c.qdrantClient == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Qdrant client is not initialized",
			})
			return
		}

		searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
		if !ok || searcher == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Qdrant searcher is not initialized",
			})
			return
		}

		manager := tasks.GetManager()
		taskID, err := manager.StartTask(tasks.TaskTypeQdrantSync, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			return tasks.NewQdrantSyncTask(id, tasks.TaskMode(req.Mode), c.contentApi, c.qdrantClient, searcher)
		})

		if err != nil {
			r.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": err.Error(),
			})
			return
		}

		r.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Task started",
			"data":    gin.H{"id": taskID},
		})
		return
	default:
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Unknown task type",
		})
		return
	}
}

func (c *Api) stopTask(r *gin.Context) {
	id := r.Param("id")
	manager := tasks.GetManager()
	if err := manager.StopTask(id); err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Task stopped",
	})
}

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

func convertContentBooks(books []content.Book) ([]Book, error) {
	// Use the centralized EnrichBooks method from content package
	enrichedBooks := content.EnrichBooks(books)

	// Convert content.Book to calibre.Book
	result := make([]Book, len(enrichedBooks))
	for i, b := range enrichedBooks {
		result[i] = Book{
			AuthorSort:   b.AuthorSort,
			Authors:      b.Authors,
			Comments:     b.Comments,
			ID:           b.ID,
			Isbn:         b.Isbn,
			Languages:    b.Languages,
			LastModified: b.LastModified,
			PubDate:      b.PubDate,
			Publisher:    b.Publisher,
			SeriesIndex:  b.SeriesIndex,
			Size:         b.Size,
			Title:        b.Title,
			Tags:         b.Tags,
			Rating:       b.Rating,
			Identifiers:  b.Identifiers,
			Cover:        b.Cover,
			FilePath:     b.FilePath,
		}
	}
	return result, nil
}

func convertContentToBooks(content map[string]content.Content) ([]Book, error) {
	var books []Book
	for id, c := range content {
		i, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		book := Book{
			// Map fields from Content to Book
			AuthorSort:   c.AuthorSort,
			Authors:      c.Authors,
			Comments:     c.Comments,
			ID:           i,
			Isbn:         c.Isbn,
			Languages:    c.Languages,
			LastModified: c.LastModified,
			PubDate:      c.PubDate,
			Publisher:    c.Publisher,
			SeriesIndex:  c.SeriesIndex,
			Size:         c.Size,
			Title:        c.Title,
			Tags:         c.Tags,
			Rating:       c.Rating,
			Identifiers:  c.Identifiers,
			Cover:        "/api/get/cover/" + strconv.FormatInt(i, 10) + ".jpg",
			FilePath:     "/api/download/book/" + strconv.FormatInt(i, 10) + ".epub",
		}
		books = append(books, book)
	}
	return books, nil
}

func (c *Api) getIsbn(c2 *gin.Context) {
	isbn := c2.Param("isbn")
	var jsonData map[string]interface{}
	resp, err := c.http.R().SetResult(&jsonData).Get(c.config.Metadata.DoubanUrl + "/v2/book/isbn/" + isbn)
	log.Infof(resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}

func (c *Api) queryMetadata(c2 *gin.Context) {
	query := c2.Query("query")
	//url encode
	var jsonData map[string]interface{}
	resp, err := c.http.R().SetResult(&jsonData).SetQueryParam("q", query).Get(c.config.Metadata.DoubanUrl + "/v2/book/search")
	log.Infof(resp.Request.URL)
	if err != nil {
		c2.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c2.JSON(http.StatusOK, resp.Result())
}

func (c *Api) proxyCover(r *gin.Context) {
	path := strings.TrimPrefix(r.Param("path"), "/")
	log.Infof("proxy cover: %s", path)
	response, err := c.http.R().SetDoNotParseResponse(true).
		SetHeader("Referer", "https://book.douban.com/").
		SetHeader("Content-Type", "image/jpeg").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3573.0 Safari/537.36").
		SetQueryParamsFromValues(r.Request.URL.Query()).
		Get(path)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := response.RawResponse
	length := resp.ContentLength
	reader := resp.Body
	defer reader.Close()
	r.DataFromReader(http.StatusOK, length, "image/jpeg", reader, nil)
}

func (c *Api) listPublisher(context *gin.Context) {
	publishers, err := c.contentApi.GetAllPublisher()

	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": publishers,
	})

}

func parseParams(book *Book, oldBook *Book) map[string]interface{} {
	///cdb/delete-books/264728/library
	metadata := map[string]interface{}{}
	if book.Comments != "" {
		metadata["comments"] = book.Comments
	}
	if book.Isbn != "" {
		identifiers := oldBook.Identifiers
		identifiers["isbn"] = book.Isbn
		metadata["identifiers"] = identifiers
	}
	if book.Title != "" {
		metadata["title"] = book.Title
	}
	if book.Publisher != "" {
		metadata["publisher"] = book.Publisher
	}
	//pubdate:"2024-05-01T12:00:00+00:00"

	if !book.PubDate.IsZero() {
		metadata["pubdate"] = book.PubDate.Format("2006-01-02T15:04:05+00:00")
	}
	if book.Authors != nil {
		metadata["authors"] = book.Authors
	}
	if book.Tags != nil {
		metadata["tags"] = book.Tags
	}
	if book.Rating > 0 {
		metadata["rating"] = book.Rating
	}
	return metadata
}

// getEnhancedTools 获取增强工具列表
func (c *Api) getEnhancedTools(context *gin.Context) {
	// 创建增强工具管理器
	etm := NewEnhancedToolManager(c)
	tools := etm.GetEnhancedTools()

	context.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": tools,
	})
}

// executeEnhancedTool 执行增强工具
func (c *Api) executeEnhancedTool(context *gin.Context) {
	toolName := context.Param("tool")

	// 解析请求参数
	var args map[string]interface{}
	if err := context.ShouldBindJSON(&args); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	// 创建增强工具管理器
	etm := NewEnhancedToolManager(c)

	// 执行工具
	result, err := etm.ExecuteEnhancedTool(toolName, args)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "工具执行失败: " + err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}
