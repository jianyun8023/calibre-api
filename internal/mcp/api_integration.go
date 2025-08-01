package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jianyun8023/calibre-api/internal/calibre"
)

// APIIntegrationTools provides actual integration with Calibre API endpoints
type APIIntegrationTools struct {
	calibreAPI *calibre.Api
	baseURL    string
	client     *http.Client
}

func NewAPIIntegrationTools(calibreAPI *calibre.Api, baseURL string) *APIIntegrationTools {
	return &APIIntegrationTools{
		calibreAPI: calibreAPI,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// SearchBooksAPI calls the actual search API
func (t *APIIntegrationTools) SearchBooksAPI(args SearchBooksArgs) (*ToolCallResult, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("q", args.Query)

	if args.Limit > 0 {
		params.Set("limit", strconv.Itoa(args.Limit))
	}
	if args.Offset > 0 {
		params.Set("offset", strconv.Itoa(args.Offset))
	}
	if args.Sort != "" {
		params.Set("sort", args.Sort)
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/api/search?%s", t.baseURL, params.Encode())
	resp, err := t.client.Get(apiURL)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	if resp.StatusCode != http.StatusOK {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body)),
			}},
			IsError: true,
		}, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("解析响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	// Format response
	responseText := t.formatSearchResponse(&apiResp)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: responseText,
		}},
	}, nil
}

// GetBookAPI calls the actual get book API
func (t *APIIntegrationTools) GetBookAPI(args GetBookArgs) (*ToolCallResult, error) {
	apiURL := fmt.Sprintf("%s/api/book/%s", t.baseURL, args.ID)
	resp, err := t.client.Get(apiURL)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	if resp.StatusCode != http.StatusOK {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("书籍未找到或API请求失败，状态码: %d", resp.StatusCode),
			}},
			IsError: true,
		}, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("解析响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	responseText := t.formatBookResponse(&apiResp)

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: responseText,
		}},
	}, nil
}

// UpdateBookMetadataAPI calls the actual update metadata API
func (t *APIIntegrationTools) UpdateBookMetadataAPI(args UpdateBookArgs) (*ToolCallResult, error) {
	// Prepare request data
	updateData := make(map[string]interface{})

	if args.Title != "" {
		updateData["title"] = args.Title
	}
	if len(args.Authors) > 0 {
		updateData["authors"] = args.Authors
	}
	if args.Publisher != "" {
		updateData["publisher"] = args.Publisher
	}
	if args.ISBN != "" {
		updateData["isbn"] = args.ISBN
	}
	if args.Comments != "" {
		updateData["comments"] = args.Comments
	}
	if len(args.Tags) > 0 {
		updateData["tags"] = args.Tags
	}
	if args.Rating > 0 {
		updateData["rating"] = args.Rating
	}
	if !args.PubDate.IsZero() {
		updateData["pubdate"] = args.PubDate.Format(time.RFC3339)
	}

	jsonData, err := json.Marshal(updateData)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("请求数据序列化失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	// Make API request
	apiURL := fmt.Sprintf("%s/api/book/%s/update", t.baseURL, args.ID)
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("解析响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	if apiResp.Code != 200 {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("更新失败: %s", apiResp.Message),
			}},
			IsError: true,
		}, fmt.Errorf("update failed: %s", apiResp.Message)
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("书籍元数据更新成功: ID=%s", args.ID),
		}},
	}, nil
}

// DeleteBookAPI calls the actual delete book API
func (t *APIIntegrationTools) DeleteBookAPI(args DeleteBookArgs) (*ToolCallResult, error) {
	apiURL := fmt.Sprintf("%s/api/book/%s/delete", t.baseURL, args.ID)
	resp, err := t.client.Post(apiURL, "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("API 请求失败: %v", err),
			}},
			IsError: true,
		}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("读取响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("解析响应失败: %v", err),
			}},
			IsError: true,
		}, err
	}

	if apiResp.Code != 200 {
		return &ToolCallResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("删除失败: %s", apiResp.Message),
			}},
			IsError: true,
		}, fmt.Errorf("delete failed: %s", apiResp.Message)
	}

	return &ToolCallResult{
		Content: []Content{{
			Type: "text",
			Text: fmt.Sprintf("书籍删除成功: ID=%s", args.ID),
		}},
	}, nil
}

// Helper methods for formatting responses
func (t *APIIntegrationTools) formatSearchResponse(apiResp *APIResponse) string {
	if apiResp.Code != 200 {
		return fmt.Sprintf("搜索失败: %s", apiResp.Message)
	}

	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return "搜索响应格式错误"
	}

	records, ok := data["records"].([]interface{})
	if !ok {
		return "搜索结果格式错误"
	}

	total, _ := data["total"].(float64)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("搜索结果 (共 %.0f 本书):\n\n", total))

	for i, record := range records {
		if book, ok := record.(map[string]interface{}); ok {
			title, _ := book["title"].(string)
			authors, _ := book["authors"].([]interface{})
			publisher, _ := book["publisher"].(string)
			id, _ := book["id"].(float64)

			result.WriteString(fmt.Sprintf("%d. 《%s》\n", i+1, title))

			if len(authors) > 0 {
				authorStrs := make([]string, len(authors))
				for j, author := range authors {
					authorStrs[j] = fmt.Sprintf("%v", author)
				}
				result.WriteString(fmt.Sprintf("   作者: %s\n", strings.Join(authorStrs, ", ")))
			}

			if publisher != "" {
				result.WriteString(fmt.Sprintf("   出版社: %s\n", publisher))
			}

			result.WriteString(fmt.Sprintf("   ID: %.0f\n\n", id))
		}
	}

	return result.String()
}

func (t *APIIntegrationTools) formatBookResponse(apiResp *APIResponse) string {
	if apiResp.Code != 200 {
		return fmt.Sprintf("获取书籍信息失败: %s", apiResp.Message)
	}

	book, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return "书籍信息格式错误"
	}

	var result strings.Builder

	if title, ok := book["title"].(string); ok {
		result.WriteString(fmt.Sprintf("标题: %s\n", title))
	}

	if authors, ok := book["authors"].([]interface{}); ok && len(authors) > 0 {
		authorStrs := make([]string, len(authors))
		for i, author := range authors {
			authorStrs[i] = fmt.Sprintf("%v", author)
		}
		result.WriteString(fmt.Sprintf("作者: %s\n", strings.Join(authorStrs, ", ")))
	}

	if publisher, ok := book["publisher"].(string); ok && publisher != "" {
		result.WriteString(fmt.Sprintf("出版社: %s\n", publisher))
	}

	if isbn, ok := book["isbn"].(string); ok && isbn != "" {
		result.WriteString(fmt.Sprintf("ISBN: %s\n", isbn))
	}

	if comments, ok := book["comments"].(string); ok && comments != "" {
		result.WriteString(fmt.Sprintf("简介: %s\n", comments))
	}

	if tags, ok := book["tags"].([]interface{}); ok && len(tags) > 0 {
		tagStrs := make([]string, len(tags))
		for i, tag := range tags {
			tagStrs[i] = fmt.Sprintf("%v", tag)
		}
		result.WriteString(fmt.Sprintf("标签: %s\n", strings.Join(tagStrs, ", ")))
	}

	if rating, ok := book["rating"].(float64); ok && rating > 0 {
		result.WriteString(fmt.Sprintf("评分: %.1f\n", rating))
	}

	if id, ok := book["id"].(float64); ok {
		result.WriteString(fmt.Sprintf("ID: %.0f\n", id))
	}

	return result.String()
}
