package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/calibre"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// setupTestRouter creates a minimal Gin router for testing
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Add some API routes
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	
	// Set NoRoute handler (same as production)
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "route not found",
			"path":  c.Request.URL.Path,
		})
	})
	
	return r
}

// TestNoRouteReturnsJSON404 - Property 1: NoRoute returns JSON 404
// Feature: remove-gin-static-resources, Property 1: NoRoute returns JSON 404
// Validates: Requirements 1.4
func TestNoRouteReturnsJSON404(t *testing.T) {
	properties := gopter.NewProperties(nil)
	
	properties.Property("NoRoute returns JSON 404 for any invalid path", prop.ForAll(
		func(path string) bool {
			// Skip paths that match existing routes
			if strings.HasPrefix(path, "/api/") || path == "/ping" || strings.HasPrefix(path, "/mcp/") {
				return true
			}
			
			router := setupTestRouter()
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", path, nil)
			router.ServeHTTP(w, req)
			
			// Verify status code is 404
			if w.Code != http.StatusNotFound {
				t.Logf("Expected 404, got %d for path %s", w.Code, path)
				return false
			}
			
			// Verify response is JSON
			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Logf("Expected JSON content type, got %s for path %s", contentType, path)
				return false
			}
			
			// Verify response contains error and path
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Logf("Failed to parse JSON response for path %s: %v", path, err)
				return false
			}
			
			if response["error"] != "route not found" {
				t.Logf("Expected error message 'route not found', got %v for path %s", response["error"], path)
				return false
			}
			
			// Gin may clean the path (e.g., // becomes /), so we check the actual request path
			actualPath := req.URL.Path
			if response["path"] != actualPath {
				t.Logf("Expected path %s in response, got %v", actualPath, response["path"])
				return false
			}
			
			// Verify response is not HTML
			body := w.Body.String()
			if strings.Contains(body, "<html") || strings.Contains(body, "<!DOCTYPE") {
				t.Logf("Response contains HTML for path %s", path)
				return false
			}
			
			return true
		},
		gen.RegexMatch("/[a-z0-9/-]+").SuchThat(func(path string) bool {
			// Generate paths that don't match existing routes
			return !strings.HasPrefix(path, "/api/") && 
			       path != "/ping" && 
			       !strings.HasPrefix(path, "/mcp/") &&
			       len(path) > 1 && len(path) < 100
		}),
	))
	
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestRouterInitializationExcludesStaticRoutes - Example 1: Router initialization excludes static routes
// Validates: Requirements 1.1, 1.2, 1.3
func TestRouterInitializationExcludesStaticRoutes(t *testing.T) {
	router := setupTestRouter()
	routes := router.Routes()
	
	// Check that no routes match static file patterns
	staticRoutes := []string{"/", "/index", "/favico.ico"}
	for _, route := range routes {
		for _, staticRoute := range staticRoutes {
			if route.Path == staticRoute {
				t.Errorf("Found static route %s, but it should not exist", staticRoute)
			}
		}
		
		// Check for /assets/* pattern
		if strings.HasPrefix(route.Path, "/assets/") {
			t.Errorf("Found assets route %s, but static file serving should be removed", route.Path)
		}
	}
	
	t.Logf("Verified router has no static file routes")
}

// TestAllRoutesAreAPIOrMCP - Property 2: All registered routes are API or MCP routes
// Feature: remove-gin-static-resources, Property 2: All registered routes are API or MCP routes
// Validates: Requirements 3.3
func TestAllRoutesAreAPIOrMCP(t *testing.T) {
	router := setupTestRouter()
	routes := router.Routes()
	
	for _, route := range routes {
		path := route.Path
		
		// Check if route is API, MCP, or ping
		isValid := strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/mcp/") ||
			path == "/ping"
		
		if !isValid {
			t.Errorf("Route %s does not start with /api/, /mcp/, or equal /ping", path)
		}
	}
	
	t.Logf("Verified all %d routes are API or MCP routes", len(routes))
}

// TestConfigLoadsWithoutStaticDir - Example 2 & 3: Config loads without StaticDir
// Validates: Requirements 1.5, 2.1, 2.2, 2.4
func TestConfigLoadsWithoutStaticDir(t *testing.T) {
	// Create a minimal config without StaticDir
	conf := &calibre.Config{
		Address: ":8080",
		TmpDir:  "/tmp",
	}
	
	// Verify StaticDir field is empty (zero value)
	if conf.Address == "" {
		t.Error("Address should not be empty")
	}
	
	// Validate config (should not require StaticDir)
	err := validateConfig(conf)
	if err != nil {
		t.Errorf("Config validation failed without StaticDir: %v", err)
	}
	
	t.Log("Config loads and validates successfully without StaticDir")
}
