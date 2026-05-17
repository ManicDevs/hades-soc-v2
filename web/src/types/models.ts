// Shared frontend models

export interface User {
  id: number | string
  username: string
  email?: string
  role?: string
  permissions?: string[]
  status?: 'active' | 'inactive' | string
  environment?: string
  [key: string]: unknown
}

export interface LoginCredentials {
  username?: string
  password?: string
  token?: string
  user?: User | string
  role?: string
  [key: string]: unknown
}

export interface AuthResponse {
  token?: string
  user?: User
  [key: string]: unknown
}

export interface UserFilters {
  search?: string
  role?: string
  status?: string
}

export interface UserStats {
  total_users: number
  active_users: number
  inactive_users: number
  by_role: Record<string, number>
  by_status: Record<string, number>
}
