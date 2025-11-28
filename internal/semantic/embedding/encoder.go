package embedding

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jianyun8023/calibre-api/internal/semantic"
)

// NewProvider creates a new embedding provider based on configuration
func NewProvider(config ProviderConfig) (Provider, error) {
	switch config.Provider {
	case "ollama":
		return NewOllamaProvider(config.Ollama), nil
	case "siliconflow":
		return NewSiliconFlowProvider(config.SiliconFlow), nil
	default:
		return nil, fmt.Errorf("unsupported embedding provider: %s", config.Provider)
	}
}

// CombineBookText 组合书籍文本用于向量化
// 优化策略：
// 1. 标题权重增强：标题重复一次以增加权重
// 2. 标签权重降低：标签放在最后，且不重复
// 3. 摘要截断：限制摘要长度，避免过长文本影响向量化质量
// 4. HTML 标签过滤：彻底清理 HTML 标签和实体
// 5. 文本清理：确保 UTF-8 有效，限制总长度
func CombineBookText(book semantic.Book) string {
	var parts []string

	// 1. 标题（增加权重：重复一次）
	if title := strings.TrimSpace(book.Title); title != "" {
		title = cleanText(title)
		if title != "" {
			parts = append(parts, title)
			parts = append(parts, title) // 重复标题以增加权重
		}
	}

	// 2. 作者
	if len(book.Authors) > 0 {
		authors := strings.Join(book.Authors, ", ")
		authors = cleanText(authors)
		if authors != "" {
			parts = append(parts, authors)
		}
	}

	// 3. 摘要（过滤 HTML 并截断）
	if summary := strings.TrimSpace(book.Comments); summary != "" {
		summary = stripHTML(summary)
		summary = truncateText(summary, 500) // 截断到 500 字符
		summary = cleanText(summary)
		if strings.TrimSpace(summary) != "" {
			parts = append(parts, summary)
		}
	}

	// 4. 标签（降低权重：放在最后，且不重复）
	if len(book.Tags) > 0 {
		// 如果标签过多，只取前几个
		tags := book.Tags
		if len(tags) > 5 {
			tags = tags[:5] // 最多保留 5 个标签
		}
		tagStr := strings.Join(tags, ", ")
		tagStr = cleanText(tagStr)
		if tagStr != "" {
			parts = append(parts, tagStr)
		}
	}

	result := strings.Join(parts, " | ")

	// 限制总长度（API 可能有长度限制，32K tokens 约等于 8000-10000 字符）
	// 为了安全，限制到 8000 字符
	if len(result) > 8000 {
		result = truncateText(result, 8000)
	}

	return result
}

// cleanText 清理文本，确保 UTF-8 有效并移除控制字符
func cleanText(text string) string {
	// 移除无效的 UTF-8 字符
	text = strings.ToValidUTF8(text, "")

	// 移除控制字符（保留换行符和制表符）
	var builder strings.Builder
	for _, r := range text {
		// 保留可打印字符、空格、换行符、制表符
		if r >= 32 || r == '\n' || r == '\t' {
			builder.WriteRune(r)
		}
	}

	return strings.TrimSpace(builder.String())
}

// stripHTML 移除 HTML 标签和实体
func stripHTML(html string) string {
	// 移除 HTML 标签（包括自闭合标签）
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, "")

	// 移除 XML/HTML 注释标记和 CDATA 结束标记
	text = strings.ReplaceAll(text, "]]>", "")
	text = strings.ReplaceAll(text, "<![CDATA[", "")
	text = strings.ReplaceAll(text, "<!--", "")
	text = strings.ReplaceAll(text, "-->", "")

	// 解码常见 HTML 实体
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&apos;", "'")
	text = strings.ReplaceAll(text, "&mdash;", "—")
	text = strings.ReplaceAll(text, "&ndash;", "–")
	text = strings.ReplaceAll(text, "&hellip;", "...")
	text = strings.ReplaceAll(text, "&ldquo;", "\"")
	text = strings.ReplaceAll(text, "&rdquo;", "\"")
	text = strings.ReplaceAll(text, "&lsquo;", "'")
	text = strings.ReplaceAll(text, "&rsquo;", "'")

	// 清理多余的空白字符（包括换行符、制表符等）
	text = regexp.MustCompile(`[\s\n\r\t]+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// truncateText 截断文本到指定长度，保留完整单词
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	// 在 maxLen 处截断
	truncated := text[:maxLen]

	// 尝试在最后一个空格处截断，保留完整单词
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLen*2/3 {
		truncated = truncated[:lastSpace]
	}

	// 如果截断后还有内容，添加省略号
	if len(text) > maxLen {
		truncated += "..."
	}

	return truncated
}
