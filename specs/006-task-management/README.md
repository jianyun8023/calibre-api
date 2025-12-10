---
status: complete
created: '2025-12-08'
tags: []
priority: medium
created_at: '2025-12-08T13:04:03.756Z'
updated_at: '2025-12-08T13:10:00.000Z'
completed_at: '2025-12-08T13:10:00.000Z'
completed: '2025-12-08'
transitions:
  - status: complete
    at: '2025-12-08T13:10:00.000Z'
depends_on:
  - 005-qdrant-vector-search
---

# 异步任务管理

> **Status**: ✅ Complete · **Priority**: Medium · **Created**: 2025-12-08

## Overview

**问题**: 向量化 1000+ 书籍需要 5-10 分钟，同步执行阻塞 HTTP 请求不可接受。

**解决方案**: 构建 TaskManager，支持异步执行长时任务，提供进度跟踪和状态管理。

**为什么现在**: Qdrant 同步和 TOC 提取都是耗时操作，需要统一的任务调度框架。

## Design

### 任务类型

**已支持的任务**:
1. `qdrant_sync`: 同步书籍到 Qdrant
   - 批量向量化（100 本/批次）
   - 支持全量/增量模式
   - 跳过已存在的书籍

2. `toc_extract`: 提取书籍目录
   - 批量处理 EPUB 文件
   - 缓存 TOC 到内存
   - 失败重试机制

3. `check_missing`: 检查缺失书籍
   - 对比 Calibre DB 和 Qdrant
   - 生成缺失列表报告
   - 支持自动补全模式

### 任务生命周期

**状态流转**:
```
pending → running → success
                  → failed
                  → cancelled
```

**并发控制**:
- 单任务类型同时只运行 1 个实例
- 不同任务类型可并发执行
- 使用 `sync.Map` 存储任务状态（线程安全）

**权衡**: 单实例 vs 多实例
- 选择单实例：避免资源竞争（Qdrant 写入冲突）
- 代价：无法并行处理多个同类任务

### 进度追踪

**进度数据结构**:
```go
type TaskProgress struct {
    TaskID   string
    Type     string
    Status   string    // pending | running | success | failed
    Progress int       // 0-100
    Message  string    // "Processing book 42/1000"
    Error    string    // 错误信息（如果失败）
    Start    time.Time
    End      time.Time
}
```

**实时更新机制**:
```go
// 任务内部定期更新
manager.UpdateProgress(taskID, progress, message)

// 前端轮询获取
GET /api/tasks/{id}
```

**权衡**: 轮询 vs WebSocket
- 选择轮询：实现简单，符合 RESTful
- 代价：有轻微延迟（~1s），增加服务器负载
- 未来可改为 SSE 推送

### 错误处理

**重试策略**:
```go
maxRetries := 3
for i := 0; i < maxRetries; i++ {
    if err := processBook(book); err == nil {
        break
    }
    log.Warnf("Retry %d/%d: %v", i+1, maxRetries, err)
    time.Sleep(time.Second * time.Duration(i+1))
}
```

**失败隔离**:
- 单本书籍失败不影响其他书籍
- 记录失败原因到任务日志
- 提供 "仅处理失败项" 重试模式

## Plan

- [x] 设计 TaskManager 架构
- [x] 实现任务注册和调度
- [x] 实现 qdrant_sync 任务
- [x] 实现 toc_extract 任务
- [x] 实现 check_missing 任务
- [x] 添加进度追踪 API
- [x] 添加任务取消功能
- [x] 前端任务管理 UI
- [x] 添加批量操作优化

## Test

- [x] 启动 qdrant_sync 任务
- [x] 进度从 0% → 100%
- [x] 1000 本书籍成功同步
- [x] 任务取消正确停止
- [x] 失败任务记录错误信息
- [x] 并发启动多个不同类型任务
- [x] 重复启动同类型任务被拒绝

## Notes

**性能数据** (优化前):
- qdrant_sync: ~10min (1000 本书籍，Ollama)
- toc_extract: ~13min (1000 本 EPUB，~75 本/分钟)
- check_missing: ~30s (数据库对比)

**性能数据** (优化后 v1.1.0):
- qdrant_sync: ~10min (无变化)
- toc_extract: ~2-3min (1000 本 EPUB，~400-800 本/分钟)
  - 大部分已缓存: ~800-1000 本/分钟 (提升 10x)
  - 需要下载: ~200-300 本/分钟 (提升 4-6x)
  - 混合场景: ~400-500 本/分钟 (提升 5-7x)
- check_missing: ~30s (无变化)

**TOC 提取任务优化** (v1.1.0):

1. **缓存管理器锁优化**:
   - 旧策略: 全局锁，所有操作串行（并行度 = 1）
   - 新策略: 无锁快速路径 + 双重检查
   - 缓存命中时无锁开销（99% 场景）
   - 只在需要下载时加锁
   - 提升: 吞吐量提升 10-20x（缓存命中场景）

2. **批量 Qdrant 更新**:
   - 旧策略: 每本书单独更新（1000 次网络请求）
   - 新策略: 批量更新 20 本/次（50 次网络请求）
   - 减少网络往返次数 95%
   - 降低 HTTP 连接开销
   - 提升: 网络延迟减少 95%

3. **增加 Worker 数量**:
   - 旧配置: 5 个 worker
   - 新配置: 10 个 worker
   - 并行处理能力翻倍
   - 更好利用 CPU 和 I/O 带宽
   - 提升: 处理速度提升 80-100%

**批量优化**:
    // 进度保存批次
    batchSize := 50
    
    // Qdrant 更新批次
    qdrantBatch := 20
    
    // Worker 并发数
    numWorkers := 10

**资源控制**:
- Embedding 并发限制: 5 个（避免 Ollama 过载）
- Qdrant 批量插入: 100 个/次（同步任务），20 个/次（TOC 更新）
- EPUB 解析内存占用: ~50MB/文件
- TOC 提取 CPU 使用: 60-80%（优化后，多核并行）
- TOC 提取内存使用: ~80-150MB（+30-50MB 批次缓冲）

**配置调优建议**:

小型书库 (< 1000 本):
    numWorkers:  5
    batchSize:   30
    qdrantBatch: 10

中型书库 (1000-5000 本):
    numWorkers:  10  // 当前默认
    batchSize:   50
    qdrantBatch: 20

大型书库 (> 5000 本):
    numWorkers:  15
    batchSize:   100
    qdrantBatch: 50

**已知限制**:
- 任务重启后进度丢失（内存存储）
- 不支持任务暂停（只能取消）
- 不支持任务优先级（FIFO 队列）

**未来优化**:
- 持久化任务状态（SQLite）
- 支持定时任务（Cron 表达式）
- 添加任务依赖（DAG 执行）
- 支持分布式任务（多实例协同）

**前端集成**:
```javascript
// 启动任务
POST /api/tasks/start
{ "type": "qdrant_sync" }

// 查询进度
GET /api/tasks/{id}

// 取消任务
POST /api/tasks/{id}/stop
```

**依赖项**:
- `005-qdrant-vector-search`: 向量同步目标
- `001-book-management`: 书籍数据源和 TOC 提取
