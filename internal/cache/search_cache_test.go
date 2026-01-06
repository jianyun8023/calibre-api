package cache

import (
	"testing"
	"time"
)

func TestSearchCache_BasicOperations(t *testing.T) {
	cache := NewSearchCache(3, 1*time.Second)

	// 测试 Set 和 Get
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if val, found := cache.Get("key1"); !found || val != "value1" {
		t.Errorf("Expected value1, got %v, found=%v", val, found)
	}

	if val, found := cache.Get("key2"); !found || val != "value2" {
		t.Errorf("Expected value2, got %v, found=%v", val, found)
	}

	// 测试不存在的键
	if _, found := cache.Get("nonexistent"); found {
		t.Error("Expected key not found")
	}
}

func TestSearchCache_LRUEviction(t *testing.T) {
	cache := NewSearchCache(3, 10*time.Second)

	// 添加 3 个条目
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// 访问 key1 使其成为最近使用
	cache.Get("key1")

	// 添加第 4 个条目，应该淘汰 key2（最久未使用）
	cache.Set("key4", "value4")

	// key2 应该被淘汰
	if _, found := cache.Get("key2"); found {
		t.Error("Expected key2 to be evicted")
	}

	// key1, key3, key4 应该还在
	if _, found := cache.Get("key1"); !found {
		t.Error("Expected key1 to exist")
	}
	if _, found := cache.Get("key3"); !found {
		t.Error("Expected key3 to exist")
	}
	if _, found := cache.Get("key4"); !found {
		t.Error("Expected key4 to exist")
	}
}

func TestSearchCache_Expiration(t *testing.T) {
	cache := NewSearchCache(10, 100*time.Millisecond)

	// 添加条目
	cache.Set("key1", "value1")

	// 立即获取应该成功
	if _, found := cache.Get("key1"); !found {
		t.Error("Expected key1 to exist immediately")
	}

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 现在应该过期了
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be expired")
	}
}

func TestSearchCache_Delete(t *testing.T) {
	cache := NewSearchCache(10, 1*time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// 删除 key1
	cache.Delete("key1")

	// key1 应该不存在
	if _, found := cache.Get("key1"); found {
		t.Error("Expected key1 to be deleted")
	}

	// key2 应该还在
	if _, found := cache.Get("key2"); !found {
		t.Error("Expected key2 to exist")
	}
}

func TestSearchCache_Clear(t *testing.T) {
	cache := NewSearchCache(10, 1*time.Second)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// 清空缓存
	cache.Clear()

	// 所有键都应该不存在
	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}

	if _, found := cache.Get("key1"); found {
		t.Error("Expected cache to be cleared")
	}
}

func TestSearchCache_CleanExpired(t *testing.T) {
	cache := NewSearchCache(10, 100*time.Millisecond)

	// 添加多个条目
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	// 等待过期
	time.Sleep(150 * time.Millisecond)

	// 清理过期条目
	count := cache.CleanExpired()

	if count != 3 {
		t.Errorf("Expected 3 expired entries, got %d", count)
	}

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after cleanup, got %d", cache.Size())
	}
}

func TestSearchCache_Concurrency(t *testing.T) {
	cache := NewSearchCache(100, 1*time.Second)

	// 并发写入
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				key := string(rune('a' + n*10 + j))
				cache.Set(key, n*10+j)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证缓存大小
	if cache.Size() > 100 {
		t.Errorf("Cache size exceeded max: %d", cache.Size())
	}
}

func TestHashKey(t *testing.T) {
	key1 := HashKey("search", "golang", "10")
	key2 := HashKey("search", "golang", "10")
	key3 := HashKey("search", "python", "10")

	// 相同参数应该生成相同的哈希
	if key1 != key2 {
		t.Error("Expected same hash for same parameters")
	}

	// 不同参数应该生成不同的哈希
	if key1 == key3 {
		t.Error("Expected different hash for different parameters")
	}
}

func BenchmarkSearchCache_Set(b *testing.B) {
	cache := NewSearchCache(1000, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(string(rune(i%1000)), i)
	}
}

func BenchmarkSearchCache_Get(b *testing.B) {
	cache := NewSearchCache(1000, 5*time.Minute)

	// 预填充缓存
	for i := 0; i < 1000; i++ {
		cache.Set(string(rune(i)), i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(string(rune(i % 1000)))
	}
}

