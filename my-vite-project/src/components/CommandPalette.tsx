import React, { useState, useEffect, useMemo } from 'react'
import { createPortal } from 'react-dom'
import { useNavigate } from 'react-router-dom'
import { 
  Search, 
  Plus, 
  Home, 
  Sun, 
  Moon, 
  Monitor,
  LogOut,
  Command,
  Globe,
  Settings,
  User,
  Play,
  BarChart3,
  Clock,
  CheckCircle,
  XCircle,
  ArrowRight
} from 'lucide-react'
import { useTheme } from '../contexts/ThemeContext'
import { useAuth } from '../contexts/AuthContext'
import { useUrls } from '../hooks/useUrls'
import toast from 'react-hot-toast'

interface Command {
  id: string
  title: string
  subtitle?: string
  icon: React.ComponentType<{ className?: string }>
  action: () => void
  category: 'navigation' | 'actions' | 'settings' | 'urls'
  keywords?: string[]
}

interface CommandPaletteProps {
  isOpen: boolean
  onClose: () => void
}

export default function CommandPalette({ isOpen, onClose }: CommandPaletteProps) {
  const [query, setQuery] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const navigate = useNavigate()
  const { resolvedTheme, setMode } = useTheme()
  const { user, logout } = useAuth()
  
  // Fetch recent URLs for quick access
  const { data: urlsResponse } = useUrls({
    limit: 5,
    sortBy: 'updated_at',
    sortOrder: 'desc'
  })
  
  const recentUrls = urlsResponse?.data || []

  // Base commands
  const baseCommands: Command[] = [
    // Navigation
    {
      id: 'nav-dashboard',
      title: 'Go to Dashboard',
      subtitle: 'View all crawled URLs',
      icon: Home,
      category: 'navigation',
      action: () => {
        navigate('/')
        onClose()
      },
      keywords: ['dashboard', 'home', 'overview']
    },
    {
      id: 'nav-add-url',
      title: 'Add New URL',
      subtitle: 'Start crawling a new website',
      icon: Plus,
      category: 'navigation',
      action: () => {
        navigate('/add')
        onClose()
      },
      keywords: ['add', 'new', 'url', 'crawl', 'create']
    },
    {
      id: 'nav-analytics',
      title: 'Analytics',
      subtitle: 'View detailed analytics',
      icon: BarChart3,
      category: 'navigation',
      action: () => {
        navigate('/analytics')
        onClose()
      },
      keywords: ['analytics', 'stats', 'reports', 'data']
    },
    
    // Actions
    {
      id: 'action-theme-light',
      title: 'Switch to Light Mode',
      subtitle: 'Set light theme',
      icon: Sun,
      category: 'actions',
      action: () => {
        setMode('light')
        toast.success('Switched to light mode')
        onClose()
      },
      keywords: ['theme', 'light', 'mode', 'appearance']
    },
    {
      id: 'action-theme-dark',
      title: 'Switch to Dark Mode',
      subtitle: 'Set dark theme',
      icon: Moon,
      category: 'actions',
      action: () => {
        setMode('dark')
        toast.success('Switched to dark mode')
        onClose()
      },
      keywords: ['theme', 'dark', 'mode', 'appearance']
    },
    {
      id: 'action-theme-system',
      title: 'Use System Theme',
      subtitle: 'Follow system preference',
      icon: Monitor,
      category: 'actions',
      action: () => {
        setMode('system')
        toast.success('Using system theme')
        onClose()
      },
      keywords: ['theme', 'system', 'auto', 'mode', 'appearance']
    },
    
    // Settings
    {
      id: 'settings-profile',
      title: 'Profile Settings',
      subtitle: `Signed in as ${user?.first_name} ${user?.last_name}`,
      icon: User,
      category: 'settings',
      action: () => {
        // Future: navigate to profile page
        toast('Profile settings coming soon!')
        onClose()
      },
      keywords: ['profile', 'account', 'user', 'settings']
    },
    {
      id: 'action-logout',
      title: 'Sign Out',
      subtitle: 'Log out of your account',
      icon: LogOut,
      category: 'settings',
      action: () => {
        logout()
        toast.success('Signed out successfully')
        onClose()
      },
      keywords: ['logout', 'sign out', 'exit']
    },
  ]

  // Generate URL commands from recent URLs
  const urlCommands: Command[] = recentUrls.map((url) => ({
    id: `url-${url.id}`,
    title: url.title || 'Untitled Page',
    subtitle: url.url,
    icon: url.status === 'completed' ? CheckCircle : 
          url.status === 'running' ? Clock :
          url.status === 'error' ? XCircle : Clock,
    category: 'urls' as const,
    action: () => {
      navigate(`/url/${url.id}`)
      onClose()
    },
    keywords: [url.title || '', url.url, url.status]
  }))

  // Combine all commands
  const allCommands = useMemo(() => [
    ...baseCommands,
    ...urlCommands
  ], [baseCommands, urlCommands, resolvedTheme, user])

  // Filter commands based on query
  const filteredCommands = useMemo(() => {
    if (!query.trim()) return allCommands

    const searchTerm = query.toLowerCase()
    return allCommands.filter(command => {
      const searchableText = [
        command.title,
        command.subtitle,
        ...(command.keywords || [])
      ].join(' ').toLowerCase()
      
      return searchableText.includes(searchTerm)
    })
  }, [allCommands, query])

  // Group commands by category
  const groupedCommands = useMemo(() => {
    const groups: Record<string, Command[]> = {}
    
    filteredCommands.forEach(command => {
      if (!groups[command.category]) {
        groups[command.category] = []
      }
      groups[command.category].push(command)
    })
    
    return groups
  }, [filteredCommands])

  // Reset selection when commands change
  useEffect(() => {
    setSelectedIndex(0)
  }, [filteredCommands])

  // Handle keyboard navigation
  useEffect(() => {
    if (!isOpen) return

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex(prev => 
            prev < filteredCommands.length - 1 ? prev + 1 : 0
          )
          break
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex(prev => 
            prev > 0 ? prev - 1 : filteredCommands.length - 1
          )
          break
        case 'Enter':
          e.preventDefault()
          if (filteredCommands[selectedIndex]) {
            filteredCommands[selectedIndex].action()
          }
          break
        case 'Escape':
          e.preventDefault()
          onClose()
          break
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, filteredCommands, selectedIndex, onClose])

  // Handle backdrop click
  const handleBackdropClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose()
    }
  }

  if (!isOpen) return null

  const categoryLabels = {
    navigation: 'Navigation',
    actions: 'Actions', 
    settings: 'Settings',
    urls: 'Recent URLs'
  }

  const categoryIcons = {
    navigation: Home,
    actions: Play,
    settings: Settings,
    urls: Globe
  }

  return createPortal(
    <div 
      className="fixed inset-0 z-50 bg-bg-primary/80 backdrop-blur-sm flex items-start justify-center pt-[20vh] px-4"
      onClick={handleBackdropClick}
    >
      <div 
        className="w-full max-w-2xl bg-bg-secondary/90 backdrop-blur-xl border border-border rounded-2xl shadow-elegant-lg animate-scale-in"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center px-6 py-4 border-b border-border/50">
          <Search className="w-5 h-5 text-text-tertiary mr-3" />
          <input
            type="text"
            placeholder="Search for commands, URLs, or actions..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="flex-1 bg-transparent text-text-primary placeholder:text-text-tertiary focus:outline-none text-lg"
            autoFocus
          />
          <div className="flex items-center space-x-2 text-text-tertiary text-sm">
            <kbd className="px-2 py-1 bg-bg-tertiary/50 rounded-lg border border-border text-xs">
              ↑↓
            </kbd>
            <span>navigate</span>
            <kbd className="px-2 py-1 bg-bg-tertiary/50 rounded-lg border border-border text-xs">
              ↵
            </kbd>
            <span>select</span>
            <kbd className="px-2 py-1 bg-bg-tertiary/50 rounded-lg border border-border text-xs">
              esc
            </kbd>
            <span>close</span>
          </div>
        </div>

        {/* Commands List */}
        <div className="max-h-96 overflow-y-auto">
          {filteredCommands.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-text-tertiary">
              <Search className="w-12 h-12 mb-4 opacity-50" />
              <p className="text-lg font-medium mb-2">No commands found</p>
              <p className="text-sm">Try searching for something else</p>
            </div>
          ) : (
            <div className="py-2">
              {Object.entries(groupedCommands).map(([category, commands]) => {
                const CategoryIcon = categoryIcons[category as keyof typeof categoryIcons]
                
                return (
                  <div key={category} className="mb-2">
                    {/* Category Header */}
                    <div className="flex items-center px-6 py-2 text-text-tertiary text-sm font-medium">
                      <CategoryIcon className="w-4 h-4 mr-2" />
                      {categoryLabels[category as keyof typeof categoryLabels]}
                    </div>
                    
                    {/* Commands in Category */}
                    {commands.map((command) => {
                      const globalIndex = filteredCommands.indexOf(command)
                      const isSelected = globalIndex === selectedIndex
                      const IconComponent = command.icon
                      
                      return (
                        <button
                          key={command.id}
                          onClick={command.action}
                          className={`w-full flex items-center px-6 py-3 text-left transition-all duration-150 ${
                            isSelected 
                              ? 'bg-accent/10 border-r-2 border-accent' 
                              : 'hover:bg-bg-tertiary/30'
                          }`}
                        >
                          <div className={`w-8 h-8 rounded-lg flex items-center justify-center mr-4 ${
                            isSelected 
                              ? 'bg-accent/20 text-accent' 
                              : 'bg-bg-tertiary/50 text-text-tertiary'
                          }`}>
                            <IconComponent className="w-4 h-4" />
                          </div>
                          
                          <div className="flex-1 min-w-0">
                            <div className={`font-medium ${
                              isSelected ? 'text-text-primary' : 'text-text-primary'
                            }`}>
                              {command.title}
                            </div>
                            {command.subtitle && (
                              <div className={`text-sm truncate ${
                                isSelected ? 'text-text-secondary' : 'text-text-tertiary'
                              }`}>
                                {command.subtitle}
                              </div>
                            )}
                          </div>
                          
                          {isSelected && (
                            <ArrowRight className="w-4 h-4 text-accent ml-4 flex-shrink-0" />
                          )}
                        </button>
                      )
                    })}
                  </div>
                )
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-6 py-3 border-t border-border/50 text-text-tertiary text-sm">
          <div className="flex items-center space-x-3">
            <div className="flex items-center space-x-1">
              <Command className="w-3 h-3" />
              <span>K</span>
            </div>
            <span>to open • Navigate with arrow keys</span>
          </div>
          <div className="text-xs">
            {filteredCommands.length} command{filteredCommands.length !== 1 ? 's' : ''}
          </div>
        </div>
      </div>
    </div>,
    document.body
  )
}

// Hook to manage command palette state
export function useCommandPalette() {
  const [isOpen, setIsOpen] = useState(false)

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd+K or Ctrl+K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setIsOpen(prev => !prev)
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  return {
    isOpen,
    open: () => setIsOpen(true),
    close: () => setIsOpen(false),
    toggle: () => setIsOpen(prev => !prev)
  }
} 