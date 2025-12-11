"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { useTheme } from "next-themes"
import { Moon, Sun, Monitor, Save } from "lucide-react"
import { useState, useEffect } from "react"
import { toast } from "sonner"

interface SettingsState {
  apiEndpoint: string
  readerFontSize: number
  readerFontFamily: string
  readerLineHeight: number
  searchMode: "keyword" | "semantic" | "hybrid"
  resultsPerPage: number
}

const DEFAULT_SETTINGS: SettingsState = {
  apiEndpoint: "/api",
  readerFontSize: 16,
  readerFontFamily: "sans-serif",
  readerLineHeight: 1.6,
  searchMode: "hybrid",
  resultsPerPage: 12,
}

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)
  const [settings, setSettings] = useState<SettingsState>(DEFAULT_SETTINGS)

  useEffect(() => {
    setMounted(true)
    // Load settings from localStorage
    const saved = localStorage.getItem("app-settings")
    if (saved) {
      try {
        setSettings(JSON.parse(saved))
      } catch (error) {
        console.error("Failed to load settings:", error)
      }
    }
  }, [])

  const saveSettings = () => {
    localStorage.setItem("app-settings", JSON.stringify(settings))
    toast.success("Settings saved successfully")
  }

  const resetSettings = () => {
    setSettings(DEFAULT_SETTINGS)
    localStorage.removeItem("app-settings")
    toast.success("Settings reset to defaults")
  }

  if (!mounted) {
    return null // Prevent hydration mismatch
  }

  return (
    <div className="container max-w-4xl">
      <h1 className="text-3xl font-bold mb-8">Settings</h1>
      
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
              <Label htmlFor="theme-select">Theme</Label>
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
                onChange={(e) => setSettings({ ...settings, apiEndpoint: e.target.value })}
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
            <CardTitle>Reader Settings</CardTitle>
            <CardDescription>
              Customize your reading experience
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="font-size">Font Size: {settings.readerFontSize}px</Label>
              <input
                id="font-size"
                type="range"
                min="12"
                max="24"
                step="1"
                value={settings.readerFontSize}
                onChange={(e) => setSettings({ ...settings, readerFontSize: parseInt(e.target.value) })}
                className="w-full"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="font-family">Font Family</Label>
              <Select
                value={settings.readerFontFamily}
                onValueChange={(value) => setSettings({ ...settings, readerFontFamily: value })}
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
              <Label htmlFor="line-height">Line Height: {settings.readerLineHeight}</Label>
              <input
                id="line-height"
                type="range"
                min="1.2"
                max="2.0"
                step="0.1"
                value={settings.readerLineHeight}
                onChange={(e) => setSettings({ ...settings, readerLineHeight: parseFloat(e.target.value) })}
                className="w-full"
              />
            </div>
          </CardContent>
        </Card>

        {/* Search Settings */}
        <Card className="glass">
          <CardHeader>
            <CardTitle>Search Settings</CardTitle>
            <CardDescription>
              Configure default search behavior
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="search-mode">Default Search Mode</Label>
              <Select
                value={settings.searchMode}
                onValueChange={(value) => setSettings({ ...settings, searchMode: value as any })}
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
            </div>

            <div className="space-y-2">
              <Label htmlFor="results-per-page">Results Per Page</Label>
              <Select
                value={settings.resultsPerPage.toString()}
                onValueChange={(value) => setSettings({ ...settings, resultsPerPage: parseInt(value) })}
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
          </CardContent>
        </Card>

        {/* Action Buttons */}
        <div className="flex justify-end gap-4">
          <Button variant="outline" onClick={resetSettings}>
            Reset to Defaults
          </Button>
          <Button onClick={saveSettings}>
            <Save className="w-4 h-4 mr-2" />
            Save Settings
          </Button>
        </div>
      </div>
    </div>
  )
}

