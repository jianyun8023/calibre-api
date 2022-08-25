package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	calibreAPI "github.com/jianyun8023/calibre-api/src"
	"github.com/spf13/viper"
	"log"
	"strings"
)

var db = make(map[string]string)

func setupRouter(client *calibreAPI.CalibreApi) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.GET("/get/cover/:id", client.GetCover)
	r.GET("/get/book/:id", client.GetBookFile)
	r.GET("/read/:id/toc", client.GetBookToc)
	r.GET("/read/:id/file/*path", client.GetBookContent)
	//r.GET("/read/:id/file", client.GetBookFile)
	r.GET("/book/:id", client.GetBook)
	r.GET("/search", client.Search)
	r.POST("/search", client.Search)
	//r.StaticFS("/css", http.Dir("./static/css"))
	//r.StaticFS("/js", http.Dir("./static/js"))
	//r.StaticFile("/favicon.ico", "./static/favicon.ico")
	//r.StaticFile("/", "./static/index.html")
	//r.StaticFile("/index.html", "./static/index.html")
	return r
}

func main() {
	conf := initConfig()
	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	var client = calibreAPI.NewClient(conf)
	r := setupRouter(&client)
	r.Run(conf.Address)
}

func initConfig() *calibreAPI.Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/calibre-api/")
	viper.AddConfigPath("$HOME/.calibre-api")
	viper.AddConfigPath(".")
	viper.SetDefault("address", ":8080")
	viper.SetEnvPrefix("CALIBRE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var conf calibreAPI.Config
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("bind config failed! %w", err))
	}
	marshal, _ := json.Marshal(conf)
	log.Printf("loaded config %s", marshal)
	return &conf
}
