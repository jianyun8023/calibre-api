package container

import (
	"context"
	"fmt"
	"time"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/internal/chat"
	"github.com/jianyun8023/calibre-api/internal/config"
	"github.com/jianyun8023/calibre-api/internal/governance"
	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/jianyun8023/calibre-api/internal/service"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/tmc/langchaingo/llms"
)

// Container 依赖注入容器
// 管理应用程序的所有依赖组件
type Container struct {
	configManager *config.Manager
	config        *calibre.Config

	// Core components
	contentAPI   *content.Api
	qdrantClient *qdrant.Client
	cacheManager *cache.Manager
	taskManager  *tasks.Manager
	sseManager   *tasks.SSEManager

	// Search components
	qdrantSearcher *qdrant.Searcher
	cachedSearcher *cache.CachedSearcher

	// Chat components
	chatDB    *chat.DB
	chatLLM   llms.Model
	chatAgent *chat.Agent

	// Governance components
	governanceDB      *governance.DB
	governanceService *governance.Service
	governanceHandler *governance.Handler

	// Repository layer
	bookRepo repository.BookRepository

	// Service layer
	bookService service.BookService
	bookHandler *calibre.BookHandlerV2

	// API instance
	api *calibre.Api
}

// NewContainer 创建新的依赖注入容器
func NewContainer(config *calibre.Config) (*Container, error) {
	c := &Container{
		config: config,
	}

	// 按依赖顺序初始化组件
	if err := c.initContentAPI(); err != nil {
		return nil, fmt.Errorf("failed to initialize content API: %w", err)
	}

	if err := c.initQdrant(); err != nil {
		log.Warnf("Qdrant initialization failed (optional): %v", err)
	}

	if err := c.initCache(); err != nil {
		log.Warnf("Cache initialization failed (optional): %v", err)
	}

	if err := c.initChat(); err != nil {
		log.Warnf("Chat initialization failed (optional): %v", err)
	}

	if err := c.initTasks(); err != nil {
		return nil, fmt.Errorf("failed to initialize task manager: %w", err)
	}

	if err := c.initGovernance(); err != nil {
		log.Warnf("Governance initialization failed (optional): %v", err)
	}

	return c, nil
}

// NewContainerWithConfigManager 使用配置管理器创建容器（支持热重载）
func NewContainerWithConfigManager() (*Container, error) {
	// 创建配置管理器
	configManager, err := config.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create config manager: %w", err)
	}

	c := &Container{
		configManager: configManager,
		config:        configManager.GetConfig(),
	}

	// 按依赖顺序初始化组件
	if err := c.initContentAPI(); err != nil {
		return nil, fmt.Errorf("failed to initialize content API: %w", err)
	}

	if err := c.initQdrant(); err != nil {
		log.Warnf("Qdrant initialization failed (optional): %v", err)
	}

	if err := c.initCache(); err != nil {
		log.Warnf("Cache initialization failed (optional): %v", err)
	}

	if err := c.initChat(); err != nil {
		log.Warnf("Chat initialization failed (optional): %v", err)
	}

	if err := c.initTasks(); err != nil {
		return nil, fmt.Errorf("failed to initialize task manager: %w", err)
	}

	if err := c.initGovernance(); err != nil {
		log.Warnf("Governance initialization failed (optional): %v", err)
	}

	return c, nil
}

// initContentAPI 初始化 Calibre Content Server 客户端
func (c *Container) initContentAPI() error {
	if c.config.Content.Server == "" {
		return fmt.Errorf("content server URL is required")
	}

	client, err := content.NewClient(c.config.Content.Server)
	if err != nil {
		return fmt.Errorf("failed to create content client: %w", err)
	}

	c.contentAPI = &client
	log.Infof("Content API initialized: %s", c.config.Content.Server)
	return nil
}

// initQdrant 初始化 Qdrant 向量数据库客户端和搜索器
func (c *Container) initQdrant() error {
	if c.config.Qdrant.URL == "" {
		return fmt.Errorf("Qdrant URL not configured (skipping)")
	}

	// 创建 Qdrant 客户端
	c.qdrantClient = qdrant.NewClient(
		c.config.Qdrant.URL,
		c.config.Qdrant.Collection,
		time.Duration(c.config.Qdrant.Timeout)*time.Second,
	)
	log.Infof("Qdrant client initialized: %s/%s", c.config.Qdrant.URL, c.config.Qdrant.Collection)

	// 初始化 Embedding Provider
	providerConfig := embedding.ProviderConfig{
		Provider:    c.config.Embedding.Provider,
		Ollama:      c.config.Embedding.Ollama,
		SiliconFlow: c.config.Embedding.SiliconFlow,
		VectorDim:   4096, // Qdrant 使用 4096 维度
	}

	provider, err := embedding.NewProvider(providerConfig)
	if err != nil {
		return fmt.Errorf("failed to create embedding provider: %w", err)
	}

	// 创建搜索器
	c.qdrantSearcher = qdrant.NewSearcher(provider, c.qdrantClient)
	log.Infof("Qdrant searcher initialized with provider: %s", c.config.Embedding.Provider)

	// 包装搜索器添加缓存功能
	cacheMaxSize := 1000
	cacheTTL := 300 // 5 分钟
	c.cachedSearcher = cache.WrapSearcher(c.qdrantSearcher, cacheMaxSize, cacheTTL)
	log.Infof("Search cache initialized: maxSize=%d, ttl=%ds", cacheMaxSize, cacheTTL)

	// 确保索引存在
	ctx := context.Background()
	if err := c.qdrantSearcher.EnsureIndexes(ctx); err != nil {
		log.Warnf("Failed to ensure Qdrant indexes (may already exist): %v", err)
	} else {
		log.Info("Qdrant payload indexes verified")
	}

	return nil
}

// initCache 初始化缓存管理器
func (c *Container) initCache() error {
	if c.config.Cache.Dir == "" {
		return fmt.Errorf("cache directory not configured (skipping)")
	}

	if c.contentAPI == nil {
		return fmt.Errorf("content API not initialized")
	}

	manager, err := cache.NewManager(c.config.Cache, c.contentAPI)
	if err != nil {
		return fmt.Errorf("failed to create cache manager: %w", err)
	}

	c.cacheManager = manager
	log.Infof("Cache manager initialized: dir=%s, max_size=%.1fGB",
		c.config.Cache.Dir, c.config.Cache.MaxSizeGB)
	return nil
}

// initChat 初始化聊天相关组件（数据库、LLM、Agent）
func (c *Container) initChat() error {
	// 初始化聊天数据库
	if c.config.Chat.DBPath == "" {
		return fmt.Errorf("chat database path not configured (skipping)")
	}

	db, err := chat.NewDB(c.config.Chat.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize chat database: %w", err)
	}
	c.chatDB = db
	log.Infof("Chat database initialized: %s", c.config.Chat.DBPath)

	// 初始化 LLM（需要 Qdrant 搜索器）
	if c.config.LLM.Provider == "" {
		log.Info("LLM provider not configured, chat agent will not be available")
		return nil
	}

	if c.qdrantSearcher == nil {
		log.Warn("Qdrant searcher not available, chat agent will not be initialized")
		return nil
	}

	llm, err := chat.NewLLM(c.config.LLM)
	if err != nil {
		return fmt.Errorf("failed to initialize LLM: %w", err)
	}
	c.chatLLM = llm
	log.Infof("LLM initialized with provider: %s", c.config.LLM.Provider)

	// 注意：chatAgent 需要在 API 实例创建后初始化
	// 因为它需要 TocFetcher 函数，该函数依赖于 API 实例
	// 这将在 BuildAPI() 方法中完成

	return nil
}

// initTasks 初始化任务管理器和 SSE 管理器
func (c *Container) initTasks() error {
	// 获取全局任务管理器实例
	c.taskManager = tasks.GetManager()

	// 初始化 SSE 管理器
	baseURL := c.config.MCP.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	c.sseManager = tasks.NewSSEManager(c.taskManager, 100) // 最大 100 个并发连接
	c.taskManager.SetSSEManager(c.sseManager)              // 双向连接

	log.Infof("Task manager initialized with SSE support (base URL: %s)", baseURL)
	return nil
}

func (c *Container) initGovernance() error {
	dbPath := c.config.Governance.DBPath
	if dbPath == "" {
		dbPath = "data/governance.db"
	}

	db, err := governance.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize governance database: %w", err)
	}
	c.governanceDB = db

	govConfig := governance.Config{
		DBPath: dbPath,
	}
	govConfig.Confidence.AutoApplyThreshold = 0.80
	govConfig.Confidence.ReviewThreshold = 0.50
	govConfig.Batch.DefaultLimit = 100
	govConfig.Batch.MaxLimit = 1000
	govConfig.Batch.WorkerCount = 5

	c.governanceService = governance.NewService(db, govConfig, c.contentAPI)
	c.governanceHandler = governance.NewHandler(c.governanceService)

	log.Infof("Governance system initialized: %s", dbPath)
	return nil
}

// BuildAPI 构建 API 实例
// 这是容器的主要输出，返回完全初始化的 API 实例
func (c *Container) BuildAPI() (*calibre.Api, error) {
	if c.contentAPI == nil {
		return nil, fmt.Errorf("content API not initialized")
	}

	// 创建 Repository 和 Service 层（如果 qdrantSearcher 可用）
	if c.qdrantSearcher != nil {
		// 创建 ContentAPI 适配器（实现 service.ContentAPI 接口）
		contentAPIAdapter := &contentAPIAdapter{api: c.contentAPI}

		// 创建 BookRepository
		c.bookRepo = repository.NewQdrantBookRepository(c.qdrantSearcher)

		// 创建 BookService（使用 Repository）
		c.bookService = service.NewBookServiceWithRepository(
			c.bookRepo,
			contentAPIAdapter,
			c.taskManager,
			c.qdrantSearcher, // 保留用于任务调度
		)

		// 创建 BookHandler
		c.bookHandler = calibre.NewBookHandler(c.bookService)

		log.Info("Book repository, service and handler initialized")
	}

	// 创建 API 实例
	api := &calibre.Api{}

	// 注入依赖
	if err := api.InjectDependencies(
		c.config,
		c.contentAPI,
		c.qdrantClient,
		c.qdrantSearcher,
		c.cachedSearcher,
		c.cacheManager,
		c.chatDB,
		c.sseManager,
		c.bookHandler,
	); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies: %w", err)
	}

	// 如果 LLM 和 Qdrant 搜索器都可用，初始化 Chat Agent
	if c.chatLLM != nil && c.qdrantSearcher != nil && c.chatDB != nil {
		// 创建 TocFetcher 函数
		tocFetcher := api.CreateTocFetcher()

		// 创建 Chat Agent
		c.chatAgent = chat.NewAgent(c.chatLLM, c.qdrantSearcher, tocFetcher)

		// 注入 Chat Agent 到 API
		if err := api.InjectChatAgent(c.chatAgent); err != nil {
			return nil, fmt.Errorf("failed to inject chat agent: %w", err)
		}

		log.Info("Chat agent initialized and injected")
	}

	if c.governanceHandler != nil {
		api.InjectGovernanceHandler(c.governanceHandler)
		api.InjectGovernanceService(c.governanceService)
		log.Info("Governance handler injected")
	}

	c.api = api
	return api, nil
}

// contentAPIAdapter 适配器，实现 service.ContentAPI 接口
type contentAPIAdapter struct {
	api *content.Api
}

func (a *contentAPIAdapter) DeleteBooks(ids []string, library string) error {
	return a.api.DeleteBooks(ids, library)
}

func (a *contentAPIAdapter) UpdateMetaData(id string, metadata map[string]interface{}, library string) (bool, error) {
	_, err := a.api.UpdateMetaData(id, metadata, library)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *contentAPIAdapter) GetAllPublisher() ([]string, error) {
	return a.api.GetAllPublisher()
}

// GetConfig 获取配置
func (c *Container) GetConfig() *calibre.Config {
	return c.config
}

// GetContentAPI 获取 Content API 客户端
func (c *Container) GetContentAPI() *content.Api {
	return c.contentAPI
}

// GetQdrantClient 获取 Qdrant 客户端
func (c *Container) GetQdrantClient() *qdrant.Client {
	return c.qdrantClient
}

// GetQdrantSearcher 获取 Qdrant 搜索器
func (c *Container) GetQdrantSearcher() *qdrant.Searcher {
	return c.qdrantSearcher
}

// GetCachedSearcher 获取带缓存的搜索器
func (c *Container) GetCachedSearcher() *cache.CachedSearcher {
	return c.cachedSearcher
}

// GetCacheManager 获取缓存管理器
func (c *Container) GetCacheManager() *cache.Manager {
	return c.cacheManager
}

// GetChatDB 获取聊天数据库
func (c *Container) GetChatDB() *chat.DB {
	return c.chatDB
}

// GetChatAgent 获取聊天 Agent
func (c *Container) GetChatAgent() *chat.Agent {
	return c.chatAgent
}

// GetTaskManager 获取任务管理器
func (c *Container) GetTaskManager() *tasks.Manager {
	return c.taskManager
}

// GetSSEManager 获取 SSE 管理器
func (c *Container) GetSSEManager() *tasks.SSEManager {
	return c.sseManager
}

// GetAPI 获取 API 实例
func (c *Container) GetAPI() *calibre.Api {
	return c.api
}

// GetConfigManager 获取配置管理器
func (c *Container) GetConfigManager() *config.Manager {
	return c.configManager
}

// EnableConfigHotReload 启用配置热重载
func (c *Container) EnableConfigHotReload() error {
	if c.configManager == nil {
		return fmt.Errorf("config manager not available")
	}

	// 添加配置变更监听器
	c.configManager.AddWatcherFunc(func(oldConfig, newConfig *calibre.Config) error {
		log.Infof("config changed, applying updates...")

		// 更新容器中的配置
		c.config = newConfig

		// 这里可以添加更多的配置变更处理逻辑
		// 例如：重新初始化某些组件、更新日志级别等

		// 更新日志级别
		if oldConfig.Debug != newConfig.Debug {
			log.EnableDebug = newConfig.Debug
			log.Infof("debug mode changed: %v -> %v", oldConfig.Debug, newConfig.Debug)
		}

		// 如果 Content Server 地址变更，记录警告
		if oldConfig.Content.Server != newConfig.Content.Server {
			log.Warnf("content server changed: %s -> %s (requires restart)",
				oldConfig.Content.Server, newConfig.Content.Server)
		}

		// 如果 Qdrant 配置变更，记录警告
		if oldConfig.Qdrant.URL != newConfig.Qdrant.URL {
			log.Warnf("qdrant URL changed: %s -> %s (requires restart)",
				oldConfig.Qdrant.URL, newConfig.Qdrant.URL)
		}

		return nil
	})

	// 开始监听配置文件变更
	c.configManager.StartWatching()

	log.Info("config hot reload enabled")
	return nil
}

// ReloadConfig 手动重新加载配置
func (c *Container) ReloadConfig() error {
	if c.configManager == nil {
		return fmt.Errorf("config manager not available")
	}

	return c.configManager.ReloadConfig()
}

// Close 关闭容器，清理资源
func (c *Container) Close() error {
	var errs []error

	if c.chatDB != nil {
		if err := c.chatDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close chat DB: %w", err))
		}
	}

	if c.cacheManager != nil {
		// Cache manager 可能需要清理资源
		log.Info("Cache manager cleanup (if needed)")
	}

	if c.configManager != nil {
		if err := c.configManager.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close config manager: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during container cleanup: %v", errs)
	}

	log.Info("Container closed successfully")
	return nil
}
