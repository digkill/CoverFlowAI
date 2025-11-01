import React, { createContext, useContext, useState, useEffect } from 'react'

interface User {
  id: string
  email: string
  name: string
  picture?: string
  can_generate?: boolean
  generations_remaining?: number
  free_generations_left?: number
  paid_generations?: number
}

interface AuthContextType {
  user: User | null
  isLoading: boolean
  signIn: () => Promise<void>
  signOut: () => Promise<void>
  checkGenerationLimit: () => Promise<{ remaining: number; canGenerate: boolean }>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    checkAuth()
  }, [])

  const checkAuth = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/auth/me', {
        credentials: 'include'
      })
      if (response.ok) {
        const userData = await response.json()
        setUser(userData)
      }
    } catch (error) {
      console.error('Auth check failed:', error)
    } finally {
      setIsLoading(false)
    }
  }

  const refreshUser = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/auth/me', {
        credentials: 'include'
      })
      if (response.ok) {
        const userData = await response.json()
        setUser(userData)
      }
    } catch (error) {
      console.error('Failed to refresh user:', error)
    }
  }

  const signIn = async () => {
    // Redirect to backend OAuth endpoint
    window.location.href = 'http://localhost:8080/api/auth/google'
  }

  const signOut = async () => {
    try {
      await fetch('http://localhost:8080/api/auth/logout', { 
        method: 'POST',
        credentials: 'include'
      })
      setUser(null)
    } catch (error) {
      console.error('Sign out failed:', error)
    }
  }

  const checkGenerationLimit = async (): Promise<{ remaining: number; canGenerate: boolean }> => {
    try {
      const response = await fetch('http://localhost:8080/api/user/limits', {
        credentials: 'include'
      })
      if (response.ok) {
        const data = await response.json()
        return {
          remaining: data.remaining || 0,
          canGenerate: data.canGenerate || false,
        }
      }
    } catch (error) {
      console.error('Failed to check limits:', error)
    }
    return { remaining: 0, canGenerate: false }
  }

  return (
    <AuthContext.Provider value={{ user, isLoading, signIn, signOut, checkGenerationLimit, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return context
}

