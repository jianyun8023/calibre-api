# Requirements Document

## Introduction

本规格文档定义了更新 GitHub CI workflow 和 docker-compose.yaml 配置的需求。项目已从旧的 Vue.js 前端（app/calibre-pages）迁移到新的 Next.js 前端（web-next），需要更新构建和部署配置以支持新架构。

## Glossary

- **GitHub Actions**: GitHub 提供的 CI/CD 自动化平台
- **Docker Compose**: 用于定义和运行多容器 Docker 应用的工具
- **Multi-stage Build**: Docker 多阶段构建，用于优化镜像大小
- **Next.js**: React 框架，支持 SSR 和 SSG
- **Calibre API**: 本项目的后端 API 服务
- **GHCR**: GitHub Container Registry，GitHub 的容器镜像仓库

## Requirements

### Requirement 1

**User Story:** 作为开发者，我希望 Dockerfile 使用正确的前端目录，以便构建包含最新 Next.js 前端的镜像。

#### Acceptance Criteria

1. WHEN 构建 Docker 镜像 THEN Dockerfile SHALL 从 web-next 目录复制前端代码而非 app/calibre-pages
2. WHEN 安装前端依赖 THEN 系统 SHALL 使用 web-next/package.json 中定义的依赖
3. WHEN 构建前端 THEN 系统 SHALL 执行 Next.js 的生产构建命令
4. WHEN 复制构建产物 THEN 系统 SHALL 将 Next.js 的输出目录复制到最终镜像
5. WHEN 设置环境变量 THEN 系统 SHALL 移除已废弃的 CALIBRE_TEMPLATE_DIR 和 CALIBRE_STATIC_DIR

### Requirement 2

**User Story:** 作为开发者，我希望 GitHub CI workflow 使用最新的 Actions 版本，以便获得更好的性能和安全性。

#### Acceptance Criteria

1. WHEN workflow 运行 THEN 系统 SHALL 使用 actions/checkout@v4 而非 v3
2. WHEN 设置 Docker Buildx THEN 系统 SHALL 使用 docker/setup-buildx-action@v3 而非 v2
3. WHEN 设置 QEMU THEN 系统 SHALL 使用 docker/setup-qemu-action@v3 而非 v2
4. WHEN 登录容器仓库 THEN 系统 SHALL 使用 docker/login-action@v3 而非 v2
5. WHEN 构建和推送镜像 THEN 系统 SHALL 使用 docker/build-push-action@v5 而非 v4

### Requirement 3

**User Story:** 作为开发者，我希望 docker-compose.yaml 包含所有必要的环境变量，以便正确配置应用运行环境。

#### Acceptance Criteria

1. WHEN 启动服务 THEN docker-compose SHALL 包含 Qdrant 向量数据库的连接配置
2. WHEN 配置 AI 功能 THEN docker-compose SHALL 包含 OpenAI API 相关环境变量
3. WHEN 配置存储 THEN docker-compose SHALL 包含数据卷挂载配置
4. WHEN 服务启动 THEN docker-compose SHALL 确保服务间的依赖关系正确
5. WHEN 使用网络 THEN docker-compose SHALL 定义适当的网络配置

### Requirement 4

**User Story:** 作为运维人员，我希望 docker-compose 支持可选的 Qdrant 服务，以便在需要时启动本地向量数据库。

#### Acceptance Criteria

1. WHEN 需要本地 Qdrant THEN docker-compose SHALL 定义 Qdrant 服务配置
2. WHEN Qdrant 启动 THEN 系统 SHALL 暴露必要的端口（6333, 6334）
3. WHEN 持久化数据 THEN 系统 SHALL 挂载 Qdrant 数据卷
4. WHEN 配置网络 THEN Qdrant SHALL 与 calibre-api 在同一网络中
5. WHEN 服务可选 THEN 配置 SHALL 允许通过 profile 控制是否启动 Qdrant

### Requirement 5

**User Story:** 作为开发者，我希望 CI workflow 支持缓存，以便加快构建速度。

#### Acceptance Criteria

1. WHEN 构建 Docker 镜像 THEN 系统 SHALL 使用 Docker layer 缓存
2. WHEN 安装 Go 依赖 THEN 系统 SHALL 缓存 Go modules
3. WHEN 安装前端依赖 THEN 系统 SHALL 缓存 pnpm store
4. WHEN 后续构建运行 THEN 系统 SHALL 复用缓存以减少构建时间
5. WHEN 缓存失效 THEN 系统 SHALL 自动重新构建缓存

### Requirement 6

**User Story:** 作为开发者，我希望构建过程清晰记录版本信息，以便追踪部署的代码版本。

#### Acceptance Criteria

1. WHEN 发布版本 THEN 系统 SHALL 使用 release tag 作为镜像版本号
2. WHEN 推送代码 THEN dev workflow SHALL 使用 "dev" 作为镜像标签
3. WHEN 构建镜像 THEN 系统 SHALL 将版本信息传递给构建参数
4. WHEN 推送镜像 THEN 系统 SHALL 同时推送版本标签和 latest 标签（仅 release）
5. WHEN 查看镜像 THEN 用户 SHALL 能够通过标签识别镜像版本

### Requirement 7

**User Story:** 作为开发者，我希望 Dockerfile 优化构建层级，以便减小最终镜像大小。

#### Acceptance Criteria

1. WHEN 构建前端 THEN 系统 SHALL 使用独立的构建阶段
2. WHEN 构建后端 THEN 系统 SHALL 使用独立的构建阶段
3. WHEN 创建最终镜像 THEN 系统 SHALL 仅复制必要的运行时文件
4. WHEN 安装系统依赖 THEN 系统 SHALL 清理 apt 缓存
5. WHEN 比较镜像大小 THEN 优化后的镜像 SHALL 小于未优化的镜像

### Requirement 8

**User Story:** 作为运维人员，我希望 docker-compose 支持健康检查，以便监控服务状态。

#### Acceptance Criteria

1. WHEN 服务启动 THEN docker-compose SHALL 定义 calibre-api 的健康检查
2. WHEN 检查健康状态 THEN 系统 SHALL 通过 HTTP 端点验证服务可用性
3. WHEN 服务不健康 THEN Docker SHALL 标记容器为 unhealthy 状态
4. WHEN 配置重启策略 THEN 系统 SHALL 在服务失败时自动重启
5. WHEN 查看状态 THEN 用户 SHALL 能够通过 docker-compose ps 查看健康状态
