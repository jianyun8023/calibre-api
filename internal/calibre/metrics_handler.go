package calibre

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/cache"
)

// MetricsHandler 性能指标处理器
type MetricsHandler struct {
	cachedSearcher *cache.CachedSearcher
	startTime      time.Time
}

// NewMetricsHandler 创建性能指标处理器
func NewMetricsHandler(cachedSearcher *cache.CachedSearcher) *MetricsHandler {
	return &MetricsHandler{
		cachedSearcher: cachedSearcher,
		startTime:      time.Now(),
	}
}

// GetMetrics 获取性能指标
// GET /api/metrics
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := map[string]interface{}{
		"server": map[string]interface{}{
			"uptime_seconds":   time.Since(h.startTime).Seconds(),
			"goroutines":       runtime.NumGoroutine(),
			"go_version":       runtime.Version(),
		},
		"memory": map[string]interface{}{
			"alloc_mb":        float64(m.Alloc) / 1024 / 1024,
			"total_alloc_mb":  float64(m.TotalAlloc) / 1024 / 1024,
			"sys_mb":          float64(m.Sys) / 1024 / 1024,
			"heap_alloc_mb":   float64(m.HeapAlloc) / 1024 / 1024,
			"heap_sys_mb":     float64(m.HeapSys) / 1024 / 1024,
			"heap_idle_mb":    float64(m.HeapIdle) / 1024 / 1024,
			"heap_inuse_mb":   float64(m.HeapInuse) / 1024 / 1024,
			"heap_released_mb": float64(m.HeapReleased) / 1024 / 1024,
			"gc_runs":         m.NumGC,
			"gc_pause_ns":     m.PauseNs[(m.NumGC+255)%256],
		},
		"timestamp": time.Now().Unix(),
	}

	// 添加缓存统计（如果可用）
	if h.cachedSearcher != nil {
		cacheStats := h.cachedSearcher.GetCacheStats()
		metrics["cache"] = cacheStats
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": metrics,
	})
}

// ClearCache 清除搜索缓存
// POST /api/metrics/cache/clear
func (h *MetricsHandler) ClearCache(c *gin.Context) {
	if h.cachedSearcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Cache service not available",
		})
		return
	}

	h.cachedSearcher.InvalidateCache()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Cache cleared successfully",
	})
}

// GetCacheStats 获取缓存统计
// GET /api/metrics/cache
func (h *MetricsHandler) GetCacheStats(c *gin.Context) {
	if h.cachedSearcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Cache service not available",
		})
		return
	}

	stats := h.cachedSearcher.GetCacheStats()

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
	})
}

// CleanExpiredCache 清理过期缓存
// POST /api/metrics/cache/clean
func (h *MetricsHandler) CleanExpiredCache(c *gin.Context) {
	if h.cachedSearcher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "Cache service not available",
		})
		return
	}

	count := h.cachedSearcher.CleanExpiredCache()

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Expired cache cleaned",
		"data": map[string]interface{}{
			"cleaned_entries": count,
		},
	})
}

// GetHealthCheck 健康检查
// GET /api/health
func (h *MetricsHandler) GetHealthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status": "healthy",
		"uptime": time.Since(h.startTime).Seconds(),
	}

	// 简单的健康检查
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 如果内存使用过高，标记为 degraded
	allocMB := float64(m.Alloc) / 1024 / 1024
	if allocMB > 1024 { // 超过 1GB
		health["status"] = "degraded"
		health["reason"] = "high memory usage"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": health,
	})
}

