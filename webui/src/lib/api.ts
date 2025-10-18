import { useAuth } from '@/contexts/auth-context'

const API_BASE = import.meta.env.DEV ? 'http://localhost:2906' : ''

export class ApiClient {
  private token: string | null = null

  constructor(token?: string) {
    this.token = token || null
  }

  setToken(token: string) {
    this.token = token
  }

  private async request(endpoint: string, options: RequestInit = {}): Promise<any> {
    const url = `${API_BASE}${endpoint}`
    const headers = new Headers(options.headers)
    headers.set('Content-Type', 'application/json')

    if (this.token) {
      headers.set('Authorization', `Bearer ${this.token}`)
    }

    const response = await fetch(url, {
      ...options,
      headers,
    })

    if (!response.ok) {
      throw new Error(`API request failed: ${response.status} ${response.statusText}`)
    }

    return response.json()
  }

  // Admin API methods
  async getMetrics(name: string, filters: Record<string, string> = {}, start?: number, end?: number) {
    const params = new URLSearchParams({ name, ...filters })
    if (start) params.set('start', start.toString())
    if (end) params.set('end', end.toString())
    return this.request(`/api/admin/v1/metrics?${params}`)
  }

  async getClientStats(start?: number, end?: number) {
    const params = new URLSearchParams()
    if (start) params.set('start', start.toString())
    if (end) params.set('end', end.toString())
    return this.request(`/api/admin/v1/clients?${params}`)
  }

  async getStats(groupBy: string[], filters: Record<string, string> = {}, start?: number, end?: number) {
    const params = new URLSearchParams({ group_by: groupBy.join(',') })
    Object.entries(filters).forEach(([key, value]) => params.set(key, value))
    if (start) params.set('start', start.toString())
    if (end) params.set('end', end.toString())
    return this.request(`/api/admin/v1/stats?${params}`)
  }

  async getConfig() {
    return this.request('/api/admin/v1/config')
  }

  async validateConfig(config: any) {
    return this.request('/api/admin/v1/config/validate', {
      method: 'POST',
      body: JSON.stringify(config),
    })
  }

  async updatePolicy(policy: { algorithm: string; priority: string; cache?: { enabled: boolean; ttl_seconds: number } }) {
    return this.request('/api/admin/v1/config/policy', {
      method: 'PUT',
      body: JSON.stringify(policy),
    })
  }

  // Client management
  async createClient(clientData: { client_id: string; api_key: string; description?: string; allowed_providers?: string[] }) {
    return this.request('/api/admin/v1/clients', {
      method: 'POST',
      body: JSON.stringify(clientData),
    })
  }

  async listClients() {
    return this.request('/api/admin/v1/clients/list')
  }

  async getClient(clientId: string) {
    return this.request(`/api/admin/v1/clients/${clientId}`)
  }

  async updateClient(clientId: string, updates: { description?: string; allowed_providers?: string[] }) {
    return this.request(`/api/admin/v1/clients/${clientId}`, {
      method: 'PUT',
      body: JSON.stringify(updates),
    })
  }

  async deleteClient(clientId: string) {
    return this.request(`/api/admin/v1/clients/${clientId}`, {
      method: 'DELETE',
    })
  }

  // Enhanced metrics
  async getClientMetrics(clientId: string, start?: number, end?: number) {
    const params = new URLSearchParams()
    if (start) params.set('start', start.toString())
    if (end) params.set('end', end.toString())
    return this.request(`/api/admin/v1/metrics/clients/${clientId}?${params}`)
  }

  async getProviderMetrics(providerId: string, start?: number, end?: number) {
    const params = new URLSearchParams()
    if (start) params.set('start', start.toString())
    if (end) params.set('end', end.toString())
    return this.request(`/api/admin/v1/metrics/providers/${providerId}?${params}`)
  }

  async getGlobalMetrics(start?: number, end?: number) {
    const params = new URLSearchParams()
    if (start) params.set('start', start.toString())
    if (end) params.set('end', end.toString())
    return this.request(`/api/admin/v1/metrics/global?${params}`)
  }
}

// Hook to get authenticated API client
export function useApi() {
  const { token } = useAuth()
  return new ApiClient(token || undefined)
}

// Global API client instance
export const apiClient = new ApiClient()
