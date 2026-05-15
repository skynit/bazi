import { defineStore } from 'pinia'
import { ref } from 'vue'
import client from '../api/client'

export interface User {
  id: number
  username: string
  email: string
  created_at: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>(localStorage.getItem('token') || '')
  const user = ref<User | null>(null)

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('token')
  }

  async function login(username: string, password: string) {
    const res = await client.post('/auth/login', { username, password })
    console.log('Auth store login response:', res.data)
    setToken(res.data.token)
    user.value = res.data.user
  }

  async function register(username: string, email: string, password: string) {
    const res = await client.post('/auth/register', { username, email, password })
    setToken(res.data.token)
    user.value = res.data.user
  }

  function logout() {
    clearToken()
    user.value = null
  }

  async function fetchMe() {
    const res = await client.get('/auth/me')
    user.value = res.data.user
  }

  const isLoggedIn = () => !!token.value

  return { token, user, login, register, logout, fetchMe, isLoggedIn }
})
