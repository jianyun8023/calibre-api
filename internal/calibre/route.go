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
	cacheManager     *cache.Manager
	chatDB           *chat.DB
	chatAgent        *chat.Agent
}

// SetupRouter 设置路由
func (c *Api) SetupRouter(r *gin.Engine) {
	base := r.Group("/api")

	// 书籍基本操作
	base.GET("/book/:id", c.getBook)
	base.POST("/book/:id/delete", c.deleteBook)
	base.POST("/book/:id/update", c.updateMetadata)
	base.GET("/recently", c.recently)
	base.GET("/random", c.random)
	base.GET("/books/all", c.getAllBooks)
	base.GET("/publisher", c.listPublisher)

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

	// Enhanced Tools MCP 端点
	base.GET("/mcp/tools/enhanced", c.getEnhancedTools)
	base.POST("/mcp/tools/enhanced/:tool", c.executeEnhancedTool)

	// Chat routes (智能问答)
	base.POST("/chat/conversations", c.CreateConversation)
	base.GET("/chat/conversations", c.ListConversations)
	base.GET("/chat/conversations/:id", c.GetConversation)
	base.GET("/chat/conversations/:id/messages", c.GetConversationMessages)
	base.DELETE("/chat/conversations/:id", c.DeleteConversation)
	base.DELETE("/chat/messages/:id", c.DeleteMessage)
	base.POST("/chat/conversations/:id/messages", c.SendMessage)
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

	return &api
}
