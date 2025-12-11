"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { useTheme } from "next-themes"
import { Moon, Sun, Monitor, Save, RotateCcw, Download, Upload, BookOpen, Search as SearchIcon, Layout } from "lucide-react"
import { useState, useEffect, useRef } from "react"
import { useSettings } from "@/hooks/use-settings"

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const { settings, updateSettings, resetSettings, exportSettings, importSettings, isLoading } = useSettings()
  const [mounted, setMounted] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    setMounted(true)
  }, [])

  const handleImport = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      await importSettings(file)
      // Reset file input
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  if (!mounted || isLoading) {
    return (
      <div className="container max-w-4xl">
        <h1 className="text-3xl font-bold mb-8">Settings</h1>
        <div className="text-center text-muted-foreground">Loading settings...</div>
      </div>
    )
  }

  return (
    <div className="container max-w-4xl">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold">Settings</h1>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={exportSettings}>
            <Download className="w-4 h-4 mr-2" />
            Export
          </Button>
          <Button variant="outline" size="sm" onClick={() => fileInputRef.current?.click()}>
            <Upload className="w-4 h-4 mr-2" />
            Import
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            accept=".json"
            onChange={handleImport}
            className="hidden"
          />
        </div>
      </div>
      
      <div className="grid gap-6">
        {/* Theme Settings */}
        <Card className="glass">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Sun className="w-5 h-5" />
              Appearance
            </CardTitle>
            <CardDescription>
              Customize the look and feel of the application
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <Label htmlFor="theme-select">App Theme</Label>
              <Select value={theme} onValueChange={setTheme}>
                <SelectTrigger id="theme-select" className="w-[180px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">
                    <div className="flex items-center gap-2">
                      <Sun className="w-4 h-4" />
                      Light
                    </div>
                  </SelectItem>
                  <SelectItem value="dark">
                    <div className="flex items-center gap-2">
                      <Moon className="w-4 h-4" />
                      Dark
                    </div>
                  </SelectItem>
                  <SelectItem value="system">
                    <div className="flex items-center gap-2">
                      <Monitor className="w-4 h-4" />
                      System
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* API Configuration */}
        <Card className="glass">
          <CardHeader>
            <CardTitle>API Configuration</CardTitle>
            <CardDescription>
              Configure the backend API endpoint
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="api-endpoint">API Endpoint</Label>
              <Input
                id="api-endpoint"
                type="text"
                value={settings.apiEndpoint}
                onChange={(e) => updateSettings({ apiEndpoint: e.target.value })}
                placeholder="/api"
              />
              <p className="text-sm text-muted-foreground">
                Leave as <code className="px-1 py-0.5 bg-muted rounded">/api</code> for default configuration
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Reader Settings */}
        <Card className="glass">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BookOpen className="w-5 h-5" />
              Reader Settings
            </CardTitle>
            <CardDescription>
              Customize your reading experience
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="font-size">Font Size: {settings.reader.fontSize}px</Label>
              <input
                id="font-size"
                type="range"
                min="12"
                max="24"
                step="1"
                value={settings.reader.fontSize}
                onChange={(e) => updateSettings({
                  reader: { ...settings.reader, fontSize: parseInt(e.target.value) }
                })}
                className="w-full"
              />
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>12px</span>
                <span>24px</span>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="font-family">Font Family</Label>
              <Select
                value={settings.reader.fontFamily}
                onValueChange={(value: 'sans-serif' | 'serif' | 'monospace') => 
                  updateSettings({ reader: { ...settings.reader, fontFamily: value } })
                }
              >
                <SelectTrigger id="font-family">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sans-serif">Sans Serif</SelectItem>
                  <SelectItem value="serif">Serif</SelectItem>
                  <SelectItem value="monospace">Monospace</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="line-height">Line Height: {settings.reader.lineHeight.toFixed(1)}</Label>
              <input
                id="line-height"
                type="range"
                min="1.2"
                max="2.0"
                step="0.1"
                value={settings.reader.lineHeight}
                onChange={(e) => updateSettings({
                  reader: { ...settings.reader, lineHeight: parseFloat(e.target.value) }
                })}
                className="w-full"
              />
              <div className="flex justify-between text-xs text-muted-foreground">
                <span>1.2</span>
                <span>2.0</span>
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="reader-theme">Reader Theme</Label>
              <Select
                value={settings.reader.theme}
                onValueChange={(value: 'light' | 'dark' | 'sepia') => 
                  updateSettings({ reader: { ...settings.reader, theme: value } })
                }
              >
                <SelectTrigger id="reader-theme">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">Light</SelectItem>
                  <SelectItem value="dark">Dark</SelectItem>
                  <SelectItem value="sepia">Sepia</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-muted-foreground">
                Theme for the book reader (independent of app theme)
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Search Settings */}
        <Card className="glass">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <SearchIcon className="w-5 h-5" />
              Search Settings
            </CardTitle>
            <CardDescription>
              Configure default search behavior
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="search-mode">Default Search Mode</Label>
              <Select
                value={settings.search.defaultMode}
                onValueChange={(value: 'keyword' | 'semantic' | 'hybrid') => 
                  updateSettings({ search: { ...settings.search, defaultMode: value } })
                }
              >
                <SelectTrigger id="search-mode">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="keyword">Keyword Search</SelectItem>
                  <SelectItem value="semantic">Semantic Search</SelectItem>
                  <SelectItem value="hybrid">Hybrid Search</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-muted-foreground">
                Hybrid search combines keyword and semantic search for best results
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="results-per-page">Results Per Page</Label>
              <Select
                value={settings.search.resultsPerPage.toString()}
                onValueChange={(value) => 
                  updateSettings({ search: { ...settings.search, resultsPerPage: parseInt(value) as 6 | 12 | 24 | 48 } })
                }
              >
                <SelectTrigger id="results-per-page">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="6">6</SelectItem>
                  <SelectItem value="12">12</SelectItem>
                  <SelectItem value="24">24</SelectItem>
                  <SelectItem value="48">48</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="auto-complete">Enable Auto-Complete</Label>
                <p className="text-sm text-muted-foreground">
                  Show search suggestions as you type
                </p>
              </div>
              <Switch
                id="auto-complete"
                checked={settings.search.enableAutoComplete}
                onCheckedChange={(checked) => 
                  updateSettings({ search: { ...settings.search, enableAutoComplete: checked } })
                }
              />
            </div>
          </CardContent>
        </Card>

        {/* UI Settings */}
        <Card className="glass">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Layout className="w-5 h-5" />
              Interface Settings
            </CardTitle>
            <CardDescription>
              Customize the user interface layout
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="compact-mode">Compact Mode</Label>
                <p className="text-sm text-muted-foreground">
                  Use denser layouts to show more content
                </p>
              </div>
              <Switch
                id="compact-mode"
                checked={settings.ui.compactMode}
                onCheckedChange={(checked) => 
                  updateSettings({ ui: { ...settings.ui, compactMode: checked } })
                }
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label htmlFor="show-covers">Show Cover Images</Label>
                <p className="text-sm text-muted-foreground">
                  Display book cover images in lists
                </p>
              </div>
              <Switch
                id="show-covers"
                checked={settings.ui.showCoverImages}
                onCheckedChange={(checked) => 
                  updateSettings({ ui: { ...settings.ui, showCoverImages: checked } })
                }
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="grid-columns">Grid Columns</Label>
              <Select
                value={settings.ui.gridColumns.toString()}
                onValueChange={(value) => 
                  updateSettings({ ui: { ...settings.ui, gridColumns: value === 'auto' ? 'auto' : parseInt(value) as 2 | 3 | 4 | 6 } })
                }
              >
                <SelectTrigger id="grid-columns">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="auto">Auto (Responsive)</SelectItem>
                  <SelectItem value="2">2 Columns</SelectItem>
                  <SelectItem value="3">3 Columns</SelectItem>
                  <SelectItem value="4">4 Columns</SelectItem>
                  <SelectItem value="6">6 Columns</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-sm text-muted-foreground">
                Number of columns in book grid layout
              </p>
            </div>
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex justify-end gap-4">
          <Button variant="outline" onClick={resetSettings}>
            <RotateCcw className="w-4 h-4 mr-2" />
            Reset to Defaults
          </Button>
        </div>
      </div>
    </div>
  )
}
