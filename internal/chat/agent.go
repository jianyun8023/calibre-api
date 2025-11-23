package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/jianyun8023/calibre-api/internal/semantic/qdrant"
	"github.com/tmc/langchaingo/llms"
)

// Agent 智能书库 Agent
type Agent struct {
	llm          llms.Model
	searcher     *qdrant.Searcher
	tocFetcher   TocFetcher
	systemPrompt string
}

// NewAgent 创建智能书库 Agent
func NewAgent(llm llms.Model, searcher *qdrant.Searcher, tocFetcher TocFetcher) *Agent {
	systemPrompt := `你是 Calibre 书库的智能助手。你可以帮助用户：
1. 推荐书籍：根据用户兴趣推荐相关书籍
2. 搜索书籍：理解自然语言查询，帮助用户找到想要的书
3. 总结书籍：根据书籍目录（TOC）提供内容摘要
4. 回答问题：回答关于书库的一般性问题

当用户询问书籍推荐或搜索时，你应该：
- 理解用户的需求和兴趣
- 使用语义搜索找到相关书籍
- 用友好的语言介绍推荐的书籍

当用户询问某本书的内容或目录时，你应该：
- 获取书籍的目录结构
- 根据目录总结书籍的主要内容

请用简洁、友好的语言回答用户。`

	return &Agent{
		llm:          llm,
		searcher:     searcher,
		tocFetcher:   tocFetcher,
		systemPrompt: systemPrompt,
	}
}

// Chat 处理用户消息（同步）
func (a *Agent) Chat(ctx context.Context, userMessage string, conversationHistory []*Message) (string, error) {
	// 构建提示词
	prompt := a.buildPrompt(userMessage, conversationHistory)

	// 生成回复
	response, err := GenerateCompletion(ctx, a.llm, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

// ChatStream 处理用户消息（流式）
func (a *Agent) ChatStream(ctx context.Context, userMessage string, conversationHistory []*Message, callback func(string) error) error {
	// 构建提示词
	prompt := a.buildPrompt(userMessage, conversationHistory)

	// 流式生成回复
	err := GenerateCompletionStream(ctx, a.llm, prompt, callback)
	if err != nil {
		return fmt.Errorf("failed to generate streaming response: %w", err)
	}

	return nil
}

// buildPrompt 构建提示词
func (a *Agent) buildPrompt(userMessage string, conversationHistory []*Message) string {
	var prompt strings.Builder

	// 系统提示
	prompt.WriteString(a.systemPrompt)
	prompt.WriteString("\n\n")

	// 对话历史（最近 10 条）
	historyStart := 0
	if len(conversationHistory) > 10 {
		historyStart = len(conversationHistory) - 10
	}

	for i := historyStart; i < len(conversationHistory); i++ {
		msg := conversationHistory[i]
		if msg.Role == "user" {
			prompt.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
		} else if msg.Role == "assistant" {
			prompt.WriteString(fmt.Sprintf("助手: %s\n", msg.Content))
		}
	}

	// 当前用户消息
	prompt.WriteString(fmt.Sprintf("\n用户: %s\n助手: ", userMessage))

	return prompt.String()
}

// ShouldUseSearch 判断是否需要使用搜索
func (a *Agent) ShouldUseSearch(userMessage string) bool {
	keywords := []string{"推荐", "搜索", "找", "有哪些", "关于"}
	lowerMessage := strings.ToLower(userMessage)

	for _, keyword := range keywords {
		if strings.Contains(lowerMessage, keyword) {
			return true
		}
	}
	return false
}

// ShouldUseToc 判断是否需要获取目录
func (a *Agent) ShouldUseToc(userMessage string) (bool, int64) {
	// 简单的关键词匹配，实际应用中可能需要更复杂的意图识别
	keywords := []string{"目录", "大纲", "内容", "讲了什么", "总结"}
	lowerMessage := strings.ToLower(userMessage)

	hasKeyword := false
	for _, keyword := range keywords {
		if strings.Contains(lowerMessage, keyword) {
			hasKeyword = true
			break
		}
	}

	if !hasKeyword {
		return false, 0
	}

	// 尝试从消息中提取书籍 ID（假设用户会提及 ID 或上下文中有 ID）
	// 这里简化处理，假设消息中包含 ID，或者由调用者（Handler）从上下文中提供
	// 实际中可能需要先搜索书籍，然后获取最佳匹配的书籍 ID
	// 为了演示，我们尝试从消息中解析数字作为 ID
	// 注意：这非常脆弱，仅用于演示。更好的方式是先搜索，然后让 LLM 决定查看哪本书的目录。
	// 或者，如果用户在书籍详情页提问，前端可以传递书籍 ID。
	// 目前我们先返回 false，由 Handler 决定是否调用搜索来找到书，然后再获取目录。
	// 但为了支持 "总结 ID 为 123 的书"，我们可以尝试提取数字。

	// 简单正则提取数字（这里用 fmt.Sscanf 简化）
	// 实际场景：用户说 "总结一下这本书"，如果上下文中有书，则使用上下文。
	// 这里我们暂不实现复杂的 ID 提取，而是依赖 Handler 或 Agent 的搜索结果。
	// 如果用户明确说了 ID，例如 "总结书籍 123"，我们可以提取。

	// 暂时只支持通过 SearchAndRespond 流程中的意图识别，或者在 Handler 中处理。
	// 这里返回 false，让 Handler 逻辑决定。
	// 实际上，我们可以修改 SearchAndRespond，让它在搜索到书籍后，如果用户问的是内容，就自动获取目录。
	return false, 0
}

// SearchAndRespond 搜索并生成回复（增强版，支持 TOC）
func (a *Agent) SearchAndRespond(ctx context.Context, userMessage string) (string, []map[string]interface{}, error) {
	// 1. 执行语义搜索
	results, err := a.searcher.Search(userMessage, 24) // 获取更多结果以支持"换一换"
	if err != nil {
		return "", nil, fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		return "抱歉，我没有找到相关的书籍。", nil, nil
	}

	// 2. 判断用户意图是否是询问内容/总结
	isAskingContent := false
	contentKeywords := []string{"目录", "大纲", "内容", "讲了什么", "总结", "介绍"}
	for _, kw := range contentKeywords {
		if strings.Contains(userMessage, kw) {
			isAskingContent = true
			break
		}
	}

	// 3. 提取书籍信息
	books := make([]map[string]interface{}, 0)
	var contextInfo strings.Builder

	// 构建所有书籍的元数据
	for _, result := range results {
		book := map[string]interface{}{
			"id":      result.Book.ID,
			"title":   result.Book.Title,
			"authors": result.Book.Authors,
			"score":   result.Score,
		}
		books = append(books, book)
	}

	if isAskingContent && len(results) > 0 {
		// 如果是询问内容，且找到了书，尝试获取第一本书的 TOC
		topBook := results[0].Book
		contextInfo.WriteString(fmt.Sprintf("找到最相关的书籍：《%s》\n", topBook.Title))

		if a.tocFetcher != nil {
			tocStr, err := a.tocFetcher(ctx, topBook.ID)
			if err == nil {
				// 截断 TOC 以防止过长
				if len(tocStr) > 2000 {
					tocStr = tocStr[:2000] + "...(截断)"
				}
				contextInfo.WriteString(fmt.Sprintf("目录摘要：\n%s\n", tocStr))
			} else {
				contextInfo.WriteString("(无法获取目录信息)\n")
			}
		}
	} else {
		contextInfo.WriteString("找到以下相关书籍（仅列出前 8 本）：\n")
		// 仅将前 8 本书的信息提供给 LLM，避免上下文过长
		limit := 8
		if len(results) < limit {
			limit = len(results)
		}

		for i := 0; i < limit; i++ {
			result := results[i]
			contextInfo.WriteString(fmt.Sprintf("%d. 《%s》 - %v (相关度: %.2f)\n",
				i+1, result.Book.Title, result.Book.Authors, result.Score))
		}
	}

	// 4. 生成回复
	prompt := fmt.Sprintf(`%s

%s

请根据用户的需求"%s"，利用上述信息进行回答。如果是推荐书籍，请简要介绍。如果是询问内容，请根据目录进行总结。`,
		a.systemPrompt, contextInfo.String(), userMessage)

	response, err := GenerateCompletion(ctx, a.llm, prompt)
	if err != nil {
		return "", books, fmt.Errorf("failed to generate response: %w", err)
	}

	return response, books, nil
}

// GenerateTitle 根据对话内容生成标题
func (a *Agent) GenerateTitle(ctx context.Context, userMessage, aiResponse string) (string, error) {
	prompt := fmt.Sprintf(`请根据以下对话内容，生成一个简短的标题（不超过 15 个字）。
用户：%s
AI：%s
标题：`, userMessage, aiResponse)

	// 使用简单的 Chat 接口生成
	resp, err := a.Chat(ctx, prompt, nil)
	if err != nil {
		return "", err
	}

	// 清理标题（去除引号等）
	title := strings.TrimSpace(resp)
	title = strings.Trim(title, `"'《》`)
	if len([]rune(title)) > 15 {
		title = string([]rune(title)[:15])
	}
	return title, nil
}
