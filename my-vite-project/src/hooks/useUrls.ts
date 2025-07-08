import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { urlApi, crawlApi, linksApi } from '../services/api'
import { toast } from 'react-hot-toast'

// Query keys for React Query
export const urlKeys = {
  all: ['urls'] as const,
  lists: () => [...urlKeys.all, 'list'] as const,
  list: (filters: Record<string, any>) => [...urlKeys.lists(), filters] as const,
  details: () => [...urlKeys.all, 'detail'] as const,
  detail: (id: number) => [...urlKeys.details(), id] as const,
}

export const crawlKeys = {
  all: ['crawls'] as const,
  status: (urlId: number) => [...crawlKeys.all, 'status', urlId] as const,
}

export const linkKeys = {
  all: ['links'] as const,
  urlLinks: (urlId: number, filters: Record<string, any> = {}) => [...linkKeys.all, 'url', urlId, filters] as const,
}

// Hook for fetching URLs with pagination and filters
export function useUrls(params?: {
  limit?: number
  offset?: number
  search?: string
  status?: string
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
}, options?: { enabled?: boolean, enablePolling?: boolean }) {
  const result = useQuery({
    queryKey: urlKeys.list(params || {}),
    queryFn: () => urlApi.getUrls(params),
    enabled: options?.enabled !== false,
    staleTime: 1000 * 30, // 30 seconds
    // Enable real-time polling if requested
    refetchInterval: options?.enablePolling ? 2000 : false,
    refetchIntervalInBackground: true,
  })

  return result
}

// Hook for fetching a single URL
export function useUrl(id: number, options?: { enabled?: boolean, enablePolling?: boolean }) {
  return useQuery({
    queryKey: urlKeys.detail(id),
    queryFn: () => urlApi.getUrl(id),
    enabled: !!id && (options?.enabled !== false),
    staleTime: 1000 * 60 * 5, // 5 minutes
    // Enable polling if requested
    refetchInterval: options?.enablePolling ? 3000 : false,
    refetchIntervalInBackground: true,
  })
}

// Hook for creating a new URL
export function useCreateUrl() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: urlApi.createUrl,
    onSuccess: (data) => {
      // Invalidate and refetch URLs list
      queryClient.invalidateQueries({ queryKey: urlKeys.lists() })
      toast.success('URL added successfully! Crawling will start shortly.')
      
      // Start polling for this new URL since it will likely start running soon
      if (data.data?.id) {
        queryClient.setQueryData(urlKeys.detail(data.data.id), data)
      }
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to add URL')
    },
  })
}

// Hook for deleting a URL
export function useDeleteUrl() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: urlApi.deleteUrl,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: urlKeys.lists() })
      toast.success('URL deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete URL')
    },
  })
}

// Hook for bulk deleting URLs
export function useBulkDeleteUrls() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: urlApi.bulkDeleteUrls,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: urlKeys.lists() })
      toast.success('URLs deleted successfully')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to delete URLs')
    },
  })
}

// Hook for starting a crawl
export function useStartCrawl() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: crawlApi.startCrawl,
    onSuccess: (_, urlId) => {
      // Invalidate related queries
      queryClient.invalidateQueries({ queryKey: urlKeys.detail(urlId) })
      queryClient.invalidateQueries({ queryKey: urlKeys.lists() })
      queryClient.invalidateQueries({ queryKey: crawlKeys.status(urlId) })
      toast.success('Crawl started successfully')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to start crawl')
    },
  })
}

// Hook for bulk rerunning crawls
export function useBulkRerunCrawls() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: crawlApi.bulkRerunCrawls,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: urlKeys.lists() })
      toast.success('Crawls restarted successfully')
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to restart crawls')
    },
  })
}

// Hook for getting crawl status
export function useCrawlStatus(urlId: number, options?: { enabled?: boolean, enablePolling?: boolean }) {
  return useQuery({
    queryKey: crawlKeys.status(urlId),
    queryFn: () => crawlApi.getCrawlStatus(urlId),
    enabled: !!urlId && (options?.enabled !== false),
    staleTime: 1000, // 1 second
    // Enable aggressive polling when requested
    refetchInterval: options?.enablePolling ? 1000 : false,
    refetchIntervalInBackground: true,
  })
}

// Hook for getting links for a URL
export function useUrlLinks(urlId: number, params?: {
  type?: 'all' | 'internal' | 'external' | 'broken' | 'accessible'
  limit?: number
  offset?: number
}, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: linkKeys.urlLinks(urlId, params || {}),
    queryFn: () => linksApi.getUrlLinks(urlId, params),
    enabled: !!urlId && (options?.enabled !== false),
    staleTime: 1000 * 60 * 10, // 10 minutes
  })
}

// Custom hook to check if any crawls are currently running
export function useHasRunningCrawls() {
  const { data: urlsResponse } = useUrls()
  const hasRunning = urlsResponse?.data?.some((url: any) => url.status === 'running') || false
  return hasRunning
}

// Custom hook to get real-time polling status
export function useRealTimeStatus() {
  const queryClient = useQueryClient()
  
  // Get all active queries and check if any are fetching
  const queries = queryClient.getQueryCache().getAll()
  
  // Check if any URL or crawl status queries are currently fetching
  const isPolling = queries.some(query => {
    const isRelevantQuery = query.queryKey[0] === 'urls' || query.queryKey[0] === 'crawls'
    const isFetching = query.state.fetchStatus === 'fetching'
    return isRelevantQuery && isFetching
  })
  
  return { isPolling }
} 