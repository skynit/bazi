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
  'username must be 3-32 characters and contain only letters, numbers, _, -, or .':
    '用户名需为 3–32 位，只能包含字母、数字、下划线、短横线或点。',
  'invalid email address': '邮箱格式不正确，请检查后重试。',
  'password must be at least 8 characters': '密码至少需要 8 个字符。',
  'chart not found': '命盘不存在或已被删除。',
  'user not found': '账户不存在。',
  unauthorized: '登录状态已过期，请重新登录。',
}

function translateValidationMessage(rawMessage: string): string | undefined {
  if (rawMessage.startsWith('invalid solar date:'))
    return '输入的公历日期不存在，请检查年、月、日。'
  if (rawMessage.startsWith('invalid lunar date:'))
    return '输入的农历日期无效，请检查日期和闰月选项。'
  if (rawMessage.startsWith('invalid timezone'))
    return '时区名称无效，请使用例如 Asia/Shanghai 的地区时区。'
  if (rawMessage.startsWith('birth time does not exist in timezone')) {
    return '该出生时间处于夏令时跳时区间，在所选时区中不存在，请调整时间或时区。'
  }
  if (rawMessage.startsWith('birth time is ambiguous in timezone')) {
    return '该出生时间在夏令时切换时重复出现，请在高级设置中选择对应的 UTC 偏移。'
  }
  if (rawMessage.startsWith('birth_utc_offset_seconds must be one of')) {
    return 'UTC 偏移与该出生时间不匹配，请按提示选择有效偏移。'
  }
  if (rawMessage.includes('longitude between -180 and 180')) {
    return '启用真太阳时后，经度必须填写为 -180 到 180 之间的数字。'
  }
  if (rawMessage === 'birth time is out of range') return '出生时间超出有效范围，请检查时、分、秒。'
  if (rawMessage.startsWith('uncertainty_seconds must be between')) {
    return '出生时间误差必须在 1 秒到 24 小时之间。'
  }
  if (rawMessage.startsWith('calendar_type must be')) return '历法类型无效，请选择公历或农历。'
  if (rawMessage.startsWith('gender must be')) return '性别选项无效，请重新选择。'
  if (rawMessage.startsWith('lunar_leap_month is only valid')) {
    return '闰月选项仅适用于农历日期。'
  }
  if (rawMessage.startsWith('birth-time interval crosses a four-pillar boundary')) {
    return '出生时间范围跨越排盘边界，请先预览并选择一个候选命盘。'
  }
  if (rawMessage.startsWith('candidate_id does not match')) {
    return '候选命盘已失效，请重新校验出生信息后再选择。'
  }
  return undefined
}

export function getApiErrorMessage(error: unknown, fallback = '请求失败'): string {
  if (axios.isAxiosError<ApiErrorResponse>(error)) {
    const payload = error.response?.data
    const rawMessage = String(payload?.message || payload?.error || '')
      .trim()
      .toLowerCase()
    if (rawMessage && knownApiMessages[rawMessage]) return knownApiMessages[rawMessage]
    const validationMessage = translateValidationMessage(rawMessage)
    if (validationMessage) return validationMessage
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
      window.dispatchEvent(new Event('auth:expired'))
      const currentPath = router.currentRoute.value.fullPath
      if (router.currentRoute.value.path !== '/login') {
        router.push({ path: '/login', query: { redirect: currentPath, expired: '1' } })
      }
    }
    return Promise.reject(error)
  },
)

export default client
