'use client'

import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Header } from '@/components/layout/Header'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { api } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'next/navigation'
import type { Deck, User } from '@/types'

interface PendingDecksResponse {
  decks: Deck[]
  count: number
}

interface AllDecksResponse {
  decks: Deck[]
  count: number
}

interface AdminsResponse {
  admins: User[]
  count: number
}

export default function AdminPage() {
  const router = useRouter()
  const { user } = useAuthStore()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<'decks' | 'users'>('decks')
  const [statusFilter, setStatusFilter] = useState<string>('pending')
  const [rejectReason, setRejectReason] = useState<Record<number, string>>({})
  const [showRejectModal, setShowRejectModal] = useState<number | null>(null)

  // Проверка прав доступа
  if (!user?.is_admin) {
    return (
      <>
        <Header />
        <main className="container mx-auto px-4 py-12">
          <Card variant="glass" className="max-w-md mx-auto">
            <CardContent className="p-6 text-center space-y-4">
              <div className="text-4xl">🚫</div>
              <h2 className="text-xl font-semibold">Доступ запрещен</h2>
              <p className="text-muted-foreground">
                У вас нет прав для доступа к админ-панели
              </p>
              <Button onClick={() => router.push('/')}>На главную</Button>
            </CardContent>
          </Card>
        </main>
      </>
    )
  }

  // Запросы для колод
  const { data: pendingDecks, isLoading: pendingLoading } = useQuery<PendingDecksResponse>({
    queryKey: ['admin', 'decks', 'pending'],
    queryFn: () => api.get<PendingDecksResponse>('/api/admin/decks/pending'),
    enabled: activeTab === 'decks' && statusFilter === 'pending',
  })

  const { data: allDecks, isLoading: allDecksLoading } = useQuery<AllDecksResponse>({
    queryKey: ['admin', 'decks', 'all', statusFilter],
    queryFn: () => api.get<AllDecksResponse>(`/api/admin/decks?status=${statusFilter}`),
    enabled: activeTab === 'decks' && statusFilter !== 'pending',
  })

  const decks = statusFilter === 'pending' ? pendingDecks?.decks || [] : allDecks?.decks || []
  const isLoadingDecks = statusFilter === 'pending' ? pendingLoading : allDecksLoading

  // Запрос для админов
  const { data: adminsData, isLoading: adminsLoading } = useQuery<AdminsResponse>({
    queryKey: ['admin', 'users', 'admins'],
    queryFn: () => api.get<AdminsResponse>('/api/admin/users/admins'),
    enabled: activeTab === 'users',
  })

  // Мутации для колод
  const approveDeckMutation = useMutation({
    mutationFn: (deckId: number) =>
      api.put<Deck>(`/api/admin/decks/${deckId}/approve`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'decks'] })
    },
  })

  const rejectDeckMutation = useMutation({
    mutationFn: ({ deckId, reason }: { deckId: number; reason: string }) =>
      api.put(`/api/admin/decks/${deckId}/reject`, { reason }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'decks'] })
      setShowRejectModal(null)
      setRejectReason({})
    },
  })

  // Мутации для пользователей
  const makeAdminMutation = useMutation({
    mutationFn: (userId: number) =>
      api.put<User>(`/api/admin/users/${userId}/make-admin`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
    },
  })

  const removeAdminMutation = useMutation({
    mutationFn: (userId: number) =>
      api.put<User>(`/api/admin/users/${userId}/remove-admin`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
    },
  })

  const handleApprove = (deckId: number) => {
    if (confirm('Одобрить эту колоду?')) {
      approveDeckMutation.mutate(deckId)
    }
  }

  const handleReject = (deckId: number) => {
    const reason = rejectReason[deckId]?.trim()
    if (!reason) {
      alert('Укажите причину отклонения')
      return
    }
    if (confirm('Отклонить эту колоду?')) {
      rejectDeckMutation.mutate({ deckId, reason })
    }
  }

  const handleMakeAdmin = (userId: number) => {
    if (confirm('Назначить этого пользователя администратором?')) {
      makeAdminMutation.mutate(userId)
    }
  }

  const handleRemoveAdmin = (userId: number) => {
    if (confirm('Убрать права администратора у этого пользователя?')) {
      removeAdminMutation.mutate(userId)
    }
  }

  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-12 max-w-6xl">
        <div className="space-y-6">
          {/* Header */}
          <div className="text-center space-y-4">
            <h1 className="text-4xl md:text-5xl font-bold gradient-text">
              Админ-панель
            </h1>
            {user.is_super_admin && (
              <p className="text-sm text-primary">👑 Супер-администратор</p>
            )}
          </div>

          {/* Tabs */}
          <div className="flex gap-2 border-b border-border">
            <button
              onClick={() => setActiveTab('decks')}
              className={`px-6 py-3 font-medium transition-colors ${
                activeTab === 'decks'
                  ? 'border-b-2 border-primary text-primary'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              Модерация колод
            </button>
            {user.is_super_admin && (
              <button
                onClick={() => setActiveTab('users')}
                className={`px-6 py-3 font-medium transition-colors ${
                  activeTab === 'users'
                    ? 'border-b-2 border-primary text-primary'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                Управление админами
              </button>
            )}
          </div>

          {/* Decks Tab */}
          {activeTab === 'decks' && (
            <div className="space-y-6">
              {/* Status Filter */}
              <div className="flex gap-2">
                <Button
                  variant={statusFilter === 'pending' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setStatusFilter('pending')}
                  className="hover-lift"
                >
                  На модерации ({pendingDecks?.count || 0})
                </Button>
                <Button
                  variant={statusFilter === 'approved' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setStatusFilter('approved')}
                  className="hover-lift"
                >
                  Одобренные
                </Button>
                <Button
                  variant={statusFilter === 'rejected' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setStatusFilter('rejected')}
                  className="hover-lift"
                >
                  Отклоненные
                </Button>
                <Button
                  variant={statusFilter === '' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setStatusFilter('')}
                  className="hover-lift"
                >
                  Все
                </Button>
              </div>

              {/* Decks List */}
              {isLoadingDecks ? (
                <div className="text-center py-12 text-muted-foreground">
                  Загрузка...
                </div>
              ) : decks.length === 0 ? (
                <Card variant="glass">
                  <CardContent className="p-12 text-center">
                    <div className="text-4xl mb-4">📦</div>
                    <p className="text-muted-foreground">Нет колод для отображения</p>
                  </CardContent>
                </Card>
              ) : (
                <div className="grid gap-4">
                  {decks.map((deck) => (
                    <Card key={deck.id} variant="glass" className="hover-lift">
                      <CardContent className="p-6">
                        <div className="flex items-start justify-between gap-4">
                          <div className="flex-1 space-y-3">
                            <div>
                              <h3 className="text-xl font-semibold">{deck.name}</h3>
                              <p className="text-sm text-muted-foreground">
                                Автор: {deck.author?.username || 'Неизвестен'} • 
                                Локаций: {deck.locations?.length || 0} • 
                                Статус: <span className={`${
                                  deck.status === 'approved' ? 'text-green-500' :
                                  deck.status === 'rejected' ? 'text-red-500' :
                                  'text-yellow-500'
                                }`}>
                                  {deck.status === 'approved' ? 'Одобрена' :
                                   deck.status === 'rejected' ? 'Отклонена' :
                                   'На модерации'}
                                </span>
                              </p>
                            </div>

                            {/* Locations Preview */}
                            {deck.locations && deck.locations.length > 0 && (
                              <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                                {deck.locations.slice(0, 4).map((location) => (
                                  <div key={location.id} className="relative">
                                    <img
                                      src={location.image_url}
                                      alt={location.name}
                                      className="w-full h-20 object-cover rounded-lg"
                                    />
                                    <p className="text-xs mt-1 text-center">{location.name}</p>
                                  </div>
                                ))}
                                {deck.locations.length > 4 && (
                                  <div className="flex items-center justify-center text-muted-foreground">
                                    +{deck.locations.length - 4}
                                  </div>
                                )}
                              </div>
                            )}
                          </div>

                          {/* Actions */}
                          {deck.status === 'pending' && (
                            <div className="flex flex-col gap-2">
                              <Button
                                onClick={() => handleApprove(deck.id)}
                                disabled={approveDeckMutation.isPending}
                                variant="primary"
                                size="sm"
                                className="hover-lift"
                              >
                                Одобрить
                              </Button>
                              <Button
                                onClick={() => setShowRejectModal(deck.id)}
                                variant="danger"
                                size="sm"
                                className="hover-lift"
                              >
                                Отклонить
                              </Button>
                            </div>
                          )}
                        </div>

                        {/* Reject Modal */}
                        {showRejectModal === deck.id && (
                          <div className="mt-4 p-4 bg-card/50 rounded-lg border border-border space-y-3">
                            <label className="text-sm font-medium">Причина отклонения</label>
                            <Input
                              value={rejectReason[deck.id] || ''}
                              onChange={(e) =>
                                setRejectReason({ ...rejectReason, [deck.id]: e.target.value })
                              }
                              placeholder="Укажите причину..."
                              className="w-full"
                            />
                            <div className="flex gap-2">
                              <Button
                                onClick={() => handleReject(deck.id)}
                                disabled={rejectDeckMutation.isPending}
                                variant="danger"
                                size="sm"
                              >
                                Подтвердить отклонение
                              </Button>
                              <Button
                                onClick={() => {
                                  setShowRejectModal(null)
                                  setRejectReason({ ...rejectReason, [deck.id]: '' })
                                }}
                                variant="secondary"
                                size="sm"
                              >
                                Отмена
                              </Button>
                            </div>
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </div>
          )}

          {/* Users Tab */}
          {activeTab === 'users' && user.is_super_admin && (
            <div className="space-y-6">
              {adminsLoading ? (
                <div className="text-center py-12 text-muted-foreground">
                  Загрузка...
                </div>
              ) : !adminsData || adminsData.admins.length === 0 ? (
                <Card variant="glass">
                  <CardContent className="p-12 text-center">
                    <div className="text-4xl mb-4">👥</div>
                    <p className="text-muted-foreground">Нет администраторов</p>
                  </CardContent>
                </Card>
              ) : (
                <div className="grid gap-4">
                  {adminsData.admins.map((admin) => (
                    <Card key={admin.id} variant="glass" className="hover-lift">
                      <CardContent className="p-6">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-4">
                            <div className="w-12 h-12 rounded-full overflow-hidden bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
                              {admin.avatar_url ? (
                                <img
                                  src={admin.avatar_url}
                                  alt={admin.username}
                                  className="w-full h-full object-cover"
                                />
                              ) : (
                                <span className="text-xl">{admin.username[0]?.toUpperCase()}</span>
                              )}
                            </div>
                            <div>
                              <h3 className="font-semibold">{admin.username}</h3>
                              <p className="text-sm text-muted-foreground">
                                ID: {admin.tg_id}
                                {admin.is_super_admin && (
                                  <span className="ml-2 px-2 py-0.5 rounded-full bg-primary/20 text-primary text-xs">
                                    👑 Супер-админ
                                  </span>
                                )}
                                {!admin.is_super_admin && admin.is_admin && (
                                  <span className="ml-2 px-2 py-0.5 rounded-full bg-blue-500/20 text-blue-500 text-xs">
                                    🛡️ Админ
                                  </span>
                                )}
                              </p>
                            </div>
                          </div>
                          {!admin.is_super_admin && (
                            <Button
                              onClick={() => handleRemoveAdmin(admin.id)}
                              disabled={removeAdminMutation.isPending}
                              variant="danger"
                              size="sm"
                              className="hover-lift"
                            >
                              Убрать права
                            </Button>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </main>
    </>
  )
}

