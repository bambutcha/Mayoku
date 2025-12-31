'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Header } from '@/components/layout/Header'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { api } from '@/lib/api'
import type { Deck, CreateRoomRequest, CreateRoomResponse } from '@/types'
import Link from 'next/link'

interface Room {
  id: string
  deck_name: string
  players_count: number
  max_players: number
  status: string
}

export default function LobbyPage() {
  const [selectedDeck, setSelectedDeck] = useState<number | null>(null)
  const queryClient = useQueryClient()

  const { data: decks, isLoading: decksLoading } = useQuery<Deck[]>({
    queryKey: ['decks'],
    queryFn: () => api.get<Deck[]>('/api/decks?status=approved'),
  })

  const { data: rooms, isLoading: roomsLoading } = useQuery<Room[]>({
    queryKey: ['rooms'],
    queryFn: () => api.get<Room[]>('/api/game/rooms'),
    refetchInterval: 3000, // Обновляем каждые 3 секунды
  })

  const createRoomMutation = useMutation({
    mutationFn: (data: CreateRoomRequest) =>
      api.post<CreateRoomResponse>('/api/game/rooms', data),
    onSuccess: (data) => {
      // Перенаправляем в комнату
      window.location.href = `/game/${data.room_id}`
    },
  })

  const handleCreateRoom = () => {
    if (!selectedDeck) return
    createRoomMutation.mutate({ deck_id: selectedDeck })
  }

  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-12 max-w-6xl">
        <div className="space-y-8">
          {/* Header */}
          <div className="text-center space-y-4">
            <h1 className="text-4xl md:text-5xl font-bold gradient-text">
              Игровое лобби
            </h1>
            <p className="text-muted-foreground text-lg">
              Выберите колоду и начните игру
            </p>
          </div>

          <div className="grid lg:grid-cols-3 gap-8">
            {/* Decks Selection */}
            <div className="lg:col-span-2 space-y-6">
              <Card variant="glass">
                <CardHeader>
                  <CardTitle>Выберите колоду</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {decksLoading ? (
                    <div className="text-center py-8 text-muted-foreground">
                      Загрузка колод...
                    </div>
                  ) : !decks || decks.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">
                      Нет доступных колод
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      {decks.map((deck) => (
                        <button
                          key={deck.id}
                          onClick={() => setSelectedDeck(deck.id)}
                          className={`p-4 rounded-xl border-2 transition-all duration-200 hover-lift text-left ${
                            selectedDeck === deck.id
                              ? 'border-primary bg-primary/10 glow'
                              : 'border-border bg-card/50 hover:border-primary/50'
                          }`}
                        >
                          <h3 className="font-semibold text-lg mb-1">{deck.name}</h3>
                          <p className="text-sm text-muted-foreground">
                            {deck.locations?.length || 0} локаций
                          </p>
                          {deck.author && (
                            <p className="text-xs text-muted-foreground mt-2">
                              Автор: {deck.author.username}
                            </p>
                          )}
                        </button>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>

              <div className="flex gap-4">
                <Button
                  onClick={handleCreateRoom}
                  disabled={!selectedDeck || createRoomMutation.isPending}
                  size="lg"
                  className="flex-1 hover-lift glow"
                >
                  {createRoomMutation.isPending ? 'Создание...' : 'Создать комнату'}
                </Button>
                <Link href="/deck-builder">
                  <Button variant="secondary" size="lg" className="hover-lift">
                    Создать колоду
                  </Button>
                </Link>
              </div>
            </div>

            {/* Active Rooms */}
            <div className="space-y-6">
              <Card variant="glass">
                <CardHeader>
                  <CardTitle>Активные комнаты</CardTitle>
                </CardHeader>
                <CardContent>
                  {roomsLoading ? (
                    <div className="text-center py-4 text-muted-foreground text-sm">
                      Загрузка...
                    </div>
                  ) : !rooms || rooms.length === 0 ? (
                    <div className="text-center py-8 text-muted-foreground">
                      <div className="text-4xl mb-2">🎮</div>
                      <p className="text-sm">Нет активных комнат</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {rooms.map((room) => (
                        <Link
                          key={room.id}
                          href={`/game/${room.id}`}
                          className="block"
                        >
                          <Card variant="elevated" className="hover-lift cursor-pointer">
                            <CardContent className="p-4">
                              <div className="flex items-center justify-between mb-2">
                                <h3 className="font-semibold">{room.deck_name}</h3>
                                <span className="text-xs px-2 py-1 rounded-full bg-primary/20 text-primary">
                                  {room.status}
                                </span>
                              </div>
                              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                                <span>👥 {room.players_count}/{room.max_players}</span>
                              </div>
                            </CardContent>
                          </Card>
                        </Link>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </main>
    </>
  )
}

