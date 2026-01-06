package cache

import (
	"sync/atomic"
	"time"
)

// CacheMetrics 缓存性能指标
type CacheMetrics struct {
	hits         atomic.Int64 // 缓存命中次数
	misses       atomic.Int64 // 缓存未命中次数
	sets         atomic.Int64 // 缓存设置次数
	evictions    atomic.Int64 // 缓存淘汰次数
	startTime    time.Time    // 统计开始时间
	lastResetAt  time.Time    // 上次重置时间
}

// NewCacheMetrics 创建缓存指标
func NewCacheMetrics() *CacheMetrics {
	now := time.Now()
	return &CacheMetrics{
		startTime:   now,
		lastResetAt: now,
	}
}

// RecordHit 记录缓存命中
func (m *CacheMetrics) RecordHit() {
	m.hits.Add(1)
}

// RecordMiss 记录缓存未命中
func (m *CacheMetrics) RecordMiss() {
	m.misses.Add(1)
}

// RecordSet 记录缓存设置
func (m *CacheMetrics) RecordSet() {
	m.sets.Add(1)
}

// RecordEviction 记录缓存淘汰
func (m *CacheMetrics) RecordEviction() {
	m.evictions.Add(1)
}

// Stats 返回统计信息
func (m *CacheMetrics) Stats() map[string]interface{} {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses
	
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	uptime := time.Since(m.startTime)
	timeSinceReset := time.Since(m.lastResetAt)

	return map[string]interface{}{
		"hits":                hits,
		"misses":              misses,
		"total_requests":      total,
		"hit_rate_percent":    hitRate,
		"sets":                m.sets.Load(),
		"evictions":           m.evictions.Load(),
		"uptime_seconds":      uptime.Seconds(),
		"time_since_reset_s":  timeSinceReset.Seconds(),
	}
}

// Reset 重置统计信息
func (m *CacheMetrics) Reset() {
	m.hits.Store(0)
	m.misses.Store(0)
	m.sets.Store(0)
	m.evictions.Store(0)
	m.lastResetAt = time.Now()
}

// HitRate 返回缓存命中率（0-1）
func (m *CacheMetrics) HitRate() float64 {
	hits := m.hits.Load()
	misses := m.misses.Load()
	total := hits + misses
	
	if total == 0 {
		return 0
	}
	
	return float64(hits) / float64(total)
}

