package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/internal/container"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/spf13/viper"
)

func main() {
	// 创建支持热重载的依赖注入容器
	cont, err := container.NewContainerWithConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create container: %v\n", err)
		os.Exit(1)
	}
	defer cont.Close()

	// 启用配置热重载
	if err := cont.EnableConfigHotReload(); err != nil {
		log.Warnf("Failed to enable config hot reload: %v", err)
	}

	// 获取初始配置
	conf := cont.GetConfig()
	log.EnableDebug = conf.Debug
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 构建 API 实例
	client, err := cont.BuildAPI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build API: %v\n", err)
		os.Exit(1)
	}

	// 创建 Gin 路由器
	r := gin.Default()

	// 配置 CORS 中间件（支持 MCP Inspector 等跨域请求）
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 允许所有来源，生产环境应限制具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 健康检查端点
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// 配置管理端点
	r.POST("/admin/config/reload", func(c *gin.Context) {
		if err := cont.ReloadConfig(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to reload config",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "config reloaded successfully",
		})
	})

	r.GET("/admin/config", func(c *gin.Context) {
		config := cont.GetConfig()
		c.JSON(http.StatusOK, gin.H{
			"config": config,
		})
	})

	// 设置 API 路由
	client.SetupRouter(r)

	// 注册 pprof 性能分析端点（仅在 debug 模式下）
	if conf.Debug {
		pprofHandler := calibre.NewPProfHandler()
		pprofHandler.RegisterRoutes(r)
		
		// 添加调试端点
		debugGroup := r.Group("/api/debug")
		{
			debugGroup.GET("/goroutines", pprofHandler.GetGoroutineInfo)
			debugGroup.POST("/gc", pprofHandler.TriggerGC)
			debugGroup.GET("/memstats", pprofHandler.GetMemStats)
			debugGroup.GET("/check-leak", pprofHandler.CheckGoroutineLeak)
		}
		log.Info("Debug endpoints enabled (pprof + debug APIs)")
	}

	// 初始化并挂载 MCP 服务器（必须在 setPages/NoRoute 之前）
	if conf.MCP.Enabled {
		mcpServer := calibre.NewMCPServer(client, conf.MCP)
		mcpServer.Mount(r)
		log.Infof("MCP Server enabled: transport=%s", conf.MCP.Transport)
	} else {
		log.Info("MCP Server disabled")
	}

	// 设置 NoRoute 处理器返回 JSON 404
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "route not found",
			"path":  c.Request.URL.Path,
		})
	})

	// 打印所有路由
	for _, route := range r.Routes() {
		log.Infof("route: %s %s", route.Method, route.Path)
	}

	// 创建 HTTP 服务器以支持优雅退出
	srv := &http.Server{
		Addr:    conf.Address,
		Handler: r,
	}

	// 在 goroutine 中启动服务器
	go func() {
		log.Infof("Server starting on %s", conf.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号以优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	// 关闭后台任务
	cont.Shutdown()

	// 设置 5 秒超时来完成现有请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Warnf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited gracefully")
}



func initConfig() (*calibre.Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/calibre-api/")
	viper.AddConfigPath("$HOME/.calibre-api")
	viper.AddConfigPath(".")
	viper.SetDefault("address", ":8080")
	viper.SetDefault("tmpDir", "/tmp")

	// MCP defaults
	viper.SetDefault("mcp.enabled", true)
	viper.SetDefault("mcp.server_name", "calibre-mcp-server")
	viper.SetDefault("mcp.version", "1.2.0")
	viper.SetDefault("mcp.transport", "sse")
	viper.SetDefault("mcp.sse_endpoint", "/mcp/sse")
	viper.SetDefault("mcp.message_endpoint", "/mcp/message")
	viper.SetDefault("mcp.timeout", 30)

	viper.SetEnvPrefix("CALIBRE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var conf calibre.Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&conf); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	marshal, _ := json.Marshal(conf)
	log.Infof("loaded config %s", marshal)
	return &conf, nil
}

// validateConfig validates the configuration
func validateConfig(conf *calibre.Config) error {
	if conf.Address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}
