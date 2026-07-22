import axios from 'axios'
import router from '../router'

export interface ApiErrorResponse {
  error: string
  code?: string
  message?: string
}

const apiCodeMessages: Record<string, string> = {
  UNAUTHORIZED: '登录状态已过期，请重新登录。',
  FORBIDDEN: '当前账户没有执行此操作的权限。',
  NOT_FOUND: '未找到所需内容，它可能已被删除。',
  CONFLICT: '当前信息已存在，请检查后重试。',
  SERVICE_DISABLED: '该功能暂时不可用，请稍后再试。',
  SERVICE_ERROR: '服务暂时不可用，请稍后重试。',
}

const knownApiMessages: Record<string, string> = {
  'invalid username or password': '用户名或密码不正确。',
  'username already exists': '该用户名已被使用，请换一个。',
  'invalid email address': '邮箱格式不正确，请检查后重试。',
  'password must be at least 8 characters': '密码至少需要 8 个字符。',
  'chart not found': '命盘不存在或已被删除。',
  'user not found': '账户不存在。',
  unauthorized: '登录状态已过期，请重新登录。',
}

export function getApiErrorMessage(error: unknown, fallback = '请求失败'): string {
  if (axios.isAxiosError<ApiErrorResponse>(error)) {
    const payload = error.response?.data
    const rawMessage = String(payload?.message || payload?.error || '').trim().toLowerCase()
    if (rawMessage && knownApiMessages[rawMessage]) return knownApiMessages[rawMessage]
    if (payload?.code && apiCodeMessages[payload.code]) return apiCodeMessages[payload.code]
    if (error.code === 'ECONNABORTED') return '请求超时，请稍后重试。'
    if (!error.response) return '无法连接服务器，请检查网络后重试。'
    return fallback
  }
  return fallback
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
