package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jianyun8023/calibre-api/internal/calibre"
)

func TestConfigManager(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	// 写入初始配置
	initialConfig := `
address: ":8080"
debug: false
content:
  server: "https://example.com"
`
	if err := os.WriteFile(configFile, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// 切换到临时目录
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// 创建配置管理器
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 测试初始配置
	config := manager.GetConfig()
	if config.Address != ":8080" {
		t.Errorf("expected address :8080, got %s", config.Address)
	}
	if config.Debug != false {
		t.Errorf("expected debug false, got %v", config.Debug)
	}

	// 测试配置变更监听器
	var changedConfig *calibre.Config
	manager.AddWatcherFunc(func(oldConfig, newConfig *calibre.Config) error {
		changedConfig = newConfig
		return nil
	})

	// 启动监听
	manager.StartWatching()

	// 修改配置文件
	updatedConfig := `
address: ":9090"
debug: true
content:
  server: "https://updated.com"
`
	if err := os.WriteFile(configFile, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("failed to update config file: %v", err)
	}

	// 等待配置变更
	time.Sleep(100 * time.Millisecond)

	// 验证配置已更新
	newConfig := manager.GetConfig()
	if newConfig.Address != ":9090" {
		t.Errorf("expected address :9090, got %s", newConfig.Address)
	}
	if newConfig.Debug != true {
		t.Errorf("expected debug true, got %v", newConfig.Debug)
	}

	// 验证监听器被调用
	if changedConfig == nil {
		t.Error("config watcher was not called")
	} else if changedConfig.Address != ":9090" {
		t.Errorf("watcher received wrong config, expected address :9090, got %s", changedConfig.Address)
	}
}

func TestConfigManagerGetters(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	config := `
address: ":8080"
debug: true
content:
  server: "https://example.com"
qdrant:
  url: "http://localhost:6333"
  collection: "books"
mcp:
  enabled: true
  server_name: "test-server"
`
	if err := os.WriteFile(configFile, []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// 切换到临时目录
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 测试各种 getter 方法
	if manager.GetDebugMode() != true {
		t.Error("expected debug mode true")
	}

	if manager.GetAddress() != ":8080" {
		t.Errorf("expected address :8080, got %s", manager.GetAddress())
	}

	if manager.GetContentServer() != "https://example.com" {
		t.Errorf("expected content server https://example.com, got %s", manager.GetContentServer())
	}

	qdrantConfig := manager.GetQdrantConfig()
	if qdrantConfig.URL != "http://localhost:6333" {
		t.Errorf("expected qdrant URL http://localhost:6333, got %s", qdrantConfig.URL)
	}

	mcpConfig := manager.GetMCPConfig()
	if mcpConfig.ServerName != "test-server" {
		t.Errorf("expected MCP server name test-server, got %s", mcpConfig.ServerName)
	}
}

func TestConfigManagerReload(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	
	initialConfig := `
address: ":8080"
debug: false
`
	if err := os.WriteFile(configFile, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// 切换到临时目录
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 验证初始配置
	config := manager.GetConfig()
	if config.Debug != false {
		t.Error("expected initial debug false")
	}

	// 修改配置文件
	updatedConfig := `
address: ":8080"
debug: true
`
	if err := os.WriteFile(configFile, []byte(updatedConfig), 0644); err != nil {
		t.Fatalf("failed to update config file: %v", err)
	}

	// 手动重新加载
	if err := manager.ReloadConfig(); err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// 验证配置已更新
	newConfig := manager.GetConfig()
	if newConfig.Debug != true {
		t.Error("expected debug true after reload")
	}
}