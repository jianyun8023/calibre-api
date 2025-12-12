# 前端 API 接口清理与迁移需求文档

## 简介

在 Spec 026 后端代码优化完成后，后端已实现统一的 API 响应格式和错误处理机制。前端需要相应地更新 API 接口调用逻辑，清理冗余代码，统一错误处理，并适配新的响应格式，确保前后端接口的一致性和可维护性。

## 术语表

- **Frontend**: Next.js 前端应用程序
- **API_Client**: 前端 API 调用客户端
- **Response_Format**: API 响应数据格式
- **Error_Handler**: 错误处理机制
- **Type_Definition**: TypeScript 类型定义
- **Pagination_Logic**: 分页处理逻辑

## 需求

### 需求 1

**用户故事:** 作为前端开发者，我希望有统一的 API 响应类型定义，以便在整个应用中保持类型安全和一致性。

#### 验收标准

1. WHEN 定义 API 响应类型 THEN Frontend SHALL 创建符合后端 V2 格式的 TypeScript 接口
2. WHEN 处理成功响应 THEN Frontend SHALL 使用统一的 ApiResponse<T> 接口结构
3. WHEN 处理错误响应 THEN Frontend SHALL 使用包含 ErrorInfo 对象的响应格式
4. WHEN 处理分页响应 THEN Frontend SHALL 使用新的 PaginatedResponse 格式替代旧的 records 结构
5. WHEN 导入类型定义 THEN Frontend SHALL 从统一的类型文件中导入所有 API 相关类型

### 需求 2

**用户故事:** 作为前端开发者，我希望有统一的 API 客户端，以便简化 API 调用和错误处理逻辑。

#### 验收标准

1. WHEN 发起 API 请求 THEN API_Client SHALL 使用统一的请求方法处理所有 HTTP 调用
2. WHEN 接收 API 响应 THEN API_Client SHALL 自动解析新的响应格式并提取数据
3. WHEN 遇到 API 错误 THEN API_Client SHALL 使用新的错误格式提供详细的错误信息
4. WHEN 处理网络错误 THEN API_Client SHALL 提供一致的错误处理和用户提示
5. WHEN 配置请求头 THEN API_Client SHALL 自动添加必要的 Content-Type 和其他标准头部

### 需求 3

**用户故事:** 作为前端开发者，我希望更新所有书籍相关的 API 调用，以便使用新的响应格式和错误处理。

#### 验收标准

1. WHEN 获取书籍列表 THEN Frontend SHALL 使用新的分页格式处理 recently、random、all 等接口
2. WHEN 获取单本书籍 THEN Frontend SHALL 处理新的成功和错误响应格式
3. WHEN 更新书籍信息 THEN Frontend SHALL 使用新的请求和响应格式
4. WHEN 删除书籍 THEN Frontend SHALL 处理新的操作结果响应格式
5. WHEN 搜索书籍 THEN Frontend SHALL 适配混合搜索和语义搜索的新响应格式

### 需求 4

**用户故事:** 作为前端开发者，我希望更新元数据搜索相关的 API 调用，以便保持与后端接口的一致性。

#### 验收标准

1. WHEN 通过 ISBN 搜索元数据 THEN Frontend SHALL 使用统一的错误处理机制
2. WHEN 通过关键词搜索元数据 THEN Frontend SHALL 处理标准化的响应格式
3. WHEN 元数据搜索失败 THEN Frontend SHALL 显示详细的错误信息和建议
4. WHEN 处理豆瓣 API 响应 THEN Frontend SHALL 正确映射到内部数据结构
5. WHEN 元数据搜索超时 THEN Frontend SHALL 提供用户友好的错误提示

### 需求 5

**用户故事:** 作为前端开发者，我希望统一错误处理逻辑，以便为用户提供一致的错误体验。

#### 验收标准

1. WHEN 遇到 API 错误 THEN Error_Handler SHALL 优先使用新格式的 error.message 字段
2. WHEN 显示错误信息 THEN Error_Handler SHALL 根据 error.code 提供本地化的错误提示
3. WHEN 记录错误日志 THEN Error_Handler SHALL 包含 error.details 和 error.context 信息
4. WHEN 处理网络错误 THEN Error_Handler SHALL 区分网络问题和服务器错误
5. WHEN 错误恢复 THEN Error_Handler SHALL 提供重试机制和用户指导

### 需求 6

**用户故事:** 作为前端开发者，我希望更新分页处理逻辑，以便使用新的分页参数和响应格式。

#### 验收标准

1. WHEN 请求分页数据 THEN Pagination_Logic SHALL 使用 page 和 page_size 参数替代 limit 和 offset
2. WHEN 处理分页响应 THEN Pagination_Logic SHALL 从 pagination 对象中提取分页信息
3. WHEN 计算页码信息 THEN Pagination_Logic SHALL 使用 total_pages 字段而非手动计算
4. WHEN 处理游标分页 THEN Pagination_Logic SHALL 支持 next_cursor 的无限滚动模式
5. WHEN 显示分页控件 THEN Pagination_Logic SHALL 使用新的分页数据结构

### 需求 7

**用户故事:** 作为前端开发者，我希望清理冗余的 API 接口代码，以便提高代码质量和可维护性。

#### 验收标准

1. WHEN 审查 API 调用代码 THEN Frontend SHALL 移除重复的 fetch 调用和错误处理逻辑
2. WHEN 统一 API 方法 THEN Frontend SHALL 将所有直接的 fetch 调用迁移到统一的 API 客户端
3. WHEN 清理类型定义 THEN Frontend SHALL 移除过时的接口定义和重复的类型声明
4. WHEN 重构组件 THEN Frontend SHALL 更新所有使用旧 API 格式的 React 组件
5. WHEN 优化导入 THEN Frontend SHALL 整理和优化 API 相关的模块导入

### 需求 8

**用户故事:** 作为前端开发者，我希望更新任务管理和聊天功能的 API 调用，以便与后端保持一致。

#### 验收标准

1. WHEN 获取任务列表 THEN Frontend SHALL 使用统一的 API 客户端和错误处理
2. WHEN 启动任务 THEN Frontend SHALL 处理新的任务操作响应格式
3. WHEN 停止任务 THEN Frontend SHALL 使用标准化的操作结果处理
4. WHEN 处理聊天消息 THEN Frontend SHALL 适配新的消息格式和错误处理
5. WHEN 管理对话 THEN Frontend SHALL 使用统一的 CRUD 操作接口

### 需求 9

**用户故事:** 作为前端开发者，我希望更新文件读取相关的 API 调用，以便保持接口的一致性。

#### 验收标准

1. WHEN 获取书籍目录 THEN Frontend SHALL 使用统一的错误处理机制
2. WHEN 读取章节内容 THEN Frontend SHALL 处理标准化的文件访问响应
3. WHEN 下载书籍文件 THEN Frontend SHALL 使用一致的文件操作接口
4. WHEN 访问封面图片 THEN Frontend SHALL 处理统一的资源访问格式
5. WHEN 处理文件错误 THEN Frontend SHALL 提供用户友好的文件访问错误提示

### 需求 10

**用户故事:** 作为质量保证工程师，我希望有完整的测试覆盖，以便确保 API 接口迁移的正确性。

#### 验收标准

1. WHEN 测试 API 客户端 THEN Frontend SHALL 包含成功响应、错误响应和网络错误的测试用例
2. WHEN 测试分页逻辑 THEN Frontend SHALL 验证新旧分页格式的兼容性处理
3. WHEN 测试错误处理 THEN Frontend SHALL 确保所有错误场景都有适当的用户提示
4. WHEN 测试类型安全 THEN Frontend SHALL 验证 TypeScript 类型定义的正确性
5. WHEN 进行回归测试 THEN Frontend SHALL 确保所有现有功能在迁移后正常工作

### 需求 11

**用户故事:** 作为前端开发者，我希望有向后兼容的迁移策略，以便平滑过渡到新的 API 格式。

#### 验收标准

1. WHEN 处理混合格式 THEN Frontend SHALL 支持新旧两种响应格式的自动识别
2. WHEN 创建适配器 THEN Frontend SHALL 提供旧格式到新格式的转换函数
3. WHEN 渐进式迁移 THEN Frontend SHALL 允许按模块逐步迁移到新格式
4. WHEN 回滚支持 THEN Frontend SHALL 保留快速回退到旧格式的能力
5. WHEN 监控迁移 THEN Frontend SHALL 记录迁移过程中的问题和性能指标

### 需求 12

**用户故事:** 作为前端开发者，我希望优化 API 调用的性能，以便提升用户体验。

#### 验收标准

1. WHEN 缓存 API 响应 THEN Frontend SHALL 实现适当的缓存策略减少重复请求
2. WHEN 批量操作 THEN Frontend SHALL 合并相似的 API 请求减少网络开销
3. WHEN 预加载数据 THEN Frontend SHALL 在适当时机预取用户可能需要的数据
4. WHEN 处理大数据集 THEN Frontend SHALL 使用分页和虚拟滚动优化性能
5. WHEN 监控性能 THEN Frontend SHALL 记录 API 调用的响应时间和成功率

### 需求 13

**用户故事:** 作为前端开发者，我希望有完善的文档和代码注释，以便团队成员理解新的 API 接口使用方式。

#### 验收标准

1. WHEN 编写 API 文档 THEN Frontend SHALL 提供所有 API 方法的使用示例和参数说明
2. WHEN 添加代码注释 THEN Frontend SHALL 在关键的 API 调用处添加清晰的注释
3. WHEN 创建迁移指南 THEN Frontend SHALL 提供从旧格式到新格式的详细迁移步骤
4. WHEN 记录最佳实践 THEN Frontend SHALL 文档化 API 调用的推荐模式和注意事项
5. WHEN 维护变更日志 THEN Frontend SHALL 记录所有 API 接口的变更历史和影响范围