package tasks

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Manager struct {
	tasks   map[string]Task
	history []TaskStatus
	mu      sync.RWMutex
}

var (
	instance *Manager
	once     sync.Once
)

func GetManager() *Manager {
	once.Do(func() {
		instance = &Manager{
			tasks:   make(map[string]Task),
			history: make([]TaskStatus, 0),
		}
	})
	return instance
}

func (m *Manager) StartTask(t TaskType, mode TaskMode, factory func(string) Task) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if task of same type is running
	// Exception: DeleteBook and UpdateMetadata tasks can run concurrently
	if t != TaskTypeDeleteBook && t != TaskTypeUpdateMetadata && t != TaskTypeCheckMissing {
		for _, task := range m.tasks {
			status := task.GetStatus()
			if status.Type == t && status.State == "running" {
				return "", fmt.Errorf("task of type %s is already running", t)
			}
		}
	}

	id := uuid.New().String()
	task := factory(id)
	m.tasks[id] = task

	go func() {
		err := task.Run()
		m.mu.Lock()
		defer m.mu.Unlock()

		status := task.GetStatus()
		status.EndTime = time.Now()
		if err != nil {
			status.State = "error"
			status.Error = err.Error()
		} else if status.State == "running" {
			status.State = "completed"
			status.Progress = 100
		}

		// Move to history
		m.history = append([]TaskStatus{status}, m.history...)
		if len(m.history) > 50 {
			m.history = m.history[:50]
		}
		delete(m.tasks, id)
	}()

	return id, nil
}

func (m *Manager) StopTask(id string) error {
	m.mu.RLock()
	task, exists := m.tasks[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("task not found")
	}

	task.Stop()
	return nil
}

func (m *Manager) GetTasks() []TaskStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []TaskStatus

	// Active tasks
	for _, task := range m.tasks {
		result = append(result, task.GetStatus())
	}

	// History
	result = append(result, m.history...)

	return result
}
