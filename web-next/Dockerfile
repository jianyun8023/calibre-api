FROM node:20-slim AS deps

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

# 复制依赖文件
COPY package.json pnpm-lock.yaml ./

# pnpm 会自动处理平台特定的原生依赖
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile --prod

## Builder
FROM node:20-slim AS builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

# 复制依赖文件并安装所有依赖（包括 devDependencies）
COPY package.json pnpm-lock.yaml ./

# Next.js 构建是跨平台的，不需要特殊处理
RUN --mount=type=cache,id=pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile

# 复制源代码
COPY . ./

# 构建应用（不使用 standalone 模式）
RUN pnpm build

## Runner
FROM node:20-slim

# 多架构支持
ARG TARGETPLATFORM
RUN echo "Running on $TARGETPLATFORM"

# 安装必要的系统依赖（某些架构可能需要）
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 创建非 root 用户以提高安全性
RUN groupadd -r nodejs && useradd -r -g nodejs nodejs

# 复制生产依赖
COPY --from=deps --chown=nodejs:nodejs /app/node_modules ./node_modules

# 复制构建产物
COPY --from=builder --chown=nodejs:nodejs /app/.next ./.next
COPY --from=builder --chown=nodejs:nodejs /app/public ./public
COPY --from=builder --chown=nodejs:nodejs /app/package.json ./package.json

# 复制必需的配置文件
COPY --from=builder --chown=nodejs:nodejs /app/next.config.ts ./next.config.ts
COPY --from=builder --chown=nodejs:nodejs /app/messages ./messages
COPY --from=builder --chown=nodejs:nodejs /app/src/i18n ./src/i18n

# 切换到非 root 用户
USER nodejs

# 设置环境变量
ENV NODE_ENV=production
ENV PORT=3000
ENV NEXT_TELEMETRY_DISABLED=1

EXPOSE 3000

# 健康检查（检查 Next.js 服务是否响应）
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/', (r) => {process.exit(r.statusCode < 500 ? 0 : 1)})" || exit 1

# 使用 next start 启动（支持运行时环境变量）
CMD ["npx", "next", "start"]
