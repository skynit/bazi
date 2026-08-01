import axios from 'axios'
import { describe, expect, it } from 'vitest'
import { getApiErrorMessage } from './client'

function apiError(data: unknown, status = 400) {
  return new axios.AxiosError('Request failed', 'ERR_BAD_REQUEST', undefined, undefined, {
    data,
    status,
    statusText: 'Bad Request',
    headers: {},
    config: {} as never,
  })
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

  it.each([
    ['invalid solar date: day 31 is out of range', '输入的公历日期不存在，请检查年、月、日。'],
    ['invalid lunar date: invalid lunar day', '输入的农历日期无效，请检查日期和闰月选项。'],
    ['invalid timezone "Mars/Base"', '时区名称无效，请使用例如 Asia/Shanghai 的地区时区。'],
    [
      'birth time is ambiguous in timezone America/New_York; birth_utc_offset_seconds must be one of [-14400 -18000]',
      '该出生时间在夏令时切换时重复出现，请在高级设置中选择对应的 UTC 偏移。',
    ],
  ])('maps birth validation message %s', (message, expected) => {
    expect(getApiErrorMessage(apiError({ code: 'INVALID_REQUEST', error: message }))).toBe(expected)
  })

  it('uses the contextual fallback for unknown server messages', () => {
    expect(getApiErrorMessage(apiError({ error: 'internal_rule_id=fortune.v2' }), '排盘失败')).toBe(
      '排盘失败',
    )
  })
})
