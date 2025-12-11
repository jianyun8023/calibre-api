# Design Document

## Overview

本设计文档描述了 EPUB 阅读器增强功能的技术实现。系统使用 localStorage 进行本地数据持久化，集成用户设置系统，并提供书签、高亮、笔记等阅读辅助功能。

## Architecture

### 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    Reader Page                           │
│                                                          │
│  ┌────────────────┐         ┌──────────────────┐       │
│  │  ReactReader   │◄────────┤ useReaderData    │       │
│  │  Component     │         │ Hook             │       │
│  └────────┬───────┘         └──────────┬───────┘       │
│           │                            │                │
│           │                            ▼                │
│           │                 ┌──────────────────┐       │
│           │                 │ Reader Context   │       │
│           │                 │ - Bookmarks      │       │
│           │                 │ - Highlights     │       │
│           │                 │ - Notes          │       │
│           │                 │ - Progress       │       │
│           │                 └──────────┬───────┘       │
│           │                            │                │
│           ▼                            ▼                │
│  ┌────────────────┐         ┌──────────────────┐       │
│  │  Reader        │         │  Settings        │       │
│  │  Sidebar       │         │  Context         │       │
│  └────────────────┘         └──────────────────┘       │
│                                     │                   │
└─────────────────────────────────────┼───────────────────┘
                                      │
                                      ▼
                          ┌──────────────────────┐
                          │   localStorage       │
                          │   - reader-data-{id} │
                          │   - app-settings     │
                          └──────────────────────┘
```

## Components and Interfaces

### 1. Reader Data Types

```typescript
// web-next/src/types/reader.ts

export interface ReadingProgress {
  bookId: string
  location: string  // EPUB CFI
  percentage: number
  lastRead: string  // ISO timestamp
}

export interface Bookmark {
  id: string
  bookId: string
  location: string  // EPUB CFI
  cfiRange: string
  text: string      // Preview text
  chapter: string
  createdAt: string
}

export interface Highlight {
  id: string
  bookId: string
  location: string  // EPUB CFI
  cfiRange: string
  text: string
  color: 'yellow' | 'green' | 'blue' | 'pink'
  createdAt: string
}

export interface Note {
  id: string
  bookId: string
  location: string  // EPUB CFI
  cfiRange: string
  text: string      // Selected text
  content: string   // Note content
  createdAt: string
  updatedAt: string
}

export interface ReaderData {
  bookId: string
  progress: ReadingProgress
  bookmarks: Bookmark[]
  highlights: Highlight[]
  notes: Note[]
}
```

### 2. Reader Context

```typescript
// web-next/src/contexts/reader-context.tsx

interface ReaderContextType {
  // Data
  readerData: ReaderData | null
  
  // Progress
  updateProgress: (location: string, percentage: number) => void
  
  // Bookmarks
  addBookmark: (location: string, text: string, chapter: string) => void
  removeBookmark: (id: string) => void
  getBookmarks: () => Bookmark[]
  
  // Highlights
  addHighlight: (cfiRange: string, text: string, color: HighlightColor) => void
  removeHighlight: (id: string) => void
  getHighlights: () => Highlight[]
  
  // Notes
  addNote: (cfiRange: string, text: string, content: string) => void
  updateNote: (id: string, content: string) => void
  removeNote: (id: string) => void
  getNotes: () => Note[]
  
  // Export/Import
  exportData: () => void
  importData: (file: File) => Promise<void>
  
  // Loading state
  isLoading: boolean
}
```

### 3. Reader Sidebar Component

```typescript
// web-next/src/components/reader-sidebar.tsx

interface ReaderSidebarProps {
  bookId: string
  isOpen: boolean
  onClose: () => void
  onNavigate: (location: string) => void
}

// Tabs: Bookmarks, Highlights, Notes
// Each tab shows a list of items with preview and actions
```

### 4. Reader Toolbar Component

```typescript
// web-next/src/components/reader-toolbar.tsx

interface ReaderToolbarProps {
  onAddBookmark: () => void
  onToggleSidebar: () => void
  onToggleSettings: () => void
  progress: number
  isVisible: boolean
}
```

## Data Models

### localStorage Schema

**Key**: `reader-data-{bookId}`

**Value Structure**:
```json
{
  "bookId": "123",
  "progress": {
    "location": "epubcfi(/6/4[chap01ref]!/4/2/2[para01]/1:0)",
    "percentage": 45.5,
    "lastRead": "2025-12-11T10:30:00.000Z"
  },
  "bookmarks": [
    {
      "id": "uuid-1",
      "bookId": "123",
      "location": "epubcfi(...)",
      "cfiRange": "epubcfi(...)",
      "text": "Preview text...",
      "chapter": "Chapter 1",
      "createdAt": "2025-12-11T10:00:00.000Z"
    }
  ],
  "highlights": [
    {
      "id": "uuid-2",
      "bookId": "123",
      "location": "epubcfi(...)",
      "cfiRange": "epubcfi(...)",
      "text": "Highlighted text",
      "color": "yellow",
      "createdAt": "2025-12-11T10:15:00.000Z"
    }
  ],
  "notes": [
    {
      "id": "uuid-3",
      "bookId": "123",
      "location": "epubcfi(...)",
      "cfiRange": "epubcfi(...)",
      "text": "Selected text",
      "content": "My note content",
      "createdAt": "2025-12-11T10:20:00.000Z",
      "updatedAt": "2025-12-11T10:25:00.000Z"
    }
  ]
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do.*

### Property 1: Progress Persistence

*For any* reading session, saving progress and then loading should restore the exact same location.

**Validates: Requirements 1.1, 1.2**

### Property 2: Bookmark Uniqueness

*For any* set of bookmarks, no two bookmarks should have the same ID within a book.

**Validates: Requirements 2.1, 2.2**

### Property 3: Highlight Rendering

*For any* highlight, applying it to the rendition should visually mark the specified text range.

**Validates: Requirements 3.1, 3.2**

### Property 4: Note Association

*For any* note, it should always be associated with a valid location in the book.

**Validates: Requirements 4.1, 4.2**

### Property 5: Settings Application

*For any* reader settings change, the rendition should reflect the new settings immediately.

**Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5**

### Property 6: Export-Import Round Trip

*For any* reader data, exporting and then importing should restore all bookmarks, highlights, and notes.

**Validates: Requirements 7.1, 7.2, 7.3, 7.4**

## Error Handling

### localStorage Errors

1. **QuotaExceededError**:
   - Fallback: Limit number of items (e.g., max 100 bookmarks)
   - Notify: Warn user about storage limitations

2. **SecurityError** (private browsing):
   - Fallback: Use in-memory storage (session only)
   - Notify: Inform user data won't persist

3. **Parse Error** (corrupted data):
   - Fallback: Use empty data structure
   - Action: Clear corrupted data
   - Log: Record error for debugging

### EPUB Rendering Errors

1. **Invalid CFI**:
   - Catch: Handle invalid location strings
   - Fallback: Navigate to book start
   - Notify: Show error message

2. **Selection Errors**:
   - Validate: Check if text is selected before adding highlight/note
   - Notify: Show helpful message

## Testing Strategy

### Unit Tests

1. **Reader Context Tests**:
   - Test data initialization
   - Test CRUD operations for bookmarks/highlights/notes
   - Test localStorage integration
   - Test export/import functionality

2. **Component Tests**:
   - Test sidebar rendering
   - Test toolbar interactions
   - Test keyboard shortcuts

### Integration Tests

1. **End-to-End Reader Flow**:
   - Open book → Add bookmark → Close → Reopen → Verify bookmark exists
   - Add highlight → Export → Import in new session → Verify highlight restored

## Implementation Notes

### Settings Integration

- Use `useSettings` hook to access reader settings
- Apply settings to rendition themes
- Listen for settings changes and update rendition

### Performance Considerations

1. **Debounce Progress Saves**: Save progress every 2 seconds, not on every page turn
2. **Lazy Load Sidebar**: Only render sidebar when opened
3. **Memoize Lists**: Use `useMemo` for filtered bookmark/highlight/note lists

### Accessibility

1. **Keyboard Navigation**: Full keyboard support for all features
2. **Screen Reader Support**: Proper ARIA labels
3. **Focus Management**: Maintain focus after operations

### Mobile Considerations

1. **Touch Gestures**: Support swipe for page navigation
2. **Responsive Sidebar**: Full-screen on mobile
3. **Touch-friendly Buttons**: Larger hit areas

## Security Considerations

1. **XSS Prevention**: Sanitize note content before rendering
2. **Data Validation**: Validate imported data structure
3. **File Size Limits**: Limit import file size

## Future Enhancements

1. **Cloud Sync**: Sync data across devices via backend API
2. **Collaborative Notes**: Share notes with other readers
3. **Advanced Search**: Search within highlights and notes
4. **Statistics**: Reading time, pages read, etc.
5. **Collections**: Organize highlights by topic
