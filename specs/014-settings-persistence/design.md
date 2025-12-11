# Design Document

## Overview

本设计文档描述了用户设置持久化系统的技术实现。系统使用 React Context + Hook 模式管理设置状态，使用 localStorage 进行持久化，并提供导入/导出功能以支持跨设备配置迁移。

## Architecture

### 系统架构

```
┌─────────────────────────────────────────────────────────┐
│                    React Application                     │
│                                                          │
│  ┌────────────────┐         ┌──────────────────┐       │
│  │  Settings Page │◄────────┤ useSettings Hook │       │
│  └────────────────┘         └──────────┬───────┘       │
│                                        │                │
│  ┌────────────────┐                   │                │
│  │  Reader Page   │◄──────────────────┤                │
│  └────────────────┘                   │                │
│                                        │                │
│  ┌────────────────┐                   │                │
│  │  Search Page   │◄──────────────────┤                │
│  └────────────────┘                   │                │
│                                        ▼                │
│                          ┌──────────────────────┐      │
│                          │ Settings Context     │      │
│                          │ - State Management   │      │
│                          │ - Persistence Logic  │      │
│                          └──────────┬───────────┘      │
│                                     │                   │
└─────────────────────────────────────┼───────────────────┘
                                      │
                                      ▼
                          ┌──────────────────────┐
                          │   localStorage       │
                          │   - app-settings     │
                          └──────────────────────┘
```

### 数据流

1. **初始化**: 应用启动 → 从 localStorage 读取设置 → 初始化 Context
2. **更新**: 用户修改设置 → 更新 Context 状态 → 保存到 localStorage → 触发重新渲染
3. **导出**: 用户点击导出 → 读取当前设置 → 生成 JSON 文件 → 触发下载
4. **导入**: 用户选择文件 → 验证格式 → 更新 Context → 保存到 localStorage

## Components and Interfaces

### 1. Settings Types

```typescript
// web-next/src/types/settings.ts

export interface AppSettings {
  // API Configuration
  apiEndpoint: string

  // Reader Settings
  reader: {
    fontSize: number        // 12-24
    fontFamily: 'sans-serif' | 'serif' | 'monospace'
    lineHeight: number      // 1.2-2.0
    theme: 'light' | 'dark' | 'sepia'
  }

  // Search Settings
  search: {
    defaultMode: 'keyword' | 'semantic' | 'hybrid'
    resultsPerPage: 6 | 12 | 24 | 48
    enableAutoComplete: boolean
  }

  // UI Settings
  ui: {
    compactMode: boolean
    showCoverImages: boolean
    gridColumns: 'auto' | 2 | 3 | 4 | 6
  }
}

export interface SettingsExport {
  version: string
  timestamp: string
  settings: AppSettings
}

export const DEFAULT_SETTINGS: AppSettings = {
  apiEndpoint: '/api',
  reader: {
    fontSize: 16,
    fontFamily: 'sans-serif',
    lineHeight: 1.6,
    theme: 'light',
  },
  search: {
    defaultMode: 'hybrid',
    resultsPerPage: 12,
    enableAutoComplete: true,
  },
  ui: {
    compactMode: false,
    showCoverImages: true,
    gridColumns: 'auto',
  },
}
```

### 2. Settings Context

```typescript
// web-next/src/contexts/settings-context.tsx

interface SettingsContextType {
  settings: AppSettings
  updateSettings: (updates: Partial<AppSettings>) => void
  resetSettings: () => void
  exportSettings: () => void
  importSettings: (file: File) => Promise<void>
  isLoading: boolean
}

export const SettingsContext = createContext<SettingsContextType | undefined>(undefined)
```

### 3. Settings Hook

```typescript
// web-next/src/hooks/use-settings.ts

export function useSettings() {
  const context = useContext(SettingsContext)
  if (!context) {
    throw new Error('useSettings must be used within SettingsProvider')
  }
  return context
}
```

### 4. Settings Provider

```typescript
// web-next/src/providers/settings-provider.tsx

export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<AppSettings>(DEFAULT_SETTINGS)
  const [isLoading, setIsLoading] = useState(true)

  // Load from localStorage on mount
  useEffect(() => {
    loadSettings()
  }, [])

  // Save to localStorage on change
  useEffect(() => {
    if (!isLoading) {
      saveSettings(settings)
    }
  }, [settings, isLoading])

  // ... implementation
}
```

## Data Models

### localStorage Schema

**Key**: `app-settings`

**Value Structure**:
```json
{
  "version": "1.0.0",
  "apiEndpoint": "/api",
  "reader": {
    "fontSize": 16,
    "fontFamily": "sans-serif",
    "lineHeight": 1.6,
    "theme": "light"
  },
  "search": {
    "defaultMode": "hybrid",
    "resultsPerPage": 12,
    "enableAutoComplete": true
  },
  "ui": {
    "compactMode": false,
    "showCoverImages": true,
    "gridColumns": "auto"
  }
}
```

### Export File Format

**Filename**: `calibre-settings-{timestamp}.json`

**Content**:
```json
{
  "version": "1.0.0",
  "timestamp": "2025-12-11T10:30:00.000Z",
  "settings": {
    // ... AppSettings
  }
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Settings Persistence Round Trip

*For any* valid settings object, saving to localStorage and then loading should produce an equivalent settings object.

**Validates: Requirements 1.1, 1.2**

### Property 2: Default Settings Fallback

*For any* corrupted or missing localStorage data, the system should always fall back to default settings without crashing.

**Validates: Requirements 1.3, 1.4**

### Property 3: Export-Import Round Trip

*For any* valid settings state, exporting and then importing should restore the exact same settings.

**Validates: Requirements 3.1, 3.2, 4.1, 4.2**

### Property 4: Settings Update Atomicity

*For any* settings update operation, either all changes are applied and persisted, or none are (no partial updates).

**Validates: Requirements 1.1, 8.2**

### Property 5: Theme Synchronization

*For any* theme change, the UI should reflect the new theme immediately without requiring a page reload.

**Validates: Requirements 5.1, 5.3**

### Property 6: Settings Validation

*For any* settings value, the system should only accept values within the defined valid ranges and types.

**Validates: Requirements 6.1, 6.2, 6.3, 9.1**

### Property 7: Context State Consistency

*For any* component using the settings hook, all components should see the same settings state at any given time.

**Validates: Requirements 8.1, 8.3**

## Error Handling

### localStorage Errors

1. **QuotaExceededError**: 
   - Fallback: Use in-memory storage
   - Notify: Warn user about storage limitations
   
2. **SecurityError** (private browsing):
   - Fallback: Use session storage or in-memory
   - Notify: Inform user settings won't persist

3. **Parse Error** (corrupted data):
   - Fallback: Use default settings
   - Action: Clear corrupted data
   - Log: Record error for debugging

### Import/Export Errors

1. **Invalid File Format**:
   - Validate: Check JSON structure
   - Reject: Show error message
   - Preserve: Keep current settings

2. **Version Mismatch**:
   - Attempt: Migrate old format
   - Fallback: Reject if migration fails
   - Inform: Show version compatibility message

3. **File Read Error**:
   - Catch: Handle FileReader errors
   - Notify: Show user-friendly error
   - Preserve: Keep current settings

## Testing Strategy

### Unit Tests

1. **Settings Hook Tests**:
   - Test settings initialization
   - Test settings updates
   - Test reset functionality
   - Test localStorage integration

2. **Validation Tests**:
   - Test range validation (fontSize, lineHeight)
   - Test enum validation (fontFamily, searchMode)
   - Test URL validation (apiEndpoint)

3. **Serialization Tests**:
   - Test JSON serialization/deserialization
   - Test export file generation
   - Test import file parsing

### Integration Tests

1. **End-to-End Settings Flow**:
   - Load page → Modify settings → Reload → Verify persistence
   - Export settings → Import in new session → Verify restoration

2. **Cross-Component Tests**:
   - Update settings in Settings Page → Verify Reader Page reflects changes
   - Update settings in Settings Page → Verify Search Page uses new defaults

### Property-Based Tests

Using `@fast-check/vitest`:

1. **Property 1: Persistence Round Trip**
   ```typescript
   fc.assert(
     fc.property(settingsArbitrary, (settings) => {
       saveSettings(settings)
       const loaded = loadSettings()
       expect(loaded).toEqual(settings)
     })
   )
   ```

2. **Property 3: Export-Import Round Trip**
   ```typescript
   fc.assert(
     fc.property(settingsArbitrary, (settings) => {
       const exported = exportSettings(settings)
       const imported = importSettings(exported)
       expect(imported).toEqual(settings)
     })
   )
   ```

3. **Property 6: Settings Validation**
   ```typescript
   fc.assert(
     fc.property(invalidSettingsArbitrary, (settings) => {
       expect(() => validateSettings(settings)).toThrow()
     })
   )
   ```

## Implementation Notes

### Theme Integration

- Use `next-themes` for theme management (already integrated)
- Theme preference stored separately by `next-themes`
- Settings system only manages reader theme (light/dark/sepia for reading view)

### Performance Considerations

1. **Debounce localStorage writes**: Avoid excessive writes on rapid changes
2. **Lazy loading**: Load settings only when needed
3. **Memoization**: Use `useMemo` for derived settings values

### Accessibility

1. **Keyboard Navigation**: All settings controls accessible via keyboard
2. **Screen Reader Support**: Proper ARIA labels and descriptions
3. **Focus Management**: Maintain focus after settings updates

### Migration Strategy

If settings schema changes in future versions:

```typescript
function migrateSettings(data: any, fromVersion: string): AppSettings {
  if (fromVersion === '1.0.0' && CURRENT_VERSION === '2.0.0') {
    // Perform migration
    return {
      ...data,
      // Add new fields with defaults
      // Transform old fields if needed
    }
  }
  return data
}
```

## Security Considerations

1. **XSS Prevention**: Sanitize imported settings before applying
2. **URL Validation**: Validate API endpoint URLs to prevent injection
3. **File Size Limits**: Limit import file size to prevent DoS
4. **Content Type Validation**: Verify imported file is valid JSON

## Future Enhancements

1. **Cloud Sync**: Sync settings across devices via backend API
2. **Settings Profiles**: Multiple named settings profiles
3. **Keyboard Shortcuts**: Customizable keyboard shortcuts
4. **Advanced Reader Settings**: More typography options
5. **Settings Search**: Search within settings page
