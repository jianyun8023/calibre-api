import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';
import createMiddleware from 'next-intl/middleware';
import { locales, defaultLocale } from './i18n/config';

/**
 * Proxy 用于运行时动态代理 API 和 MCP 请求
 * 同时集成 next-intl middleware 处理 locale 路由
 * 
 * 在 Standalone 模式下，next.config.ts 中的 rewrites 会在构建时固化，
 * 无法在运行时读取环境变量。因此我们使用 Proxy 来实现动态代理。
 * 
 * Next.js 16+ 已将 middleware 重命名为 proxy 以明确其用途。
 * 
 * @see https://nextjs.org/docs/app/building-your-application/routing/proxy
 * @see https://nextjs.org/docs/messages/middleware-to-proxy
 * @see https://next-intl.dev/docs/routing/middleware
 */

// 创建 next-intl middleware
const handleI18nRouting = createMiddleware({
  locales,
  defaultLocale,
  localePrefix: 'always' // 总是显示 locale 前缀
});

export default async function proxy(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  
  // 运行时动态读取 API_BASE_URL 环境变量
  const apiBaseUrl = process.env.API_BASE_URL || 'http://localhost:8080';
  
  // 代理 /api/* 请求到后端
  if (pathname.startsWith('/api/')) {
    const backendUrl = `${apiBaseUrl}${pathname}${search}`;
    
    console.log(`[Proxy] Proxying ${pathname} -> ${backendUrl}`);
    
    return NextResponse.rewrite(new URL(backendUrl));
  }
  
  // 代理 /mcp/* 请求到后端（MCP 协议支持）
  if (pathname.startsWith('/mcp/')) {
    const backendUrl = `${apiBaseUrl}${pathname}${search}`;
    
    console.log(`[Proxy] Proxying ${pathname} -> ${backendUrl}`);
    
    return NextResponse.rewrite(new URL(backendUrl));
  }
  
  // 其他请求使用 next-intl middleware 处理 locale 路由
  return handleI18nRouting(request);
}

/**
 * 配置 Proxy 匹配的路径
 * 基于 next-intl 官方文档，但允许 /api/ 和 /mcp/ 路径包含点（图片、文件）
 * @see https://next-intl.dev/docs/routing/middleware
 */
export const config = {
  matcher: [
    // 匹配 /api/ 开头的所有路径（包括图片等文件）
    '/api/:path*',
    // 匹配 /mcp/ 开头的所有路径
    '/mcp/:path*',
    // 匹配其他路径，但排除静态资源
    '/((?!api|mcp|_next|_vercel|.*\\..*).*)' 
  ]
};
