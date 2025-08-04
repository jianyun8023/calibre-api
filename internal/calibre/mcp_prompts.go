package calibre

import (
	"fmt"
	"strings"
)

// Prompt 提示模板
type Prompt struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Prompt      string            `json:"prompt"`
	Arguments   map[string]string `json:"arguments,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// PromptManager MCP 提示管理器
type PromptManager struct {
	api *Api
}

// NewPromptManager 创建提示管理器
func NewPromptManager(api *Api) *PromptManager {
	return &PromptManager{
		api: api,
	}
}

// GetPrompts 获取所有可用的提示模板
func (pm *PromptManager) GetPrompts() []Prompt {
	return []Prompt{
		// 搜索相关提示
		{
			Name:        "search_books_by_topic",
			Description: "按主题搜索书籍的提示模板",
			Prompt:      "请帮我搜索关于 {{topic}} 的书籍。我想了解这个主题的最新发展和经典著作。",
			Arguments: map[string]string{
				"topic": "搜索主题，如：机器学习、Python编程、历史小说等",
			},
			Tags: []string{"搜索", "主题", "书籍推荐"},
		},
		{
			Name:        "search_books_by_author",
			Description: "按作者搜索书籍的提示模板",
			Prompt:      "我想找到作者 {{author}} 的所有作品，包括他的代表作和最新作品。",
			Arguments: map[string]string{
				"author": "作者姓名",
			},
			Tags: []string{"搜索", "作者", "作品集"},
		},
		{
			Name:        "search_books_by_isbn",
			Description: "按ISBN搜索书籍的提示模板",
			Prompt:      "请帮我查找ISBN为 {{isbn}} 的书籍信息，并获取其详细元数据。",
			Arguments: map[string]string{
				"isbn": "ISBN号码",
			},
			Tags: []string{"搜索", "ISBN", "元数据"},
		},

		// 书籍管理相关提示
		{
			Name:        "update_book_metadata",
			Description: "更新书籍元数据的提示模板",
			Prompt:      "请帮我更新书籍ID为 {{book_id}} 的元数据，包括标题、作者、标签等信息。",
			Arguments: map[string]string{
				"book_id": "书籍ID",
			},
			Tags: []string{"管理", "元数据", "更新"},
		},
		{
			Name:        "get_book_details",
			Description: "获取书籍详细信息的提示模板",
			Prompt:      "请为我提供书籍ID为 {{book_id}} 的完整信息，包括封面、目录、元数据等。",
			Arguments: map[string]string{
				"book_id": "书籍ID",
			},
			Tags: []string{"信息", "详情", "资源"},
		},
		{
			Name:        "delete_book",
			Description: "删除书籍的提示模板",
			Prompt:      "请帮我删除书籍ID为 {{book_id}} 的书籍，确认删除操作。",
			Arguments: map[string]string{
				"book_id": "书籍ID",
			},
			Tags: []string{"管理", "删除", "确认"},
		},

		// 推荐相关提示
		{
			Name:        "get_recent_books",
			Description: "获取最近更新书籍的提示模板",
			Prompt:      "请为我推荐最近更新的 {{limit}} 本书籍，我想了解最新的内容。",
			Arguments: map[string]string{
				"limit": "推荐数量，默认10本",
			},
			Tags: []string{"推荐", "最新", "更新"},
		},
		{
			Name:        "get_random_books",
			Description: "获取随机书籍推荐的提示模板",
			Prompt:      "请为我随机推荐 {{limit}} 本有趣的书籍，我想发现一些新的阅读内容。",
			Arguments: map[string]string{
				"limit": "推荐数量，默认10本",
			},
			Tags: []string{"推荐", "随机", "发现"},
		},

		// 元数据服务相关提示
		{
			Name:        "search_metadata_online",
			Description: "在线搜索元数据的提示模板",
			Prompt:      "请帮我在线搜索关于 {{query}} 的书籍元数据，包括豆瓣等来源的信息。",
			Arguments: map[string]string{
				"query": "搜索查询词",
			},
			Tags: []string{"元数据", "在线搜索", "豆瓣"},
		},
		{
			Name:        "get_metadata_by_isbn",
			Description: "根据ISBN获取元数据的提示模板",
			Prompt:      "请帮我根据ISBN {{isbn}} 获取书籍的详细元数据信息。",
			Arguments: map[string]string{
				"isbn": "ISBN号码",
			},
			Tags: []string{"元数据", "ISBN", "详细信息"},
		},

		// 系统管理相关提示
		{
			Name:        "update_search_index",
			Description: "更新搜索索引的提示模板",
			Prompt:      "请帮我更新搜索索引，确保所有书籍都能被正确搜索到。",
			Arguments:   map[string]string{},
			Tags:        []string{"系统", "索引", "维护"},
		},
		{
			Name:        "get_publishers",
			Description: "获取出版社列表的提示模板",
			Prompt:      "请为我提供所有出版社的列表，我想了解书库中的出版社分布。",
			Arguments:   map[string]string{},
			Tags:        []string{"统计", "出版社", "列表"},
		},

		// 高级功能提示
		{
			Name:        "analyze_book_collection",
			Description: "分析书籍收藏的提示模板",
			Prompt:      "请帮我分析我的书籍收藏，包括作者分布、出版社分布、主题分类等统计信息。",
			Arguments:   map[string]string{},
			Tags:        []string{"分析", "统计", "收藏"},
		},
		{
			Name:        "find_similar_books",
			Description: "查找相似书籍的提示模板",
			Prompt:      "请帮我找到与书籍ID {{book_id}} 相似的其他书籍，基于作者、主题、标签等特征。",
			Arguments: map[string]string{
				"book_id": "参考书籍ID",
			},
			Tags: []string{"推荐", "相似", "匹配"},
		},
		{
			Name:        "export_book_list",
			Description: "导出书籍列表的提示模板",
			Prompt:      "请帮我导出书籍列表，格式为 {{format}}，包含 {{fields}} 等字段。",
			Arguments: map[string]string{
				"format": "导出格式：JSON、CSV、XML等",
				"fields": "导出字段：标题、作者、ISBN、出版社等",
			},
			Tags: []string{"导出", "列表", "数据"},
		},
	}
}

// GetPromptByName 根据名称获取提示模板
func (pm *PromptManager) GetPromptByName(name string) (*Prompt, error) {
	prompts := pm.GetPrompts()
	for _, prompt := range prompts {
		if prompt.Name == name {
			return &prompt, nil
		}
	}
	return nil, fmt.Errorf("未找到提示模板: %s", name)
}

// GetPromptsByTag 根据标签获取提示模板
func (pm *PromptManager) GetPromptsByTag(tag string) []Prompt {
	var result []Prompt
	prompts := pm.GetPrompts()
	for _, prompt := range prompts {
		for _, promptTag := range prompt.Tags {
			if strings.Contains(strings.ToLower(promptTag), strings.ToLower(tag)) {
				result = append(result, prompt)
				break
			}
		}
	}
	return result
}

// RenderPrompt 渲染提示模板
func (pm *PromptManager) RenderPrompt(name string, args map[string]string) (string, error) {
	prompt, err := pm.GetPromptByName(name)
	if err != nil {
		return "", err
	}

	result := prompt.Prompt
	for key, value := range args {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// GetPromptSuggestions 获取提示建议
func (pm *PromptManager) GetPromptSuggestions(context string) []Prompt {
	var suggestions []Prompt
	prompts := pm.GetPrompts()

	context = strings.ToLower(context)

	for _, prompt := range prompts {
		// 检查名称和描述是否匹配上下文
		if strings.Contains(strings.ToLower(prompt.Name), context) ||
			strings.Contains(strings.ToLower(prompt.Description), context) {
			suggestions = append(suggestions, prompt)
			continue
		}

		// 检查标签是否匹配
		for _, tag := range prompt.Tags {
			if strings.Contains(strings.ToLower(tag), context) {
				suggestions = append(suggestions, prompt)
				break
			}
		}
	}

	return suggestions
}

// GetPromptUsage 获取提示使用示例
func (pm *PromptManager) GetPromptUsage(name string) (string, error) {
	prompt, err := pm.GetPromptByName(name)
	if err != nil {
		return "", err
	}

	var usage strings.Builder
	usage.WriteString(fmt.Sprintf("提示模板: %s\n", prompt.Name))
	usage.WriteString(fmt.Sprintf("描述: %s\n", prompt.Description))
	usage.WriteString(fmt.Sprintf("模板: %s\n", prompt.Prompt))

	if len(prompt.Arguments) > 0 {
		usage.WriteString("参数:\n")
		for key, desc := range prompt.Arguments {
			usage.WriteString(fmt.Sprintf("  - %s: %s\n", key, desc))
		}
	}

	if len(prompt.Tags) > 0 {
		usage.WriteString(fmt.Sprintf("标签: %s\n", strings.Join(prompt.Tags, ", ")))
	}

	return usage.String(), nil
}

// GetPromptCategories 获取提示分类
func (pm *PromptManager) GetPromptCategories() map[string][]Prompt {
	categories := make(map[string][]Prompt)
	prompts := pm.GetPrompts()

	for _, prompt := range prompts {
		if len(prompt.Tags) > 0 {
			category := prompt.Tags[0] // 使用第一个标签作为分类
			categories[category] = append(categories[category], prompt)
		}
	}

	return categories
}
