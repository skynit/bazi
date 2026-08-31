import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ZiWeiOverlay from './ZiWeiOverlay.vue'
import type {
  ZiWeiOverlayAnalysis,
  ZiWeiOverlayFocusPalace,
  ZiWeiOverlayTrigger,
} from '../api/ziwei'

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

function mountOverlay(gender: string, overlayAnalysis?: ZiWeiOverlayAnalysis) {
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
      overlayAnalysis,
      availableYears: [2026],
      birthYearBranch: '辰',
      gender,
    },
  })
}

function makeTrigger(
  type: string,
  polarity: ZiWeiOverlayTrigger['polarity'],
  branch: string,
): ZiWeiOverlayTrigger {
  return {
    type,
    polarity,
    branch,
    palace: `${branch}宫`,
    meaning: `${branch}-${type}-${polarity}`,
    placement_basis: 'deterministic_rule_projection',
    interpretation_basis: 'traditional_rule_labels',
    interpretation_status: 'not_adjudicated',
    is_outcome_conclusion: false,
  }
}

function makeFocusPalace(branch: string, triggers: ZiWeiOverlayTrigger[]): ZiWeiOverlayFocusPalace {
  return {
    palace: `${branch}宫`,
    branch,
    triggers,
    main_stars: [],
    review_note: '',
    placement_basis: 'deterministic_rule_projection',
    interpretation_basis: 'traditional_rule_labels',
    interpretation_status: 'not_adjudicated',
    is_outcome_conclusion: false,
  }
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

  it('统一按约束与化忌、资源、移动、中性的优先级标记触发影响', async () => {
    const watchTriggers = [
      makeTrigger('流羊', 'constraint', '寅'),
      makeTrigger('化禄', 'resource', '寅'),
      makeTrigger('流马', 'movement', '寅'),
      makeTrigger('普通触发', 'neutral', '寅'),
    ]
    const goodTriggers = [
      makeTrigger('化禄', 'resource', '卯'),
      makeTrigger('流马', 'movement', '卯'),
      makeTrigger('普通触发', 'neutral', '卯'),
    ]
    const moveTriggers = [
      makeTrigger('流马', 'movement', '辰'),
      makeTrigger('普通触发', 'neutral', '辰'),
    ]
    const neutralTriggers = [makeTrigger('普通触发', 'neutral', '巳')]
    const huaJiTriggers = [
      makeTrigger('化忌', 'resource', '午'),
      makeTrigger('流马', 'movement', '午'),
    ]
    const allTriggers = [
      ...watchTriggers,
      ...goodTriggers,
      ...moveTriggers,
      ...neutralTriggers,
      ...huaJiTriggers,
    ]
    const overlayAnalysis: ZiWeiOverlayAnalysis = {
      year: 2026,
      gan_zhi: '丙午',
      stem: '丙',
      branch: '午',
      relation_to_ming: '',
      relation_evidence: [],
      review_note: '',
      summary: '',
      method: [],
      four_hua: allTriggers,
      annual_stars: [],
      focus_palaces: [
        makeFocusPalace('寅', watchTriggers),
        makeFocusPalace('卯', goodTriggers),
        makeFocusPalace('辰', moveTriggers),
        makeFocusPalace('巳', neutralTriggers),
        makeFocusPalace('午', huaJiTriggers),
      ],
      evidence_basis: 'mixed_deterministic_projection_and_unadjudicated_traditional_labels',
      placement_basis: 'deterministic_rule_projection',
      interpretation_basis: 'traditional_rule_labels',
      interpretation_status: 'not_adjudicated',
      validation_status: 'not_adjudicated',
      is_outcome_conclusion: false,
    }
    const wrapper = mountOverlay('男', overlayAnalysis)

    await wrapper.findAll('.zw-tab')[1].trigger('click')

    expect(wrapper.get('[data-branch="寅"]').classes()).toContain('zw-impact-watch')
    expect(wrapper.get('[data-branch="卯"]').classes()).toContain('zw-impact-good')
    expect(wrapper.get('[data-branch="辰"]').classes()).toContain('zw-impact-move')
    expect(wrapper.get('[data-branch="巳"]').classes()).toContain('zw-impact-neutral')
    expect(wrapper.get('[data-branch="午"]').classes()).toContain('zw-impact-watch')
    expect(wrapper.get('[data-branch="子"]').classes()).not.toEqual(
      expect.arrayContaining([
        'zw-impact-watch',
        'zw-impact-good',
        'zw-impact-move',
        'zw-impact-neutral',
      ]),
    )

    expect(wrapper.get('.zw-trigger-chip[title="寅-流羊-constraint"]').classes()).toContain('is-watch')
    expect(wrapper.get('.zw-trigger-chip[title="卯-化禄-resource"]').classes()).toContain('is-good')
    expect(wrapper.get('.zw-trigger-chip[title="辰-流马-movement"]').classes()).toContain('is-move')
    expect(wrapper.get('.zw-trigger-chip[title="巳-普通触发-neutral"]').classes()).toContain(
      'is-neutral',
    )
    expect(wrapper.get('.zw-trigger-chip[title="午-化忌-resource"]').classes()).toContain('is-watch')

    await wrapper.findAll('.zw-side-tab')[2].trigger('click')
    expect(wrapper.findAll('.zw-focus-card').map((card) => card.classes())).toEqual([
      expect.arrayContaining(['is-watch']),
      expect.arrayContaining(['is-good']),
      expect.arrayContaining(['is-move']),
      expect.arrayContaining(['is-neutral']),
      expect.arrayContaining(['is-watch']),
    ])
  })
})
