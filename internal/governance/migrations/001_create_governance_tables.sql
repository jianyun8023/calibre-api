-- 草稿表
CREATE TABLE IF NOT EXISTS metadata_drafts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    
    -- 书籍信息
    book_id INTEGER NOT NULL,
    book_title TEXT,
    
    -- 变更内容
    field TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT NOT NULL,
    
    -- 来源和质量
    source TEXT NOT NULL,
    confidence REAL,
    confidence_breakdown TEXT,
    flags TEXT,
    
    -- 状态
    status TEXT DEFAULT 'pending',
    suggested_action TEXT,
    
    -- 任务关联
    session_id TEXT,
    
    -- 并发控制
    version INTEGER DEFAULT 1,
    
    -- 时间戳
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    reviewed_at DATETIME,
    reviewed_by TEXT,
    applied_at DATETIME
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_drafts_status ON metadata_drafts(status);
CREATE INDEX IF NOT EXISTS idx_drafts_book ON metadata_drafts(book_id);
CREATE INDEX IF NOT EXISTS idx_drafts_session ON metadata_drafts(session_id);
CREATE INDEX IF NOT EXISTS idx_drafts_confidence ON metadata_drafts(confidence);

-- 防止同一本书同一字段重复创建 pending 草稿
CREATE UNIQUE INDEX IF NOT EXISTS idx_drafts_unique_pending 
    ON metadata_drafts(book_id, field) 
    WHERE status = 'pending';

-- 变更日志表
CREATE TABLE IF NOT EXISTS metadata_changelog (
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
    draft_id INTEGER,
    
    -- 时间戳
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    applied_by TEXT,
    
    -- 回滚信息
    reverted_at DATETIME,
    reverted_by TEXT,
    revert_reason TEXT
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_changelog_book ON metadata_changelog(book_id);
CREATE INDEX IF NOT EXISTS idx_changelog_applied ON metadata_changelog(applied_at);
CREATE INDEX IF NOT EXISTS idx_changelog_reverted ON metadata_changelog(reverted_at);

-- 抽取会话表
CREATE TABLE IF NOT EXISTS extraction_sessions (
    id TEXT PRIMARY KEY,
    
    -- 任务信息
    task_type TEXT NOT NULL,
    mode TEXT NOT NULL,
    
    -- 统计
    total_books INTEGER DEFAULT 0,
    processed INTEGER DEFAULT 0,
    success INTEGER DEFAULT 0,
    failed INTEGER DEFAULT 0,
    skipped INTEGER DEFAULT 0,
    auto_approved INTEGER DEFAULT 0,
    pending_review INTEGER DEFAULT 0,
    
    -- 状态
    state TEXT DEFAULT 'running',
    error_message TEXT,
    
    -- 时间戳
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_sessions_state ON extraction_sessions(state);
CREATE INDEX IF NOT EXISTS idx_sessions_started ON extraction_sessions(started_at);
