package calibre

import (
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/pkg/log"
)

// PProfHandler 性能分析处理器
// 提供 pprof 性能分析端点，用于诊断内存泄漏和性能问题
type PProfHandler struct{}

// NewPProfHandler 创建性能分析处理器
func NewPProfHandler() *PProfHandler {
	return &PProfHandler{}
}

// RegisterRoutes 注册 pprof 路由
// 路由: /debug/pprof/*
func (h *PProfHandler) RegisterRoutes(r *gin.Engine) {
	pprofGroup := r.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	log.Info("pprof endpoints registered at /debug/pprof")
}

// GetGoroutineInfo 获取 goroutine 信息（简化版）
// GET /api/debug/goroutines
func (h *PProfHandler) GetGoroutineInfo(c *gin.Context) {
	numGoroutines := runtime.NumGoroutine()

	// 获取 goroutine 栈信息
	buf := make([]byte, 1024*1024) // 1MB buffer
	n := runtime.Stack(buf, true)
	stackTrace := string(buf[:n])

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"goroutine_count": numGoroutines,
			"stack_trace":     stackTrace,
		},
	})
}

// TriggerGC 手动触发垃圾回收
// POST /api/debug/gc
func (h *PProfHandler) TriggerGC(c *gin.Context) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	// 触发 GC
	runtime.GC()

	runtime.ReadMemStats(&after)

	freed := before.Alloc - after.Alloc

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "GC triggered",
		"data": gin.H{
			"before_alloc_mb": float64(before.Alloc) / 1024 / 1024,
			"after_alloc_mb":  float64(after.Alloc) / 1024 / 1024,
			"freed_mb":        float64(freed) / 1024 / 1024,
			"gc_runs":         after.NumGC - before.NumGC,
		},
	})
}

// GetMemStats 获取详细内存统计
// GET /api/debug/memstats
func (h *PProfHandler) GetMemStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := map[string]interface{}{
		// 通用统计
		"alloc_mb":       float64(m.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(m.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(m.Sys) / 1024 / 1024,
		"num_gc":         m.NumGC,
		"goroutines":     runtime.NumGoroutine(),

		// 堆内存
		"heap": map[string]interface{}{
			"alloc_mb":    float64(m.HeapAlloc) / 1024 / 1024,
			"sys_mb":      float64(m.HeapSys) / 1024 / 1024,
			"idle_mb":     float64(m.HeapIdle) / 1024 / 1024,
			"inuse_mb":    float64(m.HeapInuse) / 1024 / 1024,
			"released_mb": float64(m.HeapReleased) / 1024 / 1024,
			"objects":     m.HeapObjects,
		},

		// GC 统计
		"gc": map[string]interface{}{
			"num_gc":              m.NumGC,
			"num_forced_gc":       m.NumForcedGC,
			"gc_cpu_fraction":     m.GCCPUFraction,
			"last_gc_time_unix":   m.LastGC,
			"pause_total_ns":      m.PauseTotalNs,
			"pause_recent_ns":     m.PauseNs[(m.NumGC+255)%256],
		},

		// 栈内存
		"stack": map[string]interface{}{
			"inuse_mb": float64(m.StackInuse) / 1024 / 1024,
			"sys_mb":   float64(m.StackSys) / 1024 / 1024,
		},

		// 其他
		"mallocs":    m.Mallocs,
		"frees":      m.Frees,
		"live_objs":  m.Mallocs - m.Frees,
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": stats,
	})
}

// CheckGoroutineLeak 检查潜在的 goroutine 泄漏
// GET /api/debug/check-leak
func (h *PProfHandler) CheckGoroutineLeak(c *gin.Context) {
	// 基准 goroutine 数量（主 goroutine + Gin 内部 goroutines）
	// 通常在服务启动后应该保持相对稳定
	baselineGoroutines := 10 // 可以根据实际情况调整

	currentGoroutines := runtime.NumGoroutine()
	suspicious := currentGoroutines > baselineGoroutines*2

	var status string
	var message string

	if suspicious {
		status = "warning"
		message = "Goroutine count is abnormally high, possible leak detected"
	} else {
		status = "ok"
		message = "Goroutine count is within normal range"
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"status":              status,
			"message":             message,
			"current_goroutines":  currentGoroutines,
			"baseline_goroutines": baselineGoroutines,
			"suspicious":          suspicious,
		},
	})
}

