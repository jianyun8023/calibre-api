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
 * 根据 next-intl 官方文档建议的配置
 * @see https://next-intl.dev/docs/routing/middleware
 */
export const config = {
  // Match all pathnames except for
  // - … if they start with `/_next` or `/_vercel`  
  // - … the ones containing a dot (e.g. `favicon.ico`)
  matcher: '/((?!_next|_vercel|.*\\..*).*)'
};
