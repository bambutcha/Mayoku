'use client'

import { useParams, useRouter } from 'next/navigation'
import { Header } from '@/components/layout/Header'
import { useGameWebSocket } from '@/hooks/useGameWebSocket'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { useAuthStore } from '@/stores/auth'
import { useState, useMemo } from 'react'
import type { PlayerRole } from '@/types/game'

export default function GamePage() {
  const params = useParams()
  const router = useRouter()
  const roomId = params.roomId as string
  const { user } = useAuthStore()
  const { roomState, myRole, error, isConnected, sendMessage } = useGameWebSocket(roomId)
  const [votingAnswer, setVotingAnswer] = useState<boolean | null>(null)
  const [spyGuess, setSpyGuess] = useState('')

  const players = useMemo(() => {
    if (!roomState) return []
    return Object.values(roomState.players)
  }, [roomState])

  const currentPlayer = useMemo(() => {
    if (!roomState || !user) return null
    return roomState.players[user.id]
  }, [roomState, user])

  const isRoomAdmin = useMemo(() => {
    return roomState?.created_by === user?.id
  }, [roomState, user])

  const timeLeft = useMemo(() => {
    if (!roomState?.timer_end) return null
    const now = Math.floor(Date.now() / 1000)
    const left = roomState.timer_end - now
    return left > 0 ? left : 0
  }, [roomState])

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60)
    const secs = seconds % 60
    return `${mins}:${secs.toString().padStart(2, '0')}`
  }

  const handleSetReady = (ready: boolean) => {
    sendMessage('set_ready', { ready })
  }

  const handleVoteStart = (targetUserId: number) => {
    sendMessage('vote_start', { target_user_id: targetUserId })
  }

  const handleVoteAnswer = (vote: boolean) => {
    sendMessage('vote_answer', { vote })
    setVotingAnswer(vote)
  }

  const handleSpyGuess = () => {
    if (!spyGuess.trim()) return
    sendMessage('spy_guess', { location: spyGuess })
    setSpyGuess('')
  }

  const handleKickPlayer = (targetUserId: number) => {
    if (confirm('Вы уверены, что хотите исключить этого игрока?')) {
      sendMessage('kick_player', { target_user_id: targetUserId })
    }
  }

  if (error) {
    return (
      <>
        <Header />
        <main className="container mx-auto px-4 py-12">
          <Card variant="glass" className="max-w-md mx-auto">
            <CardContent className="p-6 text-center space-y-4">
              <div className="text-4xl">❌</div>
              <h2 className="text-xl font-semibold">Ошибка</h2>
              <p className="text-muted-foreground">{error}</p>
              <Button onClick={() => router.push('/lobby')}>Вернуться в лобби</Button>
            </CardContent>
          </Card>
        </main>
      </>
    )
  }

  if (!roomState) {
    return (
      <>
        <Header />
        <main className="container mx-auto px-4 py-12">
          <div className="flex items-center justify-center min-h-[60vh]">
            <div className="text-center space-y-4">
              <div className="text-4xl animate-pulse">🕵️</div>
              <p className="text-muted-foreground">
                {isConnected ? 'Загрузка комнаты...' : 'Подключение...'}
              </p>
            </div>
          </div>
        </main>
      </>
    )
  }

  return (
    <>
      <Header />
      <main className="container mx-auto px-4 py-12 max-w-6xl">
        <div className="space-y-6">
          {/* Room Header */}
          <Card variant="glass">
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-2xl">{roomState.deck_name}</CardTitle>
                  <p className="text-sm text-muted-foreground mt-1">
                    Комната: {roomState.room_id}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
                  <span className="text-sm text-muted-foreground">
                    {isConnected ? 'Подключено' : 'Отключено'}
                  </span>
                </div>
              </div>
            </CardHeader>
            {timeLeft !== null && roomState.status === 'playing' && (
              <CardContent>
                <div className="text-center">
                  <div className="text-4xl font-bold gradient-text mb-2">
                    {formatTime(timeLeft)}
                  </div>
                  <p className="text-sm text-muted-foreground">Осталось времени</p>
                </div>
              </CardContent>
            )}
          </Card>

          {/* Game Status */}
          {roomState.status === 'waiting' && (
            <Card variant="glass">
              <CardHeader>
                <CardTitle>Ожидание игроков</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {players.map((player) => (
                    <Card key={player.user_id} variant="elevated" className="hover-lift">
                      <CardContent className="p-4 text-center">
                        <div className="w-16 h-16 rounded-full mx-auto mb-2 overflow-hidden bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
                          {player.avatar_url ? (
                            <img src={player.avatar_url} alt={player.username} className="w-full h-full object-cover" />
                          ) : (
                            <span className="text-2xl">{player.username[0]?.toUpperCase()}</span>
                          )}
                        </div>
                        <p className="font-medium text-sm">{player.username}</p>
                        <div className="mt-2">
                          {player.is_ready ? (
                            <span className="text-xs px-2 py-1 rounded-full bg-green-500/20 text-green-500">Готов</span>
                          ) : (
                            <span className="text-xs px-2 py-1 rounded-full bg-gray-500/20 text-gray-500">Не готов</span>
                          )}
                        </div>
                        {isRoomAdmin && player.user_id !== user?.id && (
                          <Button
                            variant="ghost"
                            size="sm"
                            className="mt-2 w-full"
                            onClick={() => handleKickPlayer(player.user_id)}
                          >
                            Исключить
                          </Button>
                        )}
                      </CardContent>
                    </Card>
                  ))}
                </div>
                <div className="flex justify-center gap-4 pt-4">
                  <Button
                    onClick={() => handleSetReady(!currentPlayer?.is_ready)}
                    variant={currentPlayer?.is_ready ? 'secondary' : 'primary'}
                    size="lg"
                    className="hover-lift"
                  >
                    {currentPlayer?.is_ready ? 'Не готов' : 'Готов'}
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Playing State */}
          {roomState.status === 'playing' && (
            <div className="grid md:grid-cols-2 gap-6">
              <Card variant="glass">
                <CardHeader>
                  <CardTitle>Ваша роль</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  {myRole ? (
                    <>
                      {myRole.role === 'spy' ? (
                        <div className="text-center space-y-4">
                          <div className="text-6xl mb-4">🕵️</div>
                          <h3 className="text-2xl font-bold text-red-500">Вы - Шпион!</h3>
                          <p className="text-muted-foreground">
                            Ваша задача - угадать локацию или дождаться окончания времени
                          </p>
                          <div className="mt-6 space-y-2">
                            <input
                              type="text"
                              value={spyGuess}
                              onChange={(e) => setSpyGuess(e.target.value)}
                              placeholder="Угадайте локацию..."
                              className="w-full px-4 py-2 rounded-lg bg-background border border-input"
                            />
                            <Button onClick={handleSpyGuess} className="w-full hover-lift">
                              Угадать
                            </Button>
                          </div>
                        </div>
                      ) : (
                        <div className="text-center space-y-4">
                          <div className="text-6xl mb-4">🏠</div>
                          <h3 className="text-2xl font-bold text-green-500">Вы - Местный!</h3>
                          <p className="text-lg font-semibold">{myRole.location}</p>
                          {myRole.location_role && (
                            <p className="text-muted-foreground">Роль: {myRole.location_role}</p>
                          )}
                          {roomState.location && (
                            <div className="mt-4">
                              <img
                                src={roomState.location.image_url}
                                alt={myRole.location}
                                className="w-full rounded-lg"
                              />
                            </div>
                          )}
                        </div>
                      )}
                    </>
                  ) : (
                    <p className="text-muted-foreground">Загрузка роли...</p>
                  )}
                </CardContent>
              </Card>

              <Card variant="glass">
                <CardHeader>
                  <CardTitle>Игроки ({players.length}/{roomState.max_players})</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="space-y-2">
                    {players.map((player) => (
                      <div
                        key={player.user_id}
                        className="flex items-center gap-3 p-3 rounded-lg bg-card/50"
                      >
                        <div className="w-10 h-10 rounded-full overflow-hidden bg-gradient-to-br from-primary to-primary/60 flex items-center justify-center">
                          {player.avatar_url ? (
                            <img src={player.avatar_url} alt={player.username} className="w-full h-full object-cover" />
                          ) : (
                            <span>{player.username[0]?.toUpperCase()}</span>
                          )}
                        </div>
                        <div className="flex-1">
                          <p className="font-medium">{player.username}</p>
                        </div>
                        {roomState.status === 'finished' && player.role && (
                          <span className={`text-xs px-2 py-1 rounded-full ${
                            player.role === 'spy' ? 'bg-red-500/20 text-red-500' : 'bg-green-500/20 text-green-500'
                          }`}>
                            {player.role === 'spy' ? 'Шпион' : 'Местный'}
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                  {myRole?.role === 'local' && (
                    <div className="pt-4 border-t border-border">
                      <p className="text-sm text-muted-foreground mb-3">Подозреваете кого-то? Начните голосование:</p>
                      <div className="grid grid-cols-2 gap-2">
                        {players
                          .filter(p => p.user_id !== user?.id)
                          .map((player) => (
                            <Button
                              key={player.user_id}
                              variant="secondary"
                              size="sm"
                              onClick={() => handleVoteStart(player.user_id)}
                              className="hover-lift"
                            >
                              Голосовать за {player.username}
                            </Button>
                          ))}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          )}

          {/* Voting State */}
          {roomState.status === 'voting' && roomState.voting && (
            <Card variant="glass">
              <CardHeader>
                <CardTitle>Голосование</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-center text-lg">
                  Голосуем за исключение игрока:{' '}
                  <span className="font-semibold">
                    {players.find(p => p.user_id === roomState.voting!.target_user_id)?.username}
                  </span>
                </p>
                {!currentPlayer?.is_voted && (
                  <div className="flex gap-4 justify-center">
                    <Button
                      onClick={() => handleVoteAnswer(true)}
                      variant="primary"
                      size="lg"
                      className="hover-lift"
                    >
                      За
                    </Button>
                    <Button
                      onClick={() => handleVoteAnswer(false)}
                      variant="secondary"
                      size="lg"
                      className="hover-lift"
                    >
                      Против
                    </Button>
                  </div>
                )}
                <div className="mt-4">
                  <p className="text-sm text-muted-foreground mb-2">Голоса:</p>
                  <div className="space-y-2">
                    {players.map((player) => (
                      <div key={player.user_id} className="flex items-center justify-between p-2 rounded bg-card/50">
                        <span>{player.username}</span>
                        {player.is_voted !== undefined ? (
                          <span className={`text-sm ${player.vote ? 'text-green-500' : 'text-red-500'}`}>
                            {player.vote ? '✓ За' : '✗ Против'}
                          </span>
                        ) : (
                          <span className="text-sm text-muted-foreground">Ожидание...</span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Finished State */}
          {roomState.status === 'finished' && (
            <Card variant="glass">
              <CardHeader>
                <CardTitle>Игра завершена</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="text-center">
                  <div className="text-6xl mb-4">
                    {roomState.winner === 'spy' ? '🕵️' : '🏠'}
                  </div>
                  <h3 className="text-3xl font-bold mb-2">
                    {roomState.winner === 'spy' ? 'Победили шпионы!' : 'Победили местные!'}
                  </h3>
                  {roomState.location && (
                    <div className="mt-4">
                      <p className="text-lg font-semibold mb-2">Локация была:</p>
                      <p className="text-2xl">{roomState.location.name}</p>
                      <img
                        src={roomState.location.image_url}
                        alt={roomState.location.name}
                        className="w-full max-w-md mx-auto rounded-lg mt-4"
                      />
                    </div>
                  )}
                </div>
                <div className="flex justify-center gap-4 pt-4">
                  <Button onClick={() => router.push('/lobby')} size="lg" className="hover-lift">
                    Вернуться в лобби
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </main>
    </>
  )
}

