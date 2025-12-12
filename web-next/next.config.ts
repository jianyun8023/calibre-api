import type { NextConfig } from "next";
import bundleAnalyzer from '@next/bundle-analyzer';
import createNextIntlPlugin from 'next-intl/plugin';

// Bundle Analyzer 配置
const withBundleAnalyzer = bundleAnalyzer({
  enabled: process.env.ANALYZE === 'true',
})

// next-intl 配置
const withNextIntl = createNextIntlPlugin('./src/i18n/request.ts');

// 解析 API_BASE_URL 获取后端服务器配置
const apiBaseUrl = process.env.API_BASE_URL || 'http://localhost:8080';
const apiUrl = new URL(apiBaseUrl);

const nextConfig: NextConfig = {
  // 启用 standalone 输出模式用于 Docker 部署
  output: 'standalone',
  
  // 使用 SSR 模式，支持服务端渲染和 API 代理
  // API 请求会被代理到 Go 后端
  async rewrites() {
    return [
      { 
        source: '/api/:path*', 
        destination: `${apiBaseUrl}/api/:path*`
      }
    ]
  },
  // 图片优化配置
  // 动态配置允许从后端加载图片的域名
  images: {
    formats: ['image/webp', 'image/avif'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920],
    imageSizes: [16, 32, 48, 64, 96, 128, 256],
    remotePatterns: [
      {
        protocol: apiUrl.protocol.replace(':', '') as 'http' | 'https',
        hostname: apiUrl.hostname,
        port: apiUrl.port,
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
