package tasks

import (
	"sync"
	"testing"
	"time"
)

// mockTask 模拟任务用于测试
type mockTask struct {
	id       string
	runFunc  func() error
	stopped  bool
	mu       sync.Mutex
	taskType TaskType
}

func (m *mockTask) Run() error {
	if m.runFunc != nil {
		return m.runFunc()
	}
	time.Sleep(10 * time.Millisecond)
	return nil
}

func (m *mockTask) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopped = true
}

func (m *mockTask) GetStatus() TaskStatus {
	return TaskStatus{
		ID:      m.id,
		Type:    m.taskType,
		State:   TaskStateRunning,
		Message: "running",
	}
}

// TestManagerStartTask 测试启动任务
func TestManagerStartTask(t *testing.T) {
	manager := GetManager()

	// 清理
	manager.mu.Lock()
	manager.tasks = make(map[string]Task)
	manager.history = make([]TaskStatus, 0)
	manager.mu.Unlock()

	taskID, err := manager.StartTask(TaskTypeQdrantSync, TaskModeFull, func(id string) Task {
		return &mockTask{
			id:       id,
			taskType: TaskTypeQdrantSync,
			runFunc: func() error {
				return nil
			},
		}
	})

	if err != nil {
		t.Fatalf("StartTask failed: %v", err)
	}

	if taskID == "" {
		t.Fatal("taskID should not be empty")
	}

	// 等待任务完成
	time.Sleep(50 * time.Millisecond)

	// 检查任务已移到历史记录
	tasks := manager.GetTasks()
	found := false
	for _, task := range tasks {
		if task.ID == taskID {
			found = true
			if task.State != TaskStateCompleted {
				t.Errorf("Task state = %s, want %s", task.State, TaskStateCompleted)
			}
		}
	}

	if !found {
		t.Error("Task not found in history")
	}
}

// TestManagerConcurrentTasks 测试并发任务
func TestManagerConcurrentTasks(t *testing.T) {
	manager := GetManager()

	// 清理
	manager.mu.Lock()
	manager.tasks = make(map[string]Task)
	manager.history = make([]TaskStatus, 0)
	manager.mu.Unlock()

	// 启动多个 DeleteBook 任务（允许并发）
	var wg sync.WaitGroup
	taskCount := 5

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := manager.StartTask(TaskTypeDeleteBook, TaskModeFull, func(id string) Task {
				return &mockTask{
					id:       id,
					taskType: TaskTypeDeleteBook,
					runFunc: func() error {
						time.Sleep(20 * time.Millisecond)
						return nil
					},
				}
			})
			if err != nil {
				t.Errorf("StartTask %d failed: %v", idx, err)
			}
		}(i)
	}

	wg.Wait()

	// 等待所有任务完成
	time.Sleep(100 * time.Millisecond)

	tasks := manager.GetTasks()
	if len(tasks) < taskCount {
		t.Errorf("Expected at least %d tasks, got %d", taskCount, len(tasks))
	}
}

// TestManagerStopTask 测试停止任务
func TestManagerStopTask(t *testing.T) {
	manager := GetManager()

	// 清理
	manager.mu.Lock()
	manager.tasks = make(map[string]Task)
	manager.history = make([]TaskStatus, 0)
	manager.mu.Unlock()

	taskID, err := manager.StartTask(TaskTypeQdrantSync, TaskModeFull, func(id string) Task {
		return &mockTask{
			id:       id,
			taskType: TaskTypeQdrantSync,
			runFunc: func() error {
				time.Sleep(1 * time.Second) // 长时间运行
				return nil
			},
		}
	})

	if err != nil {
		t.Fatalf("StartTask failed: %v", err)
	}

	// 等待任务开始
	time.Sleep(10 * time.Millisecond)

	// 停止任务
	err = manager.StopTask(taskID)
	if err != nil {
		t.Errorf("StopTask failed: %v", err)
	}
}

// BenchmarkManagerStartTask 基准测试启动任务的性能
func BenchmarkManagerStartTask(b *testing.B) {
	manager := GetManager()

	// 清理
	manager.mu.Lock()
	manager.tasks = make(map[string]Task)
	manager.history = make([]TaskStatus, 0)
	manager.mu.Unlock()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.StartTask(TaskTypeDeleteBook, TaskModeFull, func(id string) Task {
			return &mockTask{
				id:       id,
				taskType: TaskTypeDeleteBook,
				runFunc:  func() error { return nil },
			}
		})
	}
}
