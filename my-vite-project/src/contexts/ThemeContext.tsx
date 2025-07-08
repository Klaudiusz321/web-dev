import React, { createContext, useContext, useEffect, useState } from 'react'

type ThemeMode = 'light' | 'dark' | 'system'
type ResolvedTheme = 'light' | 'dark'

interface ThemeContextType {
  mode: ThemeMode
  resolvedTheme: ResolvedTheme
  setMode: (mode: ThemeMode) => void
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined)

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [mode, setModeState] = useState<ThemeMode>(() => {
    const stored = localStorage.getItem('theme-mode') as ThemeMode
    return stored || 'system'
  })

  const [systemTheme, setSystemTheme] = useState<ResolvedTheme>(() => {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  })

  // Listen for system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    
    const handleChange = (e: MediaQueryListEvent) => {
      setSystemTheme(e.matches ? 'dark' : 'light')
    }

    mediaQuery.addEventListener('change', handleChange)
    return () => mediaQuery.removeEventListener('change', handleChange)
  }, [])

  // Calculate resolved theme
  const resolvedTheme: ResolvedTheme = mode === 'system' ? systemTheme : mode

  useEffect(() => {
    const root = document.documentElement
    
    // Remove previous theme classes
    root.classList.remove('light', 'dark')
    
    // Add current theme class
    root.classList.add(resolvedTheme)
    
    // Update CSS custom properties for smooth transitions
    if (resolvedTheme === 'dark') {
      // Dark theme with improved colors
      root.style.setProperty('--bg-primary', '9 9 11') // zinc-900
      root.style.setProperty('--bg-secondary', '24 24 27') // zinc-800  
      root.style.setProperty('--bg-tertiary', '39 39 42') // zinc-700
      root.style.setProperty('--text-primary', '250 250 250') // zinc-50
      root.style.setProperty('--text-secondary', '212 212 216') // zinc-300
      root.style.setProperty('--text-tertiary', '161 161 170') // zinc-400
      root.style.setProperty('--border', '63 63 70') // zinc-600
      root.style.setProperty('--accent', '99 102 241') // indigo-500
      root.style.setProperty('--accent-secondary', '129 140 248') // indigo-400
      root.style.setProperty('--success', '34 197 94') // green-500
      root.style.setProperty('--warning', '251 191 36') // yellow-400
      root.style.setProperty('--error', '239 68 68') // red-500
    } else {
      // Light theme with improved colors
      root.style.setProperty('--bg-primary', '255 255 255') // white
      root.style.setProperty('--bg-secondary', '250 250 250') // zinc-50
      root.style.setProperty('--bg-tertiary', '244 244 245') // zinc-100
      root.style.setProperty('--text-primary', '24 24 27') // zinc-800
      root.style.setProperty('--text-secondary', '63 63 70') // zinc-600
      root.style.setProperty('--text-tertiary', '113 113 122') // zinc-500
      root.style.setProperty('--border', '228 228 231') // zinc-200
      root.style.setProperty('--accent', '79 70 229') // indigo-600
      root.style.setProperty('--accent-secondary', '99 102 241') // indigo-500
      root.style.setProperty('--success', '22 163 74') // green-600
      root.style.setProperty('--warning', '217 119 6') // yellow-600
      root.style.setProperty('--error', '220 38 38') // red-600
    }
    
    // Store mode in localStorage
    localStorage.setItem('theme-mode', mode)
  }, [resolvedTheme, mode])

  const setMode = (newMode: ThemeMode) => {
    setModeState(newMode)
  }

  const value: ThemeContextType = {
    mode,
    resolvedTheme,
    setMode,
  }

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  const context = useContext(ThemeContext)
  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
} 