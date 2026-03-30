package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

// Manager 配置管理器，支持热重载
type Manager struct {
	mu       sync.RWMutex
	config   *calibre.Config
	viper    *viper.Viper
	watchers []ConfigWatcher
}

// ConfigWatcher 配置变更监听器
type ConfigWatcher interface {
	OnConfigChanged(oldConfig, newConfig *calibre.Config) error
}

// ConfigWatcherFunc 配置变更监听器函数类型
type ConfigWatcherFunc func(oldConfig, newConfig *calibre.Config) error

// OnConfigChanged 实现 ConfigWatcher 接口
func (f ConfigWatcherFunc) OnConfigChanged(oldConfig, newConfig *calibre.Config) error {
	return f(oldConfig, newConfig)
}

// NewManager 创建配置管理器
func NewManager() (*Manager, error) {
	// 加载 .env 文件（优先级：当前目录 > $HOME/.calibre-api > /etc/calibre-api/）
	loadEnvFile()

	v := viper.New()

	// 配置 viper
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/calibre-api/")
	v.AddConfigPath("$HOME/.calibre-api")
	v.AddConfigPath(".")

	// 设置默认值
	setDefaults(v)

	// 设置环境变量（.env 已加载到系统环境变量中）
	v.SetEnvPrefix("CALIBRE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	manager := &Manager{
		viper:    v,
		watchers: make([]ConfigWatcher, 0),
	}

	// 初始加载配置
	if err := manager.loadConfig(); err != nil {
		return nil, fmt.Errorf("failed to load initial config: %w", err)
	}

	return manager, nil
}

// loadEnvFile 按优先级加载 .env 文件
func loadEnvFile() {
	// 定义搜索路径（优先级从高到低）
	searchPaths := []string{
		".",                         // 当前目录
		"$HOME/.calibre-api",        // 用户主目录
		"/etc/calibre-api/",         // 系统配置目录
	}

	for _, path := range searchPaths {
		// 展开环境变量
		expandedPath := os.ExpandEnv(path)
		envFilePath := filepath.Join(expandedPath, ".env")

		// 检查文件是否存在
		if _, err := os.Stat(envFilePath); err == nil {
			// 加载 .env 文件
			if err := gotenv.Load(envFilePath); err != nil {
				log.Warnf("Failed to load .env file from %s: %v", envFilePath, err)
			} else {
				log.Infof("Loaded .env file from: %s", envFilePath)
				return // 加载第一个找到的 .env 文件后返回
			}
		}
	}

	log.Info("No .env file found, using system environment variables only")
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	v.SetDefault("address", ":8080")
	v.SetDefault("tmpDir", "/tmp")
	v.SetDefault("debug", false)

	// MCP defaults
	v.SetDefault("mcp.enabled", true)
	v.SetDefault("mcp.server_name", "calibre-mcp-server")
	v.SetDefault("mcp.version", "1.2.0")
	v.SetDefault("mcp.transport", "sse")
	v.SetDefault("mcp.sse_endpoint", "/mcp/sse")
	v.SetDefault("mcp.message_endpoint", "/mcp/message")
	v.SetDefault("mcp.timeout", 30)

	// Content defaults
	v.SetDefault("content.server", "")

	// Qdrant defaults
	v.SetDefault("qdrant.url", "")
	v.SetDefault("qdrant.collection", "books")
	v.SetDefault("qdrant.timeout", 30)

	// Meilisearch defaults
	v.SetDefault("meilisearch.host", "")
	v.SetDefault("meilisearch.api_key", "")
	v.SetDefault("meilisearch.index_name", "books")

	// Cache defaults
	v.SetDefault("cache.dir", "")
	v.SetDefault("cache.max_size_gb", 10.0)
}

// loadConfig 加载配置
func (m *Manager) loadConfig() error {
	if err := m.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var conf calibre.Config
	if err := m.viper.Unmarshal(&conf); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&conf); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.mu.Lock()
	m.config = &conf
	m.mu.Unlock()

	// Mask sensitive info before logging
	logConf := conf
	if logConf.Meilisearch.APIKey != "" {
		logConf.Meilisearch.APIKey = "***"
	}
	if logConf.Embedding.SiliconFlow.APIToken != "" {
		logConf.Embedding.SiliconFlow.APIToken = "***"
	}

	marshal, _ := json.Marshal(logConf)
	log.Infof("loaded config: %s", marshal)
	return nil
}

// validateConfig 验证配置
func validateConfig(conf *calibre.Config) error {
	if conf.Address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}

// GetConfig 获取当前配置（线程安全）
func (m *Manager) GetConfig() *calibre.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回配置的副本，避免并发修改
	configCopy := *m.config
	return &configCopy
}

// AddWatcher 添加配置变更监听器
func (m *Manager) AddWatcher(watcher ConfigWatcher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.watchers = append(m.watchers, watcher)
}

// AddWatcherFunc 添加配置变更监听器函数
func (m *Manager) AddWatcherFunc(fn func(oldConfig, newConfig *calibre.Config) error) {
	m.AddWatcher(ConfigWatcherFunc(fn))
}

// StartWatching 开始监听配置文件变更
func (m *Manager) StartWatching() {
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("config file changed: %s", e.Name)

		// 重新加载配置
		oldConfig := m.GetConfig()
		if err := m.reloadConfig(); err != nil {
			log.Infof("ERROR: failed to reload config: %v", err)
			return
		}

		newConfig := m.GetConfig()

		// 通知所有监听器
		m.notifyWatchers(oldConfig, newConfig)
	})

	log.Info("config hot reload enabled")
}

// reloadConfig 重新加载配置
func (m *Manager) reloadConfig() error {
	var conf calibre.Config
	if err := m.viper.Unmarshal(&conf); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&conf); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.mu.Lock()
	m.config = &conf
	m.mu.Unlock()

	// Mask sensitive info before logging
	logConf := conf
	if logConf.Meilisearch.APIKey != "" {
		logConf.Meilisearch.APIKey = "***"
	}
	if logConf.Embedding.SiliconFlow.APIToken != "" {
		logConf.Embedding.SiliconFlow.APIToken = "***"
	}

	marshal, _ := json.Marshal(logConf)
	log.Infof("reloaded config: %s", marshal)
	return nil
}

// notifyWatchers 通知所有配置变更监听器
func (m *Manager) notifyWatchers(oldConfig, newConfig *calibre.Config) {
	m.mu.RLock()
	watchers := make([]ConfigWatcher, len(m.watchers))
	copy(watchers, m.watchers)
	m.mu.RUnlock()

	for _, watcher := range watchers {
		if err := watcher.OnConfigChanged(oldConfig, newConfig); err != nil {
			log.Infof("ERROR: config watcher error: %v", err)
		}
	}
}

// ReloadConfig 手动重新加载配置
func (m *Manager) ReloadConfig() error {
	oldConfig := m.GetConfig()

	if err := m.loadConfig(); err != nil {
		return err
	}

	newConfig := m.GetConfig()
	m.notifyWatchers(oldConfig, newConfig)

	return nil
}

// GetDebugMode 获取调试模式状态
func (m *Manager) GetDebugMode() bool {
	config := m.GetConfig()
	return config.Debug
}

// GetAddress 获取监听地址
func (m *Manager) GetAddress() string {
	config := m.GetConfig()
	return config.Address
}

// GetContentServer 获取 Content Server 地址
func (m *Manager) GetContentServer() string {
	config := m.GetConfig()
	return config.Content.Server
}

// GetQdrantConfig 获取 Qdrant 配置
func (m *Manager) GetQdrantConfig() calibre.QdrantConfig {
	config := m.GetConfig()
	return config.Qdrant
}

// GetMCPConfig 获取 MCP 配置
func (m *Manager) GetMCPConfig() calibre.MCPConfig {
	config := m.GetConfig()
	return config.MCP
}

// Close 关闭配置管理器
func (m *Manager) Close() error {
	// viper 没有显式的关闭方法，但我们可以清理资源
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchers = nil
	m.config = nil

	log.Info("config manager closed")
	return nil
}
