import { Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import { ThemeProvider } from './contexts/ThemeContext'
import AppLayout from './components/AppLayout'
import ProtectedRoute from './components/ProtectedRoute'
import Dashboard from './pages/Dashboard'
import UrlDetails from './pages/UrlDetails'
import AddUrl from './pages/AddUrl'
import Login from './pages/Login'
import Register from './pages/Register'

// Component that handles authenticated routes
function AuthenticatedApp() {
  return (
    <AppLayout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/add" element={<AddUrl />} />
        <Route path="/url/:id" element={<UrlDetails />} />
        {/* Catch all other routes and redirect to dashboard */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </AppLayout>
  )
}

// Component that handles public/auth routes
function PublicApp() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      {/* Redirect all other routes to login */}
      <Route path="*" element={<Navigate to="/login" replace />} />
    </Routes>
  )
}

// Main app router component
function AppRouter() {
  const { isAuthenticated, isLoading } = useAuth()

  // Show loading spinner while checking authentication
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return isAuthenticated ? (
    <ProtectedRoute>
      <AuthenticatedApp />
    </ProtectedRoute>
  ) : (
    <PublicApp />
  )
}

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <AppRouter />
        <Toaster 
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: 'rgb(var(--bg-secondary))',
              color: 'rgb(var(--text-primary))',
              border: '1px solid rgb(var(--border))',
              borderRadius: '12px',
              fontSize: '14px',
              fontWeight: '500',
            },
            success: {
              iconTheme: {
                primary: 'rgb(var(--success))',
                secondary: 'white',
              },
            },
            error: {
              iconTheme: {
                primary: 'rgb(var(--error))',
                secondary: 'white',
              },
            },
          }}
        />
      </AuthProvider>
    </ThemeProvider>
  )
}

export default App 