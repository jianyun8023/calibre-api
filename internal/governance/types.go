package governance

import (
	"encoding/json"
	"time"
)

// DraftStatus 草稿状态
type DraftStatus string

const (
	DraftStatusPending  DraftStatus = "pending"  // 待审核
	DraftStatusApproved DraftStatus = "approved" // 已批准
	DraftStatusRejected DraftStatus = "rejected" // 已拒绝
	DraftStatusApplied  DraftStatus = "applied"  // 已应用
	DraftStatusSkipped  DraftStatus = "skipped"  // 已跳过
)

// SuggestedAction 建议操作
type SuggestedAction string

const (
	ActionAutoApply SuggestedAction = "auto_apply" // 自动应用
	ActionReview    SuggestedAction = "review"     // 需审核
	ActionSkip      SuggestedAction = "skip"       // 建议跳过
)

// MetadataSource 元数据来源
type MetadataSource string

const (
	SourceCopyrightExtract MetadataSource = "copyright_extract" // 版权页抽取
	SourceDouban           MetadataSource = "douban"            // 豆瓣搜索
	SourceManual           MetadataSource = "manual"            // 手动编辑
)

// MetadataField 元数据字段
type MetadataField string

const (
	FieldISBN        MetadataField = "isbn"
	FieldTitle       MetadataField = "title"
	FieldAuthors     MetadataField = "authors"
	FieldPublisher   MetadataField = "publisher"
	FieldPublishDate MetadataField = "pubdate"
)

// DraftFlag 草稿标记
type DraftFlag string

const (
	FlagCollectionSuspected DraftFlag = "collection_suspected"  // 疑似合辑
	FlagMultipleISBN        DraftFlag = "multiple_isbn"         // 多个ISBN
	FlagISBNInvalidChecksum DraftFlag = "isbn_invalid_checksum" // ISBN校验失败
	FlagTitleTooLong        DraftFlag = "title_too_long"        // 标题过长
	FlagMultipleAuthors     DraftFlag = "multiple_authors"      // 多作者
	FlagMagazineSuspected   DraftFlag = "magazine_suspected"    // 疑似杂志
)

// ConfidenceBreakdown 置信度分解
type ConfidenceBreakdown struct {
	ISBNScore         float64 `json:"isbn_score"`         // ISBN可靠性分数
	ContextScore      float64 `json:"context_score"`      // 上下文质量分数
	ComplexityPenalty float64 `json:"complexity_penalty"` // 复杂度惩罚
	FinalScore        float64 `json:"final_score"`        // 最终分数
	Details           string  `json:"details,omitempty"`  // 详细说明
}

// MetadataDraft 元数据草稿
type MetadataDraft struct {
	ID        int64  `json:"id"`
	BookID    int64  `json:"book_id"`
	BookTitle string `json:"book_title"` // 冗余，方便展示

	// 变更内容
	Field    MetadataField `json:"field"`
	OldValue string        `json:"old_value"`
	NewValue string        `json:"new_value"`

	// 来源和质量
	Source              MetadataSource       `json:"source"`
	Confidence          float64              `json:"confidence"`
	ConfidenceBreakdown *ConfidenceBreakdown `json:"confidence_breakdown,omitempty"`
	Flags               []DraftFlag          `json:"flags,omitempty"`

	// 状态
	Status          DraftStatus     `json:"status"`
	SuggestedAction SuggestedAction `json:"suggested_action"`

	// 任务关联
	SessionID string `json:"session_id,omitempty"`

	// 并发控制
	Version int `json:"version"`

	// 时间戳
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy string     `json:"reviewed_by,omitempty"`
	AppliedAt  *time.Time `json:"applied_at,omitempty"`
}

// MetadataChangelog 元数据变更日志
type MetadataChangelog struct {
	ID        int64  `json:"id"`
	BookID    int64  `json:"book_id"`
	BookTitle string `json:"book_title"`

	// 变更内容
	Field    MetadataField `json:"field"`
	OldValue string        `json:"old_value"`
	NewValue string        `json:"new_value"`

	// 来源
	Source  MetadataSource `json:"source"`
	DraftID *int64         `json:"draft_id,omitempty"`

	// 时间戳
	AppliedAt time.Time `json:"applied_at"`
	AppliedBy string    `json:"applied_by,omitempty"`

	// 回滚信息
	RevertedAt   *time.Time `json:"reverted_at,omitempty"`
	RevertedBy   string     `json:"reverted_by,omitempty"`
	RevertReason string     `json:"revert_reason,omitempty"`
}

// SessionState 会话状态
type SessionState string

const (
	SessionStateRunning   SessionState = "running"
	SessionStateCompleted SessionState = "completed"
	SessionStateFailed    SessionState = "failed"
	SessionStateCancelled SessionState = "cancelled"
)

// ExtractionSession 抽取会话
type ExtractionSession struct {
	ID string `json:"id"`

	// 任务信息
	TaskType string `json:"task_type"` // copyright_extract, batch_validate
	Mode     string `json:"mode"`      // full, incremental, dry_run

	// 统计
	TotalBooks    int `json:"total_books"`
	Processed     int `json:"processed"`
	Success       int `json:"success"`
	Failed        int `json:"failed"`
	Skipped       int `json:"skipped"`
	AutoApproved  int `json:"auto_approved"`
	PendingReview int `json:"pending_review"`

	// 状态
	State        SessionState `json:"state"`
	ErrorMessage string       `json:"error_message,omitempty"`

	// 时间戳
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// DraftFilter 草稿查询过滤器
type DraftFilter struct {
	Status        *DraftStatus   `json:"status,omitempty"`
	ConfidenceMin *float64       `json:"confidence_min,omitempty"`
	ConfidenceMax *float64       `json:"confidence_max,omitempty"`
	HasFlags      *bool          `json:"has_flags,omitempty"`
	SessionID     string         `json:"session_id,omitempty"`
	BookID        *int64         `json:"book_id,omitempty"`
	Field         *MetadataField `json:"field,omitempty"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
}

// ChangelogFilter 变更日志查询过滤器
type ChangelogFilter struct {
	BookID   *int64         `json:"book_id,omitempty"`
	Field    *MetadataField `json:"field,omitempty"`
	FromDate *time.Time     `json:"from_date,omitempty"`
	ToDate   *time.Time     `json:"to_date,omitempty"`
	Reverted *bool          `json:"reverted,omitempty"`
	Limit    int            `json:"limit"`
	Offset   int            `json:"offset"`
}

// GovernanceStats 治理统计
type GovernanceStats struct {
	Drafts struct {
		Pending  int `json:"pending"`
		Approved int `json:"approved"`
		Rejected int `json:"rejected"`
		Applied  int `json:"applied"`
	} `json:"drafts"`

	ConfidenceDistribution struct {
		High   int `json:"high"`   // >= 0.8
		Medium int `json:"medium"` // 0.5 ~ 0.8
		Low    int `json:"low"`    // < 0.5
	} `json:"confidence_distribution"`

	BySource map[MetadataSource]int `json:"by_source"`

	FlagsCount map[DraftFlag]int `json:"flags_count"`
}

// BatchApproveRequest 批量审核请求
type BatchApproveRequest struct {
	Items []BatchItem `json:"items"`
}

// BatchItem 批量操作项
type BatchItem struct {
	ID      int64 `json:"id"`
	Version int   `json:"version"`
}

// BatchResult 批量操作结果
type BatchResult struct {
	Success   []int64 `json:"success"`
	Conflicts []int64 `json:"conflicts"`
	Errors    []int64 `json:"errors"`
}

// ApplyRequest 应用请求
type ApplyRequest struct {
	DraftIDs []int64 `json:"draft_ids"`
}

// RevertRequest 回滚请求
type RevertRequest struct {
	Reason string `json:"reason"`
}

// Helper functions

// FlagsToJSON 将标记列表转换为JSON字符串
func FlagsToJSON(flags []DraftFlag) string {
	if len(flags) == 0 {
		return "[]"
	}
	data, _ := json.Marshal(flags)
	return string(data)
}

// JSONToFlags 将JSON字符串转换为标记列表
func JSONToFlags(data string) []DraftFlag {
	if data == "" || data == "[]" {
		return nil
	}
	var flags []DraftFlag
	json.Unmarshal([]byte(data), &flags)
	return flags
}

// ConfidenceBreakdownToJSON 将置信度分解转换为JSON字符串
func ConfidenceBreakdownToJSON(breakdown *ConfidenceBreakdown) string {
	if breakdown == nil {
		return ""
	}
	data, _ := json.Marshal(breakdown)
	return string(data)
}

// JSONToConfidenceBreakdown 将JSON字符串转换为置信度分解
func JSONToConfidenceBreakdown(data string) *ConfidenceBreakdown {
	if data == "" {
		return nil
	}
	var breakdown ConfidenceBreakdown
	if err := json.Unmarshal([]byte(data), &breakdown); err != nil {
		return nil
	}
	return &breakdown
}
