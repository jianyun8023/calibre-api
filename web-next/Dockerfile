FROM node:24-slim AS builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV CI=true
ENV COREPACK_ENABLE_STRICT=0
RUN corepack enable

WORKDIR /app

# 复制依赖文件并安装所有依赖（包括 devDependencies，构建时需要）
COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile

# 复制源代码
COPY . ./

# 构建应用（使用 standalone 模式）
# standalone 模式会自动追踪依赖，只包含运行时需要的文件
RUN pnpm build

## Runner - Standalone 模式运行时镜像
FROM node:24-slim

# 多架构支持
ARG TARGETPLATFORM
RUN echo "Running on $TARGETPLATFORM"

# 安装必要的系统依赖
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 创建非 root 用户以提高安全性
RUN groupadd -r nodejs && useradd -r -g nodejs nodejs

# 复制 standalone 构建产物（最小化依赖）
COPY --from=builder --chown=nodejs:nodejs /app/.next/standalone ./
# 复制静态资源
COPY --from=builder --chown=nodejs:nodejs /app/.next/static ./.next/static
# 复制 public 目录
COPY --from=builder --chown=nodejs:nodejs /app/public ./public

# 切换到非 root 用户
USER nodejs

# 设置环境变量
ENV NODE_ENV=production
ENV PORT=3000
ENV NEXT_TELEMETRY_DISABLED=1

EXPOSE 3000

# 使用 standalone 模式的入口文件
CMD ["node", "server.js"]
