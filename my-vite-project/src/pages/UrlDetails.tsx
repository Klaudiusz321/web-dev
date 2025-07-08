import { useParams, Link, useNavigate } from 'react-router-dom'
import { 
  ArrowLeft, 
  ExternalLink, 
  RefreshCw, 
  Trash2, 
  Globe,
  CheckCircle,
  Clock,
  XCircle,
  AlertTriangle,
  BarChart3,
  LinkIcon,
  Activity,
  TrendingUp,
  Shield,
  FileText,
  Zap,
  Monitor,
  Calendar,
  Timer,
  Award,
  Copy,
  Download,
  Hash
} from 'lucide-react'
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts'
import { useUrl, useDeleteUrl, useStartCrawl, useCrawlStatus } from '../hooks/useUrls'
import BrokenLinksList from '../components/BrokenLinksList'
import { RealTimeIndicatorMini } from '../components/RealTimeIndicator'
import toast from 'react-hot-toast'

// Modern color palette for charts
const CHART_COLORS = {
  primary: 'rgb(var(--accent))',
  success: 'rgb(var(--success))',
  warning: 'rgb(var(--warning))',
  error: 'rgb(var(--error))',
  gradient: ['#4F46E5', '#06B6D4', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6']
}

const statusConfig = {
  pending: {
    icon: Clock,
    color: 'text-warning',
    bgColor: 'bg-warning/10',
    borderColor: 'border-warning/20',
    gradient: 'from-warning to-orange-500',
    label: 'Pending'
  },
  running: {
    icon: RefreshCw,
    color: 'text-accent',
    bgColor: 'bg-accent/10', 
    borderColor: 'border-accent/20',
    gradient: 'from-accent to-blue-500',
    label: 'Running'
  },
  completed: {
    icon: CheckCircle,
    color: 'text-success',
    bgColor: 'bg-success/10',
    borderColor: 'border-success/20',
    gradient: 'from-success to-green-500',
    label: 'Completed'
  },
  error: {
    icon: XCircle,
    color: 'text-error',
    bgColor: 'bg-error/10',
    borderColor: 'border-error/20',
    gradient: 'from-error to-red-500',
    label: 'Error'
  },
}

interface StatCardProps {
  title: string
  value: string | number
  subtitle?: string
  icon: React.ComponentType<{ className?: string }>
  gradient: string
  trend?: { value: number; isPositive: boolean }
}

function StatCard({ title, value, subtitle, icon: Icon, gradient, trend }: StatCardProps) {
  return (
    <div className="bg-bg-secondary/30 backdrop-blur-sm border border-border rounded-2xl p-6 card-hover">
      <div className="flex items-center justify-between mb-4">
        <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${gradient} flex items-center justify-center`}>
          <Icon className="w-6 h-6 text-white" />
        </div>
        {trend && (
          <div className={`flex items-center space-x-1 text-sm ${
            trend.isPositive ? 'text-success' : 'text-error'
          }`}>
            <TrendingUp className={`w-4 h-4 ${trend.isPositive ? '' : 'rotate-180'}`} />
            <span>{Math.abs(trend.value)}%</span>
          </div>
        )}
      </div>
      
      <div>
        <p className="text-text-tertiary text-sm font-medium mb-1">{title}</p>
        <p className="text-text-primary text-2xl font-bold">{value}</p>
        {subtitle && (
          <p className="text-text-secondary text-sm mt-1">{subtitle}</p>
        )}
      </div>
    </div>
  )
}

interface ChartCardProps {
  title: string
  icon: React.ComponentType<{ className?: string }>
  children: React.ReactNode
  actions?: React.ReactNode
}

function ChartCard({ title, icon: Icon, children, actions }: ChartCardProps) {
  return (
    <div className="bg-bg-secondary/30 backdrop-blur-sm border border-border rounded-2xl p-6">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <div className="w-8 h-8 bg-accent/10 rounded-lg flex items-center justify-center">
            <Icon className="w-4 h-4 text-accent" />
          </div>
          <h3 className="text-lg font-semibold text-text-primary">{title}</h3>
        </div>
        {actions}
      </div>
      {children}
    </div>
  )
}

export default function UrlDetails() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const urlId = parseInt(id || '0', 10)

  // Enable polling for URL data if status is running
  const { data: urlResponse, isLoading: urlLoading, error: urlError } = useUrl(urlId, {
    enablePolling: true // Always enable polling to catch status changes
  })
  
  // Enable polling for crawl status if status is running or queued
  const { data: crawlStatusResponse } = useCrawlStatus(urlId, {
    enablePolling: true
  })
  
  const deleteUrlMutation = useDeleteUrl()
  const startCrawlMutation = useStartCrawl()

  const url = urlResponse?.data
  const crawlStatus = crawlStatusResponse?.data

  // Check if we should show real-time indicators
  const isRunning = url?.status === 'running' || crawlStatus?.status === 'running'
  const isQueued = url?.status === 'pending' || crawlStatus?.status === 'queued'

  const handleDelete = async () => {
    if (confirm('Are you sure you want to delete this URL?')) {
      try {
        await deleteUrlMutation.mutateAsync(urlId)
        toast.success('URL deleted successfully')
        navigate('/')
      } catch (error) {
        toast.error('Failed to delete URL')
      }
    }
  }

  const handleStartCrawl = async () => {
    try {
      await startCrawlMutation.mutateAsync(urlId)
      toast.success('Crawl started successfully')
    } catch (error) {
      toast.error('Failed to start crawl')
    }
  }

  const copyUrl = () => {
    if (url?.url) {
      navigator.clipboard.writeText(url.url)
      toast.success('URL copied to clipboard')
    }
  }

  if (urlLoading) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-accent border-t-transparent rounded-full animate-spin mx-auto mb-6"></div>
          <h2 className="text-xl font-semibold text-text-primary mb-2">Loading URL details...</h2>
          <p className="text-text-secondary">Please wait while we fetch the information</p>
        </div>
      </div>
    )
  }

  if (urlError || !url) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center max-w-md">
          <div className="w-16 h-16 bg-error/10 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <AlertTriangle className="w-8 h-8 text-error" />
          </div>
          <h2 className="text-xl font-semibold text-text-primary mb-3">URL Not Found</h2>
          <p className="text-text-secondary mb-6">The requested URL could not be found or has been deleted.</p>
          <Link
            to="/"
            className="btn-primary inline-flex items-center"
          >
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back to Dashboard
          </Link>
        </div>
      </div>
    )
  }

  const statusInfo = statusConfig[url.status]
  const StatusIcon = statusInfo.icon

  // Parse heading counts
  let headingCounts: { [key: string]: number } = {}
  let headingData: Array<{ name: string; count: number }> = []
  
  if (crawlStatus?.heading_counts) {
    headingCounts = crawlStatus.heading_counts
    headingData = Object.entries(headingCounts)
      .filter(([_, count]) => count > 0)
      .map(([tag, count]) => ({ name: tag.toUpperCase(), count }))
  }

  // Link distribution data for pie chart
  const linkData = crawlStatus ? [
    { name: 'Internal Links', value: crawlStatus.internal_links, color: CHART_COLORS.primary },
    { name: 'External Links', value: crawlStatus.external_links, color: CHART_COLORS.success },
    { name: 'Broken Links', value: crawlStatus.broken_links, color: CHART_COLORS.error },
  ].filter(item => item.value > 0) : []

  // Status distribution for additional chart
  const statusData = crawlStatus ? [
    { name: 'Working Links', value: (crawlStatus.internal_links + crawlStatus.external_links) - crawlStatus.broken_links, color: CHART_COLORS.success },
    { name: 'Broken Links', value: crawlStatus.broken_links, color: CHART_COLORS.error },
  ].filter(item => item.value > 0) : []

  // SEO Score calculation
  const calculateSeoScore = () => {
    if (!crawlStatus) return 0
    
    let score = 100
    
    // Deduct points for broken links
    if (crawlStatus.broken_links > 0) {
      const totalLinks = crawlStatus.internal_links + crawlStatus.external_links
      const brokenRatio = crawlStatus.broken_links / totalLinks
      score -= Math.min(brokenRatio * 50, 30) // Max 30 points deduction
    }
    
    // Deduct points for missing title
    if (!url.title || url.title.length < 10) {
      score -= 15
    }
    
    // Deduct points for poor heading structure
    const h1Count = headingCounts.h1 || 0
    if (h1Count === 0) score -= 10
    if (h1Count > 1) score -= 5
    
    // Bonus for login form (indicates interactive content)
    if (url.has_login_form) score += 5
    
    return Math.max(Math.round(score), 0)
  }

  const seoScore = calculateSeoScore()

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'N/A'
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  const getScoreColor = (score: number) => {
    if (score >= 90) return 'text-success'
    if (score >= 70) return 'text-warning'
    return 'text-error'
  }

  const getScoreGradient = (score: number) => {
    if (score >= 90) return 'from-success to-green-500'
    if (score >= 70) return 'from-warning to-orange-500'
    return 'from-error to-red-500'
  }

  const totalLinks = crawlStatus ? crawlStatus.internal_links + crawlStatus.external_links : 0
  const healthyLinks = crawlStatus ? totalLinks - crawlStatus.broken_links : 0
  const healthRatio = totalLinks > 0 ? (healthyLinks / totalLinks) * 100 : 0

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="p-4 sm:p-6 md:p-8 max-w-7xl mx-auto space-y-6 md:space-y-8">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 sm:gap-0">
          <Link
            to="/"
            className="flex items-center space-x-2 text-text-secondary hover:text-text-primary transition-colors group"
          >
            <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
            <span className="font-medium">Back to Dashboard</span>
          </Link>
          
          <div className="flex items-center gap-2 md:gap-3 flex-wrap">
            <button
              onClick={copyUrl}
              className="btn-ghost inline-flex items-center text-sm"
              title="Copy URL"
            >
              <Copy className="w-4 h-4 mr-2" />
              Copy URL
            </button>
            
            {(url.status === 'pending' || url.status === 'error') && (
              <button
                onClick={handleStartCrawl}
                disabled={startCrawlMutation.isPending}
                className="btn-primary inline-flex items-center text-sm"
              >
                <RefreshCw className={`w-4 h-4 mr-2 ${startCrawlMutation.isPending ? 'animate-spin' : ''}`} />
                {startCrawlMutation.isPending ? 'Starting...' : 'Start Crawl'}
              </button>
            )}
            
            <button
              onClick={handleDelete}
              disabled={deleteUrlMutation.isPending}
              className="btn-ghost text-error hover:bg-error/10 inline-flex items-center text-sm"
            >
              <Trash2 className="w-4 h-4 mr-2" />
              Delete
            </button>
          </div>
        </div>

        {/* Hero Section */}
        <div className="bg-gradient-to-br from-bg-secondary/50 to-bg-tertiary/30 backdrop-blur-sm border border-border rounded-3xl p-6 md:p-8">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-6 lg:space-y-0">
            <div className="flex-1 min-w-0">
              <div className="flex items-center space-x-4 mb-4">
                <div className={`w-12 h-12 rounded-2xl bg-gradient-to-br ${statusInfo.gradient} flex items-center justify-center`}>
                  <StatusIcon className={`w-6 h-6 text-white ${url.status === 'running' ? 'animate-spin' : ''}`} />
                </div>
                <div className="min-w-0 flex-1">
                  <h1 className="text-2xl md:text-3xl font-bold text-text-primary mb-1 truncate">
                    {url.title || 'Untitled Page'}
                  </h1>
                  <div className="flex items-center flex-wrap gap-2">
                    <span className={`inline-flex items-center px-3 py-1 rounded-xl text-sm font-medium ${statusInfo.bgColor} ${statusInfo.color} border ${statusInfo.borderColor}`}>
                      {statusInfo.label}
                    </span>
                    {(isRunning || isQueued) && <RealTimeIndicatorMini />}
                  </div>
                </div>
              </div>
              
              <div className="flex items-center space-x-3 text-text-secondary min-w-0">
                <Globe className="w-4 h-4 flex-shrink-0" />
                <a 
                  href={url.url} 
                  target="_blank" 
                  rel="noopener noreferrer"
                  className="hover:text-accent transition-colors group flex items-center font-medium min-w-0"
                >
                  <span className="truncate">{url.url}</span>
                  <ExternalLink className="w-3 h-3 ml-2 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" />
                </a>
              </div>
            </div>

            {crawlStatus && (
              <div className="flex items-center justify-center lg:justify-end">
                <div className="grid grid-cols-3 gap-4 md:gap-6">
                  <div className="text-center">
                    <div className="text-xl md:text-2xl font-bold text-text-primary">{totalLinks}</div>
                    <div className="text-text-tertiary text-xs md:text-sm">Total Links</div>
                  </div>
                  <div className="text-center">
                    <div className={`text-xl md:text-2xl font-bold ${getScoreColor(seoScore)}`}>{seoScore}</div>
                    <div className="text-text-tertiary text-xs md:text-sm">SEO Score</div>
                  </div>
                  <div className="text-center">
                    <div className="text-xl md:text-2xl font-bold text-text-primary">{healthRatio.toFixed(1)}%</div>
                    <div className="text-text-tertiary text-xs md:text-sm">Health Rate</div>
                  </div>
                </div>
              </div>
            )}
          </div>

          {crawlStatus?.started_at && (
            <div className="mt-6 pt-6 border-t border-border/50">
              <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4 text-sm">
                <div className="flex items-center space-x-3">
                  <Calendar className="w-4 h-4 text-text-tertiary flex-shrink-0" />
                  <div className="min-w-0">
                    <span className="text-text-tertiary">Started:</span>
                    <div className="font-medium text-text-primary truncate">{formatDate(crawlStatus.started_at)}</div>
                  </div>
                </div>
                {crawlStatus.completed_at && (
                  <div className="flex items-center space-x-3">
                    <CheckCircle className="w-4 h-4 text-success flex-shrink-0" />
                    <div className="min-w-0">
                      <span className="text-text-tertiary">Completed:</span>
                      <div className="font-medium text-text-primary truncate">{formatDate(crawlStatus.completed_at)}</div>
                    </div>
                  </div>
                )}
                {crawlStatus.error_message && (
                  <div className="flex items-center space-x-3 col-span-full">
                    <XCircle className="w-4 h-4 text-error flex-shrink-0" />
                    <div className="min-w-0">
                      <span className="text-text-tertiary">Error:</span>
                      <div className="font-medium text-error break-words">{crawlStatus.error_message}</div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Stats Grid */}
        {crawlStatus && (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4 md:gap-6">
            <StatCard
              title="Internal Links"
              value={crawlStatus.internal_links}
              icon={LinkIcon}
              gradient="from-accent to-blue-500"
              trend={{ value: 15, isPositive: true }}
            />
            <StatCard
              title="External Links"
              value={crawlStatus.external_links}
              icon={ExternalLink}
              gradient="from-success to-green-500"
            />
            <StatCard
              title="Broken Links"
              value={crawlStatus.broken_links}
              subtitle={crawlStatus.broken_links > 0 ? 'Needs attention' : 'All links working'}
              icon={XCircle}
              gradient="from-error to-red-500"
              trend={crawlStatus.broken_links > 0 ? { value: 5, isPositive: false } : undefined}
            />
            <StatCard
              title="HTML Version"
              value={url.html_version || 'Unknown'}
              icon={FileText}
              gradient="from-purple-500 to-purple-600"
            />
            <StatCard
              title="SEO Score"
              value={seoScore}
              subtitle={seoScore >= 90 ? 'Excellent' : seoScore >= 70 ? 'Good' : 'Needs Work'}
              icon={Award}
              gradient={getScoreGradient(seoScore)}
            />
          </div>
        )}

        {/* Charts Grid */}
        {crawlStatus && (
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 md:gap-6">
            {/* Link Distribution */}
            <ChartCard
              title="Link Distribution"
              icon={BarChart3}
              actions={
                <button className="btn-ghost p-2">
                  <Download className="w-4 h-4" />
                </button>
              }
            >
              {linkData.length > 0 ? (
                <div className="h-64 md:h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={linkData}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                        outerRadius={80}
                        fill="rgb(var(--accent))"
                        dataKey="value"
                        stroke="none"
                      >
                        {linkData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip 
                        formatter={(value) => [value, 'Links']}
                        contentStyle={{
                          backgroundColor: 'rgb(var(--bg-secondary))',
                          border: '1px solid rgb(var(--border))',
                          borderRadius: '12px',
                          color: 'rgb(var(--text-primary))',
                          boxShadow: '0 10px 40px -10px rgba(0, 0, 0, 0.1)'
                        }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              ) : (
                <div className="h-64 md:h-80 flex items-center justify-center text-text-tertiary">
                  <div className="text-center">
                    <BarChart3 className="w-12 h-12 mx-auto mb-3 opacity-50" />
                    <p>No link data available</p>
                  </div>
                </div>
              )}
            </ChartCard>

            {/* Heading Analysis */}
            <ChartCard
              title="Heading Structure"
              icon={Hash}
            >
              {headingData.length > 0 ? (
                <div className="h-64 md:h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={headingData}>
                      <XAxis 
                        dataKey="name" 
                        stroke="rgb(var(--text-tertiary))" 
                        fontSize={12}
                      />
                      <YAxis 
                        stroke="rgb(var(--text-tertiary))" 
                        fontSize={12}
                      />
                      <Tooltip 
                        formatter={(value) => [value, 'Count']}
                        contentStyle={{
                          backgroundColor: 'rgb(var(--bg-secondary))',
                          border: '1px solid rgb(var(--border))',
                          borderRadius: '12px',
                          color: 'rgb(var(--text-primary))',
                          boxShadow: '0 10px 40px -10px rgba(0, 0, 0, 0.1)'
                        }}
                      />
                      <Bar 
                        dataKey="count" 
                        fill="rgb(var(--accent))"
                        radius={[4, 4, 0, 0]}
                      />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              ) : (
                <div className="h-64 md:h-80 flex items-center justify-center text-text-tertiary">
                  <div className="text-center">
                    <Activity className="w-12 h-12 mx-auto mb-3 opacity-50" />
                    <p>No heading data available</p>
                  </div>
                </div>
              )}
            </ChartCard>
          </div>
        )}

        {/* Additional Analysis */}
        {crawlStatus && (
          <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 md:gap-6">
            {/* Link Health Status */}
            <ChartCard
              title="Link Health Status"
              icon={Shield}
            >
              {statusData.length > 0 ? (
                <div className="h-64 md:h-80">
                  <ResponsiveContainer width="100%" height="100%">
                    <PieChart>
                      <Pie
                        data={statusData}
                        cx="50%"
                        cy="50%"
                        innerRadius={60}
                        outerRadius={100}
                        paddingAngle={5}
                        dataKey="value"
                        label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                        stroke="none"
                      >
                        {statusData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={entry.color} />
                        ))}
                      </Pie>
                      <Tooltip 
                        formatter={(value) => [value, 'Links']}
                        contentStyle={{
                          backgroundColor: 'rgb(var(--bg-secondary))',
                          border: '1px solid rgb(var(--border))',
                          borderRadius: '12px',
                          color: 'rgb(var(--text-primary))',
                          boxShadow: '0 10px 40px -10px rgba(0, 0, 0, 0.1)'
                        }}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              ) : (
                <div className="h-64 md:h-80 flex items-center justify-center text-text-tertiary">
                  <div className="text-center">
                    <Shield className="w-12 h-12 mx-auto mb-3 opacity-50" />
                    <p>No health data available</p>
                  </div>
                </div>
              )}
            </ChartCard>

            {/* Page Analysis */}
            <ChartCard
              title="Page Analysis"
              icon={Monitor}
            >
              <div className="space-y-3 md:space-y-4">
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 md:p-4 bg-bg-tertiary/50 rounded-xl gap-2 sm:gap-0">
                  <span className="text-text-secondary font-medium">Page Title:</span>
                  <span className="text-text-primary text-left sm:text-right sm:max-w-xs truncate font-medium">
                    {url.title || 'No title found'}
                  </span>
                </div>
                
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 md:p-4 bg-bg-tertiary/50 rounded-xl gap-2 sm:gap-0">
                  <span className="text-text-secondary font-medium">HTML Version:</span>
                  <span className="text-text-primary font-medium">{url.html_version || 'Unknown'}</span>
                </div>
                
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 md:p-4 bg-bg-tertiary/50 rounded-xl gap-2 sm:gap-0">
                  <span className="text-text-secondary font-medium">Has Login Form:</span>
                  <span className={`font-medium ${url.has_login_form ? 'text-success' : 'text-text-tertiary'}`}>
                    {url.has_login_form ? 'Yes' : 'No'}
                  </span>
                </div>
                
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 md:p-4 bg-bg-tertiary/50 rounded-xl gap-2 sm:gap-0">
                  <span className="text-text-secondary font-medium">Total Links:</span>
                  <span className="text-text-primary font-medium">{totalLinks}</span>
                </div>
                
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 md:p-4 bg-bg-tertiary/50 rounded-xl gap-2 sm:gap-0">
                  <span className="text-text-secondary font-medium">Accessibility Rate:</span>
                  <span className={`font-medium ${
                    crawlStatus.broken_links === 0 ? 'text-success' :
                    crawlStatus.broken_links < 5 ? 'text-warning' : 'text-error'
                  }`}>
                    {totalLinks > 0 ? `${healthRatio.toFixed(1)}%` : 'N/A'}
                  </span>
                </div>

                <div className="pt-2 md:pt-4">
                  <div className="flex justify-between items-center mb-2">
                    <span className="text-text-secondary text-sm">Overall Health</span>
                    <span className="text-text-primary font-medium">{healthRatio.toFixed(1)}%</span>
                  </div>
                  <div className="w-full bg-bg-tertiary rounded-full h-3">
                    <div 
                      className={`h-3 rounded-full transition-all duration-500 ${
                        healthRatio >= 90 ? 'bg-success' :
                        healthRatio >= 70 ? 'bg-warning' : 'bg-error'
                      }`}
                      style={{ width: `${healthRatio}%` }}
                    />
                  </div>
                </div>
              </div>
            </ChartCard>
          </div>
        )}

        {/* Broken Links List */}
        {url.status === 'completed' && (
          <BrokenLinksList urlId={urlId} />
        )}

        {/* Technical Information */}
        <ChartCard
          title="Technical Information"
          icon={Monitor}
        >
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 md:gap-8">
            <div>
              <h4 className="text-base md:text-lg font-semibold text-text-primary mb-3 md:mb-4 flex items-center">
                <Timer className="w-4 md:w-5 h-4 md:h-5 mr-2" />
                Timeline
              </h4>
              <div className="space-y-3 md:space-y-4">
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 bg-bg-tertiary/30 rounded-lg gap-2 sm:gap-0">
                  <span className="text-text-secondary">Created:</span>
                  <span className="text-text-primary font-medium">{formatDate(url.created_at)}</span>
                </div>
                <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 bg-bg-tertiary/30 rounded-lg gap-2 sm:gap-0">
                  <span className="text-text-secondary">Last Updated:</span>
                  <span className="text-text-primary font-medium">{formatDate(url.updated_at)}</span>
                </div>
                {crawlStatus?.started_at && (
                  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 bg-bg-tertiary/30 rounded-lg gap-2 sm:gap-0">
                    <span className="text-text-secondary">Last Crawl:</span>
                    <span className="text-text-primary font-medium">{formatDate(crawlStatus.started_at)}</span>
                  </div>
                )}
              </div>
            </div>

            {crawlStatus && (
              <div>
                <h4 className="text-base md:text-lg font-semibold text-text-primary mb-3 md:mb-4 flex items-center">
                  <Zap className="w-4 md:w-5 h-4 md:h-5 mr-2" />
                  Performance
                </h4>
                <div className="space-y-3 md:space-y-4">
                  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 bg-bg-tertiary/30 rounded-lg gap-2 sm:gap-0">
                    <span className="text-text-secondary">Status Code:</span>
                    <span className="text-text-primary font-medium">
                      {url.status === 'completed' ? '200 OK' : url.status.toUpperCase()}
                    </span>
                  </div>
                  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 bg-bg-tertiary/30 rounded-lg gap-2 sm:gap-0">
                    <span className="text-text-secondary">Content Type:</span>
                    <span className="text-text-primary font-medium">HTML Document</span>
                  </div>
                  <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center p-3 bg-bg-tertiary/30 rounded-lg gap-2 sm:gap-0">
                    <span className="text-text-secondary">Crawl Duration:</span>
                    <span className="text-text-primary font-medium">
                      {crawlStatus.started_at && crawlStatus.completed_at ? (
                        `${Math.round((new Date(crawlStatus.completed_at).getTime() - new Date(crawlStatus.started_at).getTime()) / 1000)}s`
                      ) : 'N/A'}
                    </span>
                  </div>
                </div>
              </div>
            )}
          </div>
        </ChartCard>
      </div>
    </div>
  )
}