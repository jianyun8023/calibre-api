package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/jianyun8023/calibre-api/internal/mcp"
	"github.com/spf13/viper"
)

func main() {
	// Initialize config
	conf := initConfig()

	// Create Calibre API client
	calibreClient := calibre.NewClient(conf)

	// Get base URL for API integration
	baseURL := os.Getenv("CALIBRE_MCP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default to localhost
	}

	// Create MCP server with API integration
	mcpServer := mcp.NewMCPServerWithIntegration(calibreClient, baseURL)

	log.Printf("Starting Calibre MCP Server...")
	log.Printf("Protocol Version: %s", "2024-11-05")
	log.Printf("Server Version: %s", "1.0.0")
	log.Printf("API Base URL: %s", baseURL)

	// Start MCP server
	if err := mcpServer.Start(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func initConfig() *calibre.Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/calibre-api/")
	viper.AddConfigPath("$HOME/.calibre-api")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("address", ":8080")
	viper.SetDefault("staticDir", "./static")
	viper.SetDefault("tmpDir", "/tmp")
	viper.SetDefault("debug", false)

	// Environment variable support
	viper.SetEnvPrefix("CALIBRE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		// Create a basic config if none exists
		log.Printf("Warning: Could not read config file: %v", err)
		log.Printf("Using default configuration")

		// Set minimal required config
		viper.SetDefault("content.server", "http://localhost:8083")
		viper.SetDefault("search.host", "http://localhost:7700")
		viper.SetDefault("search.index", "books")
		viper.SetDefault("metadata.doubanurl", "https://api.douban.com")
	}

	var conf calibre.Config
	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("failed to bind config: %w", err))
	}

	if conf.Debug {
		marshal, _ := json.Marshal(conf)
		log.Printf("Loaded config: %s", marshal)
	}

	return &conf
}
