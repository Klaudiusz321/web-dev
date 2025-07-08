import { ActivityIcon, Wifi, RefreshCw } from 'lucide-react'
import { useHasRunningCrawls, useRealTimeStatus } from '../hooks/useUrls'

interface RealTimeIndicatorProps {
  className?: string
  showText?: boolean
}

export default function RealTimeIndicator({ className = '', showText = true }: RealTimeIndicatorProps) {
  const hasRunningCrawls = useHasRunningCrawls()
  const { isPolling } = useRealTimeStatus()

  if (!hasRunningCrawls) {
    return null
  }

  return (
    <div className={`flex items-center space-x-2 ${className}`}>
      <div className="relative">
        <Wifi className="h-4 w-4 text-green-600" />
        {isPolling && (
          <div className="absolute -top-1 -right-1">
            <RefreshCw className="h-2 w-2 text-blue-600 animate-spin" />
          </div>
        )}
      </div>
      
      {showText && (
        <div className="flex items-center space-x-1">
          <span className="text-sm text-green-600 font-medium">Live</span>
          <div className="flex space-x-1">
            <div className="w-1 h-1 bg-green-600 rounded-full animate-pulse"></div>
            <div className="w-1 h-1 bg-green-600 rounded-full animate-pulse" style={{ animationDelay: '0.5s' }}></div>
            <div className="w-1 h-1 bg-green-600 rounded-full animate-pulse" style={{ animationDelay: '1s' }}></div>
          </div>
        </div>
      )}
    </div>
  )
}

// Alternative minimal indicator for small spaces
export function RealTimeIndicatorMini({ className = '' }: { className?: string }) {
  const hasRunningCrawls = useHasRunningCrawls()
  const { isPolling } = useRealTimeStatus()

  if (!hasRunningCrawls) {
    return null
  }

  return (
    <div className={`flex items-center ${className}`}>
      <div className="relative">
        <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
        {isPolling && (
          <div className="absolute inset-0 w-2 h-2 bg-green-500 rounded-full animate-ping"></div>
        )}
      </div>
    </div>
  )
}

// Hook for real-time status badge
export function useRealTimeStatusBadge() {
  const hasRunningCrawls = useHasRunningCrawls()
  const { isPolling } = useRealTimeStatus()

  if (!hasRunningCrawls) {
    return null
  }

  return {
    icon: isPolling ? RefreshCw : ActivityIcon,
    text: isPolling ? 'Updating...' : 'Live',
    className: isPolling 
      ? 'text-blue-600 bg-blue-100 animate-pulse' 
      : 'text-green-600 bg-green-100',
    spinning: isPolling
  }
} 