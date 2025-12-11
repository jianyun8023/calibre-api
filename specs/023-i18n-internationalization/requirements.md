# Requirements Document

## Introduction

为 calibre-api 的 Next.js 前端添加国际化（i18n）支持，使应用能够支持多语言界面。默认语言为中文（简体中文），同时支持英文，用户可以在界面中切换语言。这将提升应用的可用性和用户体验，特别是对于中文用户群体。

## Glossary

- **i18n**: Internationalization（国际化）的缩写，指使软件能够适应不同语言和地区的过程
- **Locale**: 语言区域设置，如 zh-CN（简体中文）、en-US（美式英语）
- **next-intl**: Next.js 的国际化库，支持 App Router 和服务端组件
- **Translation Key**: 翻译键，用于标识需要翻译的文本
- **Language Switcher**: 语言切换器，允许用户在不同语言之间切换的 UI 组件
- **Frontend**: Next.js 前端应用（web-next 目录）

## Requirements

### Requirement 1

**User Story:** 作为用户，我希望应用默认显示中文界面，这样我可以更自然地使用应用。

#### Acceptance Criteria

1. WHEN a user first visits the application THEN the system SHALL display the interface in Chinese (zh-CN)
2. WHEN the browser locale is zh-CN or zh THEN the system SHALL use Chinese as the display language
3. WHEN no language preference is stored THEN the system SHALL default to Chinese
4. WHEN the application loads THEN the system SHALL detect and apply the appropriate locale before rendering content
5. WHEN Chinese is selected THEN the system SHALL display all UI text, labels, buttons, and messages in Chinese

### Requirement 2

**User Story:** 作为用户，我希望能够切换到英文界面，这样我可以根据需要选择合适的语言。

#### Acceptance Criteria

1. WHEN a user clicks the language switcher THEN the system SHALL display available language options (Chinese and English)
2. WHEN a user selects English THEN the system SHALL immediately update all UI text to English
3. WHEN a user selects a language THEN the system SHALL persist the language preference in browser storage
4. WHEN a user returns to the application THEN the system SHALL load and apply the previously selected language
5. WHEN the language changes THEN the system SHALL update the page without requiring a full page reload

### Requirement 3

**User Story:** 作为开发者，我希望有一个结构化的翻译管理系统，这样我可以轻松添加和维护多语言内容。

#### Acceptance Criteria

1. WHEN translation files are organized THEN the system SHALL store translations in separate JSON files per locale
2. WHEN a translation key is used THEN the system SHALL retrieve the corresponding text for the current locale
3. WHEN a translation key is missing THEN the system SHALL display the translation key as fallback text
4. WHEN adding new UI text THEN the system SHALL support nested translation keys for better organization
5. WHEN translations are updated THEN the system SHALL reflect changes without requiring code modifications

### Requirement 4

**User Story:** 作为开发者，我希望 i18n 系统与 Next.js App Router 完全兼容，这样我可以在服务端和客户端组件中使用翻译。

#### Acceptance Criteria

1. WHEN using translations in Server Components THEN the system SHALL provide server-side translation functions
2. WHEN using translations in Client Components THEN the system SHALL provide client-side translation hooks
3. WHEN rendering pages THEN the system SHALL support locale-based routing (e.g., /zh-CN/books, /en-US/books)
4. WHEN generating static pages THEN the system SHALL generate pages for all supported locales
5. WHEN the locale changes THEN the system SHALL update the URL to reflect the current locale

### Requirement 5

**User Story:** 作为用户，我希望所有主要功能都有完整的中英文翻译，这样我可以完整地使用应用的所有功能。

#### Acceptance Criteria

1. WHEN viewing the book list THEN the system SHALL display all book-related UI text in the selected language
2. WHEN using search functionality THEN the system SHALL display search UI and placeholders in the selected language
3. WHEN viewing book details THEN the system SHALL display metadata labels and actions in the selected language
4. WHEN using the chat feature THEN the system SHALL display chat UI and prompts in the selected language
5. WHEN viewing settings THEN the system SHALL display all settings labels and descriptions in the selected language
