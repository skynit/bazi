import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import ZiWeiInterpretation from './ZiWeiInterpretation.vue'

function section(content = '') {
  return { title: '', content, tags: [] }
}

function reading(overrides: Record<string, unknown> = {}) {
  return {
    palaceName: '命宫',
    palaceFocus: '自我定位与做事方式主题',
    summary: '命宫以巨门、太阳为核心。',
    keyPoints: ['主星组合：巨门、太阳。'],
    evidence: [
      { type: 'main_star', label: '主星', value: '巨门庙', basis: '巨门为本宫主星' },
      { type: 'main_star', label: '主星', value: '太阳旺', basis: '太阳为本宫主星' },
      { type: 'four_hua', label: '四化', value: '太阳化权', basis: '本命四化落宫' },
      { type: 'tough_star', label: '煞曜', value: '擎羊', basis: '本宫星曜位置' },
      { type: 'body_palace', label: '身宫', value: '命宫', basis: '身宫落宫位置' },
    ],
    sanfangContext: null,
    patternDetails: [],
    mainStarAnalysis: section('命宫主星为巨门、太阳。'),
    auxStarInfluence: section(),
    sihuaInfluence: section(),
    sanFangSiZheng: section(),
    patternAnnotations: section(),
    ...overrides,
  }
}

describe('ZiWeiInterpretation', () => {
  it('先给白话结论，再解释主星、四化和辅助星曜', () => {
    const wrapper = mount(ZiWeiInterpretation, { props: { palaceReading: reading() } })

    expect(wrapper.text()).toContain('先看结论')
    expect(wrapper.text()).toContain('沟通、辨析与质疑')
    expect(wrapper.text()).toContain('表达、担当与公开性')
    expect(wrapper.text()).toContain('责任、主导与执行')
    expect(wrapper.text()).toContain('阻力、竞争与推进方式')
    expect(wrapper.text()).toContain('不是吉凶分数')
    expect(wrapper.text()).toContain('不据此断定健康、婚姻、财富、职业或其他具体事件')
  })

  it('把空宫说明为需要连同对宫与三合宫阅读', () => {
    const wrapper = mount(ZiWeiInterpretation, {
      props: {
        palaceReading: reading({
          palaceName: '夫妻宫',
          palaceFocus: '亲密关系中的互动与协商主题',
          summary: '夫妻宫为空宫。',
          keyPoints: [],
          evidence: [
            {
              type: 'borrowed_star',
              label: '空宫借星',
              value: '天同、天梁',
              basis: '本宫无主星',
            },
          ],
          sanfangContext: {
            opposite: '官禄宫',
            trine1: '迁移宫',
            trine2: '福德宫',
            opposite_stars: ['天同'],
            trine1_stars: ['天梁'],
            trine2_stars: [],
            notes: [],
          },
        }),
      },
    })

    expect(wrapper.text()).toContain('本宫没有主星')
    expect(wrapper.text()).toContain('官禄宫、迁移宫、福德宫')
    expect(wrapper.text()).toContain('空宫不等于没有内容')
    expect(wrapper.text()).toContain('不是把别宫结论直接搬过来')
  })
})
