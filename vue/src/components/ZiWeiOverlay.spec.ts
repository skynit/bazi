import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ZiWeiOverlay from './ZiWeiOverlay.vue'

const branches = ['子', '丑', '寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥']

function makePalaces() {
  return branches.map((branch) => ({
    name: `${branch}宫`,
    branch,
    heavenly_stem: '',
    is_body_palace: false,
    stars: [],
    four_hua: [],
  }))
}

function mountOverlay(gender: string) {
  const palaces = makePalaces()
  return mount(ZiWeiOverlay, {
    props: {
      baseChart: {
        palaces,
        life_master: '贪狼',
        body_master: '天相',
        five_bureau: '木三局',
        earthly_branch_of_soul_palace: '午',
        earthly_branch_of_body_palace: '申',
      },
      liunianChart: {
        palaces,
        year: 2026,
        liu_nian_stars: branches.map(() => []),
      },
      dayunStages: [],
      availableYears: [2026],
      birthYearBranch: '辰',
      gender,
    },
  })
}

describe('ZiWeiOverlay', () => {
  it('男命按顺行在宫位左下角显示小限经过虚岁', () => {
    const wrapper = mountOverlay('男')

    const ageOnePalace = wrapper.get('[data-branch="戌"] .zw-palace-ages')
    expect(ageOnePalace.get('small').text()).toBe('小限')
    expect(ageOnePalace.findAll('.zw-palace-age-list b').map((item) => item.text())).toEqual([
      '1', '13', '25', '37', '49', '61', '73', '85',
    ])

    const thirdAgePalace = wrapper.get('[data-branch="子"] .zw-palace-ages')
    expect(thirdAgePalace.findAll('.zw-palace-age-list b').map((item) => item.text())).toEqual([
      '3', '15', '27', '39', '51', '63', '75', '87',
    ])
    expect(ageOnePalace.attributes('title')).toContain('小限经过年龄')
  })

  it('女命按逆行计算宫位经过虚岁', () => {
    const wrapper = mountOverlay('女')

    const ageOnePalace = wrapper.get('[data-branch="戌"] .zw-palace-ages')
    expect(ageOnePalace.findAll('.zw-palace-age-list b').map((item) => item.text())).toEqual([
      '1', '13', '25', '37', '49', '61', '73', '85',
    ])

    const twelfthAgePalace = wrapper.get('[data-branch="亥"] .zw-palace-ages')
    expect(twelfthAgePalace.findAll('.zw-palace-age-list b').map((item) => item.text())).toEqual([
      '12', '24', '36', '48', '60', '72', '84', '96',
    ])
  })
})
