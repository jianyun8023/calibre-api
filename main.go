package main

import (
	"encoding/json"
	"fmt"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/internal/lanzou"
	"github.com/jianyun8023/calibre-api/pkg/log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"strings"
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
	calibre.NewClient(conf).SetupRouter(r)
	if l, err := lanzou.NewClient(); err == nil {
		l.SetupRouter(r)
	}
	r.Run(conf.Address)
}

func initConfig() *calibre.Config {
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

	var conf calibre.Config
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("bind config failed! %w", err))
	}
	marshal, _ := json.Marshal(conf)
	log.Infof("loaded config %s", marshal)
	return &conf
}
