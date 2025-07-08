import { Link, useLocation } from 'react-router-dom'
import { Search, Plus, Home, User, LogOut, ChevronDown } from 'lucide-react'
import { useState } from 'react'
import { useAuth } from '../contexts/AuthContext'
import toast from 'react-hot-toast'

interface LayoutProps {
  children: React.ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const { user, logout } = useAuth()
  const [showUserMenu, setShowUserMenu] = useState(false)

  const isActive = (path: string) => location.pathname === path

  const handleLogout = () => {
    logout()
    toast.success('Logged out successfully')
    setShowUserMenu(false)
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center">
              <Search className="h-8 w-8 text-blue-600" />
              <h1 className="ml-2 text-xl font-bold text-gray-900">
                Web Crawler
              </h1>
            </div>
            
            <div className="flex items-center space-x-4">
              <nav className="flex space-x-4">
                <Link
                  to="/"
                  className={`inline-flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    isActive('/') 
                      ? 'bg-blue-100 text-blue-700' 
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  <Home className="h-4 w-4 mr-1" />
                  Dashboard
                </Link>
                <Link
                  to="/add"
                  className={`inline-flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                    isActive('/add') 
                      ? 'bg-blue-100 text-blue-700' 
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Add URL
                </Link>
              </nav>

              {/* User Menu */}
              <div className="relative">
                <button
                  onClick={() => setShowUserMenu(!showUserMenu)}
                  className="flex items-center text-sm rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 p-2 hover:bg-gray-100"
                >
                  <User className="h-5 w-5 text-gray-600 mr-2" />
                  <span className="text-gray-700 font-medium">
                    {user?.first_name} {user?.last_name}
                  </span>
                  <ChevronDown className="h-4 w-4 text-gray-600 ml-1" />
                </button>

                {showUserMenu && (
                  <div className="absolute right-0 mt-2 w-56 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5">
                    <div className="px-4 py-3 border-b">
                      <p className="text-sm font-medium text-gray-900">
                        {user?.first_name} {user?.last_name}
                      </p>
                      <p className="text-sm text-gray-500 truncate">
                        @{user?.username}
                      </p>
                      <p className="text-sm text-gray-500 truncate">
                        {user?.email}
                      </p>
                      {user?.is_admin && (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 mt-1">
                          Admin
                        </span>
                      )}
                    </div>
                    <div className="py-1">
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 flex items-center"
                      >
                        <LogOut className="h-4 w-4 mr-2" />
                        Sign out
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto py-6 px-4 sm:px-6 lg:px-8">
        {children}
      </main>
    </div>
  )
} 