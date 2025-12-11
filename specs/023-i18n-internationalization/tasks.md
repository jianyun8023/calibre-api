# Implementation Plan

- [ ] 1. Setup i18n infrastructure
- [x] 1.1 Install next-intl and dependencies
  - Run `pnpm add next-intl` in web-next directory
  - Run `pnpm add -D fast-check @types/jest` for testing
  - _Requirements: All_

- [x] 1.2 Create i18n configuration files
  - Create `web-next/src/i18n/config.ts` with locale definitions
  - Define supported locales: zh-CN (default), en-US
  - Define locale display names
  - _Requirements: 1.1, 1.3, 2.1_

- [x] 1.3 Create next-intl request configuration
  - Create `web-next/src/i18n/request.ts`
  - Implement getRequestConfig for message loading
  - Add locale validation
  - _Requirements: 3.1, 3.2_

- [x] 1.4 Create initial translation files
  - Create `web-next/messages/zh-CN.json` with Chinese translations
  - Create `web-next/messages/en-US.json` with English translations
  - Add common translations (buttons, labels, navigation)
  - _Requirements: 3.1, 3.2_

- [ ] 1.5 Write unit test for translation file structure
  - **Example 6: Page translation completeness (partial)**
  - **Validates: Requirements 3.1**
  - Verify both translation files exist
  - Verify both files have matching key structure

- [ ] 2. Implement locale-based routing
- [x] 2.1 Create middleware for locale detection
  - Create `web-next/src/middleware.ts`
  - Use next-intl's createMiddleware
  - Configure locale detection and routing
  - _Requirements: 1.2, 1.4, 4.3_

- [x] 2.2 Restructure app directory for locale routing
  - Move `web-next/src/app/*` to `web-next/src/app/[locale]/*`
  - Update all page.tsx files to accept locale param
  - Preserve existing page functionality
  - _Requirements: 4.3, 4.5_

- [x] 2.3 Update root layout with NextIntlClientProvider
  - Modify `web-next/src/app/[locale]/layout.tsx`
  - Wrap application with NextIntlClientProvider
  - Load messages for current locale
  - Update html lang attribute to use locale
  - _Requirements: 1.4, 4.1, 4.2_

- [ ] 2.4 Write property test for locale routing
  - **Property 5: Locale-based routing**
  - **Validates: Requirements 4.3**
  - Generate random page paths for each locale
  - Verify URLs with locale prefix are accessible
  - Verify correct locale is applied

- [ ] 2.5 Write unit test for default locale
  - **Example 1: Default Chinese on first visit**
  - **Validates: Requirements 1.1, 1.3**
  - Clear storage and visit app
  - Verify Chinese is selected
  - Verify URL contains /zh-CN/

- [ ] 3. Implement language switcher component
- [x] 3.1 Create LanguageSwitcher component
  - Create `web-next/src/components/language-switcher.tsx`
  - Use useLocale and useRouter from next-intl
  - Implement locale switching logic
  - Use existing UI components (Select, DropdownMenu)
  - _Requirements: 2.1, 2.2, 2.5_

- [x] 3.2 Add LanguageSwitcher to AppHeader
  - Import and render LanguageSwitcher in AppHeader component
  - Position in header navigation area
  - _Requirements: 2.1_

- [ ] 3.3 Write unit test for language switcher
  - **Example 2: Language switcher functionality**
  - **Validates: Requirements 2.1, 2.2, 2.4**
  - Render language switcher
  - Verify both locales shown
  - Test switching to English
  - Verify persistence after reload

- [ ] 3.4 Write property test for language persistence
  - **Property 2: Language preference persistence**
  - **Validates: Requirements 2.3**
  - For each locale, select it and verify storage
  - Reload and verify locale is restored

- [ ] 4. Migrate common components to use translations
- [x] 4.1 Migrate AppSidebar component
  - Extract hardcoded navigation labels
  - Add translation keys to translation files
  - Use useTranslations() hook
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 4.2 Migrate AppHeader component
  - Extract hardcoded text
  - Add translation keys
  - Use useTranslations() hook
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [ ] 4.3 Migrate common UI components
  - Update button labels, tooltips, placeholders
  - Add translations for common actions (save, cancel, delete, etc.)
  - _Requirements: 1.5_

- [ ] 4.4 Write unit test for component translations
  - **Example 5: Client component translations**
  - **Validates: Requirements 4.2**
  - Render client component with useTranslations
  - Verify translated text appears correctly

- [ ] 5. Migrate book-related pages
- [ ] 5.1 Migrate books list page
  - Update `web-next/src/app/[locale]/books/page.tsx`
  - Extract all hardcoded strings
  - Add book-related translations
  - Use getTranslations() for server component
  - _Requirements: 5.1_

- [ ] 5.2 Migrate book detail page
  - Update `web-next/src/app/[locale]/detail/[id]/page.tsx`
  - Extract metadata labels and actions
  - Add detail page translations
  - _Requirements: 5.3_

- [ ] 5.3 Migrate BookGrid component
  - Update `web-next/src/components/book-grid.tsx`
  - Extract book card labels
  - Add translations for rating, author, publisher labels
  - _Requirements: 5.1_

- [ ] 5.4 Write unit test for server component translations
  - **Example 4: Server component translations**
  - **Validates: Requirements 4.1**
  - Test getTranslations() in server context
  - Verify translations render on server

- [ ] 6. Migrate search functionality
- [ ] 6.1 Migrate search page
  - Update `web-next/src/app/[locale]/search/page.tsx`
  - Extract search UI text and placeholders
  - Add search-related translations
  - _Requirements: 5.2_

- [ ] 6.2 Migrate search filter components
  - Update filter components in `web-next/src/app/search/components/`
  - Extract filter labels and options
  - Add filter translations
  - _Requirements: 5.2_

- [ ] 7. Migrate chat and settings pages
- [ ] 7.1 Migrate chat page
  - Update `web-next/src/app/[locale]/chat/page.tsx`
  - Extract chat UI text and prompts
  - Add chat-related translations
  - _Requirements: 5.4_

- [ ] 7.2 Migrate settings page
  - Update `web-next/src/app/[locale]/settings/page.tsx`
  - Extract settings labels and descriptions
  - Add settings translations
  - _Requirements: 5.5_

- [ ] 7.3 Migrate tasks page
  - Update `web-next/src/app/[locale]/tasks/page.tsx`
  - Extract task-related text
  - Add task translations
  - _Requirements: 5.5_

- [ ] 8. Implement property-based tests
- [ ] 8.1 Write property test for browser locale detection
  - **Property 1: Browser locale detection**
  - **Validates: Requirements 1.2**
  - Generate various zh-* locales
  - Verify system selects Chinese for all

- [ ] 8.2 Write property test for translation key retrieval
  - **Property 3: Translation key retrieval**
  - **Validates: Requirements 3.2**
  - Generate random valid keys from translation files
  - Verify correct translations returned for each locale

- [ ] 8.3 Write property test for missing key fallback
  - **Property 4: Missing translation key fallback**
  - **Validates: Requirements 3.3**
  - Generate random invalid keys
  - Verify key itself is returned as fallback

- [ ] 9. Add URL redirects for backward compatibility
- [ ] 9.1 Update middleware to redirect old URLs
  - Add redirect logic for non-locale URLs
  - Redirect `/books` to `/zh-CN/books`
  - Preserve query parameters
  - _Requirements: 4.3_

- [ ] 9.2 Update Next.js config for redirects
  - Add permanent redirects in next.config.ts
  - Handle common old URL patterns
  - _Requirements: 4.3_

- [ ] 10. Final integration testing
- [ ] 10.1 Write example test for URL updates on language change
  - **Example 3: URL updates on language change**
  - **Validates: Requirements 4.5**
  - Switch from Chinese to English
  - Verify URL changes from /zh-CN/ to /en-US/

- [ ] 10.2 Write comprehensive page translation test
  - **Example 6: Page translation completeness**
  - **Validates: Requirements 1.5, 5.1, 5.2, 5.3, 5.4, 5.5**
  - For each major page, test both locales
  - Verify no hardcoded text remains

- [ ] 11. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.
