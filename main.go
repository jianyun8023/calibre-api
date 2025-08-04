package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	ginmcp "github.com/ckanthony/gin-mcp"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/pkg/log"
	"github.com/spf13/viper"
)

func main() {
	conf := initConfig()
	log.EnableDebug = conf.Debug
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// 2. Define your API routes (Gin-MCP will discover these)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	setPages(r, conf)
	client := calibre.NewClient(conf)
	client.SetupRouter(r)

	for _, route := range r.Routes() {
		log.Infof("route: %s %s", route.Method, route.Path)
	}

	// 3. Create and configure the MCP server
	//    Provide essential details for the MCP client.
	mcp := ginmcp.New(r, &ginmcp.Config{
		Name:        conf.MCP.ServerName,
		Description: "a Calibre API MCP Server",
		// BaseURL is crucial! It tells MCP clients where to send requests.
		BaseURL: conf.MCP.BaseURL,
		ExcludeOperations: []string{
			"/",
			"/index",
			"/favicon.ico",
			"/assets/*",
			"/api/book/:id/delete",
		},
	})

	// 注册 API 参数模式，为 MCP 工具提供详细的参数说明
	registerMCPSchemas(mcp)

	// 4. Mount the MCP server endpoint
	mcp.Mount("/mcp") // MCP clients will connect here
	log.Infof("server listen on %s", conf.Address)
	r.Run(conf.Address)
}

// registerMCPSchemas 为 gin-mcp 注册 API 参数模式
func registerMCPSchemas(mcp *ginmcp.GinMCP) {
	// 搜索相关接口
	mcp.RegisterSchema("GET", "/api/search", calibre.SearchRequest{}, nil)
	mcp.RegisterSchema("POST", "/api/search", nil, calibre.SearchRequest{})

	// 书籍管理相关接口
	mcp.RegisterSchema("POST", "/api/book/:id/update", nil, calibre.BookUpdateRequest{})

	// 元数据相关接口
	mcp.RegisterSchema("GET", "/api/metadata/search", calibre.MetadataSearchRequest{}, nil)

	// 索引管理相关接口
	mcp.RegisterSchema("POST", "/api/index/update", nil, calibre.IndexUpdateRequest{})
	mcp.RegisterSchema("POST", "/api/index/switch", nil, calibre.IndexSwitchRequest{})

	// 出版社列表接口
	mcp.RegisterSchema("GET", "/api/publisher", calibre.PublisherListRequest{}, nil)

	// 最近书籍接口
	mcp.RegisterSchema("GET", "/api/recently", calibre.RecentlyBooksRequest{}, nil)

	// 随机书籍接口
	mcp.RegisterSchema("GET", "/api/random", calibre.RandomBooksRequest{}, nil)

	// 读取书籍目录接口
	mcp.RegisterSchema("GET", "/api/read/:id/toc", calibre.BookTocRequest{}, nil)

	// 读取书籍内容接口
	mcp.RegisterSchema("GET", "/api/read/:id/file/*path", calibre.BookContentRequest{}, nil)

	// 读取书籍内容接口
	mcp.RegisterSchema("GET", "/api/book/content", calibre.BookContentByQueryRequest{}, nil)

	// 获取封面接口
	mcp.RegisterSchema("GET", "/api/get/cover/:id", calibre.GetCoverRequest{}, nil)

	// 代理封面接口
	mcp.RegisterSchema("GET", "/api/proxy/cover/*path", calibre.ProxyCoverRequest{}, nil)

	// 获取书籍文件接口
	mcp.RegisterSchema("GET", "/api/download/book/:id", calibre.GetBookFileRequest{}, nil)

	// 获取书籍信息接口
	mcp.RegisterSchema("GET", "/api/book/:id", calibre.GetBookRequest{}, nil)

	// 获取ISBN信息接口
	mcp.RegisterSchema("GET", "/api/metadata/isbn/:isbn", calibre.GetISBNRequest{}, nil)
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
