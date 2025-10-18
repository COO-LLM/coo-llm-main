
import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar as CalendarComponent } from '@/components/ui/calendar'
import { format } from 'date-fns'
import { cn } from '@/lib/utils'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  AreaChart,
  Area
} from 'recharts'
import {
  RefreshCw,
  Download,
  Filter,
  Calendar,
  TrendingUp,
  Clock,
  Zap,
  DollarSign,
  Database,
  Activity,
  Users
} from 'lucide-react'
import { useApi } from '@/lib/api'
import { useAuth } from '@/contexts/auth-context'

// Mock data for metrics
const mockMetricsData = [
  { timestamp: '2024-01-15 10:00', latency: 120, tokens: 1500, cost: 0.75, provider: 'openai', providerKey: 'key1', client: 'client_1' },
  { timestamp: '2024-01-15 10:05', latency: 95, tokens: 1200, cost: 0.60, provider: 'anthropic', providerKey: 'key2', client: 'client_2' },
  { timestamp: '2024-01-15 10:10', latency: 110, tokens: 1800, cost: 0.90, provider: 'openai', providerKey: 'key1', client: 'client_1' },
  { timestamp: '2024-01-15 10:15', latency: 85, tokens: 900, cost: 0.45, provider: 'google', providerKey: 'key3', client: 'client_3' },
  { timestamp: '2024-01-15 10:20', latency: 130, tokens: 2000, cost: 1.00, provider: 'openai', providerKey: 'key1', client: 'client_2' },
  { timestamp: '2024-01-15 10:25', latency: 105, tokens: 1400, cost: 0.70, provider: 'anthropic', providerKey: 'key2', client: 'client_1' },
  { timestamp: '2024-01-15 10:30', latency: 140, tokens: 2200, cost: 1.10, provider: 'openai', providerKey: 'key1', client: 'client_3' },
  { timestamp: '2024-01-15 10:35', latency: 90, tokens: 1100, cost: 0.55, provider: 'google', providerKey: 'key3', client: 'client_2' }
]

interface MetricDataPoint {
  timestamp: string
  latency: number
  tokens: number
  cost: number
  provider: string
  providerKey: string
  client: string
}

export default function MetricsPage() {
  const api = useApi()
  const { token } = useAuth()
  const [filters, setFilters] = useState({
    dateRange: 'last_24h',
    provider: 'all',
    providerKey: 'all',
    client: 'all',
    displayMode: 'aggregated' // 'aggregated' | 'by_provider' | 'by_provider_key' | 'by_client'
  })
  const [isLoading, setIsLoading] = useState(false)
  const [metricsData, setMetricsData] = useState<MetricDataPoint[]>([])
  const [filteredData, setFilteredData] = useState<MetricDataPoint[]>([])
  const [globalMetrics, setGlobalMetrics] = useState<any>(null)
  const [refreshInterval, setRefreshInterval] = useState('manual')
  const [lastRefresh, setLastRefresh] = useState(new Date())
  const [availableProviders, setAvailableProviders] = useState<string[]>([])
  const [availableProviderKeys, setAvailableProviderKeys] = useState<string[]>([])
  const [availableClients, setAvailableClients] = useState<string[]>([])
  const [customDateRange, setCustomDateRange] = useState<{ from?: Date; to?: Date }>({
    from: undefined,
    to: undefined
  })

  useEffect(() => {
    console.log('📊 Metrics page mounted')
    if (token) {
      fetchMetrics()
      fetchProvidersAndClients()
    }
  }, [token])

  useEffect(() => {
    applyFilters()
  }, [filters, metricsData])

  useEffect(() => {
    if (refreshInterval !== 'manual') {
      const interval = setInterval(fetchMetrics, parseInt(refreshInterval) * 1000)
      return () => clearInterval(interval)
    }
  }, [refreshInterval])

  const fetchProvidersAndClients = async () => {
    try {
      // Fetch config to get providers
      const config = await api.getConfig()
      console.log('📊 Config received:', config)

      if (config?.Providers) {
        const providerIds = config.Providers.map((p: any) => p.ID || p.id).filter(Boolean)
        setAvailableProviders(providerIds)
        console.log('📊 Available providers:', providerIds)
      }

      // Fetch clients list
      const clientsData = await api.listClients()
      console.log('📊 Clients received:', clientsData)

      if (clientsData?.clients) {
        const clientIds = clientsData.clients.map((c: any) => c.id || c.client_id).filter(Boolean)
        setAvailableClients(clientIds)
        console.log('📊 Available clients:', clientIds)
      }
    } catch (error) {
      console.error('❌ Failed to fetch providers/clients:', error)
    }
  }

  const getTimeRange = () => {
    const now = Math.floor(Date.now() / 1000)
    switch (filters.dateRange) {
      case 'last_24h':
        return { start: now - 24 * 60 * 60, end: now }
      case 'last_7d':
        return { start: now - 7 * 24 * 60 * 60, end: now }
      case 'last_30d':
        return { start: now - 30 * 24 * 60 * 60, end: now }
      case 'custom':
        // Use custom date range if set
        if (customDateRange.from && customDateRange.to) {
          return {
            start: Math.floor(customDateRange.from.getTime() / 1000),
            end: Math.floor(customDateRange.to.getTime() / 1000)
          }
        }
        // Fallback to last 24h if custom range not set
        return { start: now - 24 * 60 * 60, end: now }
      default:
        return { start: now - 24 * 60 * 60, end: now }
    }
  }

  const fetchMetrics = async () => {
    setIsLoading(true)
    console.log('📊 Fetching metrics with params:', filters)

    try {
      const { start, end } = getTimeRange()

      // Fetch global metrics (API expects seconds)
      const global = await api.getGlobalMetrics(start, end)
      setGlobalMetrics(global)
      console.log('✅ Global metrics received:', global)

      // Fetch time-series metrics data for latency, tokens, and cost (API expects seconds)
      const [latencyMetrics, tokenMetrics, costMetrics] = await Promise.all([
        api.getMetrics('latency', {}, start, end),
        api.getMetrics('tokens', {}, start, end),
        api.getMetrics('cost', {}, start, end)
      ])

      console.log('✅ Time-series metrics received:', { latencyMetrics, tokenMetrics, costMetrics })

      // Debug: Check API response structure
      console.log('📊 Latency metrics structure:', {
        hasPoints: !!latencyMetrics?.points,
        pointsCount: latencyMetrics?.points?.length || 0,
        firstPoint: latencyMetrics?.points?.[0]
      })
      console.log('📊 Token metrics structure:', {
        hasPoints: !!tokenMetrics?.points,
        pointsCount: tokenMetrics?.points?.length || 0,
        firstPoint: tokenMetrics?.points?.[0]
      })
      console.log('📊 Cost metrics structure:', {
        hasPoints: !!costMetrics?.points,
        pointsCount: costMetrics?.points?.length || 0,
        firstPoint: costMetrics?.points?.[0]
      })

      // Create a map of timestamps to combine data points
      const dataMap = new Map<string, Partial<MetricDataPoint>>()

      // Helper to safely convert timestamp
      const safeTimestamp = (ts: any): string | null => {
        try {
          // Handle various timestamp formats
          let timestamp: number
          if (typeof ts === 'string') {
            timestamp = parseInt(ts)
          } else if (typeof ts === 'number') {
            timestamp = ts
          } else {
            return null
          }

          // Validate timestamp is in valid range
          if (isNaN(timestamp) || timestamp <= 0) {
            return null
          }

          // Convert to ISO string (assumes timestamp is in seconds)
          const date = new Date(timestamp * 1000)
          if (isNaN(date.getTime())) {
            return null
          }

          return date.toISOString()
        } catch (err) {
          console.error('Error converting timestamp:', ts, err)
          return null
        }
      }

      // Process latency metrics
      if (latencyMetrics && latencyMetrics.points) {
        latencyMetrics.points.forEach((point: any) => {
          // Handle both PascalCase (backend) and camelCase
          const ts = point.Timestamp || point.timestamp
          const val = point.Value || point.value
          const tags = point.Tags || point.tags

          const timestamp = safeTimestamp(ts)
          if (!timestamp) return

          if (!dataMap.has(timestamp)) {
            dataMap.set(timestamp, { timestamp })
          }
          dataMap.get(timestamp)!.latency = val
          // Extract tags
          if (tags) {
            // Use provider ID directly (not normalized)
            dataMap.get(timestamp)!.provider = tags.provider || 'unknown'
            dataMap.get(timestamp)!.providerKey = tags.key || 'unknown'
            dataMap.get(timestamp)!.client = tags.client_key || 'unknown'
          }
        })
      }

      // Process token metrics
      if (tokenMetrics && tokenMetrics.points) {
        tokenMetrics.points.forEach((point: any) => {
          const ts = point.Timestamp || point.timestamp
          const val = point.Value || point.value

          const timestamp = safeTimestamp(ts)
          if (!timestamp) return

          if (!dataMap.has(timestamp)) {
            dataMap.set(timestamp, { timestamp })
          }
          dataMap.get(timestamp)!.tokens = (dataMap.get(timestamp)!.tokens || 0) + val
        })
      }

      // Process cost metrics
      if (costMetrics && costMetrics.points) {
        costMetrics.points.forEach((point: any) => {
          const ts = point.Timestamp || point.timestamp
          const val = point.Value || point.value

          const timestamp = safeTimestamp(ts)
          if (!timestamp) return

          if (!dataMap.has(timestamp)) {
            dataMap.set(timestamp, { timestamp })
          }
          dataMap.get(timestamp)!.cost = (dataMap.get(timestamp)!.cost || 0) + val
        })
      }

      // Convert map to array and fill in missing values
      const sortedData = Array.from(dataMap.entries())
        .sort(([a], [b]) => a.localeCompare(b))
        .map(([, data]) => ({
          timestamp: new Date(data.timestamp!).toLocaleString(),
          latency: data.latency || 0,
          tokens: data.tokens || 0,
          cost: data.cost || 0,
          provider: data.provider || 'unknown',
          providerKey: data.providerKey || 'unknown',
          client: data.client || 'unknown'
        } as MetricDataPoint))

      setMetricsData(sortedData)

      // Extract unique provider keys from data
      const providerKeys = Array.from(new Set(sortedData.map(d => d.providerKey).filter(k => k !== 'unknown')))
      setAvailableProviderKeys(providerKeys)
      console.log('📊 Available provider keys:', providerKeys)

      setLastRefresh(new Date())
      console.log('✅ Metrics data transformed and set:', sortedData.length, 'points')
    } catch (error) {
      console.error('❌ Failed to fetch metrics:', error)
      // Fallback to mock data if API fails
      setMetricsData(mockMetricsData)
    } finally {
      setIsLoading(false)
    }
  }

  const applyFilters = () => {
    let filtered = metricsData

    // Apply provider/client/providerKey filters
    if (filters.provider !== 'all') {
      filtered = filtered.filter(item => item.provider === filters.provider)
    }

    if (filters.providerKey !== 'all') {
      filtered = filtered.filter(item => item.providerKey === filters.providerKey)
    }

    if (filters.client !== 'all') {
      filtered = filtered.filter(item => item.client === filters.client)
    }

    setFilteredData(filtered)
    console.log(`📝 Filtered data points: ${filtered.length}`)
  }

  // Prepare chart data based on display mode
  const prepareChartData = (metric: 'latency' | 'tokens' | 'cost') => {
    if (filters.displayMode === 'aggregated') {
      // MODE 1: Single aggregated line
      const timeMap = new Map<string, number>()

      filteredData.forEach(item => {
        const existing = timeMap.get(item.timestamp) || 0
        timeMap.set(item.timestamp, existing + item[metric])
      })

      return Array.from(timeMap.entries())
        .map(([timestamp, value]) => ({ timestamp, value }))
        .sort((a, b) => a.timestamp.localeCompare(b.timestamp))
    } else if (filters.displayMode === 'by_provider') {
      // MODE 2: Multiple lines by provider
      const providers = Array.from(new Set(filteredData.map(item => item.provider)))
      const timeMap = new Map<string, any>()

      filteredData.forEach(item => {
        if (!timeMap.has(item.timestamp)) {
          timeMap.set(item.timestamp, { timestamp: item.timestamp })
        }
        const entry = timeMap.get(item.timestamp)!
        entry[item.provider] = (entry[item.provider] || 0) + item[metric]
      })

      return {
        data: Array.from(timeMap.values()).sort((a, b) => a.timestamp.localeCompare(b.timestamp)),
        keys: providers
      }
    } else if (filters.displayMode === 'by_provider_key') {
      // MODE 3: Multiple lines by provider key
      const providerKeys = Array.from(new Set(filteredData.map(item => item.providerKey)))
      const timeMap = new Map<string, any>()

      filteredData.forEach(item => {
        if (!timeMap.has(item.timestamp)) {
          timeMap.set(item.timestamp, { timestamp: item.timestamp })
        }
        const entry = timeMap.get(item.timestamp)!
        entry[item.providerKey] = (entry[item.providerKey] || 0) + item[metric]
      })

      return {
        data: Array.from(timeMap.values()).sort((a, b) => a.timestamp.localeCompare(b.timestamp)),
        keys: providerKeys
      }
    } else {
      // MODE 4: Multiple lines by client
      const clients = Array.from(new Set(filteredData.map(item => item.client)))
      const timeMap = new Map<string, any>()

      filteredData.forEach(item => {
        if (!timeMap.has(item.timestamp)) {
          timeMap.set(item.timestamp, { timestamp: item.timestamp })
        }
        const entry = timeMap.get(item.timestamp)!
        entry[item.client] = (entry[item.client] || 0) + item[metric]
      })

      return {
        data: Array.from(timeMap.values()).sort((a, b) => a.timestamp.localeCompare(b.timestamp)),
        keys: clients
      }
    }
  }

  // Get unique providers, provider keys, and clients from data
  const providersFromData = Array.from(new Set(metricsData.map(item => item.provider))).filter(p => p !== 'unknown')
  const providerKeysFromData = Array.from(new Set(metricsData.map(item => item.providerKey))).filter(k => k !== 'unknown')
  const clientsFromData = Array.from(new Set(metricsData.map(item => item.client))).filter(c => c !== 'unknown')

  // Merge providers: API config + data + fallback
  const allProviders = Array.from(new Set([
    ...availableProviders,
    ...providersFromData,
  ]))
  const providers = ['all', ...allProviders]

  // Merge provider keys: state + data
  const allProviderKeys = Array.from(new Set([
    ...availableProviderKeys,
    ...providerKeysFromData,
  ]))
  const providerKeys = ['all', ...allProviderKeys]

  // Merge clients: API list + data
  const allClients = Array.from(new Set([
    ...availableClients,
    ...clientsFromData,
  ]))
  const clients = ['all', ...allClients]

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({ ...prev, [key]: value }))
    console.log('🔄 Filter changed:', key, '=', value)
  }

  const exportData = () => {
    console.log('📥 Exporting metrics data...')
    const csvContent = [
      ['Timestamp', 'Latency (ms)', 'Tokens', 'Cost ($)', 'Provider', 'Client'],
      ...filteredData.map(item => [
        item.timestamp,
        item.latency,
        item.tokens,
        item.cost,
        item.provider,
        item.client
      ])
    ].map(row => row.join(',')).join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `metrics_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  return (
    <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 rounded-lg p-6 border border-slate-200 dark:border-slate-700">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Metrics</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Real-time performance metrics and detailed analytics dashboard
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Select value={refreshInterval} onValueChange={setRefreshInterval}>
              <SelectTrigger className="w-32">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="30">30s</SelectItem>
                <SelectItem value="60">1m</SelectItem>
                <SelectItem value="300">5m</SelectItem>
                <SelectItem value="manual">Manual</SelectItem>
              </SelectContent>
            </Select>
            <Button onClick={exportData} variant="outline" className="hover:bg-blue-50 dark:hover:bg-blue-900">
              <Download className="h-4 w-4 mr-2" />
              Export CSV
            </Button>
            <Button onClick={fetchMetrics} disabled={isLoading} variant="outline" size="icon">
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        {/* Summary Cards */}
        {globalMetrics && (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
                <TrendingUp className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {globalMetrics.total_requests?.toLocaleString() || '0'}
                </div>
                <p className="text-xs text-muted-foreground">
                  Across all providers
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Tokens</CardTitle>
                <Zap className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {globalMetrics.total_tokens?.toLocaleString() || '0'}
                </div>
                <p className="text-xs text-muted-foreground">
                  Tokens processed
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Total Cost</CardTitle>
                <DollarSign className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  ${globalMetrics.total_cost?.toFixed(2) || '0.00'}
                </div>
                <p className="text-xs text-muted-foreground">
                  Total expenditure
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
                <Clock className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {globalMetrics.overall_success_rate ? (globalMetrics.overall_success_rate * 100).toFixed(1) : '0.0'}%
                </div>
                <p className="text-xs text-muted-foreground">
                  Request success rate
                </p>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Filters */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Filter className="h-5 w-5" />
              Filters & Display Options
            </CardTitle>
            <CardDescription>
              Customize your metrics view with filters and visualization modes
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="space-y-4">
              {/* Row 1: Date Range (full width for better UX) */}
              <div>
                <Label htmlFor="dateRange" className="text-sm font-medium mb-2 block">Date Range</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        "w-full justify-start text-left font-normal",
                        !customDateRange.from && "text-muted-foreground"
                      )}
                    >
                      <Calendar className="mr-2 h-4 w-4" />
                      {customDateRange.from ? (
                        customDateRange.to ? (
                          <>
                            {format(customDateRange.from, "LLL dd, y")} - {format(customDateRange.to, "LLL dd, y")}
                          </>
                        ) : (
                          format(customDateRange.from, "LLL dd, y")
                        )
                      ) : filters.dateRange === 'last_24h' ? (
                        <span>Last 24 hours</span>
                      ) : filters.dateRange === 'last_7d' ? (
                        <span>Last 7 days</span>
                      ) : filters.dateRange === 'last_30d' ? (
                        <span>Last 30 days</span>
                      ) : (
                        <span>Pick a date range</span>
                      )}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <div className="flex">
                      {/* Quick select presets */}
                      <div className="border-r p-3 min-w-[140px]">
                        <div className="text-sm font-medium mb-2 text-muted-foreground">Quick Select</div>
                        <div className="flex flex-col gap-1">
                          <Button
                            variant={filters.dateRange === 'last_24h' && !customDateRange.from ? "secondary" : "ghost"}
                            size="sm"
                            className="justify-start text-xs h-8"
                            onClick={() => {
                              handleFilterChange('dateRange', 'last_24h')
                              setCustomDateRange({})
                              fetchMetrics()
                            }}
                          >
                            Last 24 hours
                          </Button>
                          <Button
                            variant={filters.dateRange === 'last_7d' && !customDateRange.from ? "secondary" : "ghost"}
                            size="sm"
                            className="justify-start text-xs h-8"
                            onClick={() => {
                              handleFilterChange('dateRange', 'last_7d')
                              setCustomDateRange({})
                              fetchMetrics()
                            }}
                          >
                            Last 7 days
                          </Button>
                          <Button
                            variant={filters.dateRange === 'last_30d' && !customDateRange.from ? "secondary" : "ghost"}
                            size="sm"
                            className="justify-start text-xs h-8"
                            onClick={() => {
                              handleFilterChange('dateRange', 'last_30d')
                              setCustomDateRange({})
                              fetchMetrics()
                            }}
                          >
                            Last 30 days
                          </Button>
                          <div className="border-t my-2"></div>
                          <div className="text-xs text-muted-foreground px-2">Or pick custom dates →</div>
                        </div>
                      </div>

                      {/* Calendar */}
                      <div>
                        <CalendarComponent
                          mode="range"
                          selected={customDateRange.from ? customDateRange as any : undefined}
                          onSelect={(range: any) => {
                            setCustomDateRange(range || {})
                            // Auto switch to custom and refresh if both dates are selected
                            if (range?.from && range?.to) {
                              handleFilterChange('dateRange', 'custom')
                              fetchMetrics()
                            }
                          }}
                          numberOfMonths={2}
                        />
                      </div>
                    </div>
                  </PopoverContent>
                </Popover>
              </div>

              {/* Row 2: Data Filters */}
              <div>
                <Label className="text-sm font-medium mb-2 block text-muted-foreground">Data Filters</Label>
                <div className="grid gap-3 md:grid-cols-3">
                  <div className="space-y-2">
                    <Label htmlFor="provider" className="text-xs">Provider</Label>
                    <Select value={filters.provider} onValueChange={(value) => handleFilterChange('provider', value)}>
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {providers.map(provider => (
                          <SelectItem key={provider} value={provider}>
                            {provider === 'all' ? 'All Providers' : provider}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="providerKey" className="text-xs">Provider Key</Label>
                    <Select value={filters.providerKey} onValueChange={(value) => handleFilterChange('providerKey', value)}>
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {providerKeys.map(key => (
                          <SelectItem key={key} value={key}>
                            {key === 'all' ? 'All Keys' : key}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="client" className="text-xs">Client</Label>
                    <Select value={filters.client} onValueChange={(value) => handleFilterChange('client', value)}>
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {clients.map(client => (
                          <SelectItem key={client} value={client}>
                            {client === 'all' ? 'All Clients' : client}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>

              {/* Row 3: Display Mode */}
              <div className="border-t pt-4">
                <Label htmlFor="displayMode" className="text-sm font-medium mb-2 block">Visualization Mode</Label>
                <Select value={filters.displayMode} onValueChange={(value) => handleFilterChange('displayMode', value)}>
                  <SelectTrigger className="h-10">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="aggregated">
                      <div className="flex items-center gap-2">
                        <Activity className="h-4 w-4" />
                        <div>
                          <div className="font-medium">Aggregated</div>
                          <div className="text-xs text-muted-foreground">Single line showing totals</div>
                        </div>
                      </div>
                    </SelectItem>
                    <SelectItem value="by_provider">
                      <div className="flex items-center gap-2">
                        <Database className="h-4 w-4" />
                        <div>
                          <div className="font-medium">By Provider</div>
                          <div className="text-xs text-muted-foreground">Multiple lines per provider</div>
                        </div>
                      </div>
                    </SelectItem>
                    <SelectItem value="by_provider_key">
                      <div className="flex items-center gap-2">
                        <TrendingUp className="h-4 w-4" />
                        <div>
                          <div className="font-medium">By Provider Key</div>
                          <div className="text-xs text-muted-foreground">Multiple lines per API key</div>
                        </div>
                      </div>
                    </SelectItem>
                    <SelectItem value="by_client">
                      <div className="flex items-center gap-2">
                        <Users className="h-4 w-4" />
                        <div>
                          <div className="font-medium">By Client</div>
                          <div className="text-xs text-muted-foreground">Multiple lines per client</div>
                        </div>
                      </div>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Charts */}
        <div className="grid gap-6 lg:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Clock className="h-5 w-5" />
                Latency Trends
              </CardTitle>
              <CardDescription>
                Response latency over time (ms)
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-6">
              {filteredData.length === 0 ? (
                <div className="flex items-center justify-center h-[250px] text-muted-foreground">
                  <div className="text-center">
                    <Clock className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p>No latency data available</p>
                    <p className="text-sm">Make some API calls to see latency metrics</p>
                  </div>
                </div>
              ) : (() => {
                const chartData = prepareChartData('latency')
                const colors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4', '#EC4899']

                if (filters.displayMode === 'aggregated') {
                  return (
                    <ResponsiveContainer width="100%" height={250}>
                      <LineChart data={chartData as any} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="timestamp" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                        <YAxis stroke="#94a3b8" tick={{ fontSize: 12 }} />
                        <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px', color: '#fff' }} />
                        <Line type="monotone" dataKey="value" stroke={colors[0]} strokeWidth={3} dot={{ r: 4 }} activeDot={{ r: 6 }} />
                      </LineChart>
                    </ResponsiveContainer>
                  )
                } else {
                  const { data, keys } = chartData as { data: any[], keys: string[] }
                  return (
                    <ResponsiveContainer width="100%" height={250}>
                      <LineChart data={data} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="timestamp" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                        <YAxis stroke="#94a3b8" tick={{ fontSize: 12 }} />
                        <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px', color: '#fff' }} />
                        {keys.map((key, index) => (
                          <Line
                            key={key}
                            type="monotone"
                            dataKey={key}
                            stroke={colors[index % colors.length]}
                            strokeWidth={2}
                            dot={{ r: 3 }}
                            name={key}
                          />
                        ))}
                      </LineChart>
                    </ResponsiveContainer>
                  )
                }
              })()}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5" />
                Token Usage
              </CardTitle>
              <CardDescription>
                Token consumption over time
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-6">
              {filteredData.length === 0 ? (
                <div className="flex items-center justify-center h-[250px] text-muted-foreground">
                  <div className="text-center">
                    <Zap className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p>No token usage data available</p>
                    <p className="text-sm">Make some API calls to see token metrics</p>
                  </div>
                </div>
              ) : (() => {
                const chartData = prepareChartData('tokens')
                const colors = ['#10B981', '#3B82F6', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4', '#EC4899']

                if (filters.displayMode === 'aggregated') {
                  return (
                    <ResponsiveContainer width="100%" height={250}>
                      <AreaChart data={chartData as any} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="timestamp" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                        <YAxis stroke="#94a3b8" tick={{ fontSize: 12 }} />
                        <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px', color: '#fff' }} />
                        <Area type="monotone" dataKey="value" stroke={colors[0]} fill={colors[0]} fillOpacity={0.15} />
                      </AreaChart>
                    </ResponsiveContainer>
                  )
                } else {
                  const { data, keys } = chartData as { data: any[], keys: string[] }
                  return (
                    <ResponsiveContainer width="100%" height={250}>
                      <AreaChart data={data} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="timestamp" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                        <YAxis stroke="#94a3b8" tick={{ fontSize: 12 }} />
                        <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px', color: '#fff' }} />
                        {keys.map((key, index) => (
                          <Area
                            key={key}
                            type="monotone"
                            dataKey={key}
                            stroke={colors[index % colors.length]}
                            fill={colors[index % colors.length]}
                            fillOpacity={0.15}
                            name={key}
                          />
                        ))}
                      </AreaChart>
                    </ResponsiveContainer>
                  )
                }
              })()}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <DollarSign className="h-5 w-5" />
                Cost Analysis
              </CardTitle>
              <CardDescription>
                Cost accumulation over time ($)
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-6">
              {filteredData.length === 0 ? (
                <div className="flex items-center justify-center h-[250px] text-muted-foreground">
                  <div className="text-center">
                    <DollarSign className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p>No cost data available</p>
                    <p className="text-sm">Make some API calls to see cost metrics</p>
                  </div>
                </div>
              ) : (() => {
                const chartData = prepareChartData('cost')
                const colors = ['#F59E0B', '#EF4444', '#3B82F6', '#10B981', '#8B5CF6', '#06B6D4', '#EC4899']

                if (filters.displayMode === 'aggregated') {
                  return (
                    <ResponsiveContainer width="100%" height={250}>
                      <AreaChart data={chartData as any} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="timestamp" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                        <YAxis stroke="#94a3b8" tick={{ fontSize: 12 }} />
                        <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px', color: '#fff' }} />
                        <Area type="monotone" dataKey="value" stroke={colors[0]} fill={colors[0]} fillOpacity={0.15} />
                      </AreaChart>
                    </ResponsiveContainer>
                  )
                } else {
                  const { data, keys } = chartData as { data: any[], keys: string[] }
                  return (
                    <ResponsiveContainer width="100%" height={250}>
                      <AreaChart data={data} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
                        <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                        <XAxis dataKey="timestamp" tick={{ fontSize: 12 }} stroke="#94a3b8" />
                        <YAxis stroke="#94a3b8" tick={{ fontSize: 12 }} />
                        <Tooltip contentStyle={{ backgroundColor: '#1e293b', border: 'none', borderRadius: '8px', color: '#fff' }} />
                        {keys.map((key, index) => (
                          <Area
                            key={key}
                            type="monotone"
                            dataKey={key}
                            stroke={colors[index % colors.length]}
                            fill={colors[index % colors.length]}
                            fillOpacity={0.15}
                            name={key}
                          />
                        ))}
                      </AreaChart>
                    </ResponsiveContainer>
                  )
                }
              })()}
            </CardContent>
          </Card>
        </div>

        {/* Data Table */}
        <Card>
          <CardHeader>
            <CardTitle>📋 Combined Metrics Data</CardTitle>
            <CardDescription>
              Detailed metrics data with all filters applied - Showing {filteredData.length} records
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
              <Table>
                <TableHeader className="bg-slate-50 dark:bg-slate-800">
                  <TableRow className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold">Timestamp</TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold text-right">Latency (ms)</TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold text-right">Tokens</TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold text-right">Cost ($)</TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold">Provider</TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold">Client</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredData.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="h-64 text-center">
                        <div className="flex flex-col items-center justify-center text-muted-foreground">
                          <Database className="h-12 w-12 mb-2 opacity-50" />
                          <p className="text-lg font-medium">No metrics data available</p>
                          <p className="text-sm">Make some API calls through the gateway to see metrics here</p>
                        </div>
                      </TableCell>
                    </TableRow>
                  ) : (
                    filteredData.map((item, index) => (
                      <TableRow
                        key={index}
                        className="border-b border-slate-100 dark:border-slate-700 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
                      >
                        <TableCell className="font-medium text-slate-900 dark:text-slate-100">{item.timestamp}</TableCell>
                        <TableCell className="text-right">
                          <Badge
                            variant={item.latency > 100 ? 'destructive' : 'default'}
                            className="font-semibold"
                          >
                            {item.latency}ms
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right font-semibold text-slate-700 dark:text-slate-300">
                          {item.tokens.toLocaleString()}
                        </TableCell>
                        <TableCell className="text-right font-semibold text-slate-700 dark:text-slate-300">
                          ${item.cost.toFixed(2)}
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline" className="bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-700">
                            {item.provider}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant="secondary" className="bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300">
                            {item.client}
                          </Badge>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
            <div className="mt-4 flex items-center justify-between">
              <div className="text-sm text-slate-600 dark:text-slate-400">
                Showing <span className="font-semibold text-slate-900 dark:text-white">{filteredData.length}</span> of <span className="font-semibold text-slate-900 dark:text-white">{metricsData.length}</span> records
              </div>
              <div className="flex items-center gap-4">
                <div className="text-xs text-slate-500 dark:text-slate-500">
                  Auto-refresh: {refreshInterval === 'manual' ? 'Manual' : `${refreshInterval}s`}
                </div>
                <div className="text-xs text-slate-500 dark:text-slate-500">
                  Last updated: {lastRefresh.toLocaleTimeString()}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
  )
}