package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/spf13/viper"
)

func main() {
	conf, err := initConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize configuration: %v\n", err)
		os.Exit(1)
	}

	log.EnableDebug = conf.Debug
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
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

	// 2. Define your API routes (Gin-MCP will discover these)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	client := calibre.NewClient(conf)
	client.SetupRouter(r)

	// 初始化并挂载 MCP 服务器（必须在 setPages/NoRoute 之前）
	if conf.MCP.Enabled {
		mcpServer := calibre.NewMCPServer(client, conf.MCP)
		mcpServer.Mount(r)
		log.Infof("MCP Server enabled: transport=%s", conf.MCP.Transport)
	} else {
		log.Info("MCP Server disabled")
	}

	// 最后设置静态文件和 NoRoute（会捕获所有未匹配的路由）
	setPages(r, conf)

	for _, route := range r.Routes() {
		log.Infof("route: %s %s", route.Method, route.Path)
	}

	log.Infof("server listen on %s", conf.Address)
	r.Run(conf.Address)
}

func setPages(r *gin.Engine, conf *calibre.Config) {
	// 配置静态文件目录
	r.Static("/assets", conf.StaticDir+"/assets")

	// 配置模板目录
	//r.LoadHTMLGlob(conf.TemplateDir + "/*")
	r.GET("/", func(c *gin.Context) {
		//c.HTML(http.StatusOK, "index.html", nil)
		c.File(conf.StaticDir + "/index.html")
	})
	r.GET("/index", func(c *gin.Context) {
		//c.HTML(http.StatusOK, "index.html", nil)
		c.File(conf.StaticDir + "/index.html")
	})
	r.GET("/favico.ico", func(c *gin.Context) {
		//c.HTML(http.StatusOK, "index.html", nil)
		c.File(conf.StaticDir + "/favico.ico")
	})

	// Serve the index.html file for all other routes
	r.NoRoute(func(c *gin.Context) {
		c.File(conf.StaticDir + "/index.html")
	})
}

func initConfig() (*calibre.Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/calibre-api/")
	viper.AddConfigPath("$HOME/.calibre-api")
	viper.AddConfigPath(".")
	viper.SetDefault("address", ":8080")
	viper.SetDefault("staticDir", "./static")
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
	if conf.StaticDir == "" {
		return fmt.Errorf("staticDir is required")
	}
	return nil
}
