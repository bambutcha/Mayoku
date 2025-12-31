'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { authenticate, isAuthenticated } from '@/lib/auth'
import { initTelegramSDK } from '@/lib/telegram'

interface AuthGuardProps {
  children: React.ReactNode
}

export function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter()
  const [isLoading, setIsLoading] = useState(true)
  const [isAuth, setIsAuth] = useState(false)

  useEffect(() => {
    async function checkAuth() {
      try {
        // Инициализируем Telegram SDK
        initTelegramSDK()

        // Проверяем, есть ли уже токен
        if (isAuthenticated()) {
          setIsAuth(true)
          setIsLoading(false)
          return
        }

        // Пытаемся аутентифицироваться через Telegram
        await authenticate()
        setIsAuth(true)
      } catch (error) {
        console.error('Authentication failed:', error)
        // В реальном приложении можно показать ошибку
        // Для разработки разрешаем доступ
        setIsAuth(true)
      } finally {
        setIsLoading(false)
      }
    }

    checkAuth()
  }, [])

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="mb-4 text-2xl">🕵️</div>
          <p className="text-lg">Загрузка...</p>
        </div>
      </div>
    )
  }

  if (!isAuth) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="mb-4 text-2xl">❌</div>
          <p className="text-lg">Ошибка аутентификации</p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}

