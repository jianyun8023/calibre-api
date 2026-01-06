package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// SearchCache 搜索结果缓存
// 使用 LRU 策略管理缓存条目，自动淘汰最久未使用的条目
type SearchCache struct {
	cache      map[string]*cacheEntry
	accessList *accessList // 双向链表跟踪访问顺序
	mu         sync.RWMutex
	maxSize    int           // 最大缓存条目数
	ttl        time.Duration // 缓存过期时间
}

// cacheEntry 缓存条目
type cacheEntry struct {
	key        string
	value      interface{}
	expireTime time.Time
	node       *accessNode // 链表节点
}

// accessNode 访问链表节点
type accessNode struct {
	key  string
	prev *accessNode
	next *accessNode
}

// accessList 访问顺序双向链表（LRU）
type accessList struct {
	head *accessNode // 最近访问
	tail *accessNode // 最久未访问
}

// NewSearchCache 创建搜索缓存
// maxSize: 最大缓存条目数，0 表示无限制
// ttl: 缓存过期时间，0 表示永不过期
func NewSearchCache(maxSize int, ttl time.Duration) *SearchCache {
	if maxSize <= 0 {
		maxSize = 1000 // 默认 1000 条
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute // 默认 5 分钟
	}

	return &SearchCache{
		cache:      make(map[string]*cacheEntry),
		accessList: &accessList{},
		maxSize:    maxSize,
		ttl:        ttl,
	}
}

// Get 获取缓存
func (c *SearchCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(entry.expireTime) {
		c.removeEntry(key)
		return nil, false
	}

	// 更新访问顺序（移到链表头部）
	c.accessList.moveToFront(entry.node)

	return entry.value, true
}

// Set 设置缓存
func (c *SearchCache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已存在
	if entry, exists := c.cache[key]; exists {
		entry.value = value
		entry.expireTime = time.Now().Add(c.ttl)
		c.accessList.moveToFront(entry.node)
		return
	}

	// 检查是否超过最大容量
	if len(c.cache) >= c.maxSize {
		// 淘汰最久未使用的条目（尾部）
		c.evictLRU()
	}

	// 创建新条目
	node := c.accessList.addToFront(key)
	entry := &cacheEntry{
		key:        key,
		value:      value,
		expireTime: time.Now().Add(c.ttl),
		node:       node,
	}
	c.cache[key] = entry
}

// Delete 删除缓存
func (c *SearchCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.removeEntry(key)
}

// Clear 清空缓存
func (c *SearchCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cacheEntry)
	c.accessList = &accessList{}
}

// Size 返回当前缓存条目数
func (c *SearchCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// Stats 返回缓存统计信息
func (c *SearchCache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"size":     len(c.cache),
		"max_size": c.maxSize,
		"ttl_ms":   c.ttl.Milliseconds(),
	}
}

// removeEntry 移除缓存条目（内部使用，需要持有锁）
func (c *SearchCache) removeEntry(key string) {
	if entry, exists := c.cache[key]; exists {
		c.accessList.remove(entry.node)
		delete(c.cache, key)
	}
}

// evictLRU 淘汰最久未使用的条目
func (c *SearchCache) evictLRU() {
	if c.accessList.tail == nil {
		return
	}

	key := c.accessList.tail.key
	c.removeEntry(key)
}

// CleanExpired 清理过期条目（可定期调用）
func (c *SearchCache) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range c.cache {
		if now.After(entry.expireTime) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		c.removeEntry(key)
	}

	return len(expiredKeys)
}

// accessList 方法

func (l *accessList) addToFront(key string) *accessNode {
	node := &accessNode{key: key}

	if l.head == nil {
		// 空链表
		l.head = node
		l.tail = node
	} else {
		// 添加到头部
		node.next = l.head
		l.head.prev = node
		l.head = node
	}

	return node
}

func (l *accessList) moveToFront(node *accessNode) {
	if node == l.head {
		return // 已经在头部
	}

	// 从当前位置移除
	l.remove(node)

	// 添加到头部
	node.prev = nil
	node.next = l.head
	l.head.prev = node
	l.head = node
}

func (l *accessList) remove(node *accessNode) {
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		// node 是头部
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		// node 是尾部
		l.tail = node.prev
	}

	node.prev = nil
	node.next = nil
}

// HashKey 生成缓存键（哈希）
// 用于将查询参数转换为缓存键
func HashKey(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte(part))
		h.Write([]byte(":"))
	}
	return hex.EncodeToString(h.Sum(nil))
}

