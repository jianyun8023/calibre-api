package calibre

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// getEnhancedTools 获取增强工具列表
func (c *Api) getEnhancedTools(context *gin.Context) {
	// 创建增强工具管理器
	etm := NewEnhancedToolManager(c)
	tools := etm.GetEnhancedTools()

	context.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": tools,
	})
}

// executeEnhancedTool 执行增强工具
func (c *Api) executeEnhancedTool(context *gin.Context) {
	toolName := context.Param("tool")

	// 解析请求参数
	var args map[string]interface{}
	if err := context.ShouldBindJSON(&args); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数解析失败: " + err.Error(),
		})
		return
	}

	// 创建增强工具管理器
	etm := NewEnhancedToolManager(c)

	// 执行工具
	result, err := etm.ExecuteEnhancedTool(toolName, args)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "工具执行失败: " + err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
	})
}
