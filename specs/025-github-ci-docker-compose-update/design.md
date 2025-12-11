# Design Document

## Overview

本设计文档描述了如何更新 GitHub CI workflow 和 docker-compose.yaml 配置，以支持新的 Next.js 前端架构，优化构建流程，并改进部署配置。

## Architecture

### 构建流程架构

```
┌─────────────────────────────────────────────────────────────┐
│                     GitHub Actions Workflow                  │
├─────────────────────────────────────────────────────────────┤
│  Trigger: Push / Release                                     │
│     ↓                                                        │
│  Checkout Code (v4)                                          │
│     ↓                                                        │
│  Setup QEMU (v3) + Buildx (v3)                              │
│     ↓                                                        │
│  Login to GHCR (v3)                                          │
│     ↓                                                        │
│  Build Multi-stage Docker Image (v5)                         │
│     ├─ Stage 1: Build Next.js Frontend                      │
│     ├─ Stage 2: Build Go Backend                            │
│     └─ Stage 3: Create Runtime Image                        │
│     ↓                                                        │
│  Push to ghcr.io                                             │
└─────────────────────────────────────────────────────────────┘
```

### Docker Compose 架构

```
┌──────────────────────────────────────────────────────────────┐
│                    Docker Compose Stack                       │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────┐         ┌──────────────────┐           │
│  │  calibre-api    │────────▶│  Qdrant          │           │
│  │  (main service) │         │  (optional)      │           │
│  │  Port: 8080     │         │  Ports: 6333,6334│           │
│  └─────────────────┘         └──────────────────┘           │
│         │                              │                     │
│         │                              │                     │
│    ┌────▼────┐                   ┌────▼────┐               │
│    │ Volumes │                   │ Volumes │               │
│    │ /data   │                   │ /qdrant │               │
│    └─────────┘                   └─────────┘               │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. Dockerfile (Multi-stage Build)

**Stage 1: Frontend Build (web-next)**
- Base: `node:20-slim`
- 工作目录: `/app`
- 构建工具: pnpm
- 输出: Next.js standalone build

**Stage 2: Backend Build (Go)**
- Base: `golang:1.23-bookworm`
- 工作目录: `/app`
- 构建产物: `/calibre-api` 二进制文件

**Stage 3: Runtime**
- Base: `debian:bookworm-slim`
- 包含: CA 证书、calibre-api 二进制、Next.js 构建产物
- 暴露端口: 8080

### 2. GitHub Actions Workflows

**build.yaml (Production)**
- 触发器: Release published, Manual dispatch
- 平台: linux/amd64, linux/arm64
- 标签: `{version}`, `latest`

**build-dev.yaml (Development)**
- 触发器: Push to any branch, Manual dispatch
- 平台: linux/amd64, linux/arm64
- 标签: `dev`

### 3. Docker Compose Services

**calibre-api (主服务)**
- 镜像: `ghcr.io/jianyun8023/calibre-api:latest`
- 端口: 8080:8080
- 重启策略: unless-stopped
- 健康检查: HTTP GET /api/health

**qdrant (可选服务)**
- 镜像: `qdrant/qdrant:latest`
- 端口: 6333:6333, 6334:6334
- Profile: qdrant
- 数据卷: qdrant_data

## Data Models

### Environment Variables

```yaml
# Calibre API 配置
CALIBRE_DEBUG: boolean              # 调试模式
CALIBRE_CONTENT_SERVER: string      # Calibre 内容服务器 URL
CALIBRE_SEARCH_INDEX: string        # 搜索索引名称
CALIBRE_TMPDIR: string              # 临时文件目录
CALIBRE_METADATA_DOUBANURL: string  # 豆瓣元数据服务 URL
CALIBRE_MCP_ENABLED: boolean        # MCP 协议启用状态

# Qdrant 配置
CALIBRE_QDRANT_URL: string          # Qdrant 服务地址
CALIBRE_QDRANT_API_KEY: string      # Qdrant API 密钥（可选）

# OpenAI 配置
OPENAI_API_KEY: string              # OpenAI API 密钥
OPENAI_BASE_URL: string             # OpenAI API 基础 URL（可选）
OPENAI_MODEL: string                # 使用的模型名称

# 系统配置
TZ: string                          # 时区设置
```

### Docker Build Arguments

```dockerfile
ARG VERSION=dev                     # 构建版本号
ARG NODE_VERSION=20                 # Node.js 版本
ARG GO_VERSION=1.23                 # Go 版本
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Dockerfile 路径正确性

*For any* Docker 构建过程，前端代码路径应该指向 `web-next` 目录，而不是已废弃的 `app/calibre-pages` 目录

**Validates: Requirements 1.1**

### Property 2: Actions 版本一致性

*For any* GitHub workflow 文件，所有 actions 应该使用最新的稳定版本（checkout@v4, buildx@v3, qemu@v3, login@v3, build-push@v5）

**Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5**

### Property 3: 环境变量完整性

*For any* docker-compose 配置，必须包含所有必要的环境变量以确保服务正常运行

**Validates: Requirements 3.1, 3.2**

### Property 4: 镜像标签正确性

*For any* release 构建，应该同时推送版本标签和 latest 标签；对于 dev 构建，应该只推送 dev 标签

**Validates: Requirements 6.1, 6.2, 6.4**

### Property 5: 多阶段构建隔离性

*For any* Docker 构建阶段，最终运行时镜像应该只包含必要的运行时文件，不包含构建工具和源代码

**Validates: Requirements 7.3**

### Property 6: 健康检查有效性

*For any* 运行中的 calibre-api 容器，健康检查应该能够正确反映服务的实际状态

**Validates: Requirements 8.1, 8.2**

### Property 7: 服务依赖正确性

*For any* docker-compose 启动过程，如果启用 Qdrant profile，calibre-api 应该能够成功连接到 Qdrant 服务

**Validates: Requirements 4.4**

## Error Handling

### Dockerfile 错误处理

1. **依赖安装失败**: 使用 `--mount=type=cache` 缓存依赖，失败时清除缓存重试
2. **构建失败**: 多阶段构建确保每个阶段独立，失败时易于定位
3. **文件不存在**: 使用 `COPY` 前验证源路径存在

### GitHub Actions 错误处理

1. **构建失败**: Workflow 自动失败，发送通知
2. **推送失败**: 重试机制，最多 3 次
3. **认证失败**: 使用 `GITHUB_TOKEN` 自动认证，无需手动配置

### Docker Compose 错误处理

1. **服务启动失败**: `restart: unless-stopped` 自动重启
2. **健康检查失败**: 标记为 unhealthy，可配置重启策略
3. **网络连接失败**: 使用 Docker 内部 DNS 解析服务名

## Testing Strategy

### 单元测试

1. **Dockerfile 语法测试**
   - 使用 `docker build --dry-run` 验证语法
   - 验证所有 COPY 路径存在

2. **GitHub Actions 语法测试**
   - 使用 `actionlint` 验证 workflow 语法
   - 验证所有 actions 版本有效

3. **Docker Compose 语法测试**
   - 使用 `docker-compose config` 验证配置
   - 验证环境变量引用正确

### 集成测试

1. **本地构建测试**
   ```bash
   docker build -t calibre-api:test .
   docker run -p 8080:8080 calibre-api:test
   curl http://localhost:8080/api/health
   ```

2. **Docker Compose 测试**
   ```bash
   docker-compose up -d
   docker-compose ps
   docker-compose logs calibre-api
   curl http://localhost:8080/api/health
   ```

3. **多平台构建测试**
   ```bash
   docker buildx build --platform linux/amd64,linux/arm64 -t test .
   ```

### 端到端测试

1. **CI Pipeline 测试**
   - 触发 dev workflow，验证构建成功
   - 验证镜像推送到 GHCR
   - 拉取镜像并运行，验证服务正常

2. **部署测试**
   - 使用 docker-compose 部署
   - 验证所有服务启动
   - 验证健康检查通过
   - 验证服务间通信正常

## Implementation Notes

### Dockerfile 关键变更

1. **前端路径更新**
   ```dockerfile
   # 旧: COPY ./app/calibre-pages/package.json
   # 新: COPY ./web-next/package.json
   ```

2. **Next.js 构建输出**
   ```dockerfile
   # Next.js standalone 模式输出
   COPY --from=app /app/.next/standalone ./
   COPY --from=app /app/.next/static ./.next/static
   COPY --from=app /app/public ./public
   ```

3. **移除废弃环境变量**
   ```dockerfile
   # 移除: ENV CALIBRE_TEMPLATE_DIR=/app/templates
   # 移除: ENV CALIBRE_STATIC_DIR=/app/static
   ```

### GitHub Actions 关键变更

1. **Actions 版本升级**
   - checkout: v3 → v4
   - setup-buildx-action: v2 → v3
   - setup-qemu-action: v2 → v3
   - login-action: v2 → v3
   - build-push-action: v4 → v5

2. **缓存配置**
   ```yaml
   cache-from: type=gha
   cache-to: type=gha,mode=max
   ```

### Docker Compose 关键变更

1. **添加 Qdrant 服务**
   ```yaml
   qdrant:
     image: qdrant/qdrant:latest
     profiles: ["qdrant"]
     ports:
       - "6333:6333"
       - "6334:6334"
     volumes:
       - qdrant_data:/qdrant/storage
   ```

2. **添加健康检查**
   ```yaml
   healthcheck:
     test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
     interval: 30s
     timeout: 10s
     retries: 3
     start_period: 40s
   ```

3. **添加数据卷**
   ```yaml
   volumes:
     calibre_data:
     qdrant_data:
   ```

## Performance Considerations

1. **构建缓存**: 使用 GitHub Actions cache 和 Docker layer cache 减少构建时间
2. **多阶段构建**: 减小最终镜像大小，提高部署速度
3. **并行构建**: 支持 amd64 和 arm64 平台并行构建
4. **pnpm 缓存**: 使用 pnpm store 缓存加速前端依赖安装

## Security Considerations

1. **最小化镜像**: 使用 debian:bookworm-slim 作为运行时基础镜像
2. **非 root 用户**: 考虑添加非 root 用户运行服务
3. **密钥管理**: 使用环境变量传递敏感信息，不硬编码
4. **镜像扫描**: 建议添加 Trivy 或 Snyk 扫描镜像漏洞
