// User types
export interface User {
  id: number
  tg_id: number
  username: string
  avatar_url: string
  created_at: string
  updated_at: string
  is_admin: boolean
  is_super_admin: boolean
  games_played: number
  wins_spy: number
  wins_local: number
  losses_spy: number
  losses_local: number
  decks_created: number
}

// Auth types
export interface AuthRequest {
  init_data: string
}

export interface AuthResponse {
  token: string
  user: User
}

// Deck types
export interface Location {
  id: number
  deck_id: number
  name: string
  image_url: string
  roles: string[]
}

export interface Deck {
  id: number
  author_id: number
  name: string
  is_public: boolean
  status: 'draft' | 'pending' | 'approved' | 'rejected'
  created_at: string
  updated_at: string
  locations?: Location[]
  author?: User
}

// Game types
export interface CreateRoomRequest {
  deck_id: number
  max_players?: number
  spy_count?: number
  duration?: number
}

export interface CreateRoomResponse {
  room_id: string
}

// WebSocket types
export interface WSMessage {
  type: string
  payload: unknown
}

export interface ClientMessage {
  type: string
  payload: unknown
}



