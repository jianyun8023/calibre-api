/**
 * Settings Types and Constants
 * 
 * Defines the structure and default values for application settings.
 */

/**
 * Reader settings for customizing the reading experience
 */
export interface ReaderSettings {
  /** Font size in pixels (12-24) */
  fontSize: number
  /** Font family for reader content */
  fontFamily: 'sans-serif' | 'serif' | 'monospace'
  /** Line height multiplier (1.2-2.0) */
  lineHeight: number
  /** Reader theme (independent of app theme) */
  theme: 'light' | 'dark' | 'sepia'
}

/**
 * Search settings for default search behavior
 */
export interface SearchSettings {
  /** Default search mode */
  defaultMode: 'keyword' | 'semantic' | 'hybrid'
  /** Number of results per page */
  resultsPerPage: 6 | 12 | 24 | 48
  /** Enable search auto-complete suggestions */
  enableAutoComplete: boolean
}

/**
 * UI settings for interface customization
 */
export interface UISettings {
  /** Enable compact mode for denser layouts */
  compactMode: boolean
  /** Show book cover images in lists */
  showCoverImages: boolean
  /** Number of columns in book grid */
  gridColumns: 'auto' | 2 | 3 | 4 | 6
}

/**
 * Complete application settings
 */
export interface AppSettings {
  /** API endpoint URL */
  apiEndpoint: string
  /** Reader customization settings */
  reader: ReaderSettings
  /** Search behavior settings */
  search: SearchSettings
  /** UI customization settings */
  ui: UISettings
}

/**
 * Settings export format with metadata
 */
export interface SettingsExport {
  /** Settings schema version (semver) */
  version: string
  /** Export timestamp (ISO 8601) */
  timestamp: string
  /** Exported settings data */
  settings: AppSettings
}

/**
 * Current settings schema version
 */
export const SETTINGS_VERSION = '1.0.0'

/**
 * localStorage key for settings
 */
export const SETTINGS_STORAGE_KEY = 'app-settings'

/**
 * Default application settings
 */
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

/**
 * Validation ranges for numeric settings
 */
export const SETTINGS_RANGES = {
  fontSize: { min: 12, max: 24 },
  lineHeight: { min: 1.2, max: 2.0 },
} as const

/**
 * Valid values for enum settings
 */
export const SETTINGS_ENUMS = {
  fontFamily: ['sans-serif', 'serif', 'monospace'] as const,
  readerTheme: ['light', 'dark', 'sepia'] as const,
  searchMode: ['keyword', 'semantic', 'hybrid'] as const,
  resultsPerPage: [6, 12, 24, 48] as const,
  gridColumns: ['auto', 2, 3, 4, 6] as const,
} as const
