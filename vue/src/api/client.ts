import axios from 'axios'
import router from '../router'

export interface ApiErrorResponse {
  error: string
  code?: string
  message?: string
}

export function getApiErrorMessage(error: unknown, fallback = '请求失败'): string {
  if (axios.isAxiosError<ApiErrorResponse>(error)) {
    const payload = error.response?.data
    return payload?.message || payload?.error || fallback
  }
  return error instanceof Error ? error.message : fallback
}

const client = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    return Promise.reject(error)
  },
)

export default client
