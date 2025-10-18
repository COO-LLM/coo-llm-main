import { useState, useEffect } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
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
  LineChart,
  Line
} from 'recharts'
import {
  RefreshCw,
  Download,
  PieChart as PieChartIcon,
  BarChart3,
  Activity,
  TrendingUp,
  Code,
  Database,
  Users,
  DollarSign,
  AlertCircle
} from 'lucide-react'
import { useApi } from '@/lib/api'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface ProviderStats {
  provider: string
  totalQueries: number
  totalTokens: number
  totalCost: number
  avgLatency: number
  clients: number
  models: string[]
}

interface ClientStats {
  client: string
  totalQueries: number
  totalTokens: number
  totalCost: number
  avgLatency: number
  providers: string[]
  topProvider: string
}

interface ModelStats {
  model: string
  queries: number
  tokens: number
  cost: number
  provider: string
}

const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899']

export default function StatisticsPage() {
  const [groupBy, setGroupBy] = useState('provider')
  const [isLoading, setIsLoading] = useState(true)
  const [activeTab, setActiveTab] = useState('overview')
  const [refreshInterval, setRefreshInterval] = useState('30')
  const [providerStats, setProviderStats] = useState<ProviderStats[]>([])
  const [clientStats, setClientStats] = useState<ClientStats[]>([])
  const [modelStats, setModelStats] = useState<ModelStats[]>([])
  const [globalMetrics, setGlobalMetrics] = useState<any>(null)
  const [error, setError] = useState<string | null>(null)

  const api = useApi()

  useEffect(() => {
    console.log('📊 Statistics page mounted')
    fetchStatistics()
  }, [])

  useEffect(() => {
    if (refreshInterval !== 'manual') {
      const interval = setInterval(fetchStatistics, parseInt(refreshInterval) * 1000)
      return () => clearInterval(interval)
    }
  }, [refreshInterval])

  const fetchStatistics = async () => {
    setIsLoading(true)
    setError(null)
    console.log('📊 Fetching statistics...')

    try {
      // Get time range for stats (last 30 days) - API expects SECONDS not milliseconds
      const endTime = Math.floor(Date.now() / 1000) // Convert to seconds
      const startTime = endTime - (30 * 24 * 60 * 60) // 30 days ago in seconds

      // Fetch statistics grouped by different dimensions + global metrics
      const [providerData, clientData, globalData] = await Promise.all([
        api.getStats(['provider'], {}, startTime, endTime),
        api.getStats(['client_key'], {}, startTime, endTime),
        api.getGlobalMetrics(startTime, endTime)
      ])

      console.log('✅ Statistics data received:', { providerData, clientData, globalData })

      // Store global metrics for summary cards
      setGlobalMetrics(globalData)

      // Process provider stats
      const processedProviderStats: ProviderStats[] = []
      if (providerData && providerData.stats) {
        Object.entries(providerData.stats).forEach(([provider, data]: [string, any]) => {
          processedProviderStats.push({
            provider,
            totalQueries: data.queries || data.requests || 0,
            totalTokens: data.tokens || 0,
            totalCost: data.cost || 0,
            avgLatency: data.avg_latency || 0,
            clients: 1, // Will be calculated from actual data
            models: [] // TODO: Get from model-specific stats
          })
        })
      }
      setProviderStats(processedProviderStats)

      // Process client stats
      const processedClientStats: ClientStats[] = []
      if (clientData && clientData.stats) {
        Object.entries(clientData.stats).forEach(([client, data]: [string, any]) => {
          // Aggregate all providers for this client
          let totalQueries = 0
          let totalTokens = 0
          let totalCost = 0
          const providers: string[] = []

          // If data is nested by provider
          if (typeof data === 'object' && !data.queries) {
            Object.entries(data).forEach(([provider, providerData]: [string, any]) => {
              totalQueries += providerData.queries || providerData.requests || 0
              totalTokens += providerData.tokens || 0
              totalCost += providerData.cost || 0
              providers.push(provider)
            })
          } else {
            // Direct stats
            totalQueries = data.queries || data.requests || 0
            totalTokens = data.tokens || 0
            totalCost = data.cost || 0
          }

          const topProvider = providers.length > 0 ? providers[0] : 'unknown'

          processedClientStats.push({
            client,
            totalQueries,
            totalTokens,
            totalCost,
            avgLatency: data.avg_latency || 0,
            providers,
            topProvider
          })
        })
      }
      setClientStats(processedClientStats)

      // For now, keep model stats as mock data since we don't have model-level grouping yet
      setModelStats([])

      console.log('✅ Statistics data processed')
    } catch (error) {
      console.error('❌ Failed to fetch statistics:', error)
      setError('Failed to load statistics data')
    } finally {
      setIsLoading(false)
    }
  }

  const getCurrentData = () => {
    switch (groupBy) {
      case 'provider':
        return providerStats
      case 'client':
        return clientStats
      case 'model':
        return modelStats
      default:
        return providerStats
    }
  }

  const exportData = () => {
    console.log('📥 Exporting statistics data...')
    const data = getCurrentData()
    const csvContent = [
      Object.keys(data[0]).join(','),
      ...data.map(row => Object.values(row).join(','))
    ].join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `statistics_${groupBy}_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  const formatNumber = (num: number) => {
    if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
    if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
    return num.toString()
  }

  if (error) {
    return (
      <div className="space-y-6">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
        <Button onClick={fetchStatistics} variant="outline">
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
            <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">Statistics</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Comprehensive analytics and insights
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
            <Select value={groupBy} onValueChange={setGroupBy}>
              <SelectTrigger className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="provider">Group by Provider</SelectItem>
                <SelectItem value="client">Group by Client</SelectItem>
                <SelectItem value="model">Group by Model</SelectItem>
              </SelectContent>
            </Select>
            <Button onClick={fetchStatistics} disabled={isLoading} variant="outline" size="icon">
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {globalMetrics?.total_requests?.toLocaleString() || '0'}
              </div>
              <p className="text-xs text-muted-foreground">
                Last 30 days
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Tokens</CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {globalMetrics?.total_tokens ? formatNumber(globalMetrics.total_tokens) : '0'}
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
                ${globalMetrics?.total_cost?.toFixed(2) || '0.00'}
              </div>
              <p className="text-xs text-muted-foreground">
                Total spent
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Avg Latency</CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {globalMetrics?.avg_latency ? Math.round(globalMetrics.avg_latency) : '0'}ms
              </div>
              <p className="text-xs text-muted-foreground">
                Average response time
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Charts and Data */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="comparison">Comparison</TabsTrigger>
            <TabsTrigger value="raw">Raw Data</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            <div className="grid gap-6 lg:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <BarChart3 className="h-5 w-5" />
                    Query Distribution
                  </CardTitle>
                  <CardDescription>
                    Distribution of queries by {groupBy}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={getCurrentData()}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey={groupBy === 'model' ? 'model' : groupBy === 'client' ? 'client' : 'provider'} />
                      <YAxis />
                      <Tooltip />
                      <Bar dataKey="totalQueries" fill="#3B82F6" />
                    </BarChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <PieChartIcon className="h-5 w-5" />
                    Cost Distribution
                  </CardTitle>
                  <CardDescription>
                    Cost breakdown by {groupBy}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={getCurrentData()}
                        cx="50%"
                        cy="50%"
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="totalCost"
                        label={({ name, totalCost }) => `${name}: $${totalCost.toFixed(0)}`}
                      >
                        {getCurrentData().map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>
    </div>
          </TabsContent>

          <TabsContent value="comparison" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Performance Comparison</CardTitle>
                <CardDescription>
                  Compare latency and cost across different {groupBy}s
                </CardDescription>
              </CardHeader>
              <CardContent>
                <ResponsiveContainer width="100%" height={400}>
                  <LineChart data={getCurrentData()}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey={groupBy === 'model' ? 'model' : groupBy === 'client' ? 'client' : 'provider'} />
                    <YAxis yAxisId="left" />
                    <YAxis yAxisId="right" orientation="right" />
                    <Tooltip />
                    <Bar yAxisId="left" dataKey="avgLatency" fill="#3B82F6" name="Avg Latency (ms)" />
                    <Line yAxisId="right" type="monotone" dataKey="totalCost" stroke="#EF4444" strokeWidth={2} name="Total Cost ($)" />
                  </LineChart>
                </ResponsiveContainer>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="raw" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Code className="h-5 w-5" />
                  Raw Data
                </CardTitle>
                <CardDescription>
                  Raw statistics data in JSON format
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="rounded-md bg-muted p-4">
                  <pre className="text-sm overflow-auto max-h-96">
                    <code>{JSON.stringify(getCurrentData(), null, 2)}</code>
                  </pre>
    </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Detailed Table</CardTitle>
                <CardDescription>
                  Tabular view of all statistics
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>
                          {groupBy === 'model' ? 'Model' : groupBy === 'client' ? 'Client' : 'Provider'}
                        </TableHead>
                        <TableHead>Queries</TableHead>
                        <TableHead>Tokens</TableHead>
                        <TableHead>Cost</TableHead>
                        <TableHead>Avg Latency</TableHead>
                        {groupBy === 'provider' && <TableHead>Clients</TableHead>}
                        {groupBy === 'client' && <TableHead>Providers</TableHead>}
                        {groupBy === 'model' && <TableHead>Provider</TableHead>}
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {getCurrentData().map((item, index) => (
                        <TableRow key={index}>
                          <TableCell className="font-medium">
                            {groupBy === 'model' ? item.model : groupBy === 'client' ? item.client : item.provider}
                          </TableCell>
                          <TableCell>{item.totalQueries?.toLocaleString()}</TableCell>
                          <TableCell>{item.totalTokens?.toLocaleString()}</TableCell>
                          <TableCell>${item.totalCost?.toFixed(2)}</TableCell>
                          <TableCell>{item.avgLatency}ms</TableCell>
                          {groupBy === 'provider' && (
                            <TableCell>
                              <Badge variant="secondary">{item.clients}</Badge>
                            </TableCell>
                          )}
                          {groupBy === 'client' && (
                            <TableCell>
                              <div className="flex gap-1 flex-wrap">
                                {item.providers?.map((provider: string) => (
                                  <Badge key={provider} variant="outline" className="text-xs">
                                    {provider}
                                  </Badge>
                                ))}
    </div>
                            </TableCell>
                          )}
                          {groupBy === 'model' && (
                            <TableCell>
                              <Badge variant="outline">{item.provider}</Badge>
                            </TableCell>
                          )}
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
  )
}