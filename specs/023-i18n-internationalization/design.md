# Design Document

## Overview

本设计为 calibre-api 的 Next.js 前端实现国际化（i18n）支持。采用 next-intl 库，这是专为 Next.js App Router 设计的国际化解决方案，支持服务端组件和客户端组件。默认语言为简体中文（zh-CN），同时支持英文（en-US），用户可通过语言切换器在两种语言间切换。

## Architecture

### Current Architecture
```
┌─────────────────────────────────────┐
│      Next.js Frontend (web-next)    │
│  ┌──────────────────────────────┐  │
│  │   App Router Pages           │  │
│  │   - Hardcoded English text   │  │
│  │   - No i18n support          │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

### Target Architecture
```
┌─────────────────────────────────────────────────┐
│         Next.js Frontend (web-next)             │
│  ┌──────────────────────────────────────────┐  │
│  │   Locale-based Routing                   │  │
│  │   /[locale]/... (e.g., /zh-CN/books)     │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │   next-intl Provider                     │  │
│  │   - Locale detection                     │  │
│  │   - Translation loading                  │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │   Translation Files                      │  │
│  │   messages/zh-CN.json                    │  │
│  │   messages/en-US.json                    │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │   Components with Translations           │  │
│  │   - useTranslations() hook               │  │
│  │   - getTranslations() server function    │  │
│  └──────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────┐  │
│  │   Language Switcher Component            │  │
│  │   - Locale selection UI                  │  │
│  │   - Cookie-based persistence             │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. i18n Configuration

**File:** `web-next/src/i18n/config.ts`

```typescript
export const locales = ['zh-CN', 'en-US'] as const;
export type Locale = (typeof locales)[number];

export const defaultLocale: Locale = 'zh-CN';

export const localeNames: Record<Locale, string> = {
  'zh-CN': '简体中文',
  'en-US': 'English',
};
```

### 2. next-intl Request Configuration

**File:** `web-next/src/i18n/request.ts`

Provides locale detection and message loading for server components:

```typescript
import { getRequestConfig } from 'next-intl/server';
import { notFound } from 'next/navigation';
import { locales } from './config';

export default getRequestConfig(async ({ locale }) => {
  if (!locales.includes(locale as any)) notFound();
  
  return {
    messages: (await import(`../../messages/${locale}.json`)).default
  };
});
```

### 3. Root Layout with Locale

**File:** `web-next/src/app/[locale]/layout.tsx`

Wraps the application with NextIntlClientProvider:

```typescript
import { NextIntlClientProvider } from 'next-intl';
import { getMessages } from 'next-intl/server';

export default async function LocaleLayout({
  children,
  params: { locale }
}: {
  children: React.ReactNode;
  params: { locale: string };
}) {
  const messages = await getMessages();
  
  return (
    <html lang={locale}>
      <body>
        <NextIntlClientProvider messages={messages}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  );
}
```

### 4. Language Switcher Component

**File:** `web-next/src/components/language-switcher.tsx`

Client component for language selection:

```typescript
'use client';

import { useLocale } from 'next-intl';
import { useRouter, usePathname } from 'next/navigation';
import { locales, localeNames } from '@/i18n/config';

export function LanguageSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  
  const switchLocale = (newLocale: string) => {
    const newPath = pathname.replace(`/${locale}`, `/${newLocale}`);
    router.push(newPath);
  };
  
  return (
    <Select value={locale} onValueChange={switchLocale}>
      {locales.map(loc => (
        <SelectItem key={loc} value={loc}>
          {localeNames[loc]}
        </SelectItem>
      ))}
    </Select>
  );
}
```

### 5. Translation Files Structure

**Directory:** `web-next/messages/`

**File:** `zh-CN.json`
```json
{
  "common": {
    "search": "搜索",
    "loading": "加载中...",
    "error": "错误",
    "save": "保存",
    "cancel": "取消"
  },
  "nav": {
    "books": "书籍",
    "search": "搜索",
    "chat": "对话",
    "settings": "设置"
  },
  "books": {
    "title": "书籍列表",
    "author": "作者",
    "publisher": "出版社",
    "rating": "评分"
  }
}
```

**File:** `en-US.json`
```json
{
  "common": {
    "search": "Search",
    "loading": "Loading...",
    "error": "Error",
    "save": "Save",
    "cancel": "Cancel"
  },
  "nav": {
    "books": "Books",
    "search": "Search",
    "chat": "Chat",
    "settings": "Settings"
  },
  "books": {
    "title": "Book List",
    "author": "Author",
    "publisher": "Publisher",
    "rating": "Rating"
  }
}
```

### 6. Middleware for Locale Detection

**File:** `web-next/src/middleware.ts`

Handles locale detection and routing:

```typescript
import createMiddleware from 'next-intl/middleware';
import { locales, defaultLocale } from './i18n/config';

export default createMiddleware({
  locales,
  defaultLocale,
  localePrefix: 'always' // Always show locale in URL
});

export const config = {
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)']
};
```

## Data Models

### Locale Type
```typescript
type Locale = 'zh-CN' | 'en-US';
```

### Translation Message Structure
```typescript
interface Messages {
  [namespace: string]: {
    [key: string]: string | Messages;
  };
}
```

### Locale Configuration
```typescript
interface LocaleConfig {
  locales: readonly Locale[];
  defaultLocale: Locale;
  localeNames: Record<Locale, string>;
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

分析所有验收标准后，识别出以下合并：
- 标准 1.1 和 1.3 重复（都测试默认中文）
- 标准 1.5, 5.1-5.5 都测试页面翻译完整性，可合并为一个综合示例
- 标准 2.1, 2.2, 2.4 都测试语言切换流程，可合并
- 标准 1.2, 2.3, 3.2, 3.3, 4.3 是真正的属性，适用于多个输入

### Properties

Property 1: Browser locale detection
*For any* browser locale starting with 'zh' (e.g., zh-CN, zh-TW, zh-HK), the system should detect and use Chinese (zh-CN) as the display language
**Validates: Requirements 1.2**

Property 2: Language preference persistence
*For any* selected locale, when a user selects that locale, the system should store it in browser storage (cookie or localStorage) and retrieve it on subsequent visits
**Validates: Requirements 2.3**

Property 3: Translation key retrieval
*For any* valid translation key and supported locale, the system should return the corresponding translated text for that locale
**Validates: Requirements 3.2**

Property 4: Missing translation key fallback
*For any* translation key that does not exist in the translation files, the system should display the key itself as fallback text
**Validates: Requirements 3.3**

Property 5: Locale-based routing
*For any* supported locale (zh-CN, en-US), the system should support accessing pages with that locale prefix in the URL (e.g., /[locale]/books)
**Validates: Requirements 4.3**

### Examples

Example 1: Default Chinese on first visit
When a user visits the application for the first time with no stored language preference, verify that:
- The interface displays in Chinese (zh-CN)
- The URL includes /zh-CN/ prefix
- All visible UI text is in Chinese
**Validates: Requirements 1.1, 1.3**

Example 2: Language switcher functionality
When testing the language switcher:
- Click the language switcher and verify both Chinese and English options are displayed
- Select English and verify all UI text immediately updates to English
- Reload the page and verify English is still selected
**Validates: Requirements 2.1, 2.2, 2.4**

Example 3: URL updates on language change
When a user switches from Chinese to English:
- Verify the URL changes from /zh-CN/[page] to /en-US/[page]
- Verify the page content updates without full reload
**Validates: Requirements 4.5**

Example 4: Server component translations
When rendering a server component that uses translations:
- Verify the component can call getTranslations() function
- Verify translated text is rendered correctly on the server
**Validates: Requirements 4.1**

Example 5: Client component translations
When rendering a client component that uses translations:
- Verify the component can use useTranslations() hook
- Verify translated text is rendered correctly in the browser
**Validates: Requirements 4.2**

Example 6: Page translation completeness
For each major page (books, search, detail, chat, settings):
- Switch to Chinese and verify all UI elements display Chinese text
- Switch to English and verify all UI elements display English text
- Verify no hardcoded text remains untranslated
**Validates: Requirements 1.5, 5.1, 5.2, 5.3, 5.4, 5.5**

## Error Handling

### Missing Translation Keys

**Behavior:**
- Display the translation key itself as fallback
- Log warning in development mode
- Example: If key "books.newFeature" is missing, display "books.newFeature"

### Invalid Locale

**Behavior:**
- Redirect to default locale (zh-CN)
- Return 404 if locale is not in supported list
- Middleware handles this before reaching page components

### Translation File Loading Errors

**Behavior:**
- Catch errors during message import
- Fall back to empty messages object
- Log error for debugging

## Testing Strategy

### Unit Tests

Unit tests will verify specific examples and integration points:

1. **Default Locale Test**
   - Clear all storage
   - Load application
   - Verify Chinese is selected
   - Verify URL contains /zh-CN/

2. **Language Switcher Test**
   - Render language switcher component
   - Verify both locales are shown
   - Click English option
   - Verify locale changes

3. **Translation Function Test**
   - Test getTranslations() in server context
   - Test useTranslations() in client context
   - Verify correct translations are returned

4. **Page Translation Test**
   - For each major page, verify key UI elements are translated
   - Test both Chinese and English locales

### Property-Based Tests

Property-based tests will verify universal properties across many inputs:

1. **Property Test: Browser Locale Detection**
   - Generate various zh-* locales (zh-CN, zh-TW, zh-HK, zh-SG)
   - For each, verify system selects Chinese
   - Verify non-zh locales default to Chinese

2. **Property Test: Translation Key Retrieval**
   - Generate random valid translation keys from translation files
   - For each key and each locale, verify correct translation is returned
   - Verify translations are non-empty strings

3. **Property Test: Missing Key Fallback**
   - Generate random invalid translation keys
   - For each key, verify the key itself is returned as fallback
   - Verify no errors are thrown

4. **Property Test: Locale Routing**
   - For each supported locale, generate random page paths
   - Verify URLs with locale prefix are accessible
   - Verify correct locale is applied to the page

### Testing Framework

- **Unit Testing**: Jest + React Testing Library
- **Property-Based Testing**: [fast-check](https://github.com/dubzzz/fast-check) - JavaScript property testing library
- **E2E Testing**: Playwright for full user flow testing
- **Minimum Iterations**: 100 iterations per property test

## Implementation Notes

### Migration Strategy

1. **Phase 1: Setup Infrastructure**
   - Install next-intl
   - Create i18n configuration
   - Set up middleware
   - Create translation files with initial keys

2. **Phase 2: Update Routing**
   - Move app directory to app/[locale]
   - Update layout to use NextIntlClientProvider
   - Test routing works with locale prefix

3. **Phase 3: Component Migration**
   - Start with common components (header, sidebar, navigation)
   - Extract hardcoded strings to translation keys
   - Add translations to both zh-CN.json and en-US.json
   - Test each component after migration

4. **Phase 4: Page Migration**
   - Migrate pages one by one
   - Priority: books → search → detail → chat → settings
   - Verify each page is fully translated

5. **Phase 5: Language Switcher**
   - Implement language switcher component
   - Add to header/navigation
   - Test switching between languages

### Translation Key Naming Convention

Use nested structure for organization:

```
{namespace}.{component}.{element}
```

Examples:
- `common.button.save` - Common save button
- `books.list.title` - Book list page title
- `search.filter.author` - Search filter author label
- `chat.input.placeholder` - Chat input placeholder

### Backward Compatibility

**Breaking Changes:**
- URLs will change from `/books` to `/zh-CN/books`
- Need to set up redirects for old URLs

**Migration Path:**
- Add middleware redirect from old paths to new locale-prefixed paths
- Update any hardcoded links in external documentation

### Performance Considerations

- Translation files are loaded per-locale, not all at once
- Server components can use translations without client-side JavaScript
- Consider code-splitting for large translation files
- Cache translation files in production

## Dependencies

**Depends on:**
- Spec 009-frontend-migration (completed) - Next.js App Router must be in place

**Blocks:**
- None - This is an enhancement

**External Dependencies:**
- next-intl: ^3.0.0 (Next.js 13+ App Router support)
- fast-check: ^3.0.0 (for property-based testing)

## References

- next-intl Documentation: https://next-intl-docs.vercel.app/
- Next.js Internationalization: https://nextjs.org/docs/app/building-your-application/routing/internationalization
- fast-check Documentation: https://fast-check.dev/
- React Testing Library: https://testing-library.com/docs/react-testing-library/intro/
