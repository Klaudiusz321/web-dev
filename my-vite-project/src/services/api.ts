// API base configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

// Types for API responses
export interface URL {
  id: number
  url: string
  title: string
  html_version: string
  status: 'pending' | 'running' | 'completed' | 'error'
  has_login_form: boolean
  created_at: string
  updated_at: string
  crawls?: Crawl[]
  links?: Link[]
}

export interface Crawl {
  id: number
  url_id: number
  status: 'queued' | 'running' | 'completed' | 'error'
  started_at?: string
  completed_at?: string
  error_message: string
  internal_links: number
  external_links: number
  broken_links: number
  heading_counts: string
  created_at: string
  updated_at: string
}

export interface Link {
  id: number
  url_id: number
  crawl_id: number
  link_url: string
  link_text: string
  link_type: 'internal' | 'external'
  status_code: number
  is_accessible: boolean
  created_at: string
}

export interface CrawlStatusResponse {
  id: number
  url: string
  status: string
  internal_links: number
  external_links: number
  broken_links: number
  heading_counts: {
    h1: number
    h2: number
    h3: number
    h4: number
    h5: number
    h6: number
  }
  started_at?: string
  completed_at?: string
  error_message?: string
}

export interface PaginationResponse<T> {
  data: T[]
  pagination: {
    total: number
    limit: number
    offset: number
  }
}

export interface ApiResponse<T> {
  data?: T
  message?: string
  error?: string
}

// Helper function to get auth token
function getAuthToken(): string | null {
  return localStorage.getItem('token')
}

// Helper function for API calls
async function apiCall<T>(
  endpoint: string, 
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  const token = getAuthToken()
  
  const defaultOptions: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` }),
      ...options.headers,
    },
    ...options,
  }

  try {
    const response = await fetch(url, defaultOptions)
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
    }

    return await response.json()
  } catch (error) {
    console.error(`API call failed: ${endpoint}`, error)
    throw error
  }
}

// URL CRUD Operations
export const urlApi = {
  // Get all URLs with pagination and filters
  getUrls: async (params?: {
    limit?: number
    offset?: number
    search?: string
    status?: string
    sortBy?: string
    sortOrder?: 'asc' | 'desc'
  }): Promise<PaginationResponse<URL>> => {
    const searchParams = new URLSearchParams()
    
    if (params?.limit) searchParams.append('limit', params.limit.toString())
    if (params?.offset) searchParams.append('offset', params.offset.toString())
    if (params?.search) searchParams.append('search', params.search)
    if (params?.status) searchParams.append('status', params.status)
    if (params?.sortBy) searchParams.append('sortBy', params.sortBy)
    if (params?.sortOrder) searchParams.append('sortOrder', params.sortOrder)

    const query = searchParams.toString()
    return apiCall<PaginationResponse<URL>>(`/urls${query ? `?${query}` : ''}`)
  },

  // Get single URL by ID
  getUrl: async (id: number): Promise<ApiResponse<URL>> => {
    return apiCall<ApiResponse<URL>>(`/urls/${id}`)
  },

  // Create new URL
  createUrl: async (url: string): Promise<ApiResponse<URL>> => {
    return apiCall<ApiResponse<URL>>('/urls', {
      method: 'POST',
      body: JSON.stringify({ url }),
    })
  },

  // Delete URL
  deleteUrl: async (id: number): Promise<ApiResponse<void>> => {
    return apiCall<ApiResponse<void>>(`/urls/${id}`, {
      method: 'DELETE',
    })
  },

  // Bulk delete URLs
  bulkDeleteUrls: async (ids: number[]): Promise<ApiResponse<void>> => {
    return apiCall<ApiResponse<void>>('/urls/bulk-delete', {
      method: 'POST',
      body: JSON.stringify({ ids }),
    })
  },
}

// Crawl Operations
export const crawlApi = {
  // Start crawling for a URL
  startCrawl: async (urlId: number): Promise<ApiResponse<void>> => {
    return apiCall<ApiResponse<void>>(`/crawl/${urlId}`, {
      method: 'POST',
    })
  },

  // Get crawl status for a URL
  getCrawlStatus: async (urlId: number): Promise<ApiResponse<CrawlStatusResponse>> => {
    return apiCall<ApiResponse<CrawlStatusResponse>>(`/crawl/status/${urlId}`)
  },

  // Bulk rerun crawls
  bulkRerunCrawls: async (urlIds: number[]): Promise<ApiResponse<void>> => {
    return apiCall<ApiResponse<void>>('/crawl/bulk-rerun', {
      method: 'POST',
      body: JSON.stringify({ ids: urlIds }),
    })
  },
}

// Health check
export const healthApi = {
  check: async (): Promise<{ status: string }> => {
    return apiCall<{ status: string }>('/health')
  },
}

// Links Operations
export const linksApi = {
  // Get links for a specific URL
  getUrlLinks: async (urlId: number, params?: {
    type?: 'all' | 'internal' | 'external' | 'broken' | 'accessible'
    limit?: number
    offset?: number
  }): Promise<PaginationResponse<Link>> => {
    const searchParams = new URLSearchParams()
    
    if (params?.type) searchParams.append('type', params.type)
    if (params?.limit) searchParams.append('limit', params.limit.toString())
    if (params?.offset) searchParams.append('offset', params.offset.toString())

    const query = searchParams.toString()
    return apiCall<PaginationResponse<Link>>(`/urls/${urlId}/links${query ? `?${query}` : ''}`)
  },
} 