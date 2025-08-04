package calibre

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeCover    ResourceType = "cover"    // 封面图片
	ResourceTypeContent  ResourceType = "content"  // 书籍内容
	ResourceTypeToc      ResourceType = "toc"      // 目录
	ResourceTypeMetadata ResourceType = "metadata" // 元数据
	ResourceTypeFile     ResourceType = "file"     // 文件
)

// Resource 资源信息
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MimeType    string                 `json:"mimeType"`
	Size        int64                  `json:"size,omitempty"`
	Created     time.Time              `json:"created,omitempty"`
	Updated     time.Time              `json:"updated,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// ResourceManager MCP 资源管理器
type ResourceManager struct {
	api *Api
}

// NewResourceManager 创建资源管理器
func NewResourceManager(api *Api) *ResourceManager {
	return &ResourceManager{
		api: api,
	}
}

// ListResources 列出可用资源
func (rm *ResourceManager) ListResources(bookID string) ([]Resource, error) {
	var resources []Resource

	// 获取书籍信息
	book, err := rm.api.getBookByID(bookID)
	if err != nil {
		return nil, fmt.Errorf("获取书籍信息失败: %w", err)
	}

	// 封面资源
	if book.Cover != "" {
		resources = append(resources, Resource{
			URI:         fmt.Sprintf("calibre://books/%s/cover", bookID),
			Name:        fmt.Sprintf("%s_cover", book.Title),
			Description: fmt.Sprintf("《%s》的封面图片", book.Title),
			MimeType:    "image/jpeg",
			Created:     book.LastModified,
			Updated:     book.LastModified,
		})
	}

	// 目录资源
	resources = append(resources, Resource{
		URI:         fmt.Sprintf("calibre://books/%s/toc", bookID),
		Name:        fmt.Sprintf("%s_toc", book.Title),
		Description: fmt.Sprintf("《%s》的目录结构", book.Title),
		MimeType:    "application/json",
		Created:     book.LastModified,
		Updated:     book.LastModified,
	})

	// 元数据资源
	resources = append(resources, Resource{
		URI:         fmt.Sprintf("calibre://books/%s/metadata", bookID),
		Name:        fmt.Sprintf("%s_metadata", book.Title),
		Description: fmt.Sprintf("《%s》的完整元数据", book.Title),
		MimeType:    "application/json",
		Created:     book.LastModified,
		Updated:     book.LastModified,
	})

	// 书籍文件资源
	if book.FilePath != "" {
		ext := strings.ToLower(filepath.Ext(book.FilePath))
		var mimeType string
		switch ext {
		case ".epub":
			mimeType = "application/epub+zip"
		case ".pdf":
			mimeType = "application/pdf"
		case ".mobi":
			mimeType = "application/x-mobipocket-ebook"
		default:
			mimeType = "application/octet-stream"
		}

		resources = append(resources, Resource{
			URI:         fmt.Sprintf("calibre://books/%s/file", bookID),
			Name:        fmt.Sprintf("%s_file%s", book.Title, ext),
			Description: fmt.Sprintf("《%s》的电子书文件", book.Title),
			MimeType:    mimeType,
			Size:        book.Size,
			Created:     book.LastModified,
			Updated:     book.LastModified,
		})
	}

	return resources, nil
}

// ReadResource 读取资源内容
func (rm *ResourceManager) ReadResource(uri string) (*Resource, error) {
	// 解析 URI: calibre://books/{id}/{type}
	parts := strings.Split(strings.TrimPrefix(uri, "calibre://"), "/")
	if len(parts) != 4 || parts[0] != "" || parts[1] != "books" {
		return nil, fmt.Errorf("无效的资源 URI: %s", uri)
	}

	bookID := parts[2]
	resourceType := ResourceType(parts[3])

	// 获取书籍信息
	book, err := rm.api.getBookByID(bookID)
	if err != nil {
		return nil, fmt.Errorf("获取书籍信息失败: %w", err)
	}

	switch resourceType {
	case ResourceTypeCover:
		return rm.readCoverResource(book, bookID)
	case ResourceTypeToc:
		return rm.readTocResource(book, bookID)
	case ResourceTypeMetadata:
		return rm.readMetadataResource(book, bookID)
	case ResourceTypeFile:
		return rm.readFileResource(book, bookID)
	default:
		return nil, fmt.Errorf("不支持的资源类型: %s", resourceType)
	}
}

// readCoverResource 读取封面资源
func (rm *ResourceManager) readCoverResource(book *Book, bookID string) (*Resource, error) {
	if book.Cover == "" {
		return nil, fmt.Errorf("书籍没有封面")
	}

	// 获取封面数据
	coverData, err := rm.api.getCoverData(bookID)
	if err != nil {
		return nil, fmt.Errorf("获取封面数据失败: %w", err)
	}

	return &Resource{
		URI:         fmt.Sprintf("calibre://books/%s/cover", bookID),
		Name:        fmt.Sprintf("%s_cover", book.Title),
		Description: fmt.Sprintf("《%s》的封面图片", book.Title),
		MimeType:    "image/jpeg",
		Created:     book.LastModified,
		Updated:     book.LastModified,
		Data: map[string]interface{}{
			"base64": base64.StdEncoding.EncodeToString(coverData),
		},
	}, nil
}

// readTocResource 读取目录资源
func (rm *ResourceManager) readTocResource(book *Book, bookID string) (*Resource, error) {
	// 获取目录数据
	tocData, err := rm.api.getBookTocData(bookID)
	if err != nil {
		return nil, fmt.Errorf("获取目录数据失败: %w", err)
	}

	return &Resource{
		URI:         fmt.Sprintf("calibre://books/%s/toc", bookID),
		Name:        fmt.Sprintf("%s_toc", book.Title),
		Description: fmt.Sprintf("《%s》的目录结构", book.Title),
		MimeType:    "application/json",
		Created:     book.LastModified,
		Updated:     book.LastModified,
		Data: map[string]interface{}{
			"toc": tocData,
		},
	}, nil
}

// readMetadataResource 读取元数据资源
func (rm *ResourceManager) readMetadataResource(book *Book, bookID string) (*Resource, error) {
	return &Resource{
		URI:         fmt.Sprintf("calibre://books/%s/metadata", bookID),
		Name:        fmt.Sprintf("%s_metadata", book.Title),
		Description: fmt.Sprintf("《%s》的完整元数据", book.Title),
		MimeType:    "application/json",
		Created:     book.LastModified,
		Updated:     book.LastModified,
		Data: map[string]interface{}{
			"metadata": book,
		},
	}, nil
}

// readFileResource 读取文件资源
func (rm *ResourceManager) readFileResource(book *Book, bookID string) (*Resource, error) {
	if book.FilePath == "" {
		return nil, fmt.Errorf("书籍没有文件")
	}

	// 获取文件数据
	fileData, err := rm.api.getBookFileData(bookID)
	if err != nil {
		return nil, fmt.Errorf("获取文件数据失败: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(book.FilePath))
	var mimeType string
	switch ext {
	case ".epub":
		mimeType = "application/epub+zip"
	case ".pdf":
		mimeType = "application/pdf"
	case ".mobi":
		mimeType = "application/x-mobipocket-ebook"
	default:
		mimeType = "application/octet-stream"
	}

	return &Resource{
		URI:         fmt.Sprintf("calibre://books/%s/file", bookID),
		Name:        fmt.Sprintf("%s_file%s", book.Title, ext),
		Description: fmt.Sprintf("《%s》的电子书文件", book.Title),
		MimeType:    mimeType,
		Size:        book.Size,
		Created:     book.LastModified,
		Updated:     book.LastModified,
		Data: map[string]interface{}{
			"base64": base64.StdEncoding.EncodeToString(fileData),
		},
	}, nil
}

// 辅助方法 - 获取书籍信息
func (api *Api) getBookByID(id string) (*Book, error) {
	var book Book
	err := api.currentIndex().GetDocument(id, nil, &book)
	if err != nil {
		return nil, err
	}
	return &book, nil
}

// 辅助方法 - 获取封面数据
func (api *Api) getCoverData(bookID string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/get/cover/%s", api.config.MCP.BaseURL, bookID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// 辅助方法 - 获取目录数据
func (api *Api) getBookTocData(bookID string) (interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/read/%s/toc", api.config.MCP.BaseURL, bookID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// 辅助方法 - 获取文件数据
func (api *Api) getBookFileData(bookID string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/get/book/%s", api.config.MCP.BaseURL, bookID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
