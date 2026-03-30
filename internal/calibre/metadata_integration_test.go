package calibre

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/content"
	"github.com/jianyun8023/calibre-api/pkg/douban"
)

// TestMetadataLocalMode 测试本地模式
func TestMetadataLocalMode(t *testing.T) {
	// 创建配置
	config := &Config{
		Content: Content{
			Server: "https://lib.pve.icu", // 使用真实服务器或 mock
		},
		Metadata: Metadata{
			DoubanMode: "local",
		},
	}
	config.Metadata.ApplyDoubanDefaults()

	// 创建 content API 客户端
	contentClient, _ := content.NewClient(config.Content.Server)

	// 创建 Api 实例
	api := &Api{
		config:     config,
		http:       contentClient.Client,
		doubanMode: config.Metadata.DoubanMode,
	}

	// 初始化 douban 服务
	parser := douban.NewDoubanBookParser()
	loader := douban.NewDoubanBookLoader(parser, config.Metadata.DoubanConfig.BaseUrl)
	api.doubanService = douban.NewService(loader, config.Metadata.DoubanConfig.SearchUrl, config.Metadata.DoubanConfig.BaseUrl)

	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metadata/isbn/:isbn", api.getIsbn)
	router.GET("/metadata/search", api.queryMetadata)

	// 测试 ISBN 查询
	t.Run("GetISBN_LocalMode", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metadata/isbn/9787111633082", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 200 or 404, got %d", w.Code)
		}

		var result map[string]interface{}
		if w.Code == http.StatusOK {
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}
			t.Logf("ISBN query result: %+v", result)
		}
	})

	// 测试搜索
	t.Run("Search_LocalMode", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metadata/search?query=深入理解计算机系统", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 200 or 404, got %d", w.Code)
		}

		if w.Code == http.StatusOK {
			var result douban.SearchResult
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}
			t.Logf("Search found %d books", result.Count)
		}
	})
}

// TestMetadataHTTPMode 测试 HTTP 模式
func TestMetadataHTTPMode(t *testing.T) {
	// 创建 mock HTTP 服务器
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/book/isbn/9787111633082" {
			response := map[string]interface{}{
				"title":     "深入理解计算机系统",
				"author":    []string{"Randal E. Bryant"},
				"publisher": "机械工业出版社",
				"isbn13":    "9787111633082",
			}
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/v2/book/search" {
			response := douban.SearchResult{
				Success: true,
				Count:   1,
				Books: []douban.Book{
					{
						Title:     "测试书籍",
						Author:    []string{"测试作者"},
						Publisher: "测试出版社",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// 创建配置
	config := &Config{
		Content: Content{Server: mockServer.URL},
		Metadata: Metadata{
			DoubanUrl:  mockServer.URL,
			DoubanMode: "http",
		},
	}
	config.Metadata.ApplyDoubanDefaults()

	// 创建 Api 实例
	contentClient, _ := content.NewClient(mockServer.URL)
	api := &Api{
		config:     config,
		http:       contentClient.Client,
		doubanMode: config.Metadata.DoubanMode,
	}

	// 设置 Gin 测试模式
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/metadata/isbn/:isbn", api.getIsbn)
	router.GET("/metadata/search", api.queryMetadata)

	// 测试 ISBN 查询
	t.Run("GetISBN_HTTPMode", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metadata/isbn/9787111633082", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}

		if result["title"] != "深入理解计算机系统" {
			t.Errorf("Expected title '深入理解计算机系统', got '%s'", result["title"])
		}
	})

	// 测试搜索
	t.Run("Search_HTTPMode", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/metadata/search?query=测试", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestMetadataConfigDefaults 测试配置默认值
func TestMetadataConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    Metadata
		expected Metadata
	}{
		{
			name: "Empty config with local mode",
			input: Metadata{
				DoubanMode: "local",
			},
			expected: Metadata{
				DoubanMode: "local",
				DoubanConfig: DoubanConfig{
					BaseUrl:   DefaultDoubanBaseUrl,
					SearchUrl: DefaultDoubanSearchUrl,
					IsbnUrl:   DefaultDoubanIsbnUrl,
					DetailUrl: DefaultDoubanDetailUrl,
				},
			},
		},
		{
			name:  "Empty config defaults to http mode",
			input: Metadata{},
			expected: Metadata{
				DoubanMode: "http",
				DoubanConfig: DoubanConfig{
					BaseUrl:   "",
					SearchUrl: "",
					IsbnUrl:   "",
					DetailUrl: "",
				},
			},
		},
		{
			name: "Partial config with local mode",
			input: Metadata{
				DoubanMode: "local",
				DoubanConfig: DoubanConfig{
					BaseUrl: "https://custom.douban.com/",
				},
			},
			expected: Metadata{
				DoubanMode: "local",
				DoubanConfig: DoubanConfig{
					BaseUrl:   "https://custom.douban.com/",
					SearchUrl: DefaultDoubanSearchUrl,
					IsbnUrl:   DefaultDoubanIsbnUrl,
					DetailUrl: DefaultDoubanDetailUrl,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input
			config.ApplyDoubanDefaults()

			if config.DoubanMode != tt.expected.DoubanMode {
				t.Errorf("DoubanMode: expected %s, got %s", tt.expected.DoubanMode, config.DoubanMode)
			}

			if config.DoubanConfig.BaseUrl != tt.expected.DoubanConfig.BaseUrl {
				t.Errorf("BaseUrl: expected %s, got %s", tt.expected.DoubanConfig.BaseUrl, config.DoubanConfig.BaseUrl)
			}

			if tt.expected.DoubanMode == "local" {
				if config.DoubanConfig.SearchUrl != tt.expected.DoubanConfig.SearchUrl {
					t.Errorf("SearchUrl: expected %s, got %s", tt.expected.DoubanConfig.SearchUrl, config.DoubanConfig.SearchUrl)
				}
				if config.DoubanConfig.IsbnUrl != tt.expected.DoubanConfig.IsbnUrl {
					t.Errorf("IsbnUrl: expected %s, got %s", tt.expected.DoubanConfig.IsbnUrl, config.DoubanConfig.IsbnUrl)
				}
				if config.DoubanConfig.DetailUrl != tt.expected.DoubanConfig.DetailUrl {
					t.Errorf("DetailUrl: expected %s, got %s", tt.expected.DoubanConfig.DetailUrl, config.DoubanConfig.DetailUrl)
				}
			}
		})
	}
}
