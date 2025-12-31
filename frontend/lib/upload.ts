import { api } from './api'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

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

  const response = await fetch(`${API_BASE_URL}/api/upload`, {
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
    // В продакшене нужно настроить правильный URL
    return `${API_BASE_URL.replace(':8080', ':9000')}${data.url}`
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

