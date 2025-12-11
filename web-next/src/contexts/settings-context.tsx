'use client'

import React, { createContext, useState, useEffect, useCallback } from 'react'
import { 
  AppSettings, 
  DEFAULT_SETTINGS, 
  SETTINGS_STORAGE_KEY,
  SETTINGS_VERSION,
  SettingsExport 
} from '@/types/settings'
import { toast } from 'sonner'

/**
 * Settings Context Type
 */
export interface SettingsContextType {
  /** Current settings state */
  settings: AppSettings
  /** Update settings (partial update supported) */
  updateSettings: (updates: Partial<AppSettings> | ((prev: AppSettings) => AppSettings)) => void
  /** Reset all settings to defaults */
  resetSettings: () => void
  /** Export settings to JSON file */
  exportSettings: () => void
  /** Import settings from JSON file */
  importSettings: (file: File) => Promise<void>
  /** Loading state during initialization */
  isLoading: boolean
}

/**
 * Settings Context
 */
export const SettingsContext = createContext<SettingsContextType | undefined>(undefined)

/**
 * Load settings from localStorage
 */
function loadSettings(): AppSettings {
  if (typeof window === 'undefined') {
    return DEFAULT_SETTINGS
  }

  try {
    const stored = localStorage.getItem(SETTINGS_STORAGE_KEY)
    if (!stored) {
      return DEFAULT_SETTINGS
    }

    const parsed = JSON.parse(stored)
    
    // Merge with defaults to handle missing fields
    return {
      ...DEFAULT_SETTINGS,
      ...parsed,
      reader: { ...DEFAULT_SETTINGS.reader, ...parsed.reader },
      search: { ...DEFAULT_SETTINGS.search, ...parsed.search },
      ui: { ...DEFAULT_SETTINGS.ui, ...parsed.ui },
    }
  } catch (error) {
    console.error('Failed to load settings from localStorage:', error)
    toast.error('Failed to load settings, using defaults')
    return DEFAULT_SETTINGS
  }
}

/**
 * Save settings to localStorage
 */
function saveSettings(settings: AppSettings): void {
  if (typeof window === 'undefined') {
    return
  }

  try {
    const data = {
      version: SETTINGS_VERSION,
      ...settings,
    }
    localStorage.setItem(SETTINGS_STORAGE_KEY, JSON.stringify(data))
  } catch (error) {
    console.error('Failed to save settings to localStorage:', error)
    
    if (error instanceof Error && error.name === 'QuotaExceededError') {
      toast.error('Storage quota exceeded. Settings may not persist.')
    } else {
      toast.error('Failed to save settings')
    }
  }
}

/**
 * Settings Provider Component
 */
export function SettingsProvider({ children }: { children: React.ReactNode }) {
  const [settings, setSettings] = useState<AppSettings>(DEFAULT_SETTINGS)
  const [isLoading, setIsLoading] = useState(true)

  // Load settings on mount
  useEffect(() => {
    const loaded = loadSettings()
    setSettings(loaded)
    setIsLoading(false)
  }, [])

  // Save settings whenever they change (after initial load)
  useEffect(() => {
    if (!isLoading) {
      saveSettings(settings)
    }
  }, [settings, isLoading])

  /**
   * Update settings (supports partial updates and updater function)
   */
  const updateSettings = useCallback((updates: Partial<AppSettings> | ((prev: AppSettings) => AppSettings)) => {
    setSettings(prev => {
      const newSettings = typeof updates === 'function' ? updates(prev) : { ...prev, ...updates }
      return newSettings
    })
  }, [])

  /**
   * Reset settings to defaults
   */
  const resetSettings = useCallback(() => {
    setSettings(DEFAULT_SETTINGS)
    
    try {
      localStorage.removeItem(SETTINGS_STORAGE_KEY)
      toast.success('Settings reset to defaults')
    } catch (error) {
      console.error('Failed to clear settings from localStorage:', error)
      toast.error('Failed to reset settings')
    }
  }, [])

  /**
   * Export settings to JSON file
   */
  const exportSettings = useCallback(() => {
    try {
      const exportData: SettingsExport = {
        version: SETTINGS_VERSION,
        timestamp: new Date().toISOString(),
        settings,
      }

      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json',
      })

      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5)
      link.download = `calibre-settings-${timestamp}.json`
      
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      
      URL.revokeObjectURL(url)
      
      toast.success('Settings exported successfully')
    } catch (error) {
      console.error('Failed to export settings:', error)
      toast.error('Failed to export settings')
    }
  }, [settings])

  /**
   * Import settings from JSON file
   */
  const importSettings = useCallback(async (file: File): Promise<void> => {
    try {
      // Validate file type
      if (!file.name.endsWith('.json')) {
        toast.error('Invalid file type. Please select a JSON file.')
        return
      }

      // Validate file size (max 1MB)
      if (file.size > 1024 * 1024) {
        toast.error('File too large. Maximum size is 1MB.')
        return
      }

      // Read file
      const text = await file.text()
      const data: SettingsExport = JSON.parse(text)

      // Validate structure
      if (!data.settings || typeof data.settings !== 'object') {
        toast.error('Invalid settings file format')
        return
      }

      // Version compatibility check (optional migration logic)
      if (data.version !== SETTINGS_VERSION) {
        console.warn(`Settings version mismatch: ${data.version} vs ${SETTINGS_VERSION}`)
        // Could implement migration logic here
      }

      // Merge imported settings with defaults
      const importedSettings: AppSettings = {
        ...DEFAULT_SETTINGS,
        ...data.settings,
        reader: { ...DEFAULT_SETTINGS.reader, ...data.settings.reader },
        search: { ...DEFAULT_SETTINGS.search, ...data.settings.search },
        ui: { ...DEFAULT_SETTINGS.ui, ...data.settings.ui },
      }

      setSettings(importedSettings)
      toast.success('Settings imported successfully')
    } catch (error) {
      console.error('Failed to import settings:', error)
      
      if (error instanceof SyntaxError) {
        toast.error('Invalid JSON file')
      } else {
        toast.error('Failed to import settings')
      }
    }
  }, [])

  const value: SettingsContextType = {
    settings,
    updateSettings,
    resetSettings,
    exportSettings,
    importSettings,
    isLoading,
  }

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  )
}
