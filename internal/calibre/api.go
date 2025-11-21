package calibre

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/kapmahc/epub"
	"github.com/meilisearch/meilisearch-go"
	"github.com/spf13/cast"

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
	semanticSearcher interface{} // Can be *qdrant.Searcher or nil
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
	base.POST("/index/update", c.updateIndex)
	base.POST("/index/switch", c.switchIndex)

	// Enhanced Tools MCP 端点
	base.GET("/mcp/tools/enhanced", c.getEnhancedTools)
	base.POST("/mcp/tools/enhanced/:tool", c.executeEnhancedTool)

	// Semantic Search Endpoints
	base.POST("/search/semantic/index", c.semanticIndexStart)
	base.POST("/search/semantic/index/stop", c.semanticIndexStop)
	base.GET("/search/semantic/index/status", c.semanticIndexStatus)
	base.GET("/search/semantic", c.semanticSearch)

	// Task Management Endpoints
	base.GET("/tasks", c.listTasks)
	base.POST("/tasks/start", c.startTask)
	base.POST("/tasks/:id/stop", c.stopTask)
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
		providerConfig := embedding.ProviderConfig{
			Provider:    config.Embedding.Provider,
			Ollama:      config.Embedding.Ollama,
			SiliconFlow: config.Embedding.SiliconFlow,
			VectorDim:   4096, // Qdrant uses 4096 dimensions
		}

		provider, err := embedding.NewProvider(providerConfig)
		if err != nil {
			log.Warnf("Failed to create embedding provider: %v", err)
		} else {
			qdrantSearcher = qdrant.NewSearcher(provider, qdrantClient)
			log.Info("Qdrant searcher initialized successfully")
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

// ensureIndexExists checks if a Meilisearch index exists, and if not, creates it and updates its settings.
//
// Parameters:
// - client: A pointer to the Meilisearch client.
// - indexName: The name of the index to check or create.
//
// Returns:
// - A pointer to the Meilisearch index.
// - An error if the index creation or settings update fails.
func ensureIndexExists(client *meilisearch.Client, indexName string) (*meilisearch.Index, error) {
	index := client.Index(indexName)

	// Fetch index information to check if it exists
	log.Infof("Checking if index %q exists", indexName)
	_, err := index.FetchInfo()
	if err != nil {
		log.Infof("Failed to fetch index info for %q: %v", indexName, err)
		// Index does not exist, create it
		log.Infof("Creating index %q", indexName)
		_, err = client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        indexName,
			PrimaryKey: "id",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create index: %w", err)
		}
		log.Infof("Index %q created", indexName)
		// Update index settings
		log.Infof("Updating index settings for %q", indexName)
		_, err = index.UpdateSettings(&meilisearch.Settings{
			//RankingRules:         []string{"typo", "words", "proximity", "attribute", "exactness"},
			DisplayedAttributes:  []string{"*"},
			FilterableAttributes: []string{"authors", "file_path", "id", "last_modified", "pubdate", "publisher", "isbn", "tags"},
			SearchableAttributes: []string{"title", "authors", "isbn", "publisher"},
			SortableAttributes:   []string{"authors_sort", "id", "last_modified", "pubdate", "publisher"},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update index settings: %w", err)
		}
	}
	return index, nil
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

	log.Infof("Qdrant search query: %s, filter: %s, limit: %d, offset: %d", q, filterType, limit, offset)

	// Check if Qdrant searcher is available
	searcher, ok := c.semanticSearcher.(*qdrant.Searcher)
	if !ok || searcher == nil {
		log.Error("Qdrant searcher not available")
		r.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search service not available",
			"code":  500,
		})
		return
	}

	// Perform keyword search using Qdrant
	books, total, err := searcher.SearchByKeyword(q, filterType, limit, offset)
	if err != nil {
		log.Errorf("Qdrant search failed: %v", err)
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
			Pubdate:      book.Pubdate,
			ISBN:         book.ISBN,
			Tags:         book.Tags,
			Rating:       book.Rating,
			SeriesIndex:  book.SeriesIndex,
			Comments:     book.Comments,
			Languages:    book.Languages,
			LastModified: book.LastModified,
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
	var book Book
	err := c.currentIndex().GetDocument(id, nil, &book)

	if err != nil {
		// 返回文件找不到
		r.JSON(http.StatusOK, gin.H{
			"message": "book not found" + err.Error(),
			"code":    http.StatusNotFound,
		})
		return
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
	_, err = c.currentIndex().DeleteDocument(id)
	if err != nil {
		// 返回文件找不到
		r.JSON(http.StatusOK, gin.H{
			"message": "book not found" + err.Error(),
			"code":    http.StatusNotFound,
		})
		return
	}
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
	var book Book
	err := c.currentIndex().GetDocument(id, nil, &book)
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

	var book Book
	err := c.currentIndex().GetDocument(id, nil, &book)
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

func (c *Api) switchIndex(c2 *gin.Context) {

	resp, err := c.client.GetTasks(&meilisearch.TasksQuery{
		Limit:     2,
		Statuses:  []string{"enqueued", "processing"},
		IndexUIDS: []string{c.config.Search.Index, c.config.Search.Index + "-bak"},
	})
	if err != nil {
		log.Warn(err)
		c2.JSON(http.StatusOK, gin.H{"code": 500, "error": err.Error()})
		return
	}
	if len(resp.Results) != 0 {
		log.Warn(err)
		c2.JSON(http.StatusOK, gin.H{"code": 400, "error": "有任务正在执行，请稍后再试"})
		return
	}

	targetIndex := ""
	if c.useIndex == c.config.Search.Index {
		targetIndex = c.config.Search.Index + "-bak"
	} else {
		targetIndex = c.config.Search.Index
	}

	// Check if target index exists and has documents
	stats, err := c.client.Index(targetIndex).GetStats()
	if err != nil {
		log.Warnf("Failed to get stats for target index %s: %v", targetIndex, err)
		c2.JSON(http.StatusOK, gin.H{"code": 400, "error": fmt.Sprintf("无法切换：目标索引 '%s' 不存在或无法访问", targetIndex)})
		return
	}

	if stats.NumberOfDocuments == 0 {
		c2.JSON(http.StatusOK, gin.H{"code": 400, "error": fmt.Sprintf("无法切换：目标索引 '%s' 为空，请先执行全量重建", targetIndex)})
		return
	}

	c.useIndex = targetIndex
	c2.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"current_index": c.useIndex,
		},
	})
}
func (c *Api) updateIndex(c2 *gin.Context) {
	booksIds, err2 := c.contentApi.GetAllBooksIds("")
	if err2 != nil {
		log.Warn(err2)
		c2.JSON(http.StatusOK, gin.H{"code": 500, "error": err2.Error()})
		return
	}
	index := c.currentIndex()
	if c.useIndex == c.config.Search.Index {
		index = c.client.Index(c.config.Search.Index + "-bak")
	} else {
		index = c.client.Index(c.config.Search.Index)
	}
	_, err := index.DeleteAllDocuments()
	if err != nil {
		log.Warn(err)
		c2.JSON(http.StatusOK, gin.H{"code": 500, "error": err.Error()})
		return
	}

	// 按 2000 分段 booksIds,查询书籍，更新索引
	var books []Book
	var taskIds []int64
	for i := 0; i < len(booksIds); i += 2000 {
		ids := booksIds[i:min(i+2000, len(booksIds))]
		log.Infof("update index %d [%d - %d]", i, ids[0], ids[len(ids)-1])

		data, err := c.contentApi.GetBookMetaDatas(ids, "")
		if err != nil {
			log.Warnf("get book metadata error: %v", err)
			c2.JSON(http.StatusOK, gin.H{"code": 500, "error": "get book metadata error: " + err.Error()})
			return
		}
		books, err = convertContentBooks(data)
		if err != nil {
			c2.JSON(http.StatusOK, gin.H{"code": 500, "error": err.Error()})
			return
		}
		task, err := index.AddDocumentsInBatches(books, len(ids))
		if err != nil {
			c2.JSON(http.StatusOK, gin.H{"code": 500, "error": err.Error()})
			return
		}
		for _, info := range task {
			taskIds = append(taskIds, info.TaskUID)
		}
	}

	err = waitForTask(c, taskIds)
	if err != nil {
		c2.JSON(http.StatusOK, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c2.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    len(booksIds),
	})
}

func waitForTask(c *Api, taskIds []int64) error {
	timeout := time.After(30 * time.Second) // Set timeout duration
	done := make(chan bool)

	go func() {
		for {
			resp, err := c.client.GetTasks(&meilisearch.TasksQuery{
				Limit:    2,
				Statuses: []string{"enqueued", "processing"},
				UIDS:     taskIds,
			})
			if err != nil {
				log.Warn(err)
				done <- true
				return
			}
			if len(resp.Results) == 0 {
				done <- true
				return
			}
			time.Sleep(3 * time.Second) // Add a small delay to avoid tight loop
		}
	}()

	select {
	case <-timeout:
		log.Warn("Timeout reached while waiting for tasks to complete")
		return fmt.Errorf("Timeout reached while waiting for tasks to complete")
	case <-done:
		log.Info("Tasks completed successfully")
		if c.useIndex == c.config.Search.Index {
			c.useIndex = c.config.Search.Index + "-bak"
		} else {
			c.useIndex = c.config.Search.Index
		}
		return nil
	}
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

	searchRequest := meilisearch.SearchRequest{
		Sort:   []string{"id:desc"},
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	c.currentIndex().Search("", &searchRequest)

	search, err := c.currentIndex().Search("", &searchRequest)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	books := make([]Book, len(search.Hits))
	for i := range search.Hits {
		tmp := search.Hits[i].(map[string]interface{})
		jsonb, err := json.Marshal(tmp)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		book := Book{}
		if err := json.Unmarshal(jsonb, &book); err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books[i] = book
	}

	r.JSON(http.StatusOK, gin.H{
		"data": map[string]interface{}{
			"records": &books,
			"total":   search.EstimatedTotalHits,
			"limit":   search.Limit,
			"offset":  search.Offset,
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

	RandomInt := func(min, max int) int {
		return min + rand.Intn(max-min)
	}
	offset := RandomInt(0, 1000)

	// 使用随机排序
	searchRequest := meilisearch.SearchRequest{
		Limit:  int64(limit),
		Offset: int64(offset),
	}

	c.currentIndex().Search("", &searchRequest)
	search, err := c.currentIndex().Search("", &searchRequest)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	books := make([]Book, len(search.Hits))
	for i := range search.Hits {
		tmp := search.Hits[i].(map[string]interface{})
		jsonb, err := json.Marshal(tmp)
		if err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		book := Book{}
		if err := json.Unmarshal(jsonb, &book); err != nil {
			r.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		books[i] = book
	}

	r.JSON(http.StatusOK, gin.H{
		"data": &books,
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

	oldBook := &Book{}
	err = c.currentIndex().GetDocument(id, nil, oldBook)
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

	data, err := c.contentApi.GetBookMetaDatas([]int64{cast.ToInt64(id)}, "")
	if err != nil {
		log.Warnf("get book metadata error: %v", err)
		r.JSON(http.StatusOK, gin.H{
			"code":    500,
			"data":    false,
			"message": "元数据更新成功，但是查询元数据失败",
		})
		return
	}
	books, err := convertContentBooks(data)
	if err != nil {
		r.JSON(http.StatusOK, gin.H{
			"code":    500,
			"data":    false,
			"message": "书籍元数据翻译失败，请刷新索引",
		})
		return
	}
	_, err = c.currentIndex().AddDocuments(books)
	if err != nil {
		// 返回文件找不到
		r.JSON(http.StatusOK, gin.H{
			"code":    500,
			"data":    false,
			"message": "元数据更新成功，但是索引更新失败，请刷新索引",
		})
		return
	}
	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    &books[0],
	})

}

func (c *Api) semanticIndexStart(r *gin.Context) {
	if c.semanticIndexer == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Semantic search is not initialized",
		})
		return
	}

	if err := c.semanticIndexer.Start(); err != nil {
		r.JSON(http.StatusConflict, gin.H{
			"code":    409,
			"message": err.Error(),
		})
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Semantic indexing started",
	})
}

func (c *Api) semanticIndexStop(r *gin.Context) {
	if c.semanticIndexer == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Semantic search is not initialized",
		})
		return
	}

	c.semanticIndexer.Stop()
	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Semantic indexing stopped",
	})
}

func (c *Api) semanticIndexStatus(r *gin.Context) {
	if c.semanticIndexer == nil {
		r.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Semantic search is not initialized",
		})
		return
	}

	status := c.semanticIndexer.GetStatus()
	r.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": status,
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
	Type string `json:"type"` // "meilisearch_sync" or "vector_sync"
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

	manager := tasks.GetManager()
	var taskID string
	var err error

	switch tasks.TaskType(req.Type) {
	case tasks.TaskTypeMeilisearchSync:
		targetIndex := c.useIndex
		// If Full Sync, target the inactive index (backup)
		if tasks.TaskMode(req.Mode) == tasks.TaskModeFull {
			if c.useIndex == c.config.Search.Index {
				targetIndex = c.config.Search.Index + "-bak"
			} else {
				targetIndex = c.config.Search.Index
			}
		}

		taskID, err = manager.StartTask(tasks.TaskTypeMeilisearchSync, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			return tasks.NewMeilisearchTask(id, tasks.TaskMode(req.Mode), c.contentApi, c.client, targetIndex)
		})
	case tasks.TaskTypeVectorSync:
		if c.semanticIndexer == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Semantic search is not initialized",
			})
			return
		}
		taskID, err = manager.StartTask(tasks.TaskTypeVectorSync, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			return tasks.NewVectorTask(id, tasks.TaskMode(req.Mode), c.semanticIndexer)
		})
	case tasks.TaskTypeQdrantMigration:
		if c.milvusClient == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Milvus client is not initialized",
			})
			return
		}
		if c.qdrantClient == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Qdrant client is not initialized",
			})
			return
		}
		taskID, err = manager.StartTask(tasks.TaskTypeQdrantMigration, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			return tasks.NewQdrantMigrationTask(c.milvusClient, c.qdrantClient, c.contentApi)
		})
	default:
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Unknown task type",
		})
		return
	}

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

	results, err := c.semanticSearcher.Search(q, limit)
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Search failed: " + err.Error(),
		})
		return
	}

	// Extract book IDs from semantic search results
	bookIDs := make([]int64, len(results))
	scoreMap := make(map[int64]float32)
	rankMap := make(map[int64]int)

	for i, result := range results {
		bookIDs[i] = result.Book.ID
		scoreMap[result.Book.ID] = result.Score
		rankMap[result.Book.ID] = result.Rank
	}

	// Fetch full book data from Meilisearch
	if len(bookIDs) == 0 {
		r.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": []interface{}{},
		})
		return
	}

	// Build filter for Meilisearch: id IN [id1, id2, id3]
	idFilters := make([]string, len(bookIDs))
	for i, id := range bookIDs {
		idFilters[i] = fmt.Sprintf("id = %d", id)
	}
	filter := strings.Join(idFilters, " OR ")

	// Search in Meilisearch with filter
	searchReq := &meilisearch.SearchRequest{
		Limit:  int64(limit),
		Filter: filter,
	}

	searchResp, err := c.currentIndex().Search("", searchReq)
	if err != nil {
		log.Warnf("Failed to fetch books from Meilisearch: %v", err)
		// Fallback to basic results from Milvus
		r.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": results,
		})
		return
	}

	// Convert Meilisearch results to Books
	books := make([]Book, 0, len(searchResp.Hits))
	for _, hit := range searchResp.Hits {
		// Convert hit (map[string]interface{}) to Book
		hitBytes, err := json.Marshal(hit)
		if err != nil {
			continue
		}

		var book Book
		if err := json.Unmarshal(hitBytes, &book); err != nil {
			continue
		}

		books = append(books, book)
	}

	// Sort by rank to maintain semantic search order
	sort.Slice(books, func(i, j int) bool {
		rankI := rankMap[books[i].ID]
		rankJ := rankMap[books[j].ID]
		return rankI < rankJ
	})

	// Return in the same format as keyword search
	r.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"records": books,
			"total":   len(books),
			"limit":   limit,
			"offset":  0,
		},
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
