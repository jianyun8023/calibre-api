package calibre

import "time"

// SearchRequest 搜索请求参数
type SearchRequest struct {
	Q                       string `form:"q" json:"q" jsonschema:"description=搜索关键词,required"`
	Limit                   int    `form:"limit,default=20" json:"limit,omitempty" jsonschema:"description=每页结果数量,minimum=1,maximum=100"`
	Offset                  int    `form:"offset,default=0" json:"offset,omitempty" jsonschema:"description=结果偏移量,minimum=0"`
	Filter                  string `form:"filter" json:"filter,omitempty" jsonschema:"description=过滤条件"`
	Sort                    string `form:"sort" json:"sort,omitempty" jsonschema:"description=排序字段"`
	Facets                  string `form:"facets" json:"facets,omitempty" jsonschema:"description=分面搜索字段"`
	Highlight               string `form:"highlight" json:"highlight,omitempty" jsonschema:"description=高亮字段"`
	Attributes              string `form:"attributes" json:"attributes,omitempty" jsonschema:"description=返回属性"`
	AttributesToHighlight   string `form:"attributesToHighlight" json:"attributesToHighlight,omitempty" jsonschema:"description=高亮属性"`
	AttributesToCrop        string `form:"attributesToCrop" json:"attributesToCrop,omitempty" jsonschema:"description=裁剪属性"`
	CropLength              int    `form:"cropLength,default=50" json:"cropLength,omitempty" jsonschema:"description=裁剪长度"`
	AttributesToRetrieve    string `form:"attributesToRetrieve" json:"attributesToRetrieve,omitempty" jsonschema:"description=检索属性"`
	ShowRankingScore        bool   `form:"showRankingScore" json:"showRankingScore,omitempty" jsonschema:"description=显示排名分数"`
	ShowRankingScoreDetails bool   `form:"showRankingScoreDetails" json:"showRankingScoreDetails,omitempty" jsonschema:"description=显示排名分数详情"`
	Vector                  string `form:"vector" json:"vector,omitempty" jsonschema:"description=向量搜索"`
	Hybrid                  string `form:"hybrid" json:"hybrid,omitempty" jsonschema:"description=混合搜索参数"`
}

// BookUpdateRequest 书籍更新请求参数
type BookUpdateRequest struct {
	Title       string            `json:"title,omitempty" jsonschema:"description=书籍标题"`
	Authors     []string          `json:"authors,omitempty" jsonschema:"description=作者列表"`
	AuthorSort  string            `json:"author_sort,omitempty" jsonschema:"description=作者排序"`
	Comments    string            `json:"comments,omitempty" jsonschema:"description=书籍评论"`
	Publisher   string            `json:"publisher,omitempty" jsonschema:"description=出版社"`
	PubDate     time.Time         `json:"pubdate,omitempty" jsonschema:"description=出版日期"`
	Isbn        string            `json:"isbn,omitempty" jsonschema:"description=ISBN号"`
	Languages   []string          `json:"languages,omitempty" jsonschema:"description=语言列表"`
	Tags        []string          `json:"tags,omitempty" jsonschema:"description=标签列表"`
	SeriesIndex float64           `json:"series_index,omitempty" jsonschema:"description=系列索引"`
	Rating      float64           `json:"rating,omitempty" jsonschema:"description=评分,minimum=0,maximum=5"`
	Identifiers map[string]string `json:"identifiers,omitempty" jsonschema:"description=标识符映射"`
}

// MetadataSearchRequest 元数据搜索请求参数
type MetadataSearchRequest struct {
	Query string `form:"query" json:"query" jsonschema:"description=搜索查询,required"`
	Limit int    `form:"limit,default=10" json:"limit,omitempty" jsonschema:"description=结果数量限制,minimum=1,maximum=50"`
}

// IndexUpdateRequest 索引更新请求参数
type IndexUpdateRequest struct {
	Force bool `form:"force" json:"force,omitempty" jsonschema:"description=强制更新索引"`
}

// IndexSwitchRequest 索引切换请求参数
type IndexSwitchRequest struct {
	Index string `form:"index" json:"index" jsonschema:"description=目标索引名称,required"`
}

// PublisherListRequest 出版社列表请求参数
type PublisherListRequest struct {
	Limit  int `form:"limit,default=50" json:"limit,omitempty" jsonschema:"description=结果数量限制,minimum=1,maximum=100"`
	Offset int `form:"offset,default=0" json:"offset,omitempty" jsonschema:"description=结果偏移量,minimum=0"`
}

// RecentlyBooksRequest 最近书籍请求参数
type RecentlyBooksRequest struct {
	Limit  int `form:"limit,default=10" json:"limit,omitempty" jsonschema:"description=结果数量限制,minimum=1,maximum=50"`
	Offset int `form:"offset,default=0" json:"offset,omitempty" jsonschema:"description=结果偏移量,minimum=0"`
}

// RandomBooksRequest 随机书籍请求参数
type RandomBooksRequest struct {
	Limit  int `form:"limit,default=10" json:"limit,omitempty" jsonschema:"description=结果数量限制,minimum=1,maximum=50"`
	Offset int `form:"offset,default=0" json:"offset,omitempty" jsonschema:"description=结果偏移量,minimum=0"`
}

// BookIDParam 书籍ID路径参数
type BookIDParam struct {
	ID string `uri:"id" json:"id" jsonschema:"description=书籍ID,required"`
}

// CoverPathParam 封面路径参数
type CoverPathParam struct {
	Path string `uri:"*path" json:"path" jsonschema:"description=封面文件路径,required"`
}

// BookContentPathParam 书籍内容路径参数
type BookContentPathParam struct {
	ID   string `uri:"id" json:"id" jsonschema:"description=书籍ID,required"`
	Path string `uri:"*path" json:"path" jsonschema:"description=内容文件路径,required"`
}

// ISBNParam ISBN参数
type ISBNParam struct {
	ISBN string `uri:"isbn" json:"isbn" jsonschema:"description=ISBN号,required"`
}
