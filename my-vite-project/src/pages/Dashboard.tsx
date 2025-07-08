import { useState } from 'react'
import { Link } from 'react-router-dom'
import { 
  Search, 
  Filter, 
  Plus, 
  RefreshCw, 
  Trash2, 
  Eye,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  Globe,
  Clock,
  CheckCircle,
  XCircle,
  Play,
  Activity,
  TrendingUp,
  AlertTriangle,
  Calendar,
  Link as LinkIcon,
  Target,
  Zap
} from 'lucide-react'
import { useUrls, useDeleteUrl, useBulkDeleteUrls, useStartCrawl, useBulkRerunCrawls, useHasRunningCrawls } from '../hooks/useUrls'
import RealTimeIndicator from '../components/RealTimeIndicator'
import type { URL } from '../services/api'

const statusConfig = {
  pending: {
    icon: Clock,
    color: 'text-warning',
    bgColor: 'bg-warning/10',
    borderColor: 'border-warning/20',
    label: 'Pending'
  },
  running: {
    icon: RefreshCw,
    color: 'text-accent',
    bgColor: 'bg-accent/10', 
    borderColor: 'border-accent/20',
    label: 'Running'
  },
  completed: {
    icon: CheckCircle,
    color: 'text-success',
    bgColor: 'bg-success/10',
    borderColor: 'border-success/20',
    label: 'Completed'
  },
  error: {
    icon: XCircle,
    color: 'text-error',
    bgColor: 'bg-error/10',
    borderColor: 'border-error/20',
    label: 'Error'
  },
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

interface StatsCardProps {
  title: string
  value: number
  icon: React.ComponentType<{ className?: string }>
  color: string
  change?: { value: number; label: string }
}

function StatsCard({ title, value, icon: Icon, color, change }: StatsCardProps) {
  return (
    <div className="bg-bg-secondary/50 backdrop-blur-sm border border-border rounded-2xl p-6 card-hover">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-text-tertiary text-sm font-medium">{title}</p>
          <p className="text-text-primary text-3xl font-bold mt-2">{value.toLocaleString()}</p>
          {change && (
            <p className={`text-sm mt-2 flex items-center ${
              change.value >= 0 ? 'text-success' : 'text-error'
            }`}>
              <TrendingUp className={`w-4 h-4 mr-1 ${change.value < 0 ? 'rotate-180' : ''}`} />
              {change.value >= 0 ? '+' : ''}{change.value}% {change.label}
            </p>
          )}
        </div>
        <div className={`w-12 h-12 rounded-xl ${color} flex items-center justify-center`}>
          <Icon className="w-6 h-6 text-white" />
        </div>
      </div>
    </div>
  )
}

interface URLCardProps {
  url: URL
  isSelected: boolean
  onSelect: () => void
  onDelete: () => void
  onRerun: () => void
}

function URLCard({ url, isSelected, onSelect, onDelete, onRerun }: URLCardProps) {
  const statusInfo = statusConfig[url.status] || {
    icon: Clock,
    color: 'text-text-tertiary',
    bgColor: 'bg-bg-tertiary/10',
    borderColor: 'border-border',
    label: url.status || 'Unknown'
  }
  const StatusIcon = statusInfo.icon

  return (
    <div className={`
      bg-bg-secondary/30 backdrop-blur-sm border rounded-2xl p-6 transition-all duration-300
      ${isSelected 
        ? 'border-accent shadow-lg ring-2 ring-accent/20' 
        : 'border-border hover:border-accent/30 hover:shadow-elegant'
      }
    `}>
      <div className="flex items-start justify-between">
        <div className="flex items-start space-x-3 flex-1">
          <input
            type="checkbox"
            checked={isSelected}
            onChange={onSelect}
            className="mt-1 w-4 h-4 text-accent bg-bg-primary border-border rounded focus:ring-accent/50"
          />
          
          <div className="flex-1 min-w-0">
            {/* URL and Title */}
            <div className="flex items-center space-x-2 mb-3">
              <LinkIcon className="w-4 h-4 text-text-tertiary flex-shrink-0" />
              <a 
                href={url.url} 
                target="_blank" 
                rel="noopener noreferrer"
                className="text-accent hover:text-accent-secondary font-medium truncate flex items-center group"
                title={url.url}
              >
                <span className="truncate">
                  {url.url.length > 60 ? `${url.url.substring(0, 57)}...` : url.url}
                </span>
                <ExternalLink className="w-3 h-3 ml-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" />
              </a>
            </div>

            {url.title && (
              <h3 className="text-text-primary font-semibold text-lg mb-3 line-clamp-2">
                {url.title}
              </h3>
            )}

            {/* Status and Meta Info */}
            <div className="flex flex-wrap items-center gap-3 mb-4">
              <div className={`
                inline-flex items-center px-3 py-1.5 rounded-xl text-sm font-medium flex-shrink-0
                ${statusInfo.bgColor} ${statusInfo.borderColor} ${statusInfo.color} border
              `}>
                <StatusIcon className={`w-4 h-4 mr-2 ${url.status === 'running' ? 'animate-spin' : ''}`} />
                {statusInfo.label}
              </div>
              
              {url.html_version && (
                <div className="flex items-center text-text-tertiary text-sm flex-shrink-0">
                  <Globe className="w-4 h-4 mr-1" />
                  HTML {url.html_version}
                </div>
              )}
            </div>

            {/* Timestamps */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-sm text-text-tertiary">
              <div className="flex items-start space-x-2">
                <Calendar className="w-4 h-4 text-text-tertiary flex-shrink-0 mt-0.5" />
                <div className="min-w-0 flex-1">
                  <span className="block font-medium text-text-secondary">Created</span>
                  <span className="text-text-tertiary truncate block">{formatDate(url.created_at)}</span>
                </div>
              </div>
              <div className="flex items-start space-x-2">
                <Clock className="w-4 h-4 text-text-tertiary flex-shrink-0 mt-0.5" />
                <div className="min-w-0 flex-1">
                  <span className="block font-medium text-text-secondary">Updated</span>
                  <span className="text-text-tertiary truncate block">{formatDate(url.updated_at)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center space-x-2 ml-4">
          <Link
            to={`/url/${url.id}`}
            className="p-2 text-text-tertiary hover:text-text-primary hover:bg-bg-tertiary rounded-lg transition-colors"
            title="View Details"
          >
            <Eye className="w-4 h-4" />
          </Link>
          
          <button
            onClick={onRerun}
            className="p-2 text-text-tertiary hover:text-success hover:bg-success/10 rounded-lg transition-colors"
            title="Start New Crawl"
          >
            <Play className="w-4 h-4" />
          </button>
          
          <button
            onClick={onDelete}
            className="p-2 text-text-tertiary hover:text-error hover:bg-error/10 rounded-lg transition-colors"
            title="Delete URL"
          >
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  )
}

export default function Dashboard() {
  const [currentPage, setCurrentPage] = useState(1)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [selectedUrls, setSelectedUrls] = useState<number[]>([])
  
  const pageSize = 12 // Increased for card layout
  const hasRunningCrawls = useHasRunningCrawls()

  // Fetch URLs with polling enabled if there are running crawls
  const { data: urlsResponse, isLoading, error, refetch } = useUrls({
    limit: pageSize,
    offset: (currentPage - 1) * pageSize,
    search: search || undefined,
    status: statusFilter || undefined,
    sortBy: 'updated_at',
    sortOrder: 'desc',
  }, {
    enablePolling: hasRunningCrawls
  })

  const deleteUrlMutation = useDeleteUrl()
  const bulkDeleteMutation = useBulkDeleteUrls()
  const startCrawlMutation = useStartCrawl()
  const bulkRerunMutation = useBulkRerunCrawls()

  const urls = urlsResponse?.data || []
  const totalUrls = urlsResponse?.pagination?.total || 0
  const totalPages = Math.ceil(totalUrls / pageSize)

  // Calculate stats
  const stats = {
    total: totalUrls,
    completed: urls.filter(url => url.status === 'completed').length,
    running: urls.filter(url => url.status === 'running').length,
    errors: urls.filter(url => url.status === 'error').length,
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setCurrentPage(1)
  }

  const handleSelectUrl = (id: number) => {
    setSelectedUrls(prev => 
      prev.includes(id) 
        ? prev.filter(urlId => urlId !== id)
        : [...prev, id]
    )
  }

  const handleSelectAll = () => {
    if (selectedUrls.length === urls.length) {
      setSelectedUrls([])
    } else {
      setSelectedUrls(urls.map(url => url.id))
    }
  }

  const handleBulkDelete = async () => {
    if (selectedUrls.length === 0) return
    
    if (confirm(`Are you sure you want to delete ${selectedUrls.length} URL${selectedUrls.length !== 1 ? 's' : ''}? This action cannot be undone.`)) {
      await bulkDeleteMutation.mutateAsync(selectedUrls)
      setSelectedUrls([])
    }
  }

  const handleBulkRerun = async () => {
    if (selectedUrls.length === 0) return
    
    await bulkRerunMutation.mutateAsync(selectedUrls)
    setSelectedUrls([])
  }

  const clearFilters = () => {
    setSearch('')
    setStatusFilter('')
    setCurrentPage(1)
  }

  if (error) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center max-w-md">
          <div className="w-16 h-16 bg-error/10 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <XCircle className="w-8 h-8 text-error" />
          </div>
          <h2 className="text-xl font-semibold text-text-primary mb-3">Unable to Load URLs</h2>
          <p className="text-text-secondary mb-6">There was an error loading your URLs. Please check your connection and try again.</p>
          <button
            onClick={() => refetch()}
            className="btn-primary"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Try Again
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="p-8 max-w-7xl mx-auto space-y-8">
        {/* Header */}
        <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between">
          <div className="flex items-center space-x-4 mb-6 lg:mb-0">
            <div className="w-12 h-12 bg-gradient-to-br from-accent to-accent-secondary rounded-2xl flex items-center justify-center">
              <Activity className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-text-primary">Dashboard</h1>
              <p className="text-text-secondary">Monitor and manage your web crawling operations</p>
            </div>
            <RealTimeIndicator />
          </div>
          
          <Link
            to="/add"
            className="btn-primary inline-flex items-center"
          >
            <Plus className="w-4 h-4 mr-2" />
            Add URL
          </Link>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6">
          <StatsCard
            title="Total URLs"
            value={stats.total}
            icon={Target}
            color="bg-gradient-to-br from-accent to-accent-secondary"
          />
          <StatsCard
            title="Completed"
            value={stats.completed}
            icon={CheckCircle}
            color="bg-gradient-to-br from-success to-green-600"
          />
          <StatsCard
            title="Running"
            value={stats.running}
            icon={Zap}
            color="bg-gradient-to-br from-warning to-orange-600"
          />
          <StatsCard
            title="Errors"
            value={stats.errors}
            icon={AlertTriangle}
            color="bg-gradient-to-br from-error to-red-600"
          />
        </div>

        {/* Search and Filters */}
        <div className="bg-bg-secondary backdrop-blur-sm border border-border rounded-2xl p-6">
          <div className="flex flex-col lg:flex-row gap-4 lg:gap-6">
            {/* Search */}
            <div className="flex-1 max-w-md">
              <form onSubmit={handleSearch}>
                <div className="relative">
                  <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 w-5 h-5 text-position-left" />
                  <input
                    type="text"
                    placeholder="Search..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                    className="modern-input pl-16 w-full"
                  />
                </div>
              </form>
            </div>

            {/* Filters and Actions */}
            <div className="flex flex-wrap items-center gap-3">
              <div className="relative min-w-0">
                <Filter className="absolute left-4 top-1/2 transform -translate-y-1/2 w-4 h-4 text-text-tertiary" />
                <select
                  value={statusFilter}
                  onChange={(e) => {
                    setStatusFilter(e.target.value)
                    setCurrentPage(1)
                  }}
                  className="modern-input pl-14 pr-10 min-w-[140px]"
                >
                  <option value="">All Statuses</option>
                  <option value="pending">Pending</option>
                  <option value="running">Running</option>
                  <option value="completed">Completed</option>
                  <option value="error">Error</option>
                </select>
              </div>

              <button
                onClick={clearFilters}
                className="btn-ghost flex-shrink-0"
              >
                Clear All Filters
              </button>

              <button
                onClick={() => refetch()}
                disabled={isLoading}
                className={`p-3 rounded-xl border border-border hover:bg-bg-tertiary transition-colors flex-shrink-0
                  ${isLoading ? 'opacity-50 cursor-not-allowed' : 'hover:border-accent/30'}
                `}
                title="Refresh Data"
              >
                <RefreshCw className={`w-4 h-4 text-text-secondary ${isLoading ? 'animate-spin' : ''}`} />
              </button>
            </div>
          </div>

          {/* Bulk Actions */}
          {selectedUrls.length > 0 && (
            <div className="mt-6 p-4 bg-accent/10 border border-accent/20 rounded-xl">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                <div className="flex items-center space-x-4">
                  <div className="w-8 h-8 bg-accent rounded-lg flex items-center justify-center flex-shrink-0">
                    <CheckCircle className="w-4 h-4 text-white" />
                  </div>
                  <div className="min-w-0">
                    <p className="font-medium text-text-primary">
                      {selectedUrls.length} URL{selectedUrls.length !== 1 ? 's' : ''} selected
                    </p>
                    <p className="text-sm text-text-secondary">
                      Choose an action to apply to the selected items
                    </p>
                  </div>
                </div>
                
                <div className="flex gap-3 flex-shrink-0">
                  <button
                    onClick={handleBulkRerun}
                    disabled={bulkRerunMutation.isPending}
                    className="inline-flex items-center px-4 py-2 bg-success hover:bg-green-600 text-white font-medium rounded-xl transition-colors disabled:opacity-50"
                  >
                    <Play className="w-4 h-4 mr-2" />
                    Re-crawl Selected
                  </button>
                  <button
                    onClick={handleBulkDelete}
                    disabled={bulkDeleteMutation.isPending}
                    className="inline-flex items-center px-4 py-2 bg-error hover:bg-red-600 text-white font-medium rounded-xl transition-colors disabled:opacity-50"
                  >
                    <Trash2 className="w-4 h-4 mr-2" />
                    Delete Selected
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* URLs Grid */}
        {isLoading ? (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {[...Array(6)].map((_, i) => (
              <div key={i} className="bg-bg-secondary/30 border border-border rounded-2xl p-6 animate-pulse">
                <div className="space-y-4">
                  <div className="h-4 bg-bg-tertiary rounded-lg w-3/4"></div>
                  <div className="h-6 bg-bg-tertiary rounded-lg w-1/2"></div>
                  <div className="flex space-x-4">
                    <div className="h-8 bg-bg-tertiary rounded-xl w-20"></div>
                    <div className="h-8 bg-bg-tertiary rounded-xl w-24"></div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : urls.length === 0 ? (
          <div className="text-center py-16">
            <div className="w-20 h-20 bg-bg-secondary rounded-2xl flex items-center justify-center mx-auto mb-6">
              <Globe className="w-10 h-10 text-text-tertiary" />
            </div>
            <h3 className="text-xl font-semibold text-text-primary mb-3">No URLs Found</h3>
            <p className="text-text-secondary mb-8 max-w-md mx-auto">
              {search || statusFilter ? 'No URLs match your current search criteria. Try adjusting your filters to find what you\'re looking for.' : 'Get started by adding your first URL to crawl and analyze.'}
            </p>
            {!search && !statusFilter && (
              <Link
                to="/add"
                className="btn-primary inline-flex items-center"
              >
                <Plus className="w-4 h-4 mr-2" />
                Add Your First URL
              </Link>
            )}
          </div>
        ) : (
          <>
            {/* Select All Option */}
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-0">
              <div className="flex items-center space-x-3">
                <input
                  type="checkbox"
                  checked={selectedUrls.length === urls.length && urls.length > 0}
                  onChange={handleSelectAll}
                  className="w-4 h-4 text-accent bg-bg-primary border-border rounded focus:ring-accent/50 flex-shrink-0"
                />
                <span className="text-text-secondary text-sm">
                  Select all {urls.length} URLs on this page
                </span>
              </div>
              
              <div className="text-text-tertiary text-sm">
                Showing {((currentPage - 1) * pageSize) + 1}–{Math.min(currentPage * pageSize, totalUrls)} of {totalUrls} URLs
              </div>
            </div>

            {/* URL Cards Grid */}
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
              {urls.map((url) => (
                <URLCard
                  key={url.id}
                  url={url}
                  isSelected={selectedUrls.includes(url.id)}
                  onSelect={() => handleSelectUrl(url.id)}
                  onDelete={() => deleteUrlMutation.mutate(url.id)}
                  onRerun={() => startCrawlMutation.mutate(url.id)}
                />
              ))}
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div className="flex items-center justify-center space-x-2">
                <button
                  onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                  disabled={currentPage === 1}
                  className="p-2 rounded-xl border border-border disabled:opacity-50 disabled:cursor-not-allowed hover:bg-bg-tertiary transition-colors"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                
                <div className="flex space-x-1">
                  {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                    let pageNum: number
                    if (totalPages <= 7) {
                      pageNum = i + 1
                    } else if (currentPage <= 4) {
                      pageNum = i + 1
                    } else if (currentPage >= totalPages - 3) {
                      pageNum = totalPages - 6 + i
                    } else {
                      pageNum = currentPage - 3 + i
                    }
                    
                    return (
                      <button
                        key={pageNum}
                        onClick={() => setCurrentPage(pageNum)}
                        className={`w-10 h-10 rounded-xl font-medium transition-colors ${
                          currentPage === pageNum
                            ? 'bg-accent text-white'
                            : 'hover:bg-bg-tertiary text-text-secondary'
                        }`}
                      >
                        {pageNum}
                      </button>
                    )
                  })}
                </div>
                
                <button
                  onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                  disabled={currentPage === totalPages}
                  className="p-2 rounded-xl border border-border disabled:opacity-50 disabled:cursor-not-allowed hover:bg-bg-tertiary transition-colors"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
} 