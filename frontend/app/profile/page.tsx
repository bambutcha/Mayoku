'use client'

import { useQuery } from '@tanstack/react-query'
import { Header } from '@/components/layout/Header'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { api } from '@/lib/api'
import type { User } from '@/types'
import Link from 'next/link'

export default function ProfilePage() {
  const { user: authUser } = useAuthStore()

  const { data: user, isLoading } = useQuery<User>({
    queryKey: ['user', 'me'],
    queryFn: () => api.get<User>('/api/user/me'),
    enabled: !!authUser,
  })

  const stats = user ? {
    games_played: user.games_played || 0,
    wins_spy: user.wins_spy || 0,
    wins_local: user.wins_local || 0,
    losses_spy: user.losses_spy || 0,
    losses_local: user.losses_local || 0,
    decks_created: user.decks_created || 0,
    win_rate: user.games_played > 0 
      ? ((user.wins_spy + user.wins_local) / user.games_played) * 100 
      : 0,
  } : {
    games_played: 0,
    wins_spy: 0,
    wins_local: 0,
    losses_spy: 0,
    losses_local: 0,
    decks_created: 0,
    win_rate: 0,
  }

  if (isLoading) {
    return (
      <>
        <Header />
        <main className="container mx-auto px-4 py-12">
          <div className="flex items-center justify-center min-h-[60vh]">
            <div className="text-center space-y-4">
              <div className="text-4xl animate-pulse">🕵️</div>
              <p className="text-muted-foreground">Загрузка...</p>
            </div>
          </div>
        </main>
      </>
    )
  }

  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-12 max-w-4xl">
        <div className="space-y-8">
          {/* Profile Header */}
          <Card variant="glass" className="overflow-hidden">
            <div className="relative h-32 bg-gradient-to-r from-primary/20 to-primary/5">
              <div className="absolute bottom-0 left-1/2 transform -translate-x-1/2 translate-y-1/2">
                <div className="w-24 h-24 rounded-full bg-card border-4 border-card overflow-hidden glow">
                  {user?.avatar_url ? (
                    <img
                      src={user.avatar_url}
                      alt={user.username}
                      className="w-full h-full object-cover"
                    />
                  ) : (
                    <div className="w-full h-full flex items-center justify-center text-4xl bg-gradient-to-br from-primary to-primary/60">
                      {user?.username?.[0]?.toUpperCase() || '?'}
                    </div>
                  )}
                </div>
              </div>
            </div>
            <CardContent className="pt-16 pb-6 text-center">
              <h1 className="text-3xl font-bold mb-2">{user?.username || 'Пользователь'}</h1>
              <p className="text-muted-foreground">
                ID: {user?.tg_id || 'N/A'}
              </p>
              {user?.is_admin && (
                <div className="mt-4 inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-primary/20 text-primary text-sm font-medium">
                  {user.is_super_admin ? '👑 Супер-админ' : '🛡️ Админ'}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Statistics Grid */}
          <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
            <Card variant="glass" className="hover-lift">
              <CardContent className="p-6 text-center space-y-2">
                <div className="text-3xl font-bold gradient-text">{stats.games_played}</div>
                <div className="text-sm text-muted-foreground">Игр сыграно</div>
              </CardContent>
            </Card>

            <Card variant="glass" className="hover-lift">
              <CardContent className="p-6 text-center space-y-2">
                <div className="text-3xl font-bold gradient-text">{Math.round(stats.win_rate)}%</div>
                <div className="text-sm text-muted-foreground">Винрейт</div>
              </CardContent>
            </Card>

            <Card variant="glass" className="hover-lift">
              <CardContent className="p-6 text-center space-y-2">
                <div className="text-3xl font-bold gradient-text">{stats.wins_spy + stats.wins_local}</div>
                <div className="text-sm text-muted-foreground">Побед</div>
              </CardContent>
            </Card>

            <Card variant="glass" className="hover-lift">
              <CardContent className="p-6 text-center space-y-2">
                <div className="text-3xl font-bold text-primary">{stats.wins_spy}</div>
                <div className="text-sm text-muted-foreground">Побед как шпион</div>
              </CardContent>
            </Card>

            <Card variant="glass" className="hover-lift">
              <CardContent className="p-6 text-center space-y-2">
                <div className="text-3xl font-bold text-primary">{stats.wins_local}</div>
                <div className="text-sm text-muted-foreground">Побед как местный</div>
              </CardContent>
            </Card>

            <Card variant="glass" className="hover-lift">
              <CardContent className="p-6 text-center space-y-2">
                <div className="text-3xl font-bold">{stats.decks_created}</div>
                <div className="text-sm text-muted-foreground">Колод создано</div>
              </CardContent>
            </Card>
          </div>

          {/* Actions */}
          <div className="flex flex-col sm:flex-row gap-4">
            <Link href="/lobby" className="flex-1">
              <Button variant="primary" size="lg" className="w-full hover-lift">
                Найти игру
              </Button>
            </Link>
            <Link href="/deck-builder" className="flex-1">
              <Button variant="secondary" size="lg" className="w-full hover-lift">
                Создать колоду
              </Button>
            </Link>
          </div>
        </div>
      </main>
    </>
  )
}

