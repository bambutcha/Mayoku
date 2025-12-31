import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types'

interface AuthState {
  token: string | null
  user: User | null
  setAuth: (token: string, user: User) => void
  clearAuth: () => void
  isAuthenticated: () => boolean
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: null,
      user: null,
      setAuth: (token, user) => {
        set({ token, user })
        if (typeof window !== 'undefined') {
          localStorage.setItem('jwt_token', token)
        }
      },
      clearAuth: () => {
        set({ token: null, user: null })
        if (typeof window !== 'undefined') {
          localStorage.removeItem('jwt_token')
        }
      },
      isAuthenticated: () => {
        const { token } = get()
        return token !== null
      },
    }),
    {
      name: 'auth-storage',
    }
  )
)



