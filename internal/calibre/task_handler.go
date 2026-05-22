package calibre

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jianyun8023/calibre-api/internal/tasks"
)

// StartTaskRequest 启动任务请求
type StartTaskRequest struct {
	Type string `json:"type"` // "semantic_sync"
	Mode string `json:"mode"` // "full" or "incremental"
}

// listTasks 获取任务列表
func (c *Api) listTasks(r *gin.Context) {
	manager := tasks.GetManager()
	taskList := manager.GetTasks()
	r.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": taskList,
	})
}

// startTask 启动任务
func (c *Api) startTask(r *gin.Context) {
	var req StartTaskRequest
	if err := r.ShouldBindJSON(&req); err != nil {
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request body",
		})
		return
	}

	switch tasks.TaskType(req.Type) {
	case tasks.TaskTypeSemanticSync:
		if c.semanticSearcher == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Search service is not initialized",
			})
			return
		}

		manager := tasks.GetManager()
		taskID, err := manager.StartTask(tasks.TaskTypeSemanticSync, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			// Using SearchSyncTask for semantic search synchronization
			return tasks.NewSearchSyncTask(id, tasks.TaskMode(req.Mode), c.contentApi, c.semanticSearcher)
		})

		if err != nil {
			r.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": err.Error(),
			})
			return
		}

		r.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Task started",
			"data":    gin.H{"id": taskID},
		})
		return

	case tasks.TaskTypeTocExtract:
		if c.cacheManager == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Cache manager is not initialized",
			})
			return
		}

		if c.semanticSearcher == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Search service is not initialized",
			})
			return
		}

		manager := tasks.GetManager()
		taskID, err := manager.StartTask(tasks.TaskTypeTocExtract, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			return tasks.NewTocExtractTask(id, tasks.TaskMode(req.Mode), c.contentApi, c.semanticSearcher, c.cacheManager)
		})

		if err != nil {
			r.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": err.Error(),
			})
			return
		}

		r.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "TOC extraction task started",
			"data":    gin.H{"id": taskID},
		})
		return

	case tasks.TaskTypeCheckMissing:
		if c.contentApi == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Content API is not initialized",
			})
			return
		}

		if c.semanticSearcher == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Search service is not initialized",
			})
			return
		}

		manager := tasks.GetManager()
		taskID, err := manager.StartTask(tasks.TaskTypeCheckMissing, tasks.TaskModeFull, func(id string) tasks.Task {
			return tasks.NewCheckMissingTask(id, c.contentApi, c.semanticSearcher)
		})

		if err != nil {
			r.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": err.Error(),
			})
			return
		}

		r.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Check missing vectors task started",
			"data":    gin.H{"id": taskID},
		})
		return

	case tasks.TaskTypeCleanupOrphans:
		if c.contentApi == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Content API is not initialized",
			})
			return
		}

		if c.semanticSearcher == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Search service is not initialized",
			})
			return
		}

		manager := tasks.GetManager()
		taskID, err := manager.StartTask(tasks.TaskTypeCleanupOrphans, tasks.TaskModeFull, func(id string) tasks.Task {
			return tasks.NewCleanupOrphansTask(id, c.contentApi, c.semanticSearcher)
		})

		if err != nil {
			r.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": err.Error(),
			})
			return
		}

		r.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Orphan cleanup task started",
			"data":    gin.H{"id": taskID},
		})
		return

	case tasks.TaskTypeCopyrightExtract:
		if c.contentApi == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Content API is not initialized",
			})
			return
		}

		if c.cacheManager == nil {
			r.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "Cache manager is not initialized",
			})
			return
		}

		manager := tasks.GetManager()
		taskID, err := manager.StartTask(tasks.TaskTypeCopyrightExtract, tasks.TaskMode(req.Mode), func(id string) tasks.Task {
			return tasks.NewCopyrightExtractTask(id, tasks.TaskMode(req.Mode), c.contentApi, c.cacheManager)
		})

		if err != nil {
			r.JSON(http.StatusConflict, gin.H{
				"code":    409,
				"message": err.Error(),
			})
			return
		}

		r.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Copyright extraction task started",
			"data":    gin.H{"id": taskID},
		})
		return

	default:
		r.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Unknown task type",
		})
		return
	}
}

// getTask 获取单个任务状态
func (c *Api) getTask(r *gin.Context) {
	id := r.Param("id")
	manager := tasks.GetManager()
	task, err := manager.GetTask(id)
	if err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": task,
	})
}

// stopTask 停止任务
func (c *Api) stopTask(r *gin.Context) {
	id := r.Param("id")
	manager := tasks.GetManager()
	if err := manager.StopTask(id); err != nil {
		r.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": err.Error(),
		})
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Task stopped",
	})
}
