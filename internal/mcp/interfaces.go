package mcp

// CalibreAPI 定义了 MCP 工具需要的 Calibre API 接口
// 这样可以避免循环导入问题
type CalibreAPI interface {
	// 这里定义 MCP 工具需要的方法接口
	// 暂时为空，因为我们主要使用 HTTP API 集成
}

// BookSearcher 定义书籍搜索接口
type BookSearcher interface {
	SearchBooks(query string, limit, offset int) (interface{}, error)
}

// BookGetter 定义获取书籍接口
type BookGetter interface {
	GetBook(id string) (interface{}, error)
}

// MetadataSearcher 定义元数据搜索接口
type MetadataSearcher interface {
	SearchMetadata(query string) (interface{}, error)
}
