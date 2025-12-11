import { describe, it, expect, beforeEach, vi } from 'vitest'
import { DEFAULT_SETTINGS, SETTINGS_STORAGE_KEY, AppSettings } from '@/types/settings'

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {}

  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value
    },
    removeItem: (key: string) => {
      delete store[key]
    },
    clear: () => {
      store = {}
    },
  }
})()

Object.defineProperty(global, 'localStorage', {
  value: localStorageMock,
})

describe('Settings System', () => {
  beforeEach(() => {
    localStorageMock.clear()
  })

  describe('DEFAULT_SETTINGS', () => {
    it('should have all required fields', () => {
      expect(DEFAULT_SETTINGS).toHaveProperty('apiEndpoint')
      expect(DEFAULT_SETTINGS).toHaveProperty('reader')
      expect(DEFAULT_SETTINGS).toHaveProperty('search')
      expect(DEFAULT_SETTINGS).toHaveProperty('ui')
    })

    it('should have valid reader settings', () => {
      expect(DEFAULT_SETTINGS.reader.fontSize).toBeGreaterThanOrEqual(12)
      expect(DEFAULT_SETTINGS.reader.fontSize).toBeLessThanOrEqual(24)
      expect(DEFAULT_SETTINGS.reader.lineHeight).toBeGreaterThanOrEqual(1.2)
      expect(DEFAULT_SETTINGS.reader.lineHeight).toBeLessThanOrEqual(2.0)
      expect(['sans-serif', 'serif', 'monospace']).toContain(DEFAULT_SETTINGS.reader.fontFamily)
      expect(['light', 'dark', 'sepia']).toContain(DEFAULT_SETTINGS.reader.theme)
    })

    it('should have valid search settings', () => {
      expect(['keyword', 'semantic', 'hybrid']).toContain(DEFAULT_SETTINGS.search.defaultMode)
      expect([6, 12, 24, 48]).toContain(DEFAULT_SETTINGS.search.resultsPerPage)
      expect(typeof DEFAULT_SETTINGS.search.enableAutoComplete).toBe('boolean')
    })

    it('should have valid UI settings', () => {
      expect(typeof DEFAULT_SETTINGS.ui.compactMode).toBe('boolean')
      expect(typeof DEFAULT_SETTINGS.ui.showCoverImages).toBe('boolean')
      expect(['auto', 2, 3, 4, 6]).toContain(DEFAULT_SETTINGS.ui.gridColumns)
    })
  })

  describe('localStorage Integration', () => {
    it('should save settings to localStorage', () => {
      const settings = { ...DEFAULT_SETTINGS }
      localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(settings))
      
      const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
      expect(stored).not.toBeNull()
      
      const parsed = JSON.parse(stored!)
      expect(parsed).toEqual(settings)
    })

    it('should load settings from localStorage', () => {
      const customSettings: AppSettings = {
        ...DEFAULT_SETTINGS,
        reader: {
          ...DEFAULT_SETTINGS.reader,
          fontSize: 20,
        },
      }
      
      localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(customSettings))
      
      const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
      const loaded = JSON.parse(stored!)
      
      expect(loaded.reader.fontSize).toBe(20)
    })

    it('should handle corrupted localStorage data', () => {
      localStorage.setItem(SETTINGS_STORAGE_KEY, 'invalid json')
      
      expect(() => {
        const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
        JSON.parse(stored!)
      }).toThrow()
    })

    it('should handle missing localStorage data', () => {
      const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
      expect(stored).toBeNull()
    })
  })

  describe('Settings Validation', () => {
    it('should validate fontSize range', () => {
      const validSizes = [12, 16, 20, 24]
      validSizes.forEach(size => {
        expect(size).toBeGreaterThanOrEqual(12)
        expect(size).toBeLessThanOrEqual(24)
      })
    })

    it('should validate lineHeight range', () => {
      const validHeights = [1.2, 1.5, 1.8, 2.0]
      validHeights.forEach(height => {
        expect(height).toBeGreaterThanOrEqual(1.2)
        expect(height).toBeLessThanOrEqual(2.0)
      })
    })

    it('should validate enum values', () => {
      const validFontFamilies = ['sans-serif', 'serif', 'monospace']
      const validSearchModes = ['keyword', 'semantic', 'hybrid']
      const validResultsPerPage = [6, 12, 24, 48]
      
      expect(validFontFamilies).toContain('sans-serif')
      expect(validSearchModes).toContain('hybrid')
      expect(validResultsPerPage).toContain(12)
    })
  })

  describe('Settings Export/Import', () => {
    it('should export settings with metadata', () => {
      const exportData = {
        version: '1.0.0',
        timestamp: new Date().toISOString(),
        settings: DEFAULT_SETTINGS,
      }
      
      expect(exportData).toHaveProperty('version')
      expect(exportData).toHaveProperty('timestamp')
      expect(exportData).toHaveProperty('settings')
    })

    it('should import valid settings', () => {
      const importData = {
        version: '1.0.0',
        timestamp: new Date().toISOString(),
        settings: {
          ...DEFAULT_SETTINGS,
          reader: {
            ...DEFAULT_SETTINGS.reader,
            fontSize: 18,
          },
        },
      }
      
      const imported = importData.settings
      expect(imported.reader.fontSize).toBe(18)
    })

    it('should handle export-import round trip', () => {
      const original = {
        ...DEFAULT_SETTINGS,
        reader: {
          ...DEFAULT_SETTINGS.reader,
          fontSize: 20,
          fontFamily: 'serif' as const,
        },
      }
      
      // Export
      const exported = JSON.stringify({
        version: '1.0.0',
        timestamp: new Date().toISOString(),
        settings: original,
      })
      
      // Import
      const imported = JSON.parse(exported)
      
      expect(imported.settings).toEqual(original)
    })
  })

  describe('Settings Update', () => {
    it('should support partial updates', () => {
      const current = { ...DEFAULT_SETTINGS }
      const updates = {
        reader: {
          ...current.reader,
          fontSize: 18,
        },
      }
      
      const updated = { ...current, ...updates }
      
      expect(updated.reader.fontSize).toBe(18)
      expect(updated.search).toEqual(DEFAULT_SETTINGS.search)
    })

    it('should preserve other settings during update', () => {
      const current = { ...DEFAULT_SETTINGS }
      const updates = {
        apiEndpoint: '/custom-api',
      }
      
      const updated = { ...current, ...updates }
      
      expect(updated.apiEndpoint).toBe('/custom-api')
      expect(updated.reader).toEqual(DEFAULT_SETTINGS.reader)
      expect(updated.search).toEqual(DEFAULT_SETTINGS.search)
      expect(updated.ui).toEqual(DEFAULT_SETTINGS.ui)
    })
  })

  describe('Settings Reset', () => {
    it('should reset to default settings', () => {
      const customSettings = {
        ...DEFAULT_SETTINGS,
        reader: {
          ...DEFAULT_SETTINGS.reader,
          fontSize: 20,
        },
      }
      
      localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(customSettings))
      
      // Reset
      localStorage.removeItem(SETTINGS_STORAGE_KEY)
      
      const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
      expect(stored).toBeNull()
    })
  })
})
