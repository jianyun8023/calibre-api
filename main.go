package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/internal/mcp"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/spf13/viper"
)

func main() {
	// Parse command line flags
	mcpMode := flag.Bool("mcp", false, "启动 MCP 服务器模式，与 AI 助手交互")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Calibre API Server - 书籍管理和搜索服务器\n\n")
		fmt.Fprintf(os.Stderr, "使用方法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n模式说明:\n")
		fmt.Fprintf(os.Stderr, "  默认模式: 启动 HTTP API 服务器\n")
		fmt.Fprintf(os.Stderr, "  MCP 模式: 启动 MCP 服务器，与 AI 助手交互\n")
		fmt.Fprintf(os.Stderr, "\n环境变量:\n")
		fmt.Fprintf(os.Stderr, "  MCP_MODE=true        启用 MCP 模式\n")
		fmt.Fprintf(os.Stderr, "  CALIBRE_MCP_MODE=true 启用 MCP 模式\n")
		fmt.Fprintf(os.Stderr, "  CALIBRE_MCP_BASE_URL API 基础 URL\n\n")
		fmt.Fprintf(os.Stderr, "配置文件:\n")
		fmt.Fprintf(os.Stderr, "  ./config.yaml 或 ~/.calibre-api/config.yaml\n")
		fmt.Fprintf(os.Stderr, "  设置 mcp.enabled: true 启用 MCP 模式\n\n")
	}
	flag.Parse()

	conf := initConfig()
	log.EnableDebug = conf.Debug

	// Create Calibre API client
	client := calibre.NewClient(conf)

	// Check if should run in MCP mode
	if shouldRunMCPMode(*mcpMode, conf) {
		runMCPServer(client, conf)
	} else {
		runHTTPServer(client, conf)
	}
}

func shouldRunMCPMode(mcpFlag bool, conf *calibre.Config) bool {
	// 1. 命令行参数优先
	if mcpFlag {
		return true
	}

	// 2. 检查环境变量
	if os.Getenv("MCP_MODE") == "true" || os.Getenv("CALIBRE_MCP_MODE") == "true" {
		return true
	}

	// 3. 检查配置文件
	if conf.MCP.Enabled {
		return true
	}

	return false
}

func runMCPServer(client *calibre.Api, conf *calibre.Config) {
	log.Infof("Starting Calibre MCP Server...")
	log.Infof("Protocol Version: %s", "2024-11-05")
	log.Infof("Server Version: %s", getServerVersion(conf))

	// Get base URL for API integration
	baseURL := conf.MCP.BaseURL
	if baseURL == "" {
		baseURL = os.Getenv("CALIBRE_MCP_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost%s", conf.Address)
		}
	}

	log.Infof("API Base URL: %s", baseURL)

	// Create and start MCP server
	mcpServer := mcp.NewMCPServerWithIntegration(client, baseURL)
	if err := mcpServer.Start(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func runHTTPServer(client *calibre.Api, conf *calibre.Config) {
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	setPages(r, conf)
	client.SetupRouter(r)

	for _, route := range r.Routes() {
		log.Infof("route: %s %s", route.Method, route.Path)
	}

	log.Infof("server listen on %s", conf.Address)
	r.Run(conf.Address)
}

func getServerVersion(conf *calibre.Config) string {
	if conf.MCP.Version != "" {
		return conf.MCP.Version
	}
	return "1.0.0"
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

	//// Serve the settings page
	//r.GET("/setting", func(c *gin.Context) {
	//	c.File(conf.TemplateDir + "/setting.html")
	//	//c.HTML(http.StatusOK, "setting.html", nil)
	//})
	//
	//r.GET("/books", func(c *gin.Context) {
	//	c.File(conf.TemplateDir + "/books.html")
	//	//c.HTML(http.StatusOK, "setting.html", nil)
	//})
	//
	//r.GET("/search", func(c *gin.Context) {
	//	c.File(conf.TemplateDir + "/search.html")
	//	//c.HTML(http.StatusOK, "search.html", nil)
	//})
	//r.GET("/detail/:id", func(c *gin.Context) {
	//	c.File(conf.TemplateDir + "/detail.html")
	//})
}

func initConfig() *calibre.Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/calibre-api/")
	viper.AddConfigPath("$HOME/.calibre-api")
	viper.AddConfigPath(".")
	viper.SetDefault("address", ":8080")
	viper.SetDefault("staticDir", "./static")
	viper.SetDefault("tmpDir", "/tmp")

	// MCP defaults
	viper.SetDefault("mcp.enabled", false)
	viper.SetDefault("mcp.server_name", "calibre-mcp-server")
	viper.SetDefault("mcp.version", "1.0.0")
	viper.SetDefault("mcp.timeout", 30)

	viper.SetEnvPrefix("CALIBRE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var conf calibre.Config
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("bind config failed! %w", err))
	}
	marshal, _ := json.Marshal(conf)
	log.Infof("loaded config %s", marshal)
	return &conf
}
