# Implementation Plan

- [ ] 1. 更新 Dockerfile 支持 Next.js 前端
  - 更新前端构建阶段，从 web-next 目录复制代码
  - 修改 package.json 路径为 web-next/package.json
  - 更新构建命令以支持 Next.js standalone 模式
  - 复制 Next.js 构建产物（.next/standalone, .next/static, public）
  - 移除废弃的 CALIBRE_TEMPLATE_DIR 和 CALIBRE_STATIC_DIR 环境变量
  - 更新 Go 版本从 1.24.10 到 1.23（修正版本号错误）
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

- [ ] 2. 升级 GitHub Actions workflows 到最新版本
  - 更新 build.yaml 中的 actions 版本
  - 更新 build-dev.yaml 中的 actions 版本
  - 升级 checkout 从 v3 到 v4
  - 升级 setup-buildx-action 从 v2 到 v3
  - 升级 setup-qemu-action 从 v2 到 v3
  - 升级 login-action 从 v2 到 v3
  - 升级 build-push-action 从 v4 到 v5
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 3. 添加 GitHub Actions 构建缓存
  - 在 build.yaml 中添加 cache-from 和 cache-to 配置
  - 在 build-dev.yaml 中添加 cache-from 和 cache-to 配置
  - 配置 GitHub Actions cache 类型
  - 设置 cache mode 为 max 以最大化缓存效果
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 4. 更新 docker-compose.yaml 配置
  - 添加完整的环境变量配置（Qdrant, OpenAI）
  - 添加 calibre-api 健康检查配置
  - 添加数据卷配置（calibre_data）
  - 更新网络配置
  - 确保重启策略正确设置
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 8.1, 8.2, 8.4_

- [ ] 5. 添加可选的 Qdrant 服务
  - 在 docker-compose.yaml 中定义 Qdrant 服务
  - 配置 Qdrant 端口映射（6333, 6334）
  - 添加 Qdrant 数据卷挂载
  - 设置 profile 为 "qdrant" 使其可选
  - 确保 Qdrant 与 calibre-api 在同一网络
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [ ] 6. 本地测试 Dockerfile 构建
  - 执行 docker build 验证构建成功
  - 验证前端文件正确复制
  - 验证后端二进制文件正确生成
  - 检查最终镜像大小
  - 运行容器并验证服务启动
  - _Requirements: 7.1, 7.2, 7.3, 7.4_

- [ ] 7. 测试 docker-compose 配置
  - 执行 docker-compose config 验证语法
  - 启动服务并验证 calibre-api 正常运行
  - 验证健康检查状态
  - 测试使用 --profile qdrant 启动 Qdrant 服务
  - 验证服务间网络连接
  - _Requirements: 3.4, 4.4, 8.3, 8.5_

- [ ] 8. 创建 .dockerignore 文件（如果不存在）
  - 添加 node_modules 到忽略列表
  - 添加 .git 到忽略列表
  - 添加 .next 到忽略列表
  - 添加其他不需要的文件和目录
  - _Requirements: 7.5_

- [ ] 9. 更新项目文档
  - 更新 README.md 中的部署说明
  - 添加 docker-compose 使用示例
  - 添加环境变量配置说明
  - 添加 Qdrant 可选服务说明
  - _Requirements: 3.1, 4.5_

- [ ] 10. Checkpoint - 验证所有配置
  - 确保所有文件语法正确
  - 确保本地构建和运行成功
  - 确保 docker-compose 启动成功
  - 询问用户是否有问题
