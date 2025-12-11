---
status: planned
created: '2025-12-11'
tags:
  - devops
  - ci-cd
  - docker
  - github-actions
priority: medium
created_at: '2025-12-11T14:58:40.023Z'
---

# GitHub CI 和 Docker Compose 配置更新

> **Status**: 📅 Planned · **Priority**: Medium · **Created**: 2025-12-11

## Overview

更新 GitHub CI workflow 和 docker-compose.yaml 配置，支持新的 Next.js 前端架构（web-next），优化构建流程，添加完整的环境变量配置和可选的 Qdrant 服务支持。

## Documents

- [requirements.md](requirements.md) - 需求文档和验收标准（8 个主要需求）
- [design.md](design.md) - 技术设计和架构决策（多阶段构建、CI/CD 流程）
- [tasks.md](tasks.md) - 实施任务分解（10 个任务）

## Status

📋 **Planned** - 需求和设计已完成，准备开始实施

**主要变更**:
- Dockerfile: 从 app/calibre-pages 迁移到 web-next
- GitHub Actions: 升级所有 actions 到最新版本
- Docker Compose: 添加完整环境变量和可选 Qdrant 服务
- 优化: 添加构建缓存和健康检查
