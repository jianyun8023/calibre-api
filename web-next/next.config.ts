import type { NextConfig } from "next";

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

export default nextConfig;
