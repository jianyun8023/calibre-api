package main

import (
	"encoding/json"
	"fmt"
	server "github.com/ckanthony/gin-mcp"
	"net/http"
	"strings"

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
	mcp := server.New(r, &server.Config{
		Name:        conf.MCP.ServerName,
		Description: "a Calibre API MCP Server",
		// BaseURL is crucial! It tells MCP clients where to send requests.
		BaseURL: conf.MCP.BaseURL,
	})

	// 4. Mount the MCP server endpoint
	mcp.Mount("/mcp") // MCP clients will connect here
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
