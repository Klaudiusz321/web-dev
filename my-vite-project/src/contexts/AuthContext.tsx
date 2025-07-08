import React, { createContext, useContext, useEffect, useState } from 'react'
import { authApi } from '../services/authApi'

interface User {
  id: number
  username: string
  email: string
  first_name: string
  last_name: string
  is_admin: boolean
}

interface AuthContextType {
  user: User | null
  token: string | null
  login: (username: string, password: string) => Promise<void>
  register: (userData: RegisterData) => Promise<void>
  logout: () => void
  isLoading: boolean
  isAuthenticated: boolean
}

interface RegisterData {
  username: string
  email: string
  password: string
  first_name: string
  last_name: string
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'))
  const [isLoading, setIsLoading] = useState(true)

  const isAuthenticated = !!user && !!token

  // Validate token on app start
  useEffect(() => {
    const validateToken = async () => {
      if (token) {
        try {
          const response = await authApi.validateToken(token)
          if (response.valid) {
            // Get user profile
            const profileResponse = await authApi.getProfile(token)
            setUser(profileResponse.user)
          } else {
            // Token is invalid, clear it
            localStorage.removeItem('token')
            setToken(null)
          }
        } catch (error) {
          console.error('Token validation failed:', error)
          localStorage.removeItem('token')
          setToken(null)
        }
      }
      setIsLoading(false)
    }

    validateToken()
  }, [token])

  const login = async (username: string, password: string) => {
    setIsLoading(true)
    try {
      const response = await authApi.login({ username, password })
      const { token: newToken, user: newUser } = response
      
      setToken(newToken)
      setUser(newUser)
      localStorage.setItem('token', newToken)
    } catch (error) {
      console.error('Login failed:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const register = async (userData: RegisterData) => {
    setIsLoading(true)
    try {
      await authApi.register(userData)
      // After successful registration, log the user in
      await login(userData.username, userData.password)
    } catch (error) {
      console.error('Registration failed:', error)
      throw error
    } finally {
      setIsLoading(false)
    }
  }

  const logout = () => {
    setUser(null)
    setToken(null)
    localStorage.removeItem('token')
    
    // Call logout endpoint (optional, since JWT is stateless)
    if (token) {
      authApi.logout(token).catch(console.error)
    }
  }

  const value: AuthContextType = {
    user,
    token,
    login,
    register,
    logout,
    isLoading,
    isAuthenticated,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
} 