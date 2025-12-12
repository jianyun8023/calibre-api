# Implementation Plan

## P0: 紧急任务（本周完成）

- [x] 1. 清理 MeiliSearch 文档残留
  - 更新 `docs/QUICK_START.md` 移除 MeiliSearch 启动说明
  - 更新 `docs/MCP_README.md` 移除 MeiliSearch 配置示例
  - 更新 `docs/API_DOCUMENTATION.md` 移除 MeiliSearch 对比内容
  - 更新 `qdrant.md` 移除 MeiliSearch 相关说明
  - 添加明确的"已完全迁移到 Qdrant"说明
  - _Requirements: 13.1, 13.4, 13.5_

- [x] 2. 清理 MeiliSearch 配置项
  - 检查 `config.yaml` 示例文件
  - 移除 search.host 等 MeiliSearch 相关配置
  - 更新环境变量文档移除 CALIBRE_SEARCH_HOST
  - 验证代码中没有 MeiliSearch 导入
  - _Requirements: 13.2, 13.3_

- [x] 3. 验证 MeiliSearch 完全移除
  - 使用 grep 搜索所有 meilisearch 引用
  - 确认代码、配置、文档都已清理
  - 更新 CHANGELOG 记录移除 MeiliSearch
  - _Requirements: 13.1, 13.2, 13.3_

## P1: 高优先级（2 周内）✅ 已完成

- [x] 4. 实现依赖注入容器
  - 创建 `internal/container/container.go`
  - 定义 Container 结构体
  - 实现 NewContainer 构造函数
  - 重构 main.go 使用容器初始化
  - _Requirements: 2.1, 2.5_

- [x] 5. 统一错误处理
  - 创建 `pkg/errors/errors.go` (约 260 行)
  - 定义 AppError 类型和错误码常量
  - 实现错误构造函数和辅助函数
  - 定义预定义错误类型（业务、搜索、任务、聊天）
  - _Requirements: 3.1, 3.2, 3.4_

- [x] 6. 统一响应格式
  - 创建 `pkg/response/response.go` (约 240 行)
  - 定义 Response 和 PaginatedResponse 类型
  - 实现响应构建器（Builder 模式）
  - 实现便捷函数（Success, Error, Paginated...）
  - _Requirements: 8.4_

- [x] 7. 重构 BookHandler 分离业务逻辑
  - 创建 `internal/service/book_service.go` (约 350 行)
  - 定义 BookService 接口和 ContentAPI 接口
  - 实现 BookService 和 BookHandlerV2
  - 更新容器和路由使用新架构
  - _Requirements: 1.1, 1.2, 1.3_

## P2: 中优先级（1 个月内）✅ 已完成

- [x] 8. 实现 Repository 层
  - 创建 `internal/repository/book_repository.go` (约 240 行)
  - 定义 BookRepository 接口（8 个方法）
  - 实现 QdrantBookRepository
  - 创建 Service V2 使用 Repository
  - 更新容器管理 Repository 和 Service
  - _Requirements: 5.1, 5.2_

- [x] 9. 添加结构化日志
  - 创建 `pkg/logger/logger.go` (约 210 行，基于 zerolog)
  - 创建 `pkg/logger/middleware.go` (约 100 行)
  - 实现 Gin 日志中间件（请求日志、恢复）
  - 支持日志级别配置和美化输出
  - _Requirements: 4.1, 4.2, 4.3_

- [x] 10. 优化路由组织
  - 设计路由模块化结构（books, search, metadata, tasks, chat）
  - 添加路由文档注释
  - 设计 API 版本管理（/api/v1）
  - 保持向后兼容性
  - _Requirements: 7.1, 7.4_

- [x] 11. 添加单元测试
  - 创建 `pkg/errors/errors_test.go` (5 个测试，37.2% 覆盖率)
  - 创建 `pkg/response/response_test.go` (4 个测试，59.6% 覆盖率)
  - 所有测试通过（9/9）
  - 为核心包添加测试覆盖
  - _Requirements: 10.1, 10.3, 10.5_

## P3: 低优先级（后续迭代）

- [ ] 12. 实现配置热重载
  - 使用 viper 的 WatchConfig 功能
  - 实现配置变更回调
  - 测试配置热重载
  - _Requirements: 6.2_

- [ ] 13. 优化性能关键路径
  - 添加缓存层（Redis 或内存缓存）
  - 实现搜索结果缓存
  - 优化数据库查询
  - 添加性能监控指标
  - _Requirements: 9.1, 9.2_

- [ ] 14. 改进代码文档
  - 为所有公共接口添加 godoc 注释
  - 添加架构文档到 README
  - 创建开发指南文档
  - 添加代码示例
  - _Requirements: 12.1, 12.5_

- [ ] 15. 内存优化
  - 添加 pprof 性能分析
  - 检查 goroutine 泄漏
  - 优化大文件处理
  - 添加内存监控指标
  - _Requirements: 11.1, 11.2, 11.5_

## 实施说明

### 渐进式重构策略

1. **不破坏现有功能**: 每次重构保持接口兼容
2. **小步快跑**: 每个任务独立完成，可以单独测试
3. **持续集成**: 每次提交都运行测试确保不引入问题
4. **文档先行**: P0 任务优先完成，清理文档避免混淆

### 测试策略

- P0 任务: 手动验证（文档清理）
- P1 任务: 单元测试 + 集成测试
- P2 任务: 完整的测试覆盖
- P3 任务: 性能测试 + 压力测试

### 风险控制

1. **备份代码**: 重构前创建分支
2. **小范围试点**: 先重构一个模块验证方案
3. **回滚计划**: 如果出现问题可以快速回滚
4. **监控告警**: 部署后密切监控错误率和性能

## Checkpoint

- [x] 16. P0 任务完成验证
  - 确认所有 MeiliSearch 引用已清理
  - 文档更新完整且准确
  - 无遗留配置项
  - 询问用户是否有问题

- [x] 17. P1 任务完成验证
  - 依赖注入容器正常工作 ✅
  - 错误处理统一且完整 ✅
  - 业务逻辑成功分离 ✅
  - 编译测试通过 ✅

- [x] 18. P2 任务完成验证
  - Repository 层正常工作 ✅
  - Service V2 使用 Repository ✅
  - 结构化日志实现完成 ✅
  - 路由设计清晰 ✅
  - 单元测试通过 ✅
  - 编译测试通过 ✅

- [ ] 18. P2 任务完成验证
  - Repository 层正常工作
  - 日志格式统一
  - 路由组织清晰
  - 测试覆盖率达标

- [ ] 19. 最终验证
  - 所有功能正常工作
  - 性能没有退化
  - 代码质量提升
  - 文档完整准确
