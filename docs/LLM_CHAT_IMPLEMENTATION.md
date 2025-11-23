# Calibre API 大模型问答功能实施方案

基于 LangChainGo 框架，支持思考模式，实现对话持久化的完整方案。

## 技术栈

### 后端
- **LangChainGo** - LLM 框架
- **SQLite** - 对话历史存储
- **Ollama/OpenAI** - LLM 提供商
- **SSE** - 流式响应

### 前端
- **@aivue/chatbot** - AI 聊天组件
- **vue-markdown-x** - Markdown 渲染
- **@vueuse/core** - SSE 工具
- **Element Plus** - UI 组件

## 核心功能

1. **智能书籍推荐** - 基于自然语言推荐相关书籍
2. **语义搜索** - 理解自然语言查询
3. **TOC 内容总结** - 根据书籍目录生成摘要
4. **思考模式** - 显示 AI 推理过程
5. **对话持久化** - SQLite 存储对话历史

## 实施步骤

### 阶段 1：后端基础设施（1-2天）

#### 1.1 数据库层
- [ ] 创建 `internal/chat/db.go`
- [ ] 创建 `internal/chat/migrations/001_create_chat_tables.sql`
- [ ] 实现对话和消息的 CRUD 操作

#### 1.2 LangChainGo 集成
- [ ] 安装依赖：`go get github.com/tmc/langchaingo@latest`
- [ ] 创建 `internal/chat/llm.go` - LLM 客户端
- [ ] 创建 `internal/chat/agent.go` - Agent 封装
- [ ] 创建 `internal/chat/tools.go` - 工具实现

#### 1.3 思考模式支持
- [ ] 创建 `internal/chat/thinking.go`
- [ ] 实现思考标签解析

### 阶段 2：API 层（1天）

#### 2.1 HTTP 处理函数
- [ ] 创建 `internal/calibre/chat_handler.go`
- [ ] 实现 `CreateConversation`
- [ ] 实现 `ListConversations`
- [ ] 实现 `GetConversationMessages`
- [ ] 实现 `SendMessage`（SSE 流式）

#### 2.2 路由注册
- [ ] 更新 `internal/calibre/route.go`
- [ ] 添加聊天相关路由

#### 2.3 配置更新
- [ ] 更新 `config.yaml` 添加 LLM 配置

### 阶段 3：前端实现（2-3天）

#### 3.1 依赖安装
```bash
cd app/calibre-pages
npm install @aivue/chatbot vue-markdown-x
```

#### 3.2 核心组件
- [ ] 创建 `src/views/Chat.vue` - 主聊天界面
- [ ] 创建 `src/components/BookCard.vue` - 书籍卡片
- [ ] 创建 `src/api/chat.ts` - API 客户端

#### 3.3 样式定制
- [ ] 更新 `src/styles/index.scss`
- [ ] 统一 Element Plus 风格

#### 3.4 路由集成
- [ ] 更新 `src/router/index.ts`
- [ ] 更新 `src/components/Sidebar.vue`

### 阶段 4：测试验证（1天）

#### 4.1 后端测试
- [ ] 单元测试（数据库、Agent、工具）
- [ ] API 集成测试

#### 4.2 前端测试
- [ ] 界面交互测试
- [ ] 流式响应测试
- [ ] 思考过程显示测试

#### 4.3 端到端测试
- [ ] 完整对话流程
- [ ] 书籍推荐功能
- [ ] TOC 总结功能

### 阶段 5：优化与标准化（进行中）

#### 5.1 组件化重构（下一步重点）
- [ ] **布局标准化**：使用 `el-container`, `el-main`, `el-footer` 重构 `Chat.vue` 布局，替代自定义 CSS Flex 布局，确保跨设备兼容性和可维护性。
- [ ] **组件拆分**：
    - `src/views/chat/components/MessageList.vue`：负责消息列表渲染和滚动控制。
    - `src/views/chat/components/ChatInput.vue`：负责输入框逻辑和发送事件。
    - `src/views/chat/components/BookCard.vue`：独立书籍卡片组件，包含“总结”按钮逻辑。
    - `src/views/chat/components/ThinkingProcess.vue`：独立思考过程展示组件。
- [ ] **逻辑抽离**：创建 `src/composables/useChat.ts`，管理消息状态、发送逻辑、SSE 连接和分页逻辑。

#### 5.2 功能增强
- [ ] **Markdown 渲染优化**：引入 `markdown-it` 替代简单的正则替换，支持代码高亮、表格等丰富格式。
- [ ] **错误处理标准化**：统一使用 `ElMessage` 和 `ElNotification` 处理网络错误和业务异常。
- [ ] **加载状态优化**：在书籍卡片加载图片时添加骨架屏（Skeleton）。

## 已完成功能（v1.0）

- [x] **基础对话**：支持流式响应和思考模式。
- [x] **书籍推荐**：
    - 支持语义搜索推荐书籍。
    - **卡片展示**：显示封面、标题、作者。
    - **分页浏览**：支持“换一换”功能（每页 8 本）。
- [x] **内容总结**：
    - 基于 TOC 的一键总结功能。
- [x] **UI 适配**：
    - 输入框固定底部。
    - 消息自动滚动。
    - 移动端/桌面端自适应布局。

## 参考文档
- [LangChainGo 官方文档](https://tmc.github.io/langchaingo/docs/)
- [Element Plus Container 布局](https://element-plus.org/zh-CN/component/container.html)
- [Vue 3 Composables](https://vuejs.org/guide/reusability/composables.html)
- [实施计划详细版](./implementation_plan.md)
