import axios from 'axios'
import { describe, expect, it } from 'vitest'
import { getApiErrorMessage } from './client'

function apiError(data: unknown, status = 400) {
  return new axios.AxiosError(
    'Request failed',
    'ERR_BAD_REQUEST',
    undefined,
    undefined,
    { data, status, statusText: 'Bad Request', headers: {}, config: {} as never },
  )
}

describe('getApiErrorMessage', () => {
  it('maps known backend messages to readable Chinese', () => {
    expect(getApiErrorMessage(apiError({ error: 'invalid username or password' }))).toBe(
      '用户名或密码不正确。',
    )
  })

  it('maps API codes without exposing backend details', () => {
    expect(
      getApiErrorMessage(apiError({ code: 'SERVICE_ERROR', error: 'database connection refused' })),
    ).toBe('服务暂时不可用，请稍后重试。')
  })

  it('uses the contextual fallback for unknown server messages', () => {
    expect(getApiErrorMessage(apiError({ error: 'internal_rule_id=fortune.v2' }), '排盘失败')).toBe(
      '排盘失败',
    )
  })
})
