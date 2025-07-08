import React, { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { 
  LayoutGrid, 
  Plus, 
  Moon, 
  Sun, 
  Monitor,
  LogOut,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  Globe,
  BarChart3,
  Check
} from 'lucide-react'
import { useTheme } from '../contexts/ThemeContext'
import { useAuth } from '../contexts/AuthContext'
import toast from 'react-hot-toast'

interface NavItem {
  label: string
  href: string
  icon: React.ComponentType<{ className?: string }>
  badge?: number
}

export default function Sidebar() {
  const [isCollapsed, setIsCollapsed] = useState(false)
  const [showUserMenu, setShowUserMenu] = useState(false)
  const [showThemeMenu, setShowThemeMenu] = useState(false)
  const location = useLocation()
  const { mode, resolvedTheme, setMode } = useTheme()
  const { user, logout } = useAuth()

  const navigationItems: NavItem[] = [
    {
      label: 'Dashboard',
      href: '/',
      icon: LayoutGrid,
    },
    {
      label: 'Add URL',
      href: '/add',
      icon: Plus,
    },
    {
      label: 'Analytics',
      href: '/analytics',
      icon: BarChart3,
    },
  ]

  const isActive = (path: string) => location.pathname === path

  const handleLogout = () => {
    logout()
    toast.success('Logged out successfully')
    setShowUserMenu(false)
  }

  return (
    <div className={`
      ${isCollapsed ? 'w-16' : 'w-64'} 
      transition-all duration-300 ease-in-out
      h-screen bg-bg-secondary/80 backdrop-blur-xl border-r border-border
      flex flex-col relative
      shadow-elegant
    `}>
      {/* Header */}
      <div className="p-4 border-b border-border/50">
        <div className="flex items-center justify-between">
          {!isCollapsed && (
            <div className="flex items-center space-x-3 animate-fade-in">
              <div className="w-8 h-8 bg-gradient-to-br from-accent to-accent-secondary rounded-lg flex items-center justify-center">
                <Globe className="w-5 h-5 text-white" />
              </div>
              <div>
                <h1 className="text-lg font-bold text-text-primary">WebCrawler</h1>
                <p className="text-xs text-text-tertiary">Pro Analytics</p>
              </div>
            </div>
          )}
          
          <button
            onClick={() => setIsCollapsed(!isCollapsed)}
            className="p-1.5 rounded-lg hover:bg-bg-tertiary transition-colors text-text-secondary hover:text-text-primary"
          >
            {isCollapsed ? (
              <ChevronRight className="w-4 h-4" />
            ) : (
              <ChevronLeft className="w-4 h-4" />
            )}
          </button>
        </div>
      </div>

      {/* Navigation */}
      <div className="flex-1 p-3 space-y-1 overflow-y-auto">
        {navigationItems.map((item) => {
          const IconComponent = item.icon
          const active = isActive(item.href)
          
          return (
            <Link
              key={item.href}
              to={item.href}
              className={`
                group flex items-center space-x-3 px-3 py-2.5 rounded-xl
                transition-all duration-200 relative overflow-hidden
                ${active 
                  ? 'bg-accent text-white shadow-lg' 
                  : 'text-text-secondary hover:text-text-primary hover:bg-bg-tertiary/70'
                }
              `}
            >
              <IconComponent className={`
                w-5 h-5 transition-transform duration-200
                ${active ? 'scale-110' : 'group-hover:scale-105'}
              `} />
              
              {!isCollapsed && (
                <span className="font-medium animate-fade-in">
                  {item.label}
                </span>
              )}
              
              {item.badge && !isCollapsed && (
                <span className="ml-auto bg-accent-secondary text-white text-xs px-2 py-0.5 rounded-full">
                  {item.badge}
                </span>
              )}
              
              {/* Active indicator */}
              {active && (
                <div className="absolute inset-0 bg-gradient-to-r from-accent/20 to-accent-secondary/20 rounded-xl" />
              )}
            </Link>
          )
        })}
      </div>

      {/* Divider */}
      <div className="px-3">
        <div className="h-px bg-border/50" />
      </div>

      

      {/* Theme Toggle */}
      <div className="p-3 relative">
        <button
          onClick={() => setShowThemeMenu(!showThemeMenu)}
          className={`
            w-full flex items-center space-x-3 px-3 py-2.5 rounded-xl
            text-text-secondary hover:text-text-primary hover:bg-bg-tertiary/70
            transition-all duration-200 group
          `}
        >
          {resolvedTheme === 'dark' ? (
            <Moon className="w-5 h-5 group-hover:-rotate-12 transition-transform duration-200" />
          ) : (
            <Sun className="w-5 h-5 group-hover:rotate-12 transition-transform duration-200" />
          )}
          
          {!isCollapsed && (
            <div className="flex-1 flex items-center justify-between animate-fade-in">
              <span className="font-medium">
                {mode === 'system' ? 'System Theme' : mode === 'dark' ? 'Dark Mode' : 'Light Mode'}
              </span>
              <ChevronDown className={`w-4 h-4 transition-transform duration-200 ${showThemeMenu ? 'rotate-180' : ''}`} />
            </div>
          )}
        </button>

        {/* Theme Dropdown */}
        {showThemeMenu && !isCollapsed && (
          <div className="absolute bottom-full left-3 right-3 mb-2 bg-bg-primary border border-border rounded-xl shadow-elegant-lg animate-scale-in">
            <div className="p-1">
              {[
                { mode: 'light' as const, label: 'Light', icon: Sun },
                { mode: 'dark' as const, label: 'Dark', icon: Moon },
                { mode: 'system' as const, label: 'System', icon: Monitor },
              ].map((option) => (
                <button
                  key={option.mode}
                  onClick={() => {
                    setMode(option.mode)
                    setShowThemeMenu(false)
                  }}
                  className={`
                    w-full flex items-center space-x-3 px-3 py-2 rounded-lg transition-colors
                    ${mode === option.mode 
                      ? 'bg-accent text-white' 
                      : 'text-text-secondary hover:text-text-primary hover:bg-bg-tertiary/70'
                    }
                  `}
                >
                  <option.icon className="w-4 h-4" />
                  <span className="flex-1 text-left text-sm font-medium">{option.label}</span>
                  {mode === option.mode && <Check className="w-4 h-4" />}
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* User Section */}
      <div className="p-3 border-t border-border/50">
        <div className="relative">
          <button
            onClick={() => setShowUserMenu(!showUserMenu)}
            className={`
              w-full flex items-center space-x-3 px-3 py-3 rounded-xl
              text-text-secondary hover:text-text-primary hover:bg-bg-tertiary/70
              transition-all duration-200 group
            `}
          >
            <div className="w-8 h-8 bg-gradient-to-br from-accent to-accent-secondary rounded-full flex items-center justify-center">
              <span className="text-white font-semibold text-sm">
                {user?.first_name?.[0]}{user?.last_name?.[0]}
              </span>
            </div>
            
            {!isCollapsed && (
              <div className="flex-1 text-left animate-fade-in">
                <p className="text-sm font-medium text-text-primary">
                  {user?.first_name} {user?.last_name}
                </p>
                <p className="text-xs text-text-tertiary">
                  @{user?.username}
                </p>
              </div>
            )}
          </button>

          {/* User Dropdown */}
          {showUserMenu && !isCollapsed && (
            <div className="absolute bottom-full left-0 right-0 mb-2 bg-bg-primary border border-border rounded-xl shadow-elegant-lg animate-scale-in">
              <div className="p-3 border-b border-border/50">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-gradient-to-br from-accent to-accent-secondary rounded-full flex items-center justify-center">
                    <span className="text-white font-semibold">
                      {user?.first_name?.[0]}{user?.last_name?.[0]}
                    </span>
                  </div>
                  <div>
                    <p className="font-medium text-text-primary">
                      {user?.first_name} {user?.last_name}
                    </p>
                    <p className="text-sm text-text-secondary">
                      {user?.email}
                    </p>
                    {user?.is_admin && (
                      <span className="inline-block mt-1 px-2 py-0.5 bg-accent/20 text-accent text-xs rounded-full">
                        Admin
                      </span>
                    )}
                  </div>
                </div>
              </div>
              
              <div className="p-1">
                <button
                  onClick={handleLogout}
                  className="w-full flex items-center space-x-2 px-3 py-2 text-error hover:bg-error/10 rounded-lg transition-colors text-sm"
                >
                  <LogOut className="w-4 h-4" />
                  <span>Sign Out</span>
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
} 