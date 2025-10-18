import { useState, useEffect, useMemo } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination'
import {
  Search,
  Download,
  RefreshCw,
  Users,
  Database,
  DollarSign,
  ArrowUpDown,
  ArrowUp,
  ArrowDown,
  Filter,
  Zap,
  AlertCircle
} from 'lucide-react'
import { useApi } from '@/lib/api'
import { Alert, AlertDescription } from '@/components/ui/alert'

// Mock data for clients
const mockClientsData = [
  {
    id: 'client_1',
    apiKey: 'sk-1234567890abcdef',
    totalQueries: 5420,
    tokensConsumed: 2840000,
    costIncurred: 1420.50,
    providersUsed: ['openai', 'anthropic'],
    lastActive: '2024-01-15 10:30',
    status: 'active'
  },
  {
    id: 'client_2',
    apiKey: 'sk-0987654321fedcba',
    totalQueries: 3210,
    tokensConsumed: 1650000,
    costIncurred: 825.00,
    providersUsed: ['openai', 'google'],
    lastActive: '2024-01-15 10:25',
    status: 'active'
  },
  {
    id: 'client_3',
    apiKey: 'sk-abcdef1234567890',
    totalQueries: 1890,
    tokensConsumed: 980000,
    costIncurred: 490.00,
    providersUsed: ['anthropic', 'google'],
    lastActive: '2024-01-15 10:20',
    status: 'active'
  },
  {
    id: 'client_4',
    apiKey: 'sk-fedcba0987654321',
    totalQueries: 890,
    tokensConsumed: 450000,
    costIncurred: 225.00,
    providersUsed: ['openai'],
    lastActive: '2024-01-15 09:45',
    status: 'inactive'
  },
  {
    id: 'client_5',
    apiKey: 'sk-567890abcdef1234',
    totalQueries: 2340,
    tokensConsumed: 1200000,
    costIncurred: 600.00,
    providersUsed: ['openai', 'anthropic', 'google'],
    lastActive: '2024-01-15 10:15',
    status: 'active'
  },
  {
    id: 'client_6',
    apiKey: 'sk-1234fedcba567890',
    totalQueries: 1560,
    tokensConsumed: 800000,
    costIncurred: 400.00,
    providersUsed: ['google'],
    lastActive: '2024-01-15 10:10',
    status: 'active'
  }
]

type SortField = 'id' | 'totalQueries' | 'tokensConsumed' | 'costIncurred' | 'lastActive'
type SortDirection = 'asc' | 'desc' | null

interface ClientData {
  id: string
  apiKey: string
  description?: string
  allowedProviders?: string[]
  totalQueries: number
  tokensConsumed: number
  costIncurred: number
  providersUsed: string[]
  lastActive: string
  status: string
  createdAt: number
  lastUsed: number
}

export default function ClientsPage() {
  const [searchTerm, setSearchTerm] = useState('')
  const [sortField, setSortField] = useState<SortField>('totalQueries')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [currentPage, setCurrentPage] = useState(1)
  const [itemsPerPage] = useState(10)
  const [statusFilter, setStatusFilter] = useState('all')
  const [refreshInterval, setRefreshInterval] = useState('30')
  const [isLoading, setIsLoading] = useState(true)
  const [clientsData, setClientsData] = useState<ClientData[]>([])
  const [error, setError] = useState<string | null>(null)

  const api = useApi()

  const filteredAndSortedData = useMemo(() => {
    let filtered = clientsData.filter(client => {
      const matchesSearch = client.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
                            client.apiKey.toLowerCase().includes(searchTerm.toLowerCase())
      const matchesStatus = statusFilter === 'all' || client.status === statusFilter
      return matchesSearch && matchesStatus
    })

    // Sort data
    if (sortField && sortDirection) {
      filtered.sort((a, b) => {
        let aValue = a[sortField]
        let bValue = b[sortField]

        if (sortField === 'lastActive') {
          aValue = new Date(aValue as string).getTime()
          bValue = new Date(bValue as string).getTime()
        }

        if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1
        if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1
        return 0
      })
    }

    return filtered
  }, [searchTerm, sortField, sortDirection, statusFilter, clientsData])

  const paginatedData = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage
    return filteredAndSortedData.slice(startIndex, startIndex + itemsPerPage)
  }, [filteredAndSortedData, currentPage, itemsPerPage])

  const totalPages = Math.ceil(filteredAndSortedData.length / itemsPerPage)

  useEffect(() => {
    console.log('👥 Clients page mounted')
    fetchClients()
  }, [])

  useEffect(() => {
    if (refreshInterval !== 'manual') {
      const interval = setInterval(fetchClients, parseInt(refreshInterval) * 1000)
      return () => clearInterval(interval)
    }
  }, [refreshInterval])

  const fetchClients = async () => {
    setIsLoading(true)
    setError(null)
    console.log('👥 Fetching clients data...')

    try {
      // Get time range for stats (last 30 days) - API expects SECONDS not milliseconds
      const endTime = Math.floor(Date.now() / 1000) // Convert to seconds
      const startTime = endTime - (30 * 24 * 60 * 60) // 30 days ago in seconds

      // Fetch client list and stats in parallel
      const [clientsResponse, statsResponse] = await Promise.all([
        api.listClients(),
        api.getStats(['client_key'], {}, startTime, endTime)
      ])

      console.log('✅ Clients list received:', clientsResponse)
      console.log('✅ Client stats received:', statsResponse)

      // Process client data
      const processedClients: ClientData[] = []

      // First, process clients from the client list
      const clientMap = new Map<string, any>()
      if (clientsResponse && clientsResponse.clients) {
        clientsResponse.clients.forEach(client => {
          clientMap.set(client.api_key || client.id, client)
        })
      }

      // Then add stats data - stats response has client_key at top level
      if (statsResponse && statsResponse.stats) {
        Object.entries(statsResponse.stats).forEach(([clientKey, stats]: [string, any]) => {
          // Check if we have this client in the list, if not create a placeholder
          let client = clientMap.get(clientKey)
          if (!client) {
            // Client exists in stats but not in list - create placeholder
            client = {
              api_key: clientKey,
              id: clientKey,
              description: '',
              allowed_providers: [],
              created_at: 0,
              last_used: Math.floor(Date.now() / 1000) // Assume recently used if has stats
            }
            clientMap.set(clientKey, client)
          }

          // Stats format: direct stats object with queries/tokens/cost
          const totalQueries = stats.queries || stats.requests || 0
          const totalTokens = stats.tokens || 0
          const totalCost = stats.cost || 0

          // Determine status based on recent activity
          const lastUsed = client.last_used || 0
          const daysSinceLastUse = (Date.now() - lastUsed * 1000) / (1000 * 60 * 60 * 24)
          const status = daysSinceLastUse < 7 ? 'active' : 'inactive'

          processedClients.push({
            id: client.id || clientKey,
            apiKey: client.api_key || clientKey,
            description: client.description || '',
            allowedProviders: client.allowed_providers || [],
            totalQueries: Math.floor(totalQueries),
            tokensConsumed: Math.floor(totalTokens),
            costIncurred: totalCost,
            providersUsed: [], // We don't have provider breakdown in this format
            lastActive: lastUsed ? new Date(lastUsed * 1000).toLocaleString() : 'Never',
            status,
            createdAt: client.created_at || 0,
            lastUsed
          })
        })
      }

      // Add clients from list that don't have stats yet
      clientMap.forEach((client, clientKey) => {
        // Check if already processed
        if (!processedClients.find(c => c.apiKey === clientKey)) {
          const lastUsed = client.last_used || 0
          const daysSinceLastUse = (Date.now() - lastUsed * 1000) / (1000 * 60 * 60 * 24)
          const status = daysSinceLastUse < 7 ? 'active' : 'inactive'

          processedClients.push({
            id: client.id || clientKey,
            apiKey: client.api_key || clientKey,
            description: client.description || '',
            allowedProviders: client.allowed_providers || [],
            totalQueries: 0,
            tokensConsumed: 0,
            costIncurred: 0,
            providersUsed: [],
            lastActive: lastUsed ? new Date(lastUsed * 1000).toLocaleString() : 'Never',
            status,
            createdAt: client.created_at || 0,
            lastUsed
          })
        }
      })

      setClientsData(processedClients)
      console.log('✅ Clients data processed:', processedClients.length, 'clients')
    } catch (error) {
      console.error('❌ Failed to fetch clients:', error)
      setError('Failed to load client data')
      // Fallback to mock data if API fails
      setClientsData(mockClientsData)
    } finally {
      setIsLoading(false)
    }
  }

  const handleSort = (field: SortField) => {
    console.log('🔄 Sorting by:', field)
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : sortDirection === 'desc' ? null : 'asc')
    } else {
      setSortField(field)
      setSortDirection('asc')
    }
  }

  const getSortIcon = (field: SortField) => {
    if (sortField !== field) return <ArrowUpDown className="h-4 w-4 opacity-50" />
    if (sortDirection === 'asc') return <ArrowUp className="h-4 w-4" />
    if (sortDirection === 'desc') return <ArrowDown className="h-4 w-4" />
    return <ArrowUpDown className="h-4 w-4 opacity-50" />
  }

  const exportData = () => {
    console.log('📥 Exporting clients data...')
    const csvContent = [
      ['Client ID', 'API Key', 'Total Queries', 'Tokens Consumed', 'Cost Incurred ($)', 'Providers Used', 'Last Active', 'Status'],
      ...filteredAndSortedData.map(client => [
        client.id,
        client.apiKey,
        client.totalQueries,
        client.tokensConsumed,
        client.costIncurred,
        client.providersUsed.join(';'),
        client.lastActive,
        client.status
      ])
    ].map(row => row.join(',')).join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `clients_${new Date().toISOString().split('T')[0]}.csv`
    a.click()
    window.URL.revokeObjectURL(url)
  }

  const maskApiKey = (apiKey: string) => {
    if (apiKey.length <= 8) return apiKey
    return apiKey.slice(0, 8) + '*'.repeat(apiKey.length - 8)
  }

  return (
    <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 bg-gradient-to-r from-slate-50 to-slate-100 dark:from-slate-900 dark:to-slate-800 rounded-lg p-6 border border-slate-200 dark:border-slate-700">
          <div>
            <h1 className="text-3xl font-bold tracking-tight text-slate-900 dark:text-white">Clients</h1>
            <p className="text-slate-600 dark:text-slate-400 mt-1">
              Manage and monitor client API usage and performance metrics
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
            <Button onClick={exportData} variant="outline">
              <Download className="h-4 w-4 mr-2" />
              Export CSV
            </Button>
            <Button onClick={fetchClients} disabled={isLoading} variant="outline" size="icon">
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        {/* Stats Cards */}
        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Clients</CardTitle>
              <Users className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{clientsData.length}</div>
              <p className="text-xs text-muted-foreground">
                <span className="font-semibold">{clientsData.filter(c => c.status === 'active').length} active</span> • {clientsData.filter(c => c.status === 'inactive').length} inactive
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Queries</CardTitle>
              <Database className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {clientsData.reduce((sum, client) => sum + client.totalQueries, 0).toLocaleString()}
              </div>
              <p className="text-xs text-muted-foreground">
                Total requests across all clients
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Tokens Consumed</CardTitle>
              <Zap className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {(clientsData.reduce((sum, client) => sum + client.tokensConsumed, 0) / 1000000).toFixed(1)}M
              </div>
              <p className="text-xs text-muted-foreground">
                Total tokens processed
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
                ${clientsData.reduce((sum, client) => sum + client.costIncurred, 0).toFixed(2)}
              </div>
              <p className="text-xs text-muted-foreground">
                Total expenditure
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Filters */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Filter className="h-5 w-5" />
              Search & Filter
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="flex-1">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder="Search clients..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10"
                  />
    </div>
              </div>
              <div className="w-full sm:w-48">
                <Select value={statusFilter} onValueChange={setStatusFilter}>
                  <SelectTrigger>
                    <SelectValue placeholder="Filter by status" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Status</SelectItem>
                    <SelectItem value="active">Active</SelectItem>
                    <SelectItem value="inactive">Inactive</SelectItem>
                  </SelectContent>
                </Select>
    </div>
            </div>
          </CardContent>
        </Card>

        {/* Clients Table */}
        <Card>
          <CardHeader>
            <CardTitle>Clients Overview</CardTitle>
            <CardDescription>
              Manage all clients with sortable columns and detailed usage statistics - {filteredAndSortedData.length} results
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-6">
            <div className="rounded-lg border border-slate-200 dark:border-slate-700 overflow-hidden">
              <Table>
                <TableHeader className="bg-slate-50 dark:bg-slate-800">
                  <TableRow className="border-b border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800">
                    <TableHead
                      className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/50 text-slate-700 dark:text-slate-300 font-semibold transition-colors"
                      onClick={() => handleSort('id')}
                    >
                      <div className="flex items-center gap-2">
                        Client ID {getSortIcon('id')}
    </div>
                    </TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold">API Key</TableHead>
                    <TableHead
                      className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/50 text-slate-700 dark:text-slate-300 font-semibold transition-colors text-right"
                      onClick={() => handleSort('totalQueries')}
                    >
                      <div className="flex items-center gap-2 justify-end">
                        Queries {getSortIcon('totalQueries')}
    </div>
                    </TableHead>
                    <TableHead
                      className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/50 text-slate-700 dark:text-slate-300 font-semibold transition-colors text-right"
                      onClick={() => handleSort('tokensConsumed')}
                    >
                      <div className="flex items-center gap-2 justify-end">
                        Tokens {getSortIcon('tokensConsumed')}
    </div>
                    </TableHead>
                    <TableHead
                      className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/50 text-slate-700 dark:text-slate-300 font-semibold transition-colors text-right"
                      onClick={() => handleSort('costIncurred')}
                    >
                      <div className="flex items-center gap-2 justify-end">
                        Cost {getSortIcon('costIncurred')}
    </div>
                    </TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold">Providers</TableHead>
                    <TableHead
                      className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-700/50 text-slate-700 dark:text-slate-300 font-semibold transition-colors"
                      onClick={() => handleSort('lastActive')}
                    >
                      <div className="flex items-center gap-2">
                        Last Active {getSortIcon('lastActive')}
    </div>
                    </TableHead>
                    <TableHead className="text-slate-700 dark:text-slate-300 font-semibold">Status</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {paginatedData.map((client) => (
                    <TableRow
                      key={client.id}
                      className="border-b border-slate-100 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
                    >
                      <TableCell className="font-semibold text-slate-900 dark:text-slate-100">{client.id}</TableCell>
                      <TableCell>
                        <code className="text-xs bg-slate-100 dark:bg-slate-700 px-2 py-1 rounded font-mono text-slate-700 dark:text-slate-300">
                          {maskApiKey(client.apiKey)}
                        </code>
                      </TableCell>
                      <TableCell className="text-right font-semibold text-slate-700 dark:text-slate-300">{client.totalQueries.toLocaleString()}</TableCell>
                      <TableCell className="text-right font-semibold text-slate-700 dark:text-slate-300">{client.tokensConsumed.toLocaleString()}</TableCell>
                      <TableCell className="text-right font-semibold text-slate-700 dark:text-slate-300">${client.costIncurred.toFixed(2)}</TableCell>
                      <TableCell>
                        <div className="flex gap-1 flex-wrap">
                          {client.providersUsed.map((provider) => (
                            <Badge key={provider} variant="outline" className="text-xs bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-300 border-slate-300 dark:border-slate-600">
                              {provider}
                            </Badge>
                          ))}
    </div>
                      </TableCell>
                      <TableCell className="text-sm text-slate-600 dark:text-slate-400">{client.lastActive}</TableCell>
                      <TableCell>
                        <Badge
                          variant={client.status === 'active' ? 'default' : 'secondary'}
                          className={client.status === 'active' ? 'bg-green-500 hover:bg-green-600 text-white' : 'bg-slate-400 hover:bg-slate-500 text-white'}
                        >
                          {client.status}
                        </Badge>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
            
            {/* Pagination */}
            {totalPages > 1 && (
              <div className="mt-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <Pagination>
                  <PaginationContent>
                    <PaginationItem>
                      <PaginationPrevious
                        onClick={() => setCurrentPage(prev => Math.max(1, prev - 1))}
                        className={currentPage === 1 ? 'pointer-events-none opacity-50' : 'cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors'}
                      />
                    </PaginationItem>

                    {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
                      <PaginationItem key={page}>
                        <PaginationLink
                          onClick={() => setCurrentPage(page)}
                          isActive={currentPage === page}
                          className="cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                        >
                          {page}
                        </PaginationLink>
                      </PaginationItem>
                    ))}

                    <PaginationItem>
                      <PaginationNext
                        onClick={() => setCurrentPage(prev => Math.min(totalPages, prev + 1))}
                        className={currentPage === totalPages ? 'pointer-events-none opacity-50' : 'cursor-pointer hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors'}
                      />
                    </PaginationItem>
                  </PaginationContent>
                </Pagination>
    </div>
  )}

            <div className="mt-4 flex items-center justify-between bg-slate-50 dark:bg-slate-900/50 rounded-lg p-4">
              <div className="text-sm text-slate-600 dark:text-slate-400">
                Showing <span className="font-semibold text-slate-900 dark:text-white">{paginatedData.length}</span> of <span className="font-semibold text-slate-900 dark:text-white">{filteredAndSortedData.length}</span> clients
                {searchTerm && ` (filtered from ${clientsData.length} total)`}
     </div>
              <div className="text-xs text-slate-500 dark:text-slate-500">
                Page <span className="font-semibold">{currentPage}</span> of <span className="font-semibold">{totalPages}</span>
     </div>
            </div>
          </CardContent>
        </Card>
      </div>
  )
}