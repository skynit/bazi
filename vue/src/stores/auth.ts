import { defineStore } from 'pinia'
import { ref } from 'vue'
import {
  fetchMe as fetchCurrentUser,
  login as loginApi,
  register as registerApi,
} from '../api/auth'

export interface User {
  id: number
  username: string
  email: string
  created_at?: string
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
    const data = await loginApi(username, password)
    setToken(data.token)
    await fetchMe()
  }

  async function register(username: string, email: string, password: string) {
    const data = await registerApi(username, email, password)
    setToken(data.token)
    user.value = data.user || null
    if (!user.value) {
      await fetchMe()
    }
  }

  function logout() {
    clearToken()
    user.value = null
    localStorage.removeItem('bazi_last_birth')
    sessionStorage.removeItem('lastChart')
  }

  async function fetchMe() {
    user.value = await fetchCurrentUser()
  }

  const isLoggedIn = () => !!token.value

  return { token, user, login, register, logout, fetchMe, isLoggedIn }
})
