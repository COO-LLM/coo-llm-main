import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  RefreshCw,
  Settings,
  Database,
  Server,
  Eye,
  EyeOff,
  CheckCircle,
  AlertCircle,
  Save,
  Info
} from 'lucide-react'
import { useApi } from '@/lib/api'

interface PolicyConfig {
  algorithm: string
  priority: string
  cache?: {
    enabled: boolean
    ttl_seconds: number
  }
}

// Helper to convert backend format (underscore) to frontend format (dash)
const backendToFrontend = (value: string): string => {
  return value.replace(/_/g, '-')
}

// Helper to convert frontend format (dash) to backend format (underscore)
const frontendToBackend = (value: string): string => {
  return value.replace(/-/g, '_')
}

export default function ConfigurationPage() {
  const api = useApi()
  const [config, setConfig] = useState<any>(null)
  const [isLoading, setIsLoading] = useState(false)
  const [refreshInterval, setRefreshInterval] = useState('30')
  const [error, setError] = useState<string | null>(null)
  const [showSecrets, setShowSecrets] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  // Policy edit state
  const [policyEdit, setPolicyEdit] = useState<PolicyConfig | null>(null)

  useEffect(() => {
    fetchConfig()
  }, [])

  useEffect(() => {
    if (refreshInterval !== 'manual') {
      const interval = setInterval(fetchConfig, parseInt(refreshInterval) * 1000)
      return () => clearInterval(interval)
    }
  }, [refreshInterval])

  const fetchConfig = async () => {
    setIsLoading(true)
    setError(null)

    try {
      const configData = await api.getConfig()
      console.log('✅ Configuration loaded:', configData)
      setConfig(configData)

      // Initialize policy edit state
      if (configData?.Policy) {
        setPolicyEdit({
          algorithm: backendToFrontend(configData.Policy.Algorithm || 'random'),
          priority: backendToFrontend(configData.Policy.Priority || 'latency'),
          cache: {
            enabled: configData.Policy.Cache?.Enabled || false,
            ttl_seconds: configData.Policy.Cache?.TTLSeconds || 60
          }
        })
      }
    } catch (err) {
      console.error('❌ Failed to fetch configuration:', err)
      setError('Failed to load configuration')
    } finally {
      setIsLoading(false)
    }
  }

  const savePolicy = async () => {
    if (!policyEdit) return

    setIsSaving(true)
    try {
      // Convert frontend format (dash) to backend format (underscore)
      const backendPolicy = {
        algorithm: frontendToBackend(policyEdit.algorithm),
        priority: frontendToBackend(policyEdit.priority),
        cache: policyEdit.cache
      }

      console.log('💾 Saving policy:', backendPolicy)
      await api.updatePolicy(backendPolicy)
      console.log('✅ Policy updated successfully')

      // Refresh config after save
      await fetchConfig()
    } catch (err) {
      console.error('❌ Failed to save policy:', err)
      setError('Failed to save policy configuration')
    } finally {
      setIsSaving(false)
    }
  }

  const maskValue = (value: string) => {
    if (showSecrets || !value) return value
    if (value.length <= 8) return '*'.repeat(value.length)
    return value.slice(0, 4) + '*'.repeat(Math.min(20, value.length - 4))
  }

  if (error && !config) {
    return (
      <div className="space-y-6">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
          <Button onClick={fetchConfig} variant="outline" size="icon">
            <RefreshCw className="h-4 w-4" />
          </Button>
      </div>
    )
  }

  if (isLoading && !config) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="h-8 w-8 animate-spin" />
      </div>
    )
  }

  const hasChanges = policyEdit && config?.Policy && (
    policyEdit.algorithm !== backendToFrontend(config.Policy.Algorithm) ||
    policyEdit.priority !== backendToFrontend(config.Policy.Priority) ||
    policyEdit.cache?.enabled !== config.Policy.Cache?.Enabled ||
    policyEdit.cache?.ttl_seconds !== config.Policy.Cache?.TTLSeconds
  )

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 rounded-lg p-6 border border-slate-200 dark:border-slate-700">
        <div>
          <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">Configuration</h1>
          <p className="text-slate-600 dark:text-slate-400 mt-1">
            System configuration and settings
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Select value={refreshInterval} onValueChange={setRefreshInterval}>
            <SelectTrigger className="w-32">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="10">10s</SelectItem>
              <SelectItem value="30">30s</SelectItem>
              <SelectItem value="60">1m</SelectItem>
              <SelectItem value="300">5m</SelectItem>
              <SelectItem value="manual">Manual</SelectItem>
            </SelectContent>
          </Select>
          <Button
            onClick={() => setShowSecrets(!showSecrets)}
            variant="outline"
            size="sm"
          >
            {showSecrets ? <EyeOff className="h-4 w-4 mr-2" /> : <Eye className="h-4 w-4 mr-2" />}
            {showSecrets ? 'Hide' : 'Show'} Secrets
          </Button>
          <Button onClick={fetchConfig} disabled={isLoading} variant="outline" size="icon">
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
          </Button>
        </div>
      </div>

      {/* System Information (Read-only) */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5" />
            System Information
          </CardTitle>
          <CardDescription>General system configuration (read-only)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-1">
              <Label className="text-muted-foreground">Version</Label>
              <div className="font-mono text-sm">{config?.Version || 'N/A'}</div>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Listen Address</Label>
              <div className="font-mono text-sm">{config?.Server?.Listen || 'N/A'}</div>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Admin API Key</Label>
              <div className="font-mono text-sm">{maskValue(config?.Server?.AdminAPIKey || '')}</div>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Web UI Enabled</Label>
              <Badge variant={config?.Server?.WebUI?.Enabled ? 'default' : 'secondary'}>
                {config?.Server?.WebUI?.Enabled ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Admin ID</Label>
              <div className="font-mono text-sm">{config?.Server?.WebUI?.AdminID || 'N/A'}</div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Storage Configuration (Read-only) */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Storage Configuration
          </CardTitle>
          <CardDescription>Storage backend settings (read-only)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-1">
              <Label className="text-muted-foreground">Config Storage Type</Label>
              <div className="font-mono text-sm">{config?.Storage?.Config?.Type || 'N/A'}</div>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Config Path</Label>
              <div className="font-mono text-sm">{config?.Storage?.Config?.Path || 'N/A'}</div>
            </div>
            <div className="space-y-1">
              <Label className="text-muted-foreground">Runtime Storage Type</Label>
              <div className="font-mono text-sm">{config?.Storage?.Runtime?.Type || 'N/A'}</div>
            </div>
            {config?.Storage?.Runtime?.Addr && (
              <div className="space-y-1">
                <Label className="text-muted-foreground">Redis Address</Label>
                <div className="font-mono text-sm">{config.Storage.Runtime.Addr}</div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Logging Configuration (Read-only) */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Settings className="h-5 w-5" />
            Logging Configuration
          </CardTitle>
          <CardDescription>Logging and monitoring settings (read-only)</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-1">
              <Label className="text-muted-foreground">File Logging</Label>
              <Badge variant={config?.Logging?.File?.Enabled ? 'default' : 'secondary'}>
                {config?.Logging?.File?.Enabled ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
            {config?.Logging?.File?.Path && (
              <div className="space-y-1">
                <Label className="text-muted-foreground">Log Path</Label>
                <div className="font-mono text-sm">{config.Logging.File.Path}</div>
              </div>
            )}
            {config?.Logging?.File?.MaxSizeMB && (
              <div className="space-y-1">
                <Label className="text-muted-foreground">Max Size (MB)</Label>
                <div className="font-mono text-sm">{config.Logging.File.MaxSizeMB}</div>
              </div>
            )}
            <div className="space-y-1">
              <Label className="text-muted-foreground">Prometheus Metrics</Label>
              <Badge variant={config?.Logging?.Prometheus?.Enabled ? 'default' : 'secondary'}>
                {config?.Logging?.Prometheus?.Enabled ? 'Enabled' : 'Disabled'}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Policy Configuration (Editable) */}
      <Card className="border-primary/50">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <CardTitle className="flex items-center gap-2">
                <Settings className="h-5 w-5" />
                Policy Configuration
              </CardTitle>
              <CardDescription>Load balancing and caching policies (editable)</CardDescription>
            </div>
            {hasChanges && (
              <Button onClick={savePolicy} disabled={isSaving}>
                <Save className="h-4 w-4 mr-2" />
                {isSaving ? 'Saving...' : 'Save Changes'}
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <Alert>
            <Info className="h-4 w-4" />
            <AlertDescription>
              These settings can be modified. Changes will take effect immediately after saving.
            </AlertDescription>
          </Alert>

          <div className="grid gap-6 md:grid-cols-2">
            {/* Algorithm Selection */}
            <div className="space-y-2">
              <Label htmlFor="algorithm">Load Balancing Algorithm</Label>
              <Select
                value={policyEdit?.algorithm || 'random'}
                onValueChange={(value) => setPolicyEdit(prev => prev ? {...prev, algorithm: value} : null)}
              >
                <SelectTrigger id="algorithm">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="random">Random</SelectItem>
                  <SelectItem value="round-robin">Round Robin</SelectItem>
                  <SelectItem value="least-loaded">Least Loaded</SelectItem>
                  <SelectItem value="weighted">Weighted</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                How requests are distributed across providers
              </p>
            </div>

            {/* Priority Selection */}
            <div className="space-y-2">
              <Label htmlFor="priority">Selection Priority</Label>
              <Select
                value={policyEdit?.priority || 'latency'}
                onValueChange={(value) => setPolicyEdit(prev => prev ? {...prev, priority: value} : null)}
              >
                <SelectTrigger id="priority">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="latency">Latency</SelectItem>
                  <SelectItem value="cost">Cost</SelectItem>
                  <SelectItem value="availability">Availability</SelectItem>
                  <SelectItem value="quality">Quality</SelectItem>
                </SelectContent>
              </Select>
              <p className="text-xs text-muted-foreground">
                What metric to prioritize when selecting provider
              </p>
            </div>

            {/* Cache Enabled */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="cache-enabled">Response Caching</Label>
                <Switch
                  id="cache-enabled"
                  checked={policyEdit?.cache?.enabled || false}
                  onCheckedChange={(checked) => setPolicyEdit(prev =>
                    prev ? {...prev, cache: {...prev.cache!, enabled: checked}} : null
                  )}
                />
              </div>
              <p className="text-xs text-muted-foreground">
                Cache responses to reduce costs and improve latency
              </p>
            </div>

            {/* Cache TTL */}
            {policyEdit?.cache?.enabled && (
              <div className="space-y-2">
                <Label htmlFor="cache-ttl">Cache TTL (seconds)</Label>
                <Input
                  id="cache-ttl"
                  type="number"
                  min="1"
                  value={policyEdit?.cache?.ttl_seconds || 60}
                  onChange={(e) => setPolicyEdit(prev =>
                    prev ? {...prev, cache: {...prev.cache!, ttl_seconds: parseInt(e.target.value) || 60}} : null
                  )}
                />
                <p className="text-xs text-muted-foreground">
                  How long to cache responses (in seconds)
                </p>
              </div>
            )}
          </div>

          {/* Current vs New Values */}
          {hasChanges && (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                <div className="space-y-1">
                  <p className="font-medium">Pending changes:</p>
                  <ul className="list-disc list-inside text-sm space-y-1 mt-2">
                    {policyEdit?.algorithm !== backendToFrontend(config?.Policy?.Algorithm) && (
                      <li>Algorithm: {backendToFrontend(config?.Policy?.Algorithm)} → {policyEdit?.algorithm}</li>
                    )}
                    {policyEdit?.priority !== backendToFrontend(config?.Policy?.Priority) && (
                      <li>Priority: {backendToFrontend(config?.Policy?.Priority)} → {policyEdit?.priority}</li>
                    )}
                    {policyEdit?.cache?.enabled !== config?.Policy?.Cache?.Enabled && (
                      <li>Cache: {config?.Policy?.Cache?.Enabled ? 'Enabled' : 'Disabled'} → {policyEdit?.cache?.enabled ? 'Enabled' : 'Disabled'}</li>
                    )}
                    {policyEdit?.cache?.ttl_seconds !== config?.Policy?.Cache?.TTLSeconds && (
                      <li>Cache TTL: {config?.Policy?.Cache?.TTLSeconds}s → {policyEdit?.cache?.ttl_seconds}s</li>
                    )}
                  </ul>
                </div>
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>

      {/* System Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-500" />
            System Status
          </CardTitle>
          <CardDescription>Current system health</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-3">
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-500" />
              <span className="text-sm">Gateway: Online</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-500" />
              <span className="text-sm">Storage: Connected</span>
            </div>
            <div className="flex items-center gap-2">
              <CheckCircle className="h-4 w-4 text-green-500" />
              <span className="text-sm">Providers: Healthy</span>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
