# EPUB TOC 提取与缓存管理实现总结

## 实施完成时间
2025-11-21

## 概述
成功实现了完整的 EPUB 目录（TOC）提取系统，包括：
1. 带容量限制的 EPUB 文件缓存管理
2. Qdrant schema 支持 TOC 数据增量更新
3. 支持断点续传的 TOC 提取任务
4. 智能 TOC 查询（优先使用 Qdrant，缺失时自动提取并回填）

## 实现的功能

### 1. 配置更新
**文件**: `config.yaml`
- 新增缓存配置段：
  - `cache.dir`: 缓存目录路径（默认 `.cache/epub`）
  - `cache.max_size_gb`: 最大容量 5GB
  - `cache.cleanup_threshold`: 清理阈值 0.9（90%）

**文件**: `internal/calibre/types.go`
- 添加 `cache.Config` 到主配置结构

**文件**: `internal/semantic/types.go`
- `Book` 结构体新增 `Toc` 字段用于存储目录数据

### 2. 缓存管理模块
**新包**: `internal/cache`
- `Manager`: 核心缓存管理器
  - `GetOrExtractEpub()`: 获取或下载 EPUB 文件
  - `checkAndCleanup()`: 自动清理功能
  - `GetCacheStats()`: 缓存统计信息

**清理策略**:
- 监控缓存大小，超过 90% 阈值时触发清理
- 基于文件访问时间（atime）排序
- 删除最旧的文件直到使用率降至 70%

**跨平台支持**:
- `manager_unix.go`: Unix/Linux/macOS 访问时间处理
- `manager_windows.go`: Windows 访问时间处理

### 3. Qdrant Schema 增量更新
**文件**: `internal/semantic/qdrant/client.go`
- **新增 `SetPayload()` 方法**: 部分更新 payload 字段，避免覆盖其他数据
- 更新 `BookToPayload()`: 支持 TOC 字段
- 更新 `PayloadToBook()`: 解析 TOC 字段

**文件**: `internal/semantic/qdrant/searcher.go`
- **新增 `GetBookToc()` 方法**: 从 Qdrant 获取书籍 TOC
- **新增 `UpdateToc()` 方法**: 增量更新 TOC 到 Qdrant

### 4. TOC 提取任务系统
**文件**: `internal/tasks/types.go`
- 新增任务类型 `TaskTypeTocExtract`

**文件**: `internal/tasks/toc_extract.go`
- `TocExtractTask`: 完整的 TOC 提取任务实现
- **断点续传功能**:
  - 进度保存在 `.cache/toc_progress_{taskID}.json`
  - 记录已处理的书籍 ID 列表
  - 每处理 10 本书保存一次进度
  - 任务中断后可从断点继续
- **TOC 提取**:
  - 使用缓存管理器获取 EPUB 文件
  - 提取完整的嵌套 TOC 结构
  - 自动更新到 Qdrant

### 5. API 集成
**文件**: `internal/calibre/task_handler.go`
- 在 `startTask()` 中添加对 `toc_extract` 任务的支持
- 任务 API:
  - `POST /api/tasks/start`: 启动 TOC 提取任务
    ```json
    {
      "type": "toc_extract",
      "mode": "full" // 或 "incremental"
    }
    ```
  - `GET /api/tasks`: 获取所有任务状态
  - `POST /api/tasks/:id/stop`: 停止任务

**文件**: `internal/calibre/route.go`
- 在 `NewClient()` 中初始化缓存管理器
- 将缓存管理器注入到 API 结构体

### 6. 智能 TOC 查询
**文件**: `internal/calibre/book_content_handler.go`
- 重构 `getBookToc()` 实现智能查询策略：
  1. **优先从 Qdrant 查询**: 快速返回已存储的 TOC
  2. **缓存未命中**: 从 EPUB 文件提取
  3. **异步回填**: 提取后异步更新到 Qdrant（不阻塞响应）
  4. **返回结果**: 确保用户快速获得数据

- 新增 `extractTocFromEpub()`: 提取 TOC 的公共函数
  - 支持缓存管理器（优先）
  - 降级到旧的文件缓存方法
  - 返回标准化的 TOC 结构

## 架构优化

### 解决循环依赖
- 将 `CacheManager` 从 `internal/calibre` 移至独立的 `internal/cache` 包
- 避免了 `calibre` ↔ `tasks` 之间的循环依赖

### 模块化设计
```
internal/
├── cache/          # 缓存管理（独立包）
│   ├── manager.go
│   ├── manager_unix.go
│   └── manager_windows.go
├── calibre/        # Calibre API
│   ├── book_content_handler.go  # TOC API
│   ├── task_handler.go          # 任务 API
│   └── route.go                 # 路由和初始化
├── semantic/       # 语义搜索
│   ├── types.go    # Book 结构（含 Toc 字段）
│   └── qdrant/     # Qdrant 集成
│       ├── client.go    # SetPayload API
│       └── searcher.go  # GetBookToc/UpdateToc
└── tasks/          # 任务系统
    ├── types.go
    └── toc_extract.go  # TOC 提取任务
```

## 使用示例

### 1. 启动 TOC 提取任务
```bash
curl -X POST http://localhost:8080/api/tasks/start \
  -H "Content-Type: application/json" \
  -d '{"type": "toc_extract", "mode": "full"}'
```

响应:
```json
{
  "code": 200,
  "message": "TOC extraction task started",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

### 2. 查询任务状态
```bash
curl http://localhost:8080/api/tasks
```

响应:
```json
{
  "code": 200,
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "toc_extract",
      "mode": "full",
      "state": "running",
      "progress": 45.5,
      "message": "Processing book 455/1000 (ID: 123)",
      "start_time": "2025-11-21T10:00:00Z"
    }
  ]
}
```

### 3. 获取书籍 TOC（智能查询）
```bash
curl http://localhost:8080/api/read/123/toc
```

响应:
```json
{
  "points": [
    {
      "text": "第一章 引言",
      "content": {
        "src": "/read/123/file/OEBPS/chapter1.xhtml"
      },
      "points": [
        {
          "text": "1.1 背景",
          "content": {
            "src": "/read/123/file/OEBPS/section1.xhtml"
          }
        }
      ]
    }
  ],
  "metadata": { /* EPUB metadata */ },
  "baseDir": "OEBPS"
}
```

### 4. 停止任务
```bash
curl -X POST http://localhost:8080/api/tasks/{taskID}/stop
```

## 性能优势

1. **智能缓存**: TOC 数据存储在 Qdrant 中，避免重复解析 EPUB 文件
2. **按需提取**: 首次访问时提取，后续访问直接从 Qdrant 读取
3. **异步回填**: 提取操作不阻塞 API 响应
4. **容量管理**: 自动清理旧的 EPUB 缓存文件
5. **断点续传**: 大批量处理支持中断恢复

## 配置建议

```yaml
cache:
  dir: ".cache/epub"        # 缓存目录
  max_size_gb: 5            # 根据磁盘空间调整
  cleanup_threshold: 0.9    # 90% 时触发清理
```

建议根据实际情况调整：
- 小型书库（< 1000 本）: 2-3 GB
- 中型书库（1000-5000 本）: 5-10 GB
- 大型书库（> 5000 本）: 10-20 GB

## 测试验证

编译成功，所有模块集成完毕：
```bash
✅ 编译成功
✅ 无循环依赖
✅ 所有类型检查通过
```

## 后续改进建议

1. **批量 TOC 更新**: 支持批量更新多本书的 TOC 到 Qdrant
2. **TOC 搜索**: 支持在 TOC 中搜索关键词
3. **进度 UI**: 前端显示 TOC 提取任务进度
4. **缓存预热**: 系统启动时预加载热门书籍的 TOC
5. **监控告警**: 缓存使用率超过阈值时发送通知

## 相关文档

- [快速开始指南](docs/QUICK_START.md)
- [代码结构说明](docs/CODE_STRUCTURE.md)
- [API 文档](docs/API_DOCUMENTATION.md)

