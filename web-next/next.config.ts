import type { NextConfig } from "next";
import bundleAnalyzer from '@next/bundle-analyzer';
import createNextIntlPlugin from 'next-intl/plugin';

// Bundle Analyzer 配置
const withBundleAnalyzer = bundleAnalyzer({
  enabled: process.env.ANALYZE === 'true',
})

// next-intl 配置
const withNextIntl = createNextIntlPlugin('./src/i18n/request.ts');

const nextConfig: NextConfig = {
  // 启用 standalone 输出模式用于 Docker 部署
  output: 'standalone',
  
  // 使用 SSR 模式，支持服务端渲染和 API 代理
  // API 和 MCP 请求会被代理到 Go 后端
  async rewrites() {
    // 运行时动态获取 API_BASE_URL
    const apiBaseUrl = process.env.API_BASE_URL || 'http://localhost:8080';
    return [
      { 
        source: '/api/:path*', 
        destination: `${apiBaseUrl}/api/:path*`
      },
      // MCP 协议端点转发（用于 AI 助手集成）
      { 
        source: '/mcp/:path*', 
        destination: `${apiBaseUrl}/mcp/:path*`
      }
    ]
  },
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

export default withBundleAnalyzer(withNextIntl(nextConfig));
