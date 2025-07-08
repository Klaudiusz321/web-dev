// API base configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

// Types
export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  first_name: string
  last_name: string
}

export interface AuthResponse {
  token: string
  user: {
    id: number
    username: string
    email: string
    first_name: string
    last_name: string
    is_admin: boolean
  }
}

export interface ValidateResponse {
  valid: boolean
  user_id?: number
  username?: string
  is_admin?: boolean
  message: string
}

export interface ProfileResponse {
  user: {
    id: number
    username: string
    email: string
    first_name: string
    last_name: string
    is_admin: boolean
  }
}

export interface ApiError {
  error: string
  message: string
}

// Helper function for API calls
async function apiCall<T>(
  endpoint: string, 
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`
  
  const defaultOptions: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  }

  try {
    const response = await fetch(url, defaultOptions)
    
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Unknown error', message: 'Failed to parse error response' }))
      throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
    }

    return await response.json()
  } catch (error) {
    console.error(`API call failed: ${endpoint}`, error)
    throw error
  }
}

// Helper function for authenticated API calls
async function authenticatedApiCall<T>(
  endpoint: string,
  token: string,
  options: RequestInit = {}
): Promise<T> {
  return apiCall<T>(endpoint, {
    ...options,
    headers: {
      ...options.headers,
      'Authorization': `Bearer ${token}`,
    },
  })
}

// Auth API
export const authApi = {
  // Login user
  login: async (credentials: LoginRequest): Promise<AuthResponse> => {
    return apiCall<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    })
  },

  // Register new user
  register: async (userData: RegisterRequest): Promise<{ message: string; user: AuthResponse['user'] }> => {
    return apiCall<{ message: string; user: AuthResponse['user'] }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    })
  },

  // Validate token
  validateToken: async (token: string): Promise<ValidateResponse> => {
    return authenticatedApiCall<ValidateResponse>('/auth/validate', token)
  },

  // Get user profile
  getProfile: async (token: string): Promise<ProfileResponse> => {
    return authenticatedApiCall<ProfileResponse>('/auth/profile', token)
  },

  // Refresh token
  refreshToken: async (token: string): Promise<AuthResponse> => {
    return apiCall<AuthResponse>('/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ token }),
    })
  },

  // Logout user
  logout: async (token: string): Promise<{ message: string }> => {
    return authenticatedApiCall<{ message: string }>('/auth/logout', token, {
      method: 'POST',
    })
  },
} 