import { useContext } from 'react'
import { SettingsContext, SettingsContextType } from '@/contexts/settings-context'

/**
 * Hook to access and modify application settings
 * 
 * @throws {Error} If used outside of SettingsProvider
 * @returns Settings context with current settings and update functions
 * 
 * @example
 * ```tsx
 * function MyComponent() {
 *   const { settings, updateSettings } = useSettings()
 *   
 *   return (
 *     <div style={{ fontSize: settings.reader.fontSize }}>
 *       Content
 *     </div>
 *   )
 * }
 * ```
 */
export function useSettings(): SettingsContextType {
  const context = useContext(SettingsContext)
  
  if (context === undefined) {
    throw new Error('useSettings must be used within a SettingsProvider')
  }
  
  return context
}
