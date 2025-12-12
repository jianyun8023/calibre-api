# Deployment Guide

## 架构说明

### 服务架构

```
┌─────────────────────────────────────────────────────────────┐
│                         用户浏览器                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                    calibre-web (前端)                         │
│                    Port: 3000                                 │
│                    Next.js Standalone                         │
├─────────────────────────────────────────────────────────────┤
│  • 服务端渲染 (SSR)                                           │
│  • API 请求代理                                               │
│  • 静态资源服务                                               │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ /api/* 请求
                              │ (通过 Docker 网络)
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   calibre-api (后端)                          │
│                   Port: 8080                                  │
│                   Go + Gin Framework                          │
├─────────────────────────────────────────────────────────────┤
│  • RESTful API                                                │
│  • MCP 协议支持                                               │
│  • 书籍管理                                                   │
│  • 搜索功能                                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ↓
                    ┌─────────────────┐
                    │  Calibre Server │
                    │  Qdrant (可选)  │
                    │  OpenAI (可选)  │
                    └─────────────────┘
```

## 前端访问后端配置

### 1. Docker Compose 环境

在 Docker Compose 中，前端通过 **Docker 内部网络** 访问后端：

```yaml
calibre-web:
  environment:
    # 服务端 API 代理配置
    # 使用 Docker 服务名称作为主机名（基础 URL，代理会自动拼接 /api/:path*）
    - API_BASE_URL=http://calibre-api:8080
```

**工作原理**:
1. 用户浏览器访问 `http://localhost:3000`
2. 前端页面发起 API 请求到 `/api/books`
3. Next.js 服务端通过 rewrites 将请求代理到 `http://calibre-api:8080/api/books`
4. Docker 网络解析 `calibre-api` 为后端容器的内部 IP
5. 后端处理请求并返回数据
6. Next.js 将数据返回给浏览器

### 2. 本地开发环境

在本地开发时，前端和后端分别运行：

```bash
# 终端 1: 启动后端
cd /path/to/calibre-api
go run main.go

# 终端 2: 启动前端
cd /path/to/calibre-api/web-next
pnpm dev
```

**配置**:
```bash
# web-next/.env.local
API_BASE_URL=http://localhost:8080
```

### 3. 生产环境（Kubernetes）

在 Kubernetes 中，使用 Service 名称访问：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: calibre-web-config
data:
  API_BASE_URL: "http://calibre-api-service:8080"
```

### 4. 反向代理环境（Nginx）

如果使用 Nginx 作为反向代理：

```nginx
# Nginx 配置
upstream calibre_api {
    server calibre-api:8080;
}

upstream calibre_web {
    server calibre-web:3000;
}

server {
    listen 80;
    server_name example.com;

    # 前端
    location / {
        proxy_pass http://calibre_web;
    }

    # API 直接代理到后端（可选）
    location /api/ {
        proxy_pass http://calibre_api;
    }
}
```

## 环境变量说明

### 前端环境变量 (calibre-web)

| 变量名 | 说明 | 示例值 | 必需 |
|--------|------|--------|------|
| `API_BASE_URL` | 后端 API 基础地址（不包含路径） | `http://calibre-api:8080` | ✅ |
| `NODE_ENV` | Node.js 运行环境 | `production` | ✅ |
| `PORT` | 前端服务端口 | `3000` | ❌ |
| `TZ` | 时区设置 | `Asia/Shanghai` | ❌ |

### 后端环境变量 (calibre-api)

| 变量名 | 说明 | 示例值 | 必需 |
|--------|------|--------|------|
| `CALIBRE_CONTENT_SERVER` | Calibre 内容服务器地址 | `https://lib.pve.icu` | ✅ |
| `CALIBRE_DEBUG` | 调试模式 | `true` / `false` | ❌ |
| `CALIBRE_SEARCH_INDEX` | 搜索索引名称 | `books` | ✅ |
| `CALIBRE_QDRANT_URL` | Qdrant 服务地址 | `http://qdrant:6333` | ❌ |
| `OPENAI_API_KEY` | OpenAI API 密钥 | `sk-...` | ❌ |

## 部署步骤

### 使用 Docker Compose

#### 1. 基础部署（仅前后端）

```bash
# 克隆仓库
git clone https://github.com/jianyun8023/calibre-api.git
cd calibre-api

# 配置环境变量（可选）
cp docker-compose.yaml docker-compose.override.yaml
# 编辑 docker-compose.override.yaml 修改配置

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 访问服务
# 前端: http://localhost:3000
# 后端: http://localhost:8080
```

#### 2. 完整部署（包含 Qdrant）

```bash
# 启动所有服务（包括 Qdrant）
docker-compose --profile qdrant up -d

# 查看服务状态
docker-compose ps

# 访问 Qdrant 管理界面
# http://localhost:6333/dashboard
```

#### 3. 仅启动后端

```bash
docker-compose up -d calibre-api
```

#### 4. 仅启动前端

```bash
docker-compose up -d calibre-web
```

### 使用预构建镜像

```bash
# 拉取镜像
docker pull ghcr.io/jianyun8023/calibre-api:latest
docker pull ghcr.io/jianyun8023/calibre-web:latest

# 运行后端
docker run -d \
  --name calibre-api \
  -p 8080:8080 \
  -e CALIBRE_CONTENT_SERVER=https://lib.pve.icu \
  ghcr.io/jianyun8023/calibre-api:latest

# 运行前端
docker run -d \
  --name calibre-web \
  -p 3000:3000 \
  -e API_BASE_URL=http://calibre-api:8080 \
  --link calibre-api \
  ghcr.io/jianyun8023/calibre-web:latest
```

## 健康检查

### 后端健康检查

```bash
# 检查后端状态
curl http://localhost:8080/ping

# 预期响应
{"message":"pong"}
```

### 前端健康检查

```bash
# 检查前端状态
curl http://localhost:3000

# 预期响应: HTML 页面
```

### Docker Compose 健康检查

```bash
# 查看服务健康状态
docker-compose ps

# 预期输出
NAME            IMAGE                                    STATUS
calibre-api     ghcr.io/jianyun8023/calibre-api:latest   Up (healthy)
calibre-web     ghcr.io/jianyun8023/calibre-web:latest   Up (healthy)
```

## 故障排查

### 前端无法访问后端

**症状**: 前端页面加载正常，但 API 请求失败

**检查步骤**:

1. 检查后端是否运行
   ```bash
   docker-compose ps calibre-api
   curl http://localhost:8080/ping
   ```

2. 检查前端环境变量
   ```bash
   docker-compose exec calibre-web env | grep API_BASE_URL
   ```

3. 检查 Docker 网络
   ```bash
   docker network inspect calibre-api_calibre-network
   ```

4. 检查前端日志
   ```bash
   docker-compose logs calibre-web
   ```

**常见问题**:

- ❌ `API_BASE_URL=http://localhost:8080` - 错误！容器内 localhost 指向自己
- ❌ `API_BASE_URL=http://calibre-api:8080/api/:path*` - 错误！不应包含路径部分
- ✅ `API_BASE_URL=http://calibre-api:8080` - 正确！使用 Docker 服务名，代理会自动拼接路径

### 服务依赖问题

**症状**: 前端启动失败，提示后端不可用

**解决方案**:

1. 确保后端健康检查配置正确
2. 增加前端启动等待时间
   ```yaml
   calibre-web:
     depends_on:
       calibre-api:
         condition: service_healthy
     healthcheck:
       start_period: 60s  # 增加启动等待时间
   ```

### CORS 问题

**症状**: 浏览器控制台显示 CORS 错误

**说明**: 由于 Next.js 使用服务端代理，不应该出现 CORS 问题。如果出现，检查：

1. 前端是否正确配置了 rewrites
2. 是否直接从浏览器访问后端 API（应该通过前端代理）

## 性能优化

### 1. 启用 HTTP/2

```nginx
server {
    listen 443 ssl http2;
    # ...
}
```

### 2. 启用 Gzip 压缩

Next.js 默认启用，无需额外配置。

### 3. 配置 CDN

将静态资源上传到 CDN：

```typescript
// next.config.ts
const nextConfig: NextConfig = {
  assetPrefix: process.env.CDN_URL || '',
  // ...
}
```

### 4. 数据库连接池

后端配置数据库连接池以提高性能。

## 安全建议

1. **使用 HTTPS**: 生产环境必须使用 HTTPS
2. **限制 CORS**: 配置允许的来源域名
3. **API 认证**: 实施 JWT 或 OAuth 认证
4. **环境变量**: 不要在代码中硬编码敏感信息
5. **定期更新**: 保持依赖包和基础镜像更新

## 监控和日志

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f calibre-api
docker-compose logs -f calibre-web

# 查看最近 100 行日志
docker-compose logs --tail=100 calibre-api
```

### 监控指标

建议使用 Prometheus + Grafana 监控：

- CPU 使用率
- 内存使用率
- API 响应时间
- 错误率
- 请求量

## 备份和恢复

### 备份数据卷

```bash
# 备份 calibre_data
docker run --rm \
  -v calibre-api_calibre_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/calibre_data_backup.tar.gz /data

# 备份 qdrant_data
docker run --rm \
  -v calibre-api_qdrant_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/qdrant_data_backup.tar.gz /data
```

### 恢复数据卷

```bash
# 恢复 calibre_data
docker run --rm \
  -v calibre-api_calibre_data:/data \
  -v $(pwd):/backup \
  alpine sh -c "cd / && tar xzf /backup/calibre_data_backup.tar.gz"
```

## 更新和升级

### 更新到最新版本

```bash
# 拉取最新镜像
docker-compose pull

# 重启服务
docker-compose up -d

# 清理旧镜像
docker image prune -f
```

### 回滚到之前版本

```bash
# 使用特定版本标签
docker-compose down
# 编辑 docker-compose.yaml 修改镜像标签
docker-compose up -d
```
