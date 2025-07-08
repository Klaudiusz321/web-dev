import { useState } from 'react'
import { 
  ExternalLink, 
  AlertTriangle, 
  XCircle, 
  ChevronDown,
  ChevronUp,
  Filter,
  Search,
  RefreshCw,
  Eye
} from 'lucide-react'
import { useUrlLinks } from '../hooks/useUrls'
import type { Link } from '../services/api'

interface BrokenLinksListProps {
  urlId: number
}

const statusIcons: { [key: number]: any } = {
  404: AlertTriangle,
  500: XCircle,
  503: XCircle,
}

const statusColors: { [key: number]: string } = {
  404: 'text-warning bg-warning/10 border-warning/20',
  500: 'text-error bg-error/10 border-error/20',
  503: 'text-error bg-error/10 border-error/20',
}

export default function BrokenLinksList({ urlId }: BrokenLinksListProps) {
  const [expanded, setExpanded] = useState(false)
  const [linkType, setLinkType] = useState<'broken' | 'all'>('broken')
  const [search, setSearch] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const pageSize = 20

  const { data: linksResponse, isLoading, error, refetch } = useUrlLinks(urlId, {
    type: linkType,
    limit: pageSize,
    offset: (currentPage - 1) * pageSize,
  })

  const links = linksResponse?.data || []
  const totalLinks = linksResponse?.pagination?.total || 0
  const totalPages = Math.ceil(totalLinks / pageSize)

  // Filter links based on search
  const filteredLinks = search 
    ? links.filter(link => 
        link.link_url.toLowerCase().includes(search.toLowerCase()) ||
        link.link_text.toLowerCase().includes(search.toLowerCase())
      )
    : links

  const getStatusIcon = (statusCode: number) => {
    const Icon = statusIcons[statusCode] || XCircle
    return <Icon className="h-4 w-4" />
  }

  const getStatusColor = (statusCode: number) => {
    return statusColors[statusCode] || 'text-text-tertiary bg-bg-tertiary/50 border-border'
  }

  const formatUrl = (url: string) => {
    if (url.length > 60) {
      return url.substring(0, 57) + '...'
    }
    return url
  }

  if (error) {
    return (
      <div className="bg-bg-secondary/30 backdrop-blur-sm border border-border rounded-2xl p-6">
        <div className="text-center text-error">
          <XCircle className="h-8 w-8 mx-auto mb-2" />
          <p>Failed to load links</p>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-bg-secondary/30 backdrop-blur-sm border border-border rounded-2xl">
      {/* Header */}
      <div className="p-6 border-b border-border/50">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <h3 className="text-lg font-semibold text-text-primary">
              Links Analysis
            </h3>
            {linkType === 'broken' && totalLinks > 0 && (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-xl text-xs font-medium bg-error/10 text-error border border-error/20">
                {totalLinks} Broken
              </span>
            )}
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => refetch()}
              disabled={isLoading}
              className="p-2 text-text-tertiary hover:text-text-primary hover:bg-bg-tertiary rounded-lg transition-colors disabled:opacity-50"
              title="Refresh"
            >
              <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
            <button
              onClick={() => setExpanded(!expanded)}
              className="flex items-center space-x-1 px-3 py-2 text-sm font-medium text-text-secondary bg-bg-tertiary/50 rounded-xl hover:bg-bg-tertiary transition-colors"
            >
              <span>{expanded ? 'Collapse' : 'Expand'}</span>
              {expanded ? (
                <ChevronUp className="h-4 w-4" />
              ) : (
                <ChevronDown className="h-4 w-4" />
              )}
            </button>
          </div>
        </div>

        {expanded && (
          <div className="mt-4 space-y-4">
            {/* Filters */}
            <div className="flex flex-col sm:flex-row gap-4">
              <div className="relative flex-1 max-w-sm">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-text-tertiary" />
                <input
                  type="text"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="Search broken links by URL or text..."
                  className="modern-input pl-10"
                />
              </div>

              <div className="relative">
                <Filter className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-text-tertiary" />
                <select
                  value={linkType}
                  onChange={(e) => {
                    setLinkType(e.target.value as 'broken' | 'all')
                    setCurrentPage(1)
                  }}
                  className="modern-input pl-10 pr-8 appearance-none"
                >
                  <option value="broken">Broken Links</option>
                  <option value="all">All Links</option>
                  <option value="internal">Internal Links</option>
                  <option value="external">External Links</option>
                  <option value="accessible">Accessible Links</option>
                </select>
              </div>
            </div>

            {/* Quick Stats */}
            {linkType === 'all' && (
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 p-4 bg-bg-tertiary/30 rounded-xl">
                <div className="text-center">
                  <div className="text-sm text-text-tertiary">Total</div>
                  <div className="text-lg font-semibold text-text-primary">{totalLinks}</div>
                </div>
                <div className="text-center">
                  <div className="text-sm text-text-tertiary">Internal</div>
                  <div className="text-lg font-semibold text-accent">
                    {links.filter(l => l.link_type === 'internal').length}
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-sm text-text-tertiary">External</div>
                  <div className="text-lg font-semibold text-success">
                    {links.filter(l => l.link_type === 'external').length}
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-sm text-text-tertiary">Broken</div>
                  <div className="text-lg font-semibold text-error">
                    {links.filter(l => !l.is_accessible).length}
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Content */}
      {expanded && (
        <div className="p-6">
          {isLoading ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-accent mx-auto mb-4"></div>
              <p className="text-text-secondary">Loading links...</p>
            </div>
          ) : filteredLinks.length === 0 ? (
            <div className="text-center py-8">
              <Eye className="h-12 w-12 text-text-tertiary mx-auto mb-4" />
              <h4 className="text-lg font-medium text-text-primary mb-2">
                {search ? 'No matching links found' : 
                 linkType === 'broken' ? 'No broken links found' : 'No links found'}
              </h4>
              <p className="text-text-secondary">
                {search ? 'Try adjusting your search criteria.' :
                 linkType === 'broken' ? 'All links are working properly!' : 
                 'No links were discovered during crawling.'}
              </p>
            </div>
          ) : (
            <>
              {/* Links List */}
              <div className="space-y-4">
                {filteredLinks.map((link: Link) => (
                  <div
                    key={link.id}
                    className={`border rounded-xl p-4 transition-colors ${
                      !link.is_accessible 
                        ? 'border-error/20 bg-error/5' 
                        : 'border-border bg-bg-tertiary/20 hover:bg-bg-tertiary/40'
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center flex-wrap gap-2 mb-2">
                          <span className={`inline-flex items-center px-2 py-1 rounded-lg text-xs font-medium border ${
                            link.link_type === 'internal' 
                              ? 'bg-accent/10 text-accent border-accent/20' 
                              : 'bg-success/10 text-success border-success/20'
                          }`}>
                            {link.link_type}
                          </span>
                          {!link.is_accessible && (
                            <span className={`inline-flex items-center px-2 py-1 rounded-lg text-xs font-medium border ${getStatusColor(link.status_code)}`}>
                              {getStatusIcon(link.status_code)}
                              <span className="ml-1">{link.status_code}</span>
                            </span>
                          )}
                        </div>
                        
                        <div className="mb-2">
                          <a
                            href={link.link_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-sm font-medium text-accent hover:text-accent-secondary flex items-center group"
                            title={link.link_url}
                          >
                            <span className="truncate">{formatUrl(link.link_url)}</span>
                            <ExternalLink className="h-3 w-3 ml-1 flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity" />
                          </a>
                        </div>

                        {link.link_text && (
                          <div className="text-sm text-text-secondary">
                            <span className="font-medium">Link text:</span> {link.link_text}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Pagination */}
              {totalPages > 1 && (
                <div className="mt-6 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                  <div className="text-sm text-text-secondary">
                    Showing {(currentPage - 1) * pageSize + 1} to{' '}
                    {Math.min(currentPage * pageSize, totalLinks)} of {totalLinks} links
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                      disabled={currentPage === 1}
                      className="btn-ghost disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Previous
                    </button>
                    <button
                      onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                      disabled={currentPage === totalPages}
                      className="btn-ghost disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Next
                    </button>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
} 