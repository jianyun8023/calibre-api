package main

import (
	"encoding/json"
	"fmt"
	"github.com/bookstairs/bookhunter/internal/client"
	"github.com/bookstairs/bookhunter/internal/driver/lanzou"
	"github.com/gin-gonic/gin"
	calibreAPI "github.com/jianyun8023/calibre-api/src"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var db = make(map[string]string)

func setupRouter(c *calibreAPI.CalibreApi) *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	r.GET("/get/cover/:id", c.GetCover)
	r.GET("/get/book/:id", c.GetBookFile)
	r.GET("/read/:id/toc", c.GetBookToc)
	r.GET("/read/:id/file/*path", c.GetBookContent)
	//r.GET("/read/:id/file", client.GetBookFile)
	r.GET("/book/:id", c.GetBook)
	r.GET("/search", c.Search)
	r.POST("/search", c.Search)

	lanzouAPI, _ := lanzou.New(&client.Config{})
	r.GET("/api/lanzou", func(c *gin.Context) {
		lanzouUrl := c.Query("url")
		pwd := c.Query("pwd")

		_, err := url.Parse(lanzouUrl)
		if err != nil {
			c.JSON(http.StatusBadRequest,
				&Response{
					Msg:  "url参数错误",
					Code: 401,
				})
			return
		}

		link, err := lanzouAPI.ResolveShareURL(lanzouUrl, pwd)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				&Response{
					Msg:  err.Error(),
					Code: 500,
				})
			return
		}
		c.JSON(http.StatusOK, &Response{
			Msg:  err.Error(),
			Code: 500,
			Data: link,
		})
	})
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
	var c = calibreAPI.NewClient(conf)
	r := setupRouter(&c)
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

type Response struct {
	Code int64                 `json:"code"`
	Data []lanzou.ResponseData `json:"data"`
	Msg  string                `json:"msg"`
}
