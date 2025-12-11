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
  // 使用 SSR 模式，支持服务端渲染和 API 代理
  // API 请求会被代理到 Go 后端
  async rewrites() {
    return [
      { 
        source: '/api/:path*', 
        destination: process.env.API_BASE_URL || 'http://localhost:8080/api/:path*'
      }
    ]
  },
  // 图片优化配置
  images: {
    formats: ['image/webp', 'image/avif'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920],
    imageSizes: [16, 32, 48, 64, 96, 128, 256],
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8080',
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
