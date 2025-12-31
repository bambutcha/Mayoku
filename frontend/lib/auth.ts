import { api } from './api'
import type { AuthResponse, AuthRequest } from '@/types'
import { getTelegramInitData } from './telegram'
import { useAuthStore } from '@/stores/auth'

/**
 * Выполняет аутентификацию через Telegram initData
 */
export async function authenticate(): Promise<AuthResponse> {
  const initData = getTelegramInitData()
  
  if (!initData) {
    throw new Error('Telegram initData не найден')
  }

  const response = await api.post<AuthResponse>('/api/auth', {
    init_data: initData,
  } as AuthRequest)

  // Сохраняем токен и пользователя в store
  const { setAuth } = useAuthStore.getState()
  setAuth(response.token, response.user)

  return response
}

/**
 * Проверяет, аутентифицирован ли пользователь
 */
export function isAuthenticated(): boolean {
  return useAuthStore.getState().isAuthenticated()
}

/**
 * Выход из системы
 */
export function logout(): void {
  const { clearAuth } = useAuthStore.getState()
  clearAuth()
}

