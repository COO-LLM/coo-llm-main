import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts'
import {
  Activity,
  Users,
  Database,
  DollarSign,
  RefreshCw,
  Clock,
  AlertCircle
} from 'lucide-react'
import { useApi } from '@/lib/api'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface DashboardStats {
  systemStatus: string
  activeClients: number
  totalQueries: number
  totalCost: number
}

interface ProviderData {
  name: string
  value: number
  color: string
  [key: string]: any
}

interface UsageData {
  name: string
  queries: number
  cost: number
}

export default function Dashboard() {
  const [refreshInterval, setRefreshInterval] = useState('30')
  const [lastRefresh, setLastRefresh] = useState(new Date())
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [providerData, setProviderData] = useState<ProviderData[]>([])
  const [usageData, setUsageData] = useState<UsageData[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const api = useApi()

  const fetchDashboardData = async () => {
    try {
      setIsLoading(true)
      setError(null)

      // Fetch global metrics for the last 7 days
      const endTime = Math.floor(Date.now() / 1000) // Convert to seconds
      const startTime = endTime - (7 * 24 * 60 * 60) // 7 days ago in seconds

      const [globalMetrics, clientStats, statsData] = await Promise.all([
        api.getGlobalMetrics(startTime, endTime),
        api.getClientStats(startTime, endTime),
        api.getStats(['provider'], {}, startTime, endTime)
      ])

      // Process stats
      const totalQueries = globalMetrics?.total_requests || 0
      const totalCost = globalMetrics?.total_cost || 0
      const activeClients = clientStats?.client_stats ? Object.keys(clientStats.client_stats).length : 0

      console.log('📊 Processed dashboard stats:', { totalQueries, totalCost, activeClients, globalMetrics, clientStats })

      setStats({
        systemStatus: 'online', // TODO: Get from health check
        activeClients,
        totalQueries,
        totalCost
      })

      // Process provider data for pie chart
      console.log('📊 Stats data from API:', statsData)
      const providerStats = statsData?.stats || {}
      console.log('📊 Provider stats:', providerStats)

      const providerColors = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4']
      const processedProviderData = Object.entries(providerStats).map(([provider, data]: [string, any], index: number) => {
        console.log(`📊 Processing provider ${provider}:`, data)
        return {
          name: provider,
          value: data.queries || 0,
          color: providerColors[index % providerColors.length]
        }
      })
      console.log('📊 Processed provider data:', processedProviderData)
      setProviderData(processedProviderData.filter(p => p.value > 0))

      // Process usage data for bar chart (daily breakdown)
      // Get metrics for each day in the last 7 days
      const dailyData: UsageData[] = []

      // Fetch daily metrics in parallel
      const dailyPromises = []
      for (let i = 6; i >= 0; i--) {
        const dayEnd = Math.floor(Date.now() / 1000) - (i * 24 * 60 * 60)
        const dayStart = dayEnd - (24 * 60 * 60)
        dailyPromises.push(
          api.getGlobalMetrics(dayStart, dayEnd)
            .then(metrics => ({ dayEnd, metrics }))
            .catch(err => {
              console.error(`Failed to fetch metrics for day ${i}:`, err)
              return { dayEnd, metrics: null }
            })
        )
      }

      const dailyResults = await Promise.all(dailyPromises)

      for (const { dayEnd, metrics } of dailyResults) {
        const date = new Date(dayEnd * 1000)
        // Format as "MM/DD" for better clarity
        const dayLabel = `${(date.getMonth() + 1).toString().padStart(2, '0')}/${date.getDate().toString().padStart(2, '0')}`

        dailyData.push({
          name: dayLabel,
          queries: metrics?.total_requests || 0,
          cost: metrics?.total_cost || 0
        })
      }

      console.log('📊 Daily usage data:', dailyData)
      setUsageData(dailyData)

      setLastRefresh(new Date())
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err)
      setError('Failed to load dashboard data')
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    console.log('📊 Dashboard mounted')
    fetchDashboardData()
    return () => console.log('📊 Dashboard unmounted')
  }, [])

  useEffect(() => {
    if (refreshInterval !== 'manual') {
      const interval = setInterval(fetchDashboardData, parseInt(refreshInterval) * 1000)
      return () => clearInterval(interval)
    }
  }, [refreshInterval])

  const handleRefresh = () => {
    console.log('🔄 Refreshing dashboard data...')
    fetchDashboardData()
  }

  const formatLastRefresh = (date: Date) => {
    return date.toLocaleTimeString()
  }

  if (error) {
    return (
      <div className="space-y-6">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button onClick={handleRefresh} variant="outline">
          <RefreshCw className="h-4 w-4 mr-2" />
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 rounded-lg p-6 border border-slate-200 dark:border-slate-700">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">Dashboard</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Overview of COO-LLM Gateway System
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
            <Button onClick={handleRefresh} variant="outline" size="icon" disabled={isLoading}>
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">System Status</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                <Badge variant={stats?.systemStatus === 'online' ? 'default' : 'destructive'}>
                  {stats?.systemStatus || 'unknown'}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground">
                Gateway operational
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Active Clients</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats?.activeClients || 0}</div>
              <p className="text-xs text-muted-foreground">
                Currently active
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Queries</CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{(stats?.totalQueries || 0).toLocaleString()}</div>
              <p className="text-xs text-muted-foreground">
                Last 7 days
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Cost</CardTitle>
              <DollarSign className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">${(stats?.totalCost || 0).toFixed(2)}</div>
              <p className="text-xs text-muted-foreground">
                Last 7 days
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Charts */}
        <div className="grid gap-6 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Daily Usage</CardTitle>
              <CardDescription>
                Query volume over the last 7 days
              </CardDescription>
            </CardHeader>
            <CardContent>
              <ResponsiveContainer width="100%" height={300}>
                <BarChart data={usageData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="queries" fill="#3B82F6" />
                </BarChart>
              </ResponsiveContainer>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Provider Distribution</CardTitle>
              <CardDescription>
                Query distribution by provider
              </CardDescription>
            </CardHeader>
            <CardContent>
              {providerData.length === 0 ? (
                <div className="flex items-center justify-center h-[300px] text-muted-foreground">
                  <div className="text-center">
                    <Database className="h-12 w-12 mx-auto mb-2 opacity-50" />
                    <p>No provider data available</p>
                    <p className="text-sm">Make some API calls to see provider distribution</p>
                  </div>
                </div>
              ) : (
                <ResponsiveContainer width="100%" height={300}>
                  <PieChart>
                    <Pie
                      data={providerData}
                      cx="50%"
                      cy="50%"
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                      label={({ name, value }) => `${name}: ${value}`}
                    >
                      {providerData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Last Refresh Info */}
        <div className="flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center gap-1">
            <Clock className="h-4 w-4" />
            Last updated: {formatLastRefresh(lastRefresh)}
          </div>
          <div>Auto-refresh: {refreshInterval === 'manual' ? 'Manual' : `${refreshInterval}s`}</div>
        </div>
      </div>
  )
}