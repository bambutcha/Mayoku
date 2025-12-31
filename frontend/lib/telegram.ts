import WebApp from '@twa-dev/sdk'

/**
 * Получает initData от Telegram Mini App
 */
export function getTelegramInitData(): string | null {
  if (typeof window === 'undefined') return null
  
  try {
    return WebApp.initData || null
  } catch (error) {
    console.error('Failed to get Telegram initData:', error)
    return null
  }
}

/**
 * Получает пользователя из initData (без валидации, только для UI)
 */
export function getTelegramUser() {
  if (typeof window === 'undefined') return null
  
  try {
    return WebApp.initDataUnsafe?.user || null
  } catch (error) {
    console.error('Failed to get Telegram user:', error)
    return null
  }
}

/**
 * Инициализирует Telegram Mini App SDK
 */
export function initTelegramSDK() {
  if (typeof window === 'undefined') return
  
  try {
    // SDK автоматически инициализируется при импорте
    WebApp.ready()
    console.log('Telegram Mini App SDK initialized')
  } catch (error) {
    console.error('Failed to initialize Telegram SDK:', error)
  }
}

