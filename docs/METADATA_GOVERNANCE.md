# 元数据治理系统设计

> 状态: 设计草案  
> 版本: 0.1.0  
> 日期: 2026-01-05

## 1. 背景与目标

### 1.1 问题陈述

当前书库存在以下元数据质量问题：

| 问题 | 来源 | 影响 |
|------|------|------|
| ISBN 缺失 | EPUB 原始元数据不完整 | 无法关联外部数据源 |
| 标题污染 | 过长描述、营销文案混入标题 | 搜索和展示体验差 |
| 作者/出版社错误 | OPF 元数据填写错误 | 信息不准确 |
| 垃圾书籍 | 低质量来源 | 污染书库 |

### 1.2 目标

1. **新书入库**：自动抽取/提示补全元数据
2. **历史矫正**：批量筛选问题书籍，审核后修正
3. **安全可控**：高置信度自动应用，低置信度人工审核
4. **可追溯**：完整变更历史，支持回滚

### 1.3 约束

- 豆瓣 API 有访问限制，不能批量调用
- 仅处理中文书籍
- Calibre 为元数据权威源

---

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                      元数据来源                              │
├─────────────┬─────────────┬─────────────────────────────────┤
│ 版权页抽取   │  豆瓣搜索   │  手动编辑                        │
│ (可批量)    │  (手动触发)  │  (单本)                         │
└──────┬──────┴──────┬──────┴──────┬──────────────────────────┘
       │             │             │
       ▼             ▼             ▼
┌─────────────────────────────────────────────────────────────┐
│              草稿存储 (metadata_drafts)                      │
│  SQLite: 待审核的变更、置信度、来源、状态                      │
└──────────────────────┬──────────────────────────────────────┘
                       │
          ┌────────────┴────────────┐
          │     置信度分流           │
          ▼                         ▼
   confidence ≥ 0.8          confidence < 0.8
          │                         │
          ▼                         ▼
   ┌────────────┐            ┌────────────┐
   │ 自动应用队列 │            │  审核队列   │
   │ (approved) │            │ (pending)  │
   └──────┬─────┘            └──────┬─────┘
          │                         │
          │    ┌────────────────────┘
          │    │ 用户审核
          ▼    ▼
┌─────────────────────────────────────────────────────────────┐
│                      应用阶段                                │
├─────────────────────────────────────────────────────────────┤
│  1. 写入 Calibre (通过 Content Server API)                   │
│  2. 写入变更日志 (metadata_changelog)                        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 存储设计

**Calibre**：书籍元数据最终存储（权威源）

**SQLite**：过程数据（复用现有基础设施）

项目已有 SQLite 支持（`internal/chat/db.go`），采用相同模式：
- 使用 `database/sql` + `go-sqlite3`
- `embed.FS` 管理迁移脚本
- 配置路径：`governance.db_path`

**方案选择**：

| 方案 | 描述 | 优缺点 |
|------|------|--------|
| A. 共享数据库 | 与 chat 使用同一个 db 文件 | 简单，但模块耦合 |
| B. 独立数据库 | `governance.db` 单独文件 | 隔离，推荐 |
| C. 通用 DB 层 | 抽象公共 db 包 | 长期更好，初期过度设计 |

**推荐 B 方案**：独立数据库文件，代码结构参考 `internal/chat/`

```
internal/governance/
├── db.go                    # 数据库操作
├── migrations/
│   └── 001_create_governance_tables.sql
├── types.go                 # 数据结构
├── confidence.go            # 置信度计算
├── collection_detector.go   # 合辑检测
└── service.go               # 业务逻辑
```

存储位置：`$DATA_DIR/governance.db`（通过 `governance.db_path` 配置）

### 2.3 与现有系统的关系

```
bookimporter (CLI)          calibre-api (Web)
├── clname (标题清理)        ├── 版权页 ISBN 抽取
├── check (完整性检测)       ├── 审核队列 UI
└── 入库前处理              ├── 变更历史查看
                           └── 入库后治理
```

---

## 3. 数据模型

### 3.1 草稿表

```sql
CREATE TABLE metadata_drafts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- 书籍信息
    book_id INTEGER NOT NULL,           -- Calibre book ID
    book_title TEXT,                    -- 冗余，方便展示
    
    -- 变更内容
    field TEXT NOT NULL,                -- isbn, title, authors, publisher, pubdate
    old_value TEXT,                     -- 变更前的值
    new_value TEXT NOT NULL,            -- 变更后的值
    
    -- 来源和质量
    source TEXT NOT NULL,               -- copyright_extract, douban, manual
    confidence REAL,                    -- 0.0 ~ 1.0
    confidence_breakdown TEXT,          -- JSON: 置信度分解
    flags TEXT,                         -- JSON: ["collection_suspected", ...]
    
    -- 状态
    status TEXT DEFAULT 'pending',      -- pending, approved, rejected, applied, skipped
    suggested_action TEXT,              -- auto_apply, review, skip
    
    -- 任务关联
    session_id TEXT,                    -- 抽取任务 ID
    
    -- 并发控制
    version INTEGER DEFAULT 1,          -- 乐观锁版本号
    
    -- 时间戳
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    reviewed_at DATETIME,
    reviewed_by TEXT,
    applied_at DATETIME
);

-- 索引
CREATE INDEX idx_drafts_status ON metadata_drafts(status);
CREATE INDEX idx_drafts_book ON metadata_drafts(book_id);
CREATE INDEX idx_drafts_session ON metadata_drafts(session_id);
CREATE INDEX idx_drafts_confidence ON metadata_drafts(confidence);

-- 防止同一本书同一字段重复创建 pending 草稿
CREATE UNIQUE INDEX idx_drafts_unique_pending 
    ON metadata_drafts(book_id, field) 
    WHERE status = 'pending';
```

### 3.2 变更日志表

```sql
CREATE TABLE metadata_changelog (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- 书籍信息
    book_id INTEGER NOT NULL,
    book_title TEXT,
    
    -- 变更内容
    field TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT NOT NULL,
    
    -- 来源
    source TEXT NOT NULL,
    draft_id INTEGER,                   -- 关联的草稿 ID
    
    -- 时间戳
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    applied_by TEXT,
    
    -- 回滚信息
    reverted_at DATETIME,
    reverted_by TEXT,
    revert_reason TEXT
);

-- 索引
CREATE INDEX idx_changelog_book ON metadata_changelog(book_id);
CREATE INDEX idx_changelog_applied ON metadata_changelog(applied_at);
CREATE INDEX idx_changelog_reverted ON metadata_changelog(reverted_at);
```

### 3.3 抽取会话表

```sql
CREATE TABLE extraction_sessions (
    id TEXT PRIMARY KEY,                -- UUID
    
    -- 任务信息
    task_type TEXT NOT NULL,            -- copyright_extract, batch_validate
    mode TEXT NOT NULL,                 -- full, incremental, dry_run
    
    -- 统计
    total_books INTEGER DEFAULT 0,
    processed INTEGER DEFAULT 0,
    success INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    skipped INTEGER DEFAULT 0,
    auto_approved INTEGER DEFAULT 0,
    pending_review INTEGER DEFAULT 0,
    
    -- 状态
    state TEXT DEFAULT 'running',       -- running, completed, failed, cancelled
    error_message TEXT,
    
    -- 时间戳
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);
```

### 3.4 并发控制

采用**乐观锁**机制防止并发冲突：

#### 3.4.1 场景分析

| 场景 | 问题 | 解决方案 |
|------|------|----------|
| 多个抽取任务同时处理同一本书 | 重复创建草稿 | `UNIQUE INDEX WHERE status='pending'` |
| 多人同时审核同一草稿 | 状态覆盖 | 乐观锁 `version` 字段 |
| 审核时草稿已被应用 | 无效操作 | 状态检查 + 乐观锁 |

#### 3.4.2 乐观锁实现

```go
// 审核操作：带版本号的更新
func (s *GovernanceService) ApproveDraft(id int64, version int) error {
    result, err := s.db.Exec(`
        UPDATE metadata_drafts 
        SET status = 'approved', 
            reviewed_at = CURRENT_TIMESTAMP,
            version = version + 1
        WHERE id = ? AND version = ? AND status = 'pending'
    `, id, version)
    
    if err != nil {
        return err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrConcurrentModification
    }
    return nil
}
```

#### 3.4.3 API 层处理

前端在获取草稿时拿到 `version`，提交审核时带上：

```json
// 请求
POST /api/metadata/drafts/123/approve
{
    "version": 3
}

// 响应 - 冲突
HTTP 409 Conflict
{
    "error": "concurrent_modification",
    "message": "草稿已被其他用户修改，请刷新后重试"
}
```

#### 3.4.4 批量操作

批量审核时，单条失败不影响其他记录：

```go
func (s *GovernanceService) BatchApprove(items []BatchItem) *BatchResult {
    result := &BatchResult{}
    for _, item := range items {
        err := s.ApproveDraft(item.ID, item.Version)
        if err == ErrConcurrentModification {
            result.Conflicts = append(result.Conflicts, item.ID)
        } else if err != nil {
            result.Errors = append(result.Errors, item.ID)
        } else {
            result.Success = append(result.Success, item.ID)
        }
    }
    return result
}
```

---

## 4. 置信度计算

### 4.1 计算公式

```
置信度 = ISBN可靠性 + 抽取上下文质量 - 书籍复杂度惩罚
```

最终值归一化到 [0, 1] 区间。

### 4.2 ISBN 可靠性 (0 ~ 0.50)

| 因素 | 分值 | 说明 |
|------|------|------|
| 校验位正确 | +0.30 | ISBN-10/13 校验算法通过 |
| ISBN-13 格式 | +0.05 | 比 ISBN-10 更可靠 |
| 标准前缀 | +0.10 | 978/979 开头 |
| 非测试号段 | +0.05 | 排除已知测试 ISBN |

```go
func isbnScore(isbn string) float64 {
    score := 0.0
    clean := strings.ReplaceAll(isbn, "-", "")
    
    if len(clean) == 13 {
        score += 0.05
        if validateISBN13Checksum(clean) {
            score += 0.30
        }
        if strings.HasPrefix(clean, "978") || strings.HasPrefix(clean, "979") {
            score += 0.10
        }
    } else if len(clean) == 10 {
        if validateISBN10Checksum(clean) {
            score += 0.30
        }
    }
    
    if !isTestISBN(clean) {
        score += 0.05
    }
    
    return score
}
```

### 4.3 抽取上下文质量 (-0.20 ~ 0.36)

| 因素 | 分值 | 说明 |
|------|------|------|
| 出版社存在 | +0.04 | 版权页结构完整 |
| 作者存在 | +0.04 | 版权页结构完整 |
| 出版日期存在 | +0.04 | 版权页结构完整 |
| 书名存在 | +0.04 | 版权页结构完整 |
| ISBN 标签明确 | +0.10 | 有 "ISBN" 前缀 |
| 单一 ISBN | +0.10 | 页面只有一个 ISBN |
| 多个 ISBN | -0.20 | 可能引用其他书 |

```go
func contextScore(metadata *CopyrightMetadata, pageContent string) float64 {
    score := 0.0
    
    // 结构完整性
    if metadata.Publisher != "" { score += 0.04 }
    if metadata.Author != "" { score += 0.04 }
    if metadata.PublishDate != "" { score += 0.04 }
    if metadata.BookTitle != "" { score += 0.04 }
    
    // ISBN 标签明确
    if regexp.MustCompile(`(?i)ISBN[：:\s]`).MatchString(pageContent) {
        score += 0.10
    }
    
    // ISBN 数量
    isbnMatches := isbnPattern.FindAllString(pageContent, -1)
    uniqueISBNs := uniqueISBNSet(isbnMatches)
    if len(uniqueISBNs) > 1 {
        score -= 0.20
    } else if len(uniqueISBNs) == 1 {
        score += 0.10
    }
    
    return score
}
```

### 4.4 书籍复杂度惩罚 (0 ~ 0.65)

| 因素 | 惩罚 | 说明 |
|------|------|------|
| 强合辑关键词 | +0.50 | "套装共"、"(套装" 等 |
| 弱合辑关键词 | +0.20 | "套装"、"合集"、"全集" |
| 丛书/系列关键词 | +0.10 | "丛书"、"系列" |
| 杂志特征 | +0.30 | 标题含年月、期号 |
| 标题过长 | +0.10 | 超过 50 字符 |
| 多作者 | +0.05 | 超过 3 个作者 |

```go
func bookComplexityPenalty(book *Book, title string) float64 {
    penalty := 0.0
    
    // 强合辑关键词
    strongKeywords := []string{"套装共", "套装全", "(套装", "（套装", "合集共", "全集共"}
    for _, kw := range strongKeywords {
        if strings.Contains(title, kw) {
            return 0.50  // 直接返回高惩罚
        }
    }
    
    // 弱合辑关键词
    weakKeywords := []string{"套装", "合集", "合辑", "全集"}
    for _, kw := range weakKeywords {
        if strings.Contains(title, kw) {
            penalty += 0.20
            break
        }
    }
    
    // 丛书/系列
    seriesKeywords := []string{"丛书", "系列", "文集"}
    for _, kw := range seriesKeywords {
        if strings.Contains(title, kw) {
            penalty += 0.10
            break
        }
    }
    
    // 杂志检测
    if regexp.MustCompile(`\d{4}年|\d+期|第\d+期`).MatchString(title) {
        penalty += 0.30
    }
    
    // 标题过长
    if utf8.RuneCountInString(title) > 50 {
        penalty += 0.10
    }
    
    // 多作者
    if len(book.Authors) > 3 {
        penalty += 0.05
    }
    
    return penalty
}
```

### 4.5 阈值设定

| 置信度范围 | 建议动作 | 说明 |
|-----------|----------|------|
| ≥ 0.80 | `auto_apply` | 自动应用 |
| 0.50 ~ 0.79 | `review` | 需人工审核 |
| < 0.50 | `skip` | 建议跳过 |

阈值可通过配置调整。

---

## 5. 合辑/套装处理

### 5.1 分类

| 类型 | 特征 | ISBN 情况 | 处理策略 |
|------|------|-----------|----------|
| 单本书 | 正常书籍 | 1 个 ISBN | 正常处理 |
| 套装（物理） | 多本书打包销售 | 套装有独立 ISBN | 可抽取套装 ISBN |
| 合辑（电子） | 多本书合成一个 EPUB | 无 ISBN 或多个 | 跳过，标记 |
| 丛书中的一本 | 系列书籍 | 有独立 ISBN | 正常处理 |
| 杂志/期刊 | 周期性出版物 | ISSN 而非 ISBN | 跳过 |

### 5.2 检测方法

#### 5.2.1 标题关键词

```go
var (
    // 强合辑：高置信度跳过
    strongCollectionKeywords = []string{
        "套装共", "套装全", "(套装", "（套装",
        "[套装", "【套装", "合集共", "全集共",
    }
    
    // 弱合辑：降权但继续处理
    weakCollectionKeywords = []string{
        "套装", "合集", "合辑", "全集",
    }
    
    // 系列：轻微降权
    seriesKeywords = []string{
        "丛书", "系列", "文集",
    }
)
```

#### 5.2.2 目录结构分析

检测目录是否包含多本书的结构：

```go
func hasMultipleBookStructure(toc []TOCEntry) bool {
    patterns := []string{
        `第[一二三四五六七八九十\d]+本`,
        `第[一二三四五六七八九十\d]+部`,
        `《[^》]+》`,  // 多个书名号
        `(?i)book\s*\d+`,
        `(?i)volume\s*\d+`,
    }
    
    topLevel := filterTopLevel(toc)
    bookLikeCount := 0
    
    for _, entry := range topLevel {
        for _, pattern := range patterns {
            if regexp.MustCompile(pattern).MatchString(entry.Title) {
                bookLikeCount++
                break
            }
        }
    }
    
    return bookLikeCount >= 2
}
```

#### 5.2.3 版权页 ISBN 数量

```go
func countUniqueISBNs(content string) int {
    matches := isbnPattern.FindAllString(content, -1)
    unique := make(map[string]bool)
    for _, m := range matches {
        clean := normalizeISBN(m)
        unique[clean] = true
    }
    return len(unique)
}
```

### 5.3 决策流程

```
┌─────────────────────────────────────┐
│         开始处理书籍                 │
└──────────────┬──────────────────────┘
               ▼
┌─────────────────────────────────────┐
│    标题含强合辑关键词?               │
└──────────────┬──────────────────────┘
          Yes  │  No
    ┌──────────┴──────────┐
    ▼                     ▼
┌─────────┐    ┌─────────────────────┐
│ 跳过    │    │ 标题含弱合辑关键词?  │
│ 标记合辑 │    └──────────┬──────────┘
└─────────┘           Yes  │  No
                ┌──────────┴──────────┐
                ▼                     ▼
      ┌─────────────────┐    ┌─────────────────┐
      │  检查目录结构    │    │   正常抽取       │
      └────────┬────────┘    └────────┬────────┘
               ▼                      ▼
      ┌─────────────────┐    ┌─────────────────┐
      │ 有多本书结构?    │    │   计算置信度     │
      └────────┬────────┘    └────────┬────────┘
          Yes  │  No                  ▼
    ┌──────────┴──────────┐  ┌─────────────────┐
    ▼                     ▼  │ 按阈值分流       │
┌─────────┐    ┌─────────┐  └─────────────────┘
│ 跳过    │    │ 降权后   │
│ 标记合辑 │    │ 继续抽取 │
└─────────┘    └─────────┘
```

---

## 6. 工作流

### 6.1 批量抽取任务

```
POST /api/tasks
{
    "type": "copyright_extract",
    "mode": "dry_run",           // dry_run | execute
    "options": {
        "limit": 100,            // 每次处理数量
        "confidence_threshold": 0.8,
        "auto_apply": false      // dry_run 模式下始终 false
    }
}
```

**Dry-Run 模式**：
1. 扫描缺少 ISBN 的书籍
2. 抽取版权页元数据
3. 计算置信度
4. 写入 `metadata_drafts`（status=pending）
5. 不写入 Calibre

**Execute 模式**：
1. 处理 `metadata_drafts` 中 status=approved 的记录
2. 写入 Calibre
3. 写入 `metadata_changelog`
4. 更新 draft status=applied

### 6.2 审核流程

```
┌─────────────────────────────────────────────────────────────┐
│                    审核队列 (Inbox)                          │
├─────────────────────────────────────────────────────────────┤
│ Filters: [All] [High Confidence] [Low Confidence] [Flagged] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ☐ 《三体》                                                   │
│   ISBN: 9787536692930  置信度: 0.85  来源: 版权页抽取         │
│   [✓ 通过] [✗ 拒绝] [编辑]                                   │
│                                                             │
│ ☐ 《人类简史》                                               │
│   ISBN: 9787508660752  置信度: 0.72  来源: 版权页抽取         │
│   ⚠️ 标记: 多个ISBN                                         │
│   [✓ 通过] [✗ 拒绝] [编辑]                                   │
│                                                             │
│ ☐ 《哈利波特全集》                                           │
│   ISBN: 无法确定  置信度: 0.25  来源: 版权页抽取              │
│   ⚠️ 标记: 疑似合辑, 多个ISBN                                │
│   [跳过] [手动编辑]                                          │
│                                                             │
├─────────────────────────────────────────────────────────────┤
│ 批量操作: [全选] [通过选中] [拒绝选中]     显示 20 / 共 156   │
└─────────────────────────────────────────────────────────────┘
```

### 6.3 回滚流程

```
GET /api/metadata/changelog?book_id=123

Response:
[
    {
        "id": 456,
        "field": "isbn",
        "old_value": "",
        "new_value": "9787536692930",
        "applied_at": "2026-01-05T10:30:00Z",
        "reverted_at": null
    }
]

POST /api/metadata/changelog/456/revert
{
    "reason": "ISBN 错误"
}
```

---

## 7. API 设计

### 7.1 抽取任务

```
# 启动抽取任务
POST /api/tasks
{
    "type": "copyright_extract",
    "mode": "dry_run",
    "options": {
        "limit": 100,
        "skip_collections": true
    }
}

# 查看任务状态
GET /api/tasks/:id

# 查看任务产生的草稿
GET /api/metadata/drafts?session_id=xxx
```

### 7.2 草稿管理

```
# 查询草稿列表
GET /api/metadata/drafts
    ?status=pending
    &confidence_min=0.5
    &confidence_max=0.8
    &has_flags=true
    &limit=20
    &offset=0

# 批量审核
POST /api/metadata/drafts/batch
{
    "ids": [1, 2, 3],
    "action": "approve"  // approve | reject | skip
}

# 单个草稿操作
POST /api/metadata/drafts/:id/approve
POST /api/metadata/drafts/:id/reject
PUT /api/metadata/drafts/:id
{
    "new_value": "corrected-isbn"
}
```

### 7.3 应用变更

```
# 应用已审核的草稿
POST /api/metadata/apply
{
    "draft_ids": [1, 2, 3]
}

# 应用所有已审核的草稿
POST /api/metadata/apply-all
```

### 7.4 变更历史

```
# 查询变更历史
GET /api/metadata/changelog
    ?book_id=123
    &field=isbn
    &from=2026-01-01
    &to=2026-01-31

# 回滚变更
POST /api/metadata/changelog/:id/revert
{
    "reason": "ISBN 错误"
}
```

### 7.5 统计

```
GET /api/metadata/stats

Response:
{
    "drafts": {
        "pending": 156,
        "approved": 45,
        "rejected": 12,
        "applied": 1234
    },
    "confidence_distribution": {
        "high": 890,      // >= 0.8
        "medium": 234,    // 0.5 ~ 0.8
        "low": 156        // < 0.5
    },
    "by_source": {
        "copyright_extract": 1200,
        "douban": 50,
        "manual": 30
    },
    "flags_count": {
        "collection_suspected": 89,
        "multiple_isbn": 45,
        "isbn_invalid_checksum": 23
    }
}
```

---

## 8. 前端页面

### 8.1 页面规划

| 路径 | 功能 |
|------|------|
| `/metadata/queue` | 审核队列（收件箱） |
| `/metadata/history` | 变更历史 |
| `/metadata/stats` | 统计仪表盘 |

### 8.2 审核队列页面

**功能**：
- 筛选：状态、置信度范围、标记类型
- 排序：置信度、创建时间、书名
- 批量操作：全选、通过、拒绝
- 单条操作：通过、拒绝、编辑、查看详情
- 快捷键：j/k 上下移动，a 通过，r 拒绝

**展示信息**：
- 书名、当前值、新值
- 置信度（进度条）
- 来源
- 标记（badges）
- 抽取时间

### 8.3 变更历史页面

**功能**：
- 按书籍查看历史
- 按时间范围筛选
- 回滚操作
- 导出历史

**展示信息**：
- 书名、字段、旧值、新值
- 来源、应用时间
- 回滚状态

### 8.4 统计仪表盘

**展示**：
- 待审核数量（大数字）
- 置信度分布（饼图/柱状图）
- 按来源分布
- 标记类型分布
- 最近 7 天趋势

---

## 9. 配置

```yaml
# config.yaml
governance:
  # 数据库路径
  db_path: "data/governance.db"
  
  # 置信度配置
  confidence:
    # 分层阈值（针对不同来源设置不同阈值）
    thresholds:
      copyright_extract:     # 版权页抽取
        auto_apply: 0.85
        review: 0.50
      douban:                # 豆瓣搜索（更可信）
        auto_apply: 0.75
        review: 0.40
      default:               # 默认阈值
        auto_apply: 0.80
        review: 0.50
    
    # 动态调优配置（Phase 6 启用）
    auto_tune:
      enabled: false         # 积累足够数据后启用
      min_samples: 100       # 每个来源至少 100 条审核记录
      target_precision: 0.98 # 自动应用的目标精确率
      update_interval: "168h" # 更新频率（每周）
  
  # 合辑处理
  collection:
    skip_strong_keywords: true
    strong_keywords:
      - "套装共"
      - "(套装"
      - "（套装"
    weak_keywords:
      - "套装"
      - "合集"
      - "全集"
  
  # 批量任务
  batch:
    default_limit: 100
    max_limit: 1000
    worker_count: 5
```

### 9.1 动态阈值调优

基于历史审核数据自动计算最优阈值：

```sql
-- 分析不同置信度区间的审核通过率
SELECT 
    CASE 
        WHEN confidence >= 0.9 THEN '0.9+'
        WHEN confidence >= 0.8 THEN '0.8-0.9'
        WHEN confidence >= 0.7 THEN '0.7-0.8'
        WHEN confidence >= 0.6 THEN '0.6-0.7'
        ELSE '<0.6'
    END as confidence_range,
    COUNT(*) as total,
    SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) as approved,
    ROUND(100.0 * SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END) / COUNT(*), 2) as approval_rate
FROM metadata_drafts
WHERE status IN ('approved', 'rejected')
GROUP BY confidence_range
ORDER BY confidence_range DESC;
```

**调优策略**：将 `auto_apply` 阈值设定在通过率 ≥ 98% 的边界点

---

## 10. 实施计划

### Phase 1: 基础设施 (1-2 周) ✅

- [x] SQLite 数据库初始化
- [x] 数据模型实现
- [x] 置信度计算函数
- [x] 合辑检测函数

### Phase 2: 抽取任务改造 (1 周)

- [x] Dry-run 模式 (框架就绪)
- [x] 草稿写入
- [x] 任务统计
- [ ] 批量抽取入口 API

### Phase 3: API 实现 (1 周) ✅

- [x] 草稿管理 API
- [x] 变更应用 API
- [x] 历史查询 API
- [x] 统计 API

### Phase 4: 前端审核队列 (1-2 周) ✅

- [x] 队列页面 (`/metadata/queue`)
- [x] DraftCard 组件
- [ ] 批量操作
- [ ] 快捷键支持

### Phase 5: 历史与回滚 (1 周) ✅

- [x] 历史页面 (`/metadata/history`)
- [x] 回滚功能
- [ ] 导出功能

### Phase 6: 统计与优化 (持续)

- [x] 统计仪表盘 (`/metadata/stats`)
- [ ] 性能优化
- [ ] 置信度模型调优


---

## 11. 开放问题

1. ~~**置信度权重调优**~~：已有设计方向（见 9.1 动态阈值调优），需要实际数据验证
2. **合辑检测的误判率**：「全集」可能是完整版而非合辑
3. **豆瓣集成**：如何在限流约束下提供手动触发的豆瓣验证
4. **增量同步**：新书入库时如何自动触发抽取

---

## 附录

### A. ISBN 校验算法

```go
// ISBN-13 校验
func validateISBN13Checksum(isbn string) bool {
    if len(isbn) != 13 {
        return false
    }
    sum := 0
    for i, c := range isbn {
        digit := int(c - '0')
        if i%2 == 0 {
            sum += digit
        } else {
            sum += digit * 3
        }
    }
    return sum%10 == 0
}

// ISBN-10 校验
func validateISBN10Checksum(isbn string) bool {
    if len(isbn) != 10 {
        return false
    }
    sum := 0
    for i, c := range isbn {
        var digit int
        if c == 'X' || c == 'x' {
            digit = 10
        } else {
            digit = int(c - '0')
        }
        sum += digit * (10 - i)
    }
    return sum%11 == 0
}
```

### B. 测试 ISBN 号段

```go
var testISBNPrefixes = []string{
    "9780000000",  // 测试用
    "9790000000",  // 测试用
    "0000000000",  // 无效
}

func isTestISBN(isbn string) bool {
    for _, prefix := range testISBNPrefixes {
        if strings.HasPrefix(isbn, prefix) {
            return true
        }
    }
    return false
}
```

---

**文档维护**: jianyun8023  
**最后更新**: 2026-01-05
