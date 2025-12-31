import { api } from './api'

// In production (Docker), API is proxied through nginx at /api
// In development, use NEXT_PUBLIC_API_URL or default to localhost:8080
const getAPIBaseURL = (): string => {
  if (typeof window !== 'undefined') {
    // In browser, check if we're in production (nginx proxy) or development
    if (window.location.origin === 'http://localhost:3000' || window.location.origin.includes('localhost:3000')) {
      return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
    }
    // In production, use relative path (nginx proxy)
    return '/api'
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
}

const API_BASE_URL = getAPIBaseURL()

export interface PresignedURLResponse {
  url: string
  key: string
}

/**
 * Загружает файл через multipart/form-data
 */
export async function uploadFile(file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)

  const token = typeof window !== 'undefined' 
    ? localStorage.getItem('jwt_token') 
    : null

  const uploadURL = API_BASE_URL ? `${API_BASE_URL}/api/upload` : '/api/upload'
  const response = await fetch(uploadURL, {
    method: 'POST',
    headers: token ? {
      'Authorization': `Bearer ${token}`,
    } : {},
    body: formData,
  })

  if (!response.ok) {
    const error = await response.json().catch(() => ({}))
    throw new Error(error.error || 'Failed to upload file')
  }

  const data = await response.json()
  // API возвращает относительный путь, нужно преобразовать в полный URL
  // В продакшене это будет presigned URL или публичный URL
  if (data.url.startsWith('/')) {
    // Для локальной разработки используем MinIO напрямую
    if (API_BASE_URL && API_BASE_URL.includes('localhost:8080')) {
      return `${API_BASE_URL.replace(':8080', ':9000')}${data.url}`
    }
    // В production через nginx proxy
    return `http://localhost:9000${data.url}`
  }
  return data.url
}

/**
 * Получает presigned URL для загрузки файла
 */
export async function getPresignedURL(filename: string, contentType: string): Promise<PresignedURLResponse> {
  const response = await api.post<PresignedURLResponse>('/api/upload/presigned', {
    filename,
    content_type: contentType,
  })

  return response
}

