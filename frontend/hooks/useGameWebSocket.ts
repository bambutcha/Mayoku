'use client'

import { useEffect, useRef, useState, useCallback } from 'react'
import { useAuthStore } from '@/stores/auth'
import type { WSMessage, RoomState, GameStartedPayload } from '@/types/game'

// WebSocket URL - in production use wss://, in development ws://
const getWebSocketURL = (roomId: string): string => {
  if (typeof window === 'undefined') return ''
  
  const isProduction = window.location.protocol === 'https:'
  const protocol = isProduction ? 'wss:' : 'ws:'
  const host = window.location.host
  
  // In Docker, nginx proxies WebSocket to backend
  // In development, connect directly to backend
  if (host === 'localhost' || host.includes('localhost')) {
    return `ws://localhost:8080/api/game/ws?room_id=${roomId}`
  }
  
  return `${protocol}//${host}/api/game/ws?room_id=${roomId}`
}

export function useGameWebSocket(roomId: string) {
  const [roomState, setRoomState] = useState<RoomState | null>(null)
  const [myRole, setMyRole] = useState<GameStartedPayload['my_role'] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const { token, user } = useAuthStore()

  const connect = useCallback(() => {
    if (!token || !user) {
      setError('Необходима авторизация')
      return
    }

    try {
      const wsUrl = getWebSocketURL(roomId)
      const ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        setIsConnected(true)
        setError(null)
        
        // Отправляем сообщение о присоединении
        ws.send(JSON.stringify({
          type: 'join_room',
          payload: { room_id: roomId }
        }))
      }

      ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)

          switch (message.type) {
            case 'room_state':
              setRoomState(message.payload as RoomState)
              break

            case 'game_started':
              const gameStarted = message.payload as GameStartedPayload
              setMyRole(gameStarted.my_role)
              setRoomState((prev) => {
                if (!prev) return null
                return { ...prev, status: 'playing', timer_end: gameStarted.timer_end }
              })
              break

            case 'error':
              setError((message.payload as { message: string }).message)
              break

            case 'player_joined':
            case 'player_left':
            case 'voting_started':
            case 'voting_result':
            case 'game_finished':
              // Обновляем состояние при любых изменениях
              if (message.payload && typeof message.payload === 'object' && 'room_id' in message.payload) {
                setRoomState(message.payload as RoomState)
              }
              break

            default:
              console.log('Unknown message type:', message.type)
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }

      ws.onerror = (err) => {
        console.error('WebSocket error:', err)
        setError('Ошибка подключения')
        setIsConnected(false)
      }

      ws.onclose = () => {
        setIsConnected(false)
        
        // Автоматическое переподключение через 3 секунды
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current)
        }
        reconnectTimeoutRef.current = setTimeout(() => {
          connect()
        }, 3000)
      }

      wsRef.current = ws
    } catch (err) {
      console.error('Failed to create WebSocket:', err)
      setError('Не удалось подключиться')
    }
  }, [roomId, token, user])

  const sendMessage = useCallback((type: string, payload: unknown) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload }))
    } else {
      setError('WebSocket не подключен')
    }
  }, [])

  useEffect(() => {
    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (wsRef.current) {
        wsRef.current.close()
      }
    }
  }, [connect])

  return {
    roomState,
    myRole,
    error,
    isConnected,
    sendMessage,
    reconnect: connect,
  }
}

