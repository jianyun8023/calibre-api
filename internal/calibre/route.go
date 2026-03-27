package calibre

import (
	"context"
	"io/fs"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/middleware"
	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/meilisearch"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/client"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// Api Calibre API 服务
type Api struct {
	config           *Config
	contentApi       *content.Api
	baseDir          string
	http             *client.Client
	meiliClient      *meilisearch.Client
	semanticSearcher semantic.Searcher
	cachedSearcher   *cache.CachedSearcher // 带缓存的搜索器
	cacheManager     *cache.Manager
	sseManager       *tasks.SSEManager
	bookHandler      *BookHandlerV2  // 新的 Handler（使用 Service 层）
	metricsHandler   *MetricsHandler // 性能指标处理器
	draftHandler     *DraftHandler   // 草稿处理器
}

// InjectDependencies 注入依赖（用于依赖注入容器）
func (a *Api) InjectDependencies(
	config *Config,
	contentApi *content.Api,
	_ interface{}, // deprecated qdrantClient parameter, kept for compatibility
	searcher semantic.Searcher,
	cachedSearcher *cache.CachedSearcher,
	cacheManager *cache.Manager,
	sseManager *tasks.SSEManager,
	bookHandler *BookHandlerV2,
	draftHandler *DraftHandler,
) error {
	a.config = config
	a.contentApi = contentApi
	a.baseDir = config.TmpDir
	a.http = contentApi.Client
	a.semanticSearcher = searcher
	a.cachedSearcher = cachedSearcher
	a.cacheManager = cacheManager
	a.sseManager = sseManager
	a.bookHandler = bookHandler
	a.draftHandler = draftHandler

	// 创建性能指标处理器
	if cachedSearcher != nil {
		a.metricsHandler = NewMetricsHandler(cachedSearcher)
	}

	// 确保 baseDir 存在
	if !Exists(a.baseDir) {
		if err := os.MkdirAll(a.baseDir, fs.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// SetupRouter 设置路由
func (c *Api) SetupRouter(r *gin.Engine) {
	base := r.Group("/api")

	// 书籍基本操作（使用新的 Service 层）
	base.GET("/book/:id", c.getBookV2)
	base.POST("/book/:id/delete", c.deleteBookV2)
	base.POST("/book/:id/update", c.updateMetadataV2)
	base.GET("/recently", c.recentlyV2)
	base.GET("/random", c.randomV2)
	base.GET("/books/all", c.getAllBooksV2)
	base.GET("/publisher", c.listPublisherV2)

	// 书籍内容相关
	base.GET("/get/cover/:id", c.getCover)
	base.GET("/proxy/cover/*path", c.proxyCover)
	base.GET("/download/book/:id", c.getBookFile)
	base.GET("/read/:id/toc", c.getBookToc)
	base.GET("/read/:id/file/*path", c.getBookContent)
	base.GET("/book/content", c.getBookContentByQuery)
	base.POST("/book/:id/extract-metadata", c.extractMetadata)

	// 搜索相关
	base.GET("/search", c.search)
	base.POST("/search", c.search)
	base.GET("/search/semantic", c.semanticSearch)

	// 元数据相关
	base.GET("/metadata/isbn/:isbn", c.getIsbn)
	base.GET("/metadata/search", c.queryMetadata)

	// 任务管理
	base.GET("/tasks", c.listTasks)
	base.GET("/tasks/:id", c.getTask)
	base.POST("/tasks/start", c.startTask)
	base.POST("/tasks/:id/stop", c.stopTask)
	base.GET("/tasks/stream", c.streamTasks) // SSE 任务流

	// 草稿管理
	if c.draftHandler != nil {
		// Public endpoints for external systems (if no global auth is needed, otherwise wrap them)
		// Assuming we want to protect all draft modification endpoints:
		draftGroup := base.Group("/drafts")
		draftGroup.Use(middleware.APIKeyAuth(c.config.Auth.APIKey))
		{
			draftGroup.POST("/delete", c.draftHandler.ReceiveDeletes)
			draftGroup.POST("/update", c.draftHandler.ReceiveUpdates)
			draftGroup.GET("", c.draftHandler.GetPendingDrafts)
			draftGroup.POST("/apply", c.draftHandler.ApplyDrafts)
			draftGroup.POST("/reject", c.draftHandler.RejectDrafts)
			draftGroup.GET("/history", c.draftHandler.GetHistory)
		}
	}

	// 性能监控和指标
	if c.metricsHandler != nil {
		base.GET("/metrics", c.metricsHandler.GetMetrics)
		base.GET("/metrics/cache", c.metricsHandler.GetCacheStats)
		base.POST("/metrics/cache/clear", c.metricsHandler.ClearCache)
		base.POST("/metrics/cache/clean", c.metricsHandler.CleanExpiredCache)
		base.GET("/health", c.metricsHandler.GetHealthCheck)
	}
}

// NewClient 创建 Calibre API 客户端
func NewClient(config *Config) *Api {
	baseDir := config.TmpDir
	if !Exists(baseDir) {
		os.MkdirAll(baseDir, fs.ModePerm)
	}

	newClient, err := content.NewClient(config.Content.Server)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Searcher (MeiliSearch only)
	var meiliClient *meilisearch.Client
	var semanticSearcher semantic.Searcher

	// Initialize embedding provider for semantic search
	providerConfig := embedding.ProviderConfig{
		Provider:    config.Embedding.Provider,
		Ollama:      config.Embedding.Ollama,
		SiliconFlow: config.Embedding.SiliconFlow,
		VectorDim:   4096, // Default to 4096 if model uses it (Qwen/DeepSeek often use 1024-4096)
	}

	provider, err := embedding.NewProvider(providerConfig)
	if err != nil {
		log.Warnf("Failed to create embedding provider: %v. Semantic search capabilities will be limited.", err)
	}

	// Check if Meilisearch is configured
	if config.Meilisearch.Host != "" {
		log.Info("Initializing Meilisearch client...")
		meiliClient = meilisearch.NewClient(
			config.Meilisearch.Host,
			config.Meilisearch.APIKey,
			config.Meilisearch.IndexName,
		)
		// Pass provider to Meilisearch searcher
		semanticSearcher = meilisearch.NewSearcher(meiliClient, provider)
		log.Info("Meilisearch searcher initialized successfully")

		// Ensure indexes
		ctx := context.Background()
		if err := semanticSearcher.EnsureIndexes(ctx); err != nil {
			log.Warnf("Failed to ensure Meilisearch indexes: %v", err)
		} else {
			log.Info("Meilisearch indexes ensured successfully")
		}

	} else {
		log.Warn("MeiliSearch is not configured. Search functionality will be unavailable.")
	}

	// Initialize cache manager
	var cacheManager *cache.Manager
	if config.Cache.Dir != "" {
		cacheManager, err = cache.NewManager(config.Cache, &newClient)
		if err != nil {
			log.Warnf("Failed to initialize cache manager: %v", err)
		} else {
			log.Infof("Cache manager initialized: dir=%s, max_size=%.1fGB",
				config.Cache.Dir, config.Cache.MaxSizeGB)
		}
	}

	// Create Api instance
	api := Api{
		config:           config,
		baseDir:          config.TmpDir,
		contentApi:       &newClient,
		http:             newClient.Client,
		meiliClient:      meiliClient,
		semanticSearcher: semanticSearcher,
		cacheManager:     cacheManager,
	}

	// 初始化 SSE MCP 服务器（在 HTTP 模式下默认启用）
	// 注意：这里不依赖 config.MCP.Enabled，因为那个控制的是传统 MCP 模式
	baseURL := config.MCP.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	log.Infof("SSE MCP Server initialized with base URL: %s", baseURL)

	// 初始化 SSE 管理器并连接到任务管理器
	taskManager := tasks.GetManager()
	api.sseManager = tasks.NewSSEManager(taskManager, 100) // 最大 100 个并发连接
	taskManager.SetSSEManager(api.sseManager)              // 双向连接
	log.Info("SSE Manager initialized and connected to Task Manager for real-time updates")

	return &api
}
