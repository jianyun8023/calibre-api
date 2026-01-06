package tasks

import (
	"testing"
	"time"
)

func TestSSEManagerTaskList(t *testing.T) {
	// 获取任务管理器单例
	manager := GetManager()

	// 清理之前的状态
	manager.mu.Lock()
	manager.tasks = make(map[string]Task)
	manager.history = make([]TaskStatus, 0)
	manager.mu.Unlock()

	// 创建 SSE 管理器
	sseManager := NewSSEManager(manager, 10)
	manager.SetSSEManager(sseManager)

	// 创建一个测试任务（使用长时间运行的任务，确保它在检查时仍在运行）
	taskID, err := manager.StartTask(TaskTypeCheckMissing, TaskModeFull, func(id string) Task {
		return &mockTask{
			id:       id,
			taskType: TaskTypeCheckMissing,
			runFunc: func() error {
				time.Sleep(5 * time.Second) // 长时间运行
				return nil
			},
		}
	})
	if err != nil {
		t.Fatalf("Failed to start task: %v", err)
	}
	t.Logf("Started task: %s", taskID)

	// 等待任务开始
	time.Sleep(50 * time.Millisecond)

	// 验证任务列表
	tasks := manager.GetTasks()
	t.Logf("Task count: %d", len(tasks))
	for i, task := range tasks {
		t.Logf("Task[%d]: ID=%s, Type=%s, State=%s", i, task.ID, task.Type, task.State)
	}

	if len(tasks) == 0 {
		t.Error("Expected at least one task, got 0")
	}

	// 创建 SSE 客户端
	client := NewSSEClient("test-client")

	// 注册客户端
	if err := sseManager.RegisterClient(client); err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}

	// 等待初始任务列表消息
	select {
	case msg := <-client.Channel:
		t.Logf("Received message: type=%s, tasks=%d", msg.Type, len(msg.Tasks))
		if msg.Type != SSEMessageTypeTaskList {
			t.Errorf("Expected message type %s, got %s", SSEMessageTypeTaskList, msg.Type)
		}
		if len(msg.Tasks) == 0 {
			t.Error("Expected tasks in message, got 0")
		}
		for i, task := range msg.Tasks {
			t.Logf("Message Task[%d]: ID=%s, Type=%s, State=%s", i, task.ID, task.Type, task.State)
		}
	case <-time.After(3 * time.Second):
		t.Error("Timeout waiting for initial task list")
	}

	// 清理
	sseManager.UnregisterClient("test-client")
}
