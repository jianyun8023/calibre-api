package chat

import (
	"regexp"
	"strings"
)

// ThinkingResponse 包含思考过程和最终答案
type ThinkingResponse struct {
	Thinking string `json:"thinking,omitempty"` // 思考过程
	Answer   string `json:"answer"`             // 最终答案
}

var (
	thinkingRegex = regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`)
	answerRegex   = regexp.MustCompile(`(?s)<answer>(.*?)</answer>`)
)

// ParseThinkingResponse 解析包含思考标签的响应
func ParseThinkingResponse(content string) ThinkingResponse {
	var response ThinkingResponse

	// 提取思考过程
	if matches := thinkingRegex.FindStringSubmatch(content); len(matches) > 1 {
		response.Thinking = strings.TrimSpace(matches[1])
	}

	// 提取答案
	if matches := answerRegex.FindStringSubmatch(content); len(matches) > 1 {
		response.Answer = strings.TrimSpace(matches[1])
	} else {
		// 如果没有答案标签，使用整个内容（去除思考部分）
		response.Answer = strings.TrimSpace(thinkingRegex.ReplaceAllString(content, ""))
	}

	// 如果答案为空，使用原始内容
	if response.Answer == "" {
		response.Answer = content
	}

	return response
}

// HasThinkingTags 检查内容是否包含思考标签
func HasThinkingTags(content string) bool {
	return thinkingRegex.MatchString(content)
}

// StripThinkingTags 移除思考标签，只保留答案
func StripThinkingTags(content string) string {
	parsed := ParseThinkingResponse(content)
	return parsed.Answer
}
