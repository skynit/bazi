import client from './client'

export interface User {
  id: number
  username: string
  email: string
  created_at?: string
}

export interface AuthResponse {
  token: string
  user?: User
}

export async function login(username: string, password: string) {
  const { data } = await client.post<AuthResponse>('/auth/login', { username, password })
  return data
}

export async function register(username: string, email: string, password: string) {
  const { data } = await client.post<AuthResponse>('/auth/register', { username, email, password })
  return data
}

export async function fetchMe() {
  const { data } = await client.get<{ user: User }>('/auth/me')
  return data.user
}
