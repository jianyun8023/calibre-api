package calibre

import (
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/chat"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
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
	qdrantClient     *qdrant.Client
	semanticSearcher interface{} // *qdrant.Searcher
	cachedSearcher   *cache.CachedSearcher // 带缓存的搜索器
	cacheManager     *cache.Manager
	chatDB           *chat.DB
	chatAgent        *chat.Agent
	sseManager       *tasks.SSEManager
	bookHandler      *BookHandlerV2 // 新的 Handler（使用 Service 层）
	metricsHandler   *MetricsHandler // 性能指标处理器
}

// InjectDependencies 注入依赖（用于依赖注入容器）
func (a *Api) InjectDependencies(
	config *Config,
	contentApi *content.Api,
	qdrantClient *qdrant.Client,
	qdrantSearcher *qdrant.Searcher,
	cachedSearcher *cache.CachedSearcher,
	cacheManager *cache.Manager,
	chatDB *chat.DB,
	sseManager *tasks.SSEManager,
	bookHandler *BookHandlerV2,
) error {
	a.config = config
	a.contentApi = contentApi
	a.baseDir = config.TmpDir
	a.http = contentApi.Client
	a.qdrantClient = qdrantClient
	a.semanticSearcher = qdrantSearcher
	a.cachedSearcher = cachedSearcher
	a.cacheManager = cacheManager
	a.chatDB = chatDB
	a.sseManager = sseManager
	a.bookHandler = bookHandler

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

// InjectChatAgent 注入聊天 Agent（需要在 API 实例创建后）
func (a *Api) InjectChatAgent(agent *chat.Agent) error {
	a.chatAgent = agent
	return nil
}

// CreateTocFetcher 创建 TOC 获取函数（用于 Chat Agent）
func (a *Api) CreateTocFetcher() func(ctx context.Context, bookID int64) (string, error) {
	return func(ctx context.Context, bookID int64) (string, error) {
		tocData, err := a.GetBookTocData(strconv.FormatInt(bookID, 10))
		if err != nil {
			return "", err
		}
		// 将 tocData 转换为 JSON 字符串
		bytes, err := json.MarshalIndent(tocData, "", "  ")
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	}
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

	// 搜索相关
	base.GET("/search", c.search)
	base.POST("/search", c.search)
	base.GET("/search/semantic", c.semanticSearch)

	// 元数据相关
	base.GET("/metadata/isbn/:isbn", c.getIsbn)
	base.GET("/metadata/search", c.queryMetadata)

	// 任务管理
	base.GET("/tasks", c.listTasks)
	base.POST("/tasks/start", c.startTask)
	base.POST("/tasks/:id/stop", c.stopTask)
	base.GET("/tasks/stream", c.streamTasks) // SSE 任务流

	// Chat routes (智能问答)
	base.POST("/chat/conversations", c.CreateConversation)
	base.GET("/chat/conversations", c.ListConversations)
	base.GET("/chat/conversations/:id", c.GetConversation)
	base.GET("/chat/conversations/:id/messages", c.GetConversationMessages)
	base.DELETE("/chat/conversations/:id", c.DeleteConversation)
	base.DELETE("/chat/messages/:id", c.DeleteMessage)
	base.POST("/chat/conversations/:id/messages", c.SendMessage)
	base.POST("/chat/stream", c.ChatStream) // AI SDK stream endpoint

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

	// Initialize chat components
	var chatDB *chat.DB
	var chatAgent *chat.Agent

	if config.Chat.DBPath != "" {
		chatDB, err = chat.NewDB(config.Chat.DBPath)
		if err != nil {
			log.Warnf("Failed to initialize chat database: %v", err)
		} else {
			log.Infof("Chat database initialized: %s", config.Chat.DBPath)
		}
	}

	// Create Api instance first (without chatAgent)
	api := Api{
		config:           config,
		baseDir:          config.TmpDir,
		contentApi:       &newClient,
		http:             newClient.Client,
		qdrantClient:     qdrantClient,
		semanticSearcher: qdrantSearcher,
		cacheManager:     cacheManager,
		chatDB:           chatDB,
	}

	// Initialize LLM and Agent if chat DB is available and Qdrant searcher exists
	if chatDB != nil && qdrantSearcher != nil && config.LLM.Provider != "" {
		llm, err := chat.NewLLM(config.LLM)
		if err != nil {
			log.Warnf("Failed to initialize LLM client: %v", err)
		} else {
			// Define TocFetcher using the api instance
			tocFetcher := func(ctx context.Context, bookID int64) (string, error) {
				tocData, err := api.GetBookTocData(strconv.FormatInt(bookID, 10))
				if err != nil {
					return "", err
				}
				// Convert tocData to string summary
				// tocData is map[string]interface{} or similar structure
				// We need to format it nicely for the LLM
				bytes, err := json.MarshalIndent(tocData, "", "  ")
				if err != nil {
					return "", err
				}
				return string(bytes), nil
			}

			chatAgent = chat.NewAgent(llm, qdrantSearcher, tocFetcher)
			api.chatAgent = chatAgent // Inject agent back to api
			log.Infof("Chat agent initialized with provider: %s", config.LLM.Provider)
		}
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
