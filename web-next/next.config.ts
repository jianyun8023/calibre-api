import type { NextConfig } from "next";
import createNextIntlPlugin from 'next-intl/plugin';

// next-intl 配置
const withNextIntl = createNextIntlPlugin('./src/i18n/request.ts');

// 注意：已移除 @next/bundle-analyzer 以支持 Standalone 模式
// 如需分析包大小，请在本地开发环境使用：
// npm install --save-dev @next/bundle-analyzer
// 然后取消注释下面的代码

const nextConfig: NextConfig = {
  // 启用 Standalone 输出模式（最小化镜像）
  output: 'standalone',
  
  // 注意：Standalone 模式下 rewrites 会在构建时固化
  // 因此我们使用 Proxy 实现运行时动态代理
  // 参见：src/proxy.ts
  
  // 图片优化配置
  // 允许从多种来源加载图片（本地开发、Docker、生产环境）
  images: {
    formats: ['image/webp', 'image/avif'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920],
    imageSizes: [16, 32, 48, 64, 96, 128, 256],
    remotePatterns: [
      // 本地开发环境
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8080',
        pathname: '/api/**',
      },
      // Docker Compose 环境
      {
        protocol: 'http',
        hostname: 'calibre-api',
        port: '8080',
        pathname: '/api/**',
      },
      // 支持任意 hostname（生产环境可能需要限制）
      {
        protocol: 'http',
        hostname: '**',
        pathname: '/api/**',
      },
      {
        protocol: 'https',
        hostname: '**',
        pathname: '/api/**',
      },
    ],
  },
  // 暂时忽略 TypeScript 错误
  typescript: {
    ignoreBuildErrors: true,
  },
};

export default withNextIntl(nextConfig);
