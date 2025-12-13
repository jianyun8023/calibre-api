import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Proxy 用于运行时动态代理 API 和 MCP 请求
 * 
 * 在 Standalone 模式下，next.config.ts 中的 rewrites 会在构建时固化，
 * 无法在运行时读取环境变量。因此我们使用 Proxy 来实现动态代理。
 * 
 * Next.js 16+ 已将 middleware 重命名为 proxy 以明确其用途。
 * 
 * @see https://nextjs.org/docs/app/building-your-application/routing/proxy
 * @see https://nextjs.org/docs/messages/middleware-to-proxy
 */
export function proxy(request: NextRequest) {
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
  
  // 其他请求正常处理
  return NextResponse.next();
}

/**
 * 配置 Proxy 匹配的路径
 * 只在 /api/* 和 /mcp/* 路径上运行，提高性能
 */
export const config = {
  matcher: [
    '/api/:path*',
    '/mcp/:path*',
  ],
};
