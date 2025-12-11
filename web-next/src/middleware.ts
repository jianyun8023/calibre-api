import createMiddleware from 'next-intl/middleware';
import { locales, defaultLocale } from './i18n/config';

export default createMiddleware({
  // A list of all locales that are supported
  locales,
  
  // Used when no locale matches
  defaultLocale,
  
  // Always show locale in URL for clarity
  // Root path / will redirect to /zh-CN
  localePrefix: 'always'
});

export const config = {
  // Match all pathnames except for
  // - /api routes
  // - /_next (Next.js internals)
  // - /_vercel (Vercel internals)
  // - Static files (e.g. /favicon.ico, /sitemap.xml, /robots.txt, etc.)
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)']
};
