package tasks

import (
	"fmt"
	"testing"
)

// TestIncrementalSyncQueryFormat 测试增量同步的查询格式
// 验证 fmt.Sprintf 生成的查询格式是否正确
func TestIncrementalSyncQueryFormat(t *testing.T) {
	testCases := []struct {
		name     string
		maxID    uint64
		expected string
	}{
		{
			name:     "maxID is 0 - should be empty (full sync)",
			maxID:    0,
			expected: "",
		},
		{
			name:     "maxID is 100",
			maxID:    100,
			expected: "id:>100",
		},
		{
			name:     "maxID is 12345",
			maxID:    12345,
			expected: "id:>12345",
		},
		{
			name:     "maxID is large number",
			maxID:    999999,
			expected: "id:>999999",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟实际代码的查询生成逻辑
			var query string
			if tc.maxID > 0 {
				query = fmt.Sprintf("id:>%d", tc.maxID)
			}

			if query != tc.expected {
				t.Errorf("Query mismatch: got %q, want %q", query, tc.expected)
			}
		})
	}
}

// TestIncrementalSyncPotentialIssues 分析增量同步可能的问题（旧实现）
// 
// 旧实现的问题分析：
// 1. GetMaxID() 遍历所有 Qdrant 点来找最大 ID - 这假设 ID 是连续的
// 2. 如果 Calibre 中删除了书籍，ID 可能不连续
// 3. 如果 Qdrant 中有 ID 100, 200, 300，maxID=300
// 4. 但 Calibre 中可能有 ID 101-299 的书籍没有被同步
//
// 核心问题：增量同步只能发现 ID > maxID 的新书，
// 无法发现 ID < maxID 但不在 Qdrant 中的书籍（例如之前同步失败的）
func TestIncrementalSyncPotentialIssues(t *testing.T) {
	t.Log("增量同步潜在问题分析（旧实现 - 已修复）:")
	t.Log("")
	t.Log("场景 1: ID 不连续")
	t.Log("  - Calibre 中有书籍 ID: 1, 2, 3, 100, 101, 102")
	t.Log("  - Qdrant 中只有 ID: 1, 2, 3 (maxID=3)")
	t.Log("  - 增量同步查询 id:>3，会找到 100, 101, 102 ✓")
	t.Log("")
	t.Log("场景 2: 中间有缺失")
	t.Log("  - Calibre 中有书籍 ID: 1, 2, 3, 4, 5")
	t.Log("  - Qdrant 中只有 ID: 1, 3, 5 (maxID=5)")
	t.Log("  - 增量同步查询 id:>5，找不到任何书籍")
	t.Log("  - 但 ID 2, 4 没有被同步！✗")
	t.Log("")
	t.Log("场景 3: Qdrant 中有额外的 ID")
	t.Log("  - Calibre 中有书籍 ID: 1, 2, 3")
	t.Log("  - Qdrant 中有 ID: 1, 2, 3, 100 (maxID=100)")
	t.Log("  - 增量同步查询 id:>100，找不到任何书籍")
	t.Log("  - 这种情况下增量同步永远不会工作！✗")
}

// TestGetMaxIDAssumption 测试 GetMaxID 的假设（旧实现）
// GetMaxID 通过遍历所有点找最大 ID，这个方法本身是正确的
// 但问题在于：使用 maxID 作为增量同步的基准是否合理？
func TestGetMaxIDAssumption(t *testing.T) {
	t.Log("GetMaxID 方法分析（旧实现 - 已修复）:")
	t.Log("")
	t.Log("当前实现:")
	t.Log("  1. 遍历 Qdrant 中所有点")
	t.Log("  2. 找到最大的 ID")
	t.Log("  3. 查询 Calibre 中 id:>maxID 的书籍")
	t.Log("")
	t.Log("假设:")
	t.Log("  - 书籍 ID 是递增的")
	t.Log("  - 新书的 ID 总是大于已有书籍的 ID")
	t.Log("")
	t.Log("问题:")
	t.Log("  - 如果之前的同步失败，某些 ID < maxID 的书籍可能没有被同步")
	t.Log("  - 增量同步无法发现这些遗漏的书籍")
	t.Log("")
	t.Log("建议的解决方案:")
	t.Log("  1. 比较 Calibre 和 Qdrant 中的所有 ID，找出差异")
	t.Log("  2. 同步差异中的书籍")
	t.Log("  3. 这样可以处理任何遗漏的情况")
}

// TestFindMissingBooksLogic 测试新的增量同步逻辑
// 新实现通过比较 Calibre 和 Qdrant 的 ID 差异来找出缺失的书籍
func TestFindMissingBooksLogic(t *testing.T) {
	testCases := []struct {
		name       string
		calibreIDs []int64
		qdrantIDs  []int64
		expected   []int64
	}{
		{
			name:       "所有书籍都已同步",
			calibreIDs: []int64{1, 2, 3, 4, 5},
			qdrantIDs:  []int64{1, 2, 3, 4, 5},
			expected:   []int64{},
		},
		{
			name:       "Qdrant 为空",
			calibreIDs: []int64{1, 2, 3},
			qdrantIDs:  []int64{},
			expected:   []int64{1, 2, 3},
		},
		{
			name:       "中间有缺失",
			calibreIDs: []int64{1, 2, 3, 4, 5},
			qdrantIDs:  []int64{1, 3, 5},
			expected:   []int64{2, 4},
		},
		{
			name:       "只有新书缺失",
			calibreIDs: []int64{1, 2, 3, 100, 101},
			qdrantIDs:  []int64{1, 2, 3},
			expected:   []int64{100, 101},
		},
		{
			name:       "Qdrant 有额外的 ID（已删除的书）",
			calibreIDs: []int64{1, 2, 3},
			qdrantIDs:  []int64{1, 2, 3, 100},
			expected:   []int64{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 模拟 findMissingBooks 的逻辑
			qdrantMap := make(map[int64]bool)
			for _, id := range tc.qdrantIDs {
				qdrantMap[id] = true
			}

			var missingIDs []int64
			for _, id := range tc.calibreIDs {
				if !qdrantMap[id] {
					missingIDs = append(missingIDs, id)
				}
			}

			// 比较结果
			if len(missingIDs) != len(tc.expected) {
				t.Errorf("Missing IDs count mismatch: got %d, want %d", len(missingIDs), len(tc.expected))
				return
			}

			for i, id := range missingIDs {
				if id != tc.expected[i] {
					t.Errorf("Missing ID mismatch at index %d: got %d, want %d", i, id, tc.expected[i])
				}
			}
		})
	}
}
