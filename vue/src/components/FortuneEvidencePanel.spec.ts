import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import FortuneEvidencePanel from './FortuneEvidencePanel.vue'
import type { FortuneScoreBreakdown, ScoreEvidence } from '../api/fortune'

const supporting: ScoreEvidence[] = [
  {
    code: 'relation.stem.shengWo',
    stage: 'relation',
    category: '天干生克',
    label: '流日生扶日主',
    impact: 18,
    description: '流日五行生扶日主。',
    source: '《三命通会》十神生克规则',
    evidence_basis: 'empirical',
    validation_status: 'not_validated',
    interpretation_status: 'not_adjudicated',
    is_outcome_conclusion: false,
  },
]
const counter: ScoreEvidence[] = [
  {
    code: 'relation.branch.clash',
    stage: 'relation',
    category: '地支关系',
    label: '地支六冲',
    impact: -30,
    description: '流日支与日支命中六冲结构。',
    source: '《协纪辨方书》地支关系规则',
    evidence_basis: 'empirical',
    validation_status: 'not_validated',
    interpretation_status: 'not_adjudicated',
    is_outcome_conclusion: false,
  },
]
const breakdown: FortuneScoreBreakdown = {
  pipeline_version: 'fortune-score-pipeline-test',
  score_kind: 'structural_relation_index',
  evidence_basis: 'empirical',
  validation_status: 'not_validated',
  interpretation_status: 'not_adjudicated',
  is_outcome_probability: false,
  base_score: 50,
  relation_score: 38,
  final_score: 38,
  evidence_completeness: 90,
  supporting_evidence: supporting,
  counter_evidence: counter,
}

const commonBoundary =
  '以下内容只记录今日干支与命盘之间命中的规则关系。名称和说明来自原始证据，按原文展示；它们不表示吉凶、现实结果或发生概率，也不是行动建议。'
const professionalBoundary =
  '‘本地权重’是当前关系计算阶段用于内部比较值的加减点数。证据基于经验规则，尚未验证，尚未对现实结果含义作裁决；每条证据不是现实结果结论，内部比较值也不是结果概率。来源仅用于复核规则口径，不表示吉凶、可靠性或行动建议。'

const levels = ['basic', 'advanced', 'professional'] as const

describe('FortuneEvidencePanel', () => {
  it('all levels and empty state show fixed headings and the common boundary', () => {
    for (const level of levels) {
      const wrapper = mount(FortuneEvidencePanel, {
        props: { level, supporting: [], counter: [] },
      })

      expect(wrapper.get('.evidence-eyebrow').text()).toBe('今日干支结构')
      expect(wrapper.get('h2').text()).toBe('今天记录到的结构关系')
      expect(wrapper.text()).toContain(commonBoundary)
      expect(wrapper.text()).toContain('今天没有记录到可展示的干支关系。')
      if (level === 'professional') {
        expect(wrapper.text()).toContain(professionalBoundary)
      } else {
        expect(wrapper.text()).not.toContain('‘本地权重’是当前关系计算阶段')
      }
    }
  })

  it('keeps raw evidence text and level-specific display counts', () => {
    const basic = mount(FortuneEvidencePanel, {
      props: { level: 'basic', supporting, counter, breakdown },
    })
    expect(basic.findAll('.evidence-list li')).toHaveLength(2)
    expect(basic.text()).toContain('2 条')
    expect(basic.text()).toContain(supporting[0].label)
    expect(basic.text()).toContain(supporting[0].description)
    expect(basic.text()).toContain(counter[0].label)
    expect(basic.text()).toContain(counter[0].description)

    const advanced = mount(FortuneEvidencePanel, {
      props: { level: 'advanced', supporting, counter, breakdown },
    })
    expect(advanced.findAll('.evidence-list li')).toHaveLength(2)
    for (const item of [...supporting, ...counter]) {
      expect(advanced.text()).toContain(item.label)
      expect(advanced.text()).toContain(item.description)
    }
    expect(advanced.text()).toContain(commonBoundary)
    expect(advanced.text()).not.toContain('规则口径：')
  })

  it('professional keeps raw source and formatted impact with fixed governance meaning', () => {
    const professional = mount(FortuneEvidencePanel, {
      props: {
        level: 'professional',
        supporting,
        counter,
        breakdown,
      },
    })

    expect(professional.get('[data-testid="professional-meta"]').exists()).toBe(true)
    expect(professional.text()).toContain(commonBoundary)
    expect(professional.text()).toContain(professionalBoundary)
    expect(professional.text()).toContain('本地权重 +18')
    expect(professional.text()).toContain('本地权重 -30')
    expect(professional.text()).toContain('比较基准 50')
    expect(professional.text()).toContain('关系阶段比较值 38')
    expect(professional.text()).toContain('内部比较值 38')
    expect(professional.text()).toContain(`规则口径：${supporting[0].source}`)
    expect(professional.text()).toContain(`规则口径：${counter[0].source}`)
  })

  it('keeps internal codes, pipeline version, and raw governance tokens hidden', () => {
    for (const level of levels) {
      const wrapper = mount(FortuneEvidencePanel, {
        props: { level, supporting, counter, breakdown },
      })

      expect(wrapper.text()).not.toContain('relation.stem.shengWo')
      expect(wrapper.text()).not.toContain('fortune-score-pipeline-test')
      expect(wrapper.text()).not.toContain('empirical')
      expect(wrapper.text()).not.toContain('not_validated')
      expect(wrapper.text()).not.toContain('not_adjudicated')
    }
  })
})
