// Game types
export type GameStatus = 'waiting' | 'playing' | 'voting' | 'finished'
export type PlayerRole = 'spy' | 'local'

export interface Player {
  user_id: number
  tg_id: number
  username: string
  avatar_url: string
  role?: PlayerRole
  location?: string
  location_role?: string
  is_ready: boolean
  is_voted?: boolean
  vote?: boolean
}

export interface LocationInfo {
  name: string
  image_url: string
  roles: string[]
}

export interface VotingState {
  target_user_id: number
  votes: Record<number, boolean>
  started_at: string
}

export interface RoomState {
  room_id: string
  status: GameStatus
  players: Record<number, Player>
  location?: LocationInfo
  spy_ids?: number[]
  timer_end?: number
  voting?: VotingState
  winner?: 'spy' | 'locals'
  deck_id: number
  deck_name: string
  max_players: number
  spy_count: number
  duration: number
  created_by: number
  created_at: string
}

export interface WSMessage {
  type: string
  payload: unknown
}

export interface GameStartedPayload {
  room_id: string
  my_role: {
    role: PlayerRole
    location?: string
    location_role?: string
  }
  timer_end: number
  spy_count: number
}

// Client messages
export interface JoinRoomPayload {
  room_id: string
}

export interface SetReadyPayload {
  ready: boolean
}

export interface VoteStartPayload {
  target_user_id: number
}

export interface VoteAnswerPayload {
  vote: boolean // true = за, false = против
}

export interface SpyGuessPayload {
  location: string
}

export interface KickPlayerPayload {
  target_user_id: number
}

