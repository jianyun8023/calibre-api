package container

import (
	"context"
	"fmt"
	"time"

	"github.com/jianyun8023/calibre-api/internal/cache"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/internal/config"
	"github.com/jianyun8023/calibre-api/internal/repository"
	"github.com/jianyun8023/calibre-api/internal/semantic"
	"github.com/jianyun8023/calibre-api/internal/semantic/embedding"
	"github.com/jianyun8023/calibre-api/internal/semantic/meilisearch"
	"github.com/jianyun8023/calibre-api/internal/drafts"
	"github.com/jianyun8023/calibre-api/internal/service"
	"github.com/jianyun8023/calibre-api/internal/tasks"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// Container 依赖注入容器
// 管理应用程序的所有依赖组件
type Container struct {
	configManager *config.Manager
	config        *calibre.Config

	// Core components
	contentAPI   *content.Api
	cacheManager *cache.Manager
	taskManager  *tasks.Manager
	sseManager   *tasks.SSEManager

	// Search components
	semanticSearcher semantic.Searcher // MeiliSearch searcher
	cachedSearcher   *cache.CachedSearcher

	// Repository layer
	bookRepo repository.BookRepository

	// Service layer
	bookService service.BookService
	bookHandler *calibre.BookHandlerV2

	// Drafts layer
	draftRepo    repository.DraftRepository
	draftService service.DraftService
	draftHandler *calibre.DraftHandler

	// API instance
	api *calibre.Api

	// Background tasks control
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

// NewContainer 创建新的依赖注入容器
func NewContainer(config *calibre.Config) (*Container, error) {
	// 创建 shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	
	c := &Container{
		config:         config,
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}

	// 按依赖顺序初始化组件
	if err := c.initContentAPI(); err != nil {
		return nil, fmt.Errorf("failed to initialize content API: %w", err)
	}

	// 初始化 MeiliSearch 搜索引擎
	if err := c.initMeilisearch(); err != nil {
		log.Warnf("Meilisearch initialization failed (optional): %v", err)
	}

	if err := c.initCache(); err != nil {
		log.Warnf("Cache initialization failed (optional): %v", err)
	}

	if err := c.initTasks(); err != nil {
		return nil, fmt.Errorf("failed to initialize task manager: %w", err)
	}

	if err := c.initDrafts(); err != nil {
		log.Warnf("Drafts initialization failed: %v", err)
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

	// 创建 shutdown context
	ctx, cancel := context.WithCancel(context.Background())

	c := &Container{
		configManager:  configManager,
		config:         configManager.GetConfig(),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}

	// 按依赖顺序初始化组件
	if err := c.initContentAPI(); err != nil {
		return nil, fmt.Errorf("failed to initialize content API: %w", err)
	}

	// 初始化 MeiliSearch 搜索引擎
	if err := c.initMeilisearch(); err != nil {
		log.Warnf("Meilisearch initialization failed (optional): %v", err)
	}

	if err := c.initCache(); err != nil {
		log.Warnf("Cache initialization failed (optional): %v", err)
	}

	if err := c.initTasks(); err != nil {
		return nil, fmt.Errorf("failed to initialize task manager: %w", err)
	}

	if err := c.initDrafts(); err != nil {
		log.Warnf("Drafts initialization failed: %v", err)
	}

	return c, nil
}

// initDrafts 初始化草稿数据库和服务
func (c *Container) initDrafts() error {
	dbPath := c.config.Drafts.DBPath
	if dbPath == "" {
		dbPath = ".cache/drafts.db" // default
	}

	db, err := drafts.InitDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize drafts database: %w", err)
	}

	c.draftRepo = repository.NewSqliteDraftRepository(db)
	return nil
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

// initMeilisearch 初始化 MeiliSearch 搜索引擎
func (c *Container) initMeilisearch() error {
	if c.config.Meilisearch.Host == "" {
		return fmt.Errorf("Meilisearch host not configured (skipping)")
	}

	// 初始化 Embedding Provider（用于语义搜索）
	providerConfig := embedding.ProviderConfig{
		Provider:    c.config.Embedding.Provider,
		Ollama:      c.config.Embedding.Ollama,
		SiliconFlow: c.config.Embedding.SiliconFlow,
		VectorDim:   4096,
	}

	var provider embedding.Provider
	var err error
	provider, err = embedding.NewProvider(providerConfig)
	if err != nil {
		log.Warnf("Failed to create embedding provider for MeiliSearch: %v. Semantic search will be limited.", err)
	}

	// 创建 MeiliSearch 客户端
	meiliClient := meilisearch.NewClient(
		c.config.Meilisearch.Host,
		c.config.Meilisearch.APIKey,
		c.config.Meilisearch.IndexName,
	)

	// 创建 MeiliSearch 搜索器
	meiliSearcher := meilisearch.NewSearcher(meiliClient, provider)
	c.semanticSearcher = meiliSearcher
	log.Infof("Meilisearch searcher initialized: %s (index: %s)", c.config.Meilisearch.Host, c.config.Meilisearch.IndexName)

	// 包装搜索器添加缓存功能
	cacheMaxSize := 1000
	cacheTTL := 300 // 5 分钟
	c.cachedSearcher = cache.WrapSearcher(meiliSearcher, cacheMaxSize, cacheTTL)
	log.Infof("Search cache initialized (MeiliSearch): maxSize=%d, ttl=%ds", cacheMaxSize, cacheTTL)

	// 确保索引存在
	ctx := context.Background()
	if err := meiliSearcher.EnsureIndexes(ctx); err != nil {
		log.Warnf("Failed to ensure MeiliSearch indexes: %v", err)
	} else {
		log.Info("MeiliSearch indexes ensured successfully")
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

// BuildAPI 构建 API 实例
// 这是容器的主要输出，返回完全初始化的 API 实例
func (c *Container) BuildAPI() (*calibre.Api, error) {
	if c.contentAPI == nil {
		return nil, fmt.Errorf("content API not initialized")
	}

	// 创建 Repository 和 Service 层（如果搜索器可用）
	if c.semanticSearcher != nil {
		// 创建 ContentAPI 适配器（实现 service.ContentAPI 接口）
		contentAPIAdapter := &contentAPIAdapter{api: c.contentAPI}

		// 创建 BookRepository
		c.bookRepo = repository.NewQdrantBookRepository(c.semanticSearcher)

		// 创建 BookService（使用 Repository）
		c.bookService = service.NewBookServiceWithRepository(
			c.bookRepo,
			contentAPIAdapter,
			c.taskManager,
			c.semanticSearcher, // 保留用于任务调度
		)

		// 创建 BookHandler
		c.bookHandler = calibre.NewBookHandler(c.bookService)

		// 初始化 DraftService 和 Handler (需要 bookService 依赖)
		if c.draftRepo != nil {
			c.draftService = service.NewDraftService(c.draftRepo, c.bookService)
			c.draftHandler = calibre.NewDraftHandler(c.draftService)
			log.Info("Drafts service and handler initialized")

			// Start background cleanup task with graceful shutdown support
			expireDays := c.config.Drafts.ExpireDays
			if expireDays <= 0 {
				expireDays = 30 // Default 30 days
			}
			go c.runDraftCleanupTask(expireDays)
		}

		log.Info("Book repository, service and handler initialized")
	}

	// 创建 API 实例
	api := &calibre.Api{}

	// 注入依赖
	if err := api.InjectDependencies(
		c.config,
		c.contentAPI,
		nil, // qdrantClient removed
		c.semanticSearcher,
		c.cachedSearcher,
		c.cacheManager,
		c.sseManager,
		c.bookHandler,
		c.draftHandler,
	); err != nil {
		return nil, fmt.Errorf("failed to inject dependencies: %w", err)
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

func (a *contentAPIAdapter) GetBookDetail(id int64) (*service.Book, error) {
	contentBook, err := a.api.GetBookDetail(id)
	if err != nil {
		return nil, err
	}

	// 转换 content.Book 到 service.Book
	return &service.Book{
		ID:           contentBook.ID,
		Title:        contentBook.Title,
		Authors:      contentBook.Authors,
		Publisher:    contentBook.Publisher,
		PubDate:      contentBook.PubDate,
		Isbn:         contentBook.Isbn,
		Tags:         contentBook.Tags,
		Rating:       contentBook.Rating,
		SeriesIndex:  contentBook.SeriesIndex,
		Comments:     contentBook.Comments,
		Languages:    contentBook.Languages,
		LastModified: contentBook.LastModified,
		Cover:        contentBook.Cover,
		FilePath:     contentBook.FilePath,
		Identifiers:  contentBook.Identifiers,
		Size:         contentBook.Size,
	}, nil
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

// GetCachedSearcher 获取带缓存的搜索器
func (c *Container) GetCachedSearcher() *cache.CachedSearcher {
	return c.cachedSearcher
}

// GetCacheManager 获取缓存管理器
func (c *Container) GetCacheManager() *cache.Manager {
	return c.cacheManager
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

		// 如果 MeiliSearch 配置变更，记录警告
		if oldConfig.Meilisearch.Host != newConfig.Meilisearch.Host {
			log.Warnf("meilisearch host changed: %s -> %s (requires restart)",
				oldConfig.Meilisearch.Host, newConfig.Meilisearch.Host)
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

// runDraftCleanupTask 运行后台草稿清理任务，支持优雅退出
func (c *Container) runDraftCleanupTask(expireDays int) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run once immediately on startup
	ctx, cancel := context.WithTimeout(c.shutdownCtx, 5*time.Minute)
	if count, err := c.draftService.CleanupExpiredDrafts(ctx, expireDays); err != nil {
		if ctx.Err() == nil { // 不记录由于 shutdown 导致的错误
			log.Warnf("Failed to cleanup expired drafts on startup: %v", err)
		}
	} else if count > 0 {
		log.Infof("Cleaned up %d expired drafts on startup", count)
	}
	cancel()

	// 定期清理
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(c.shutdownCtx, 5*time.Minute)
			if count, err := c.draftService.CleanupExpiredDrafts(ctx, expireDays); err != nil {
				if ctx.Err() == nil { // 不记录由于 shutdown 导致的错误
					log.Warnf("Failed to cleanup expired drafts: %v", err)
				}
			} else if count > 0 {
				log.Infof("Cleaned up %d expired drafts", count)
			}
			cancel()
		case <-c.shutdownCtx.Done():
			log.Info("Draft cleanup task stopped gracefully")
			return
		}
	}
}

// Shutdown 优雅关闭所有后台任务
func (c *Container) Shutdown() {
	log.Info("Shutting down background tasks...")
	if c.shutdownCancel != nil {
		c.shutdownCancel()
	}
	// 给后台任务一些时间来完成清理
	time.Sleep(100 * time.Millisecond)
	log.Info("Background tasks shutdown complete")
}
