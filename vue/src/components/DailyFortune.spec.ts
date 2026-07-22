import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it } from 'vitest'
import DailyFortune from './DailyFortune.vue'

const evidence = {
  code: 'relation.stem.shengWo',
  stage: 'relation',
  category: '天干生克',
  label: '流日生扶日主',
  impact: 18,
  description: '流日五行生扶日主。',
  source: '规则来源',
  evidence_basis: 'empirical' as const,
  validation_status: 'not_validated' as const,
  interpretation_status: 'not_adjudicated' as const,
  is_outcome_conclusion: false as const,
}

beforeEach(() => localStorage.clear())

describe('DailyFortune interpretation levels', () => {
  it('defaults to concise mode and progressively reveals detailed calculation content', async () => {
    const wrapper = mount(DailyFortune, {
      props: {
        solarDate: '2026-07-10',
        dayGanZhi: '甲子',
        fortuneScore: 68,
        supportingEvidence: [evidence],
        counterEvidence: [],
        scoreBreakdown: {
          pipeline_version: 'pipeline-test',
          score_kind: 'structural_relation_index',
          evidence_basis: 'empirical',
          validation_status: 'not_validated',
          interpretation_status: 'not_adjudicated',
          is_outcome_probability: false,
          base_score: 50,
          relation_score: 68,
          final_score: 68,
          evidence_completeness: 88,
          supporting_evidence: [evidence],
          counter_evidence: [],
        },
        seasonElement: {
          rule_id: 'fortune.season-element.month-branch-v1',
          reference_stem: '甲',
          reference_element: '木',
          query_month_branch: '未',
          season: '夏',
          basis: 'reference_day_stem_element_and_query_month_branch',
          status: 'observed',
          interpretation_status: 'not_adjudicated',
        },
        tenGod: {
          rule_id: 'rikuyo.ten-god-day-stem-v1',
          reference_stem: '甲',
          query_stem: '丙',
          name: '食神',
          basis: 'reference_day_stem_and_query_day_stem',
          status: 'observed',
          interpretation_status: 'not_adjudicated',
        },
        seasonalState: {
          rule_id: 'rikuyo.seasonal-state-v1',
          query_stem: '丙',
          query_element: '火',
          query_month_branch: '未',
          season: '夏',
          state: '旺',
          basis: 'query_day_stem_element_and_query_month_branch',
          status: 'observed',
          interpretation_status: 'not_adjudicated',
        },
        hiddenStems: [
          {
            rule_id: 'rikuyo.hidden-stem-ten-god-v1',
            query_branch: '午',
            reference_stem: '甲',
            stem: '丁',
            type: '本气',
            element: '火',
            ten_god: '伤官',
            basis: 'query_branch_hidden_stem_and_reference_day_stem',
            status: 'observed',
            interpretation_status: 'not_adjudicated',
          },
        ],
        stemRelations: [
          {
            rule_id: 'rikuyo.stem-relation-v2.five_combine',
            query_stem: '甲',
            target_pillar: '年柱',
            target_stem: '己',
            type: 'five_combine',
            name: '天干五合',
            combined_element: '土',
            basis: 'query_day_stem_and_natal_stem',
            status: 'observed',
            transformation_status: 'not_adjudicated',
            interpretation_status: 'not_adjudicated',
          },
        ],
        branchRelations: [],
        fortuneLayers: {
          rule_version: 'rules-test',
          school: 'test-school',
          dayun: {
            rule_id: 'fortune.layer.dayun-v2',
            key: 'dayun',
            name: '大运',
            pillar: '乙丑',
            gan: '乙',
            zhi: '丑',
            start_age: 8,
            end_age: 17,
            ten_god: {
              rule_id: 'rikuyo.ten-god-day-stem-v1',
              reference_stem: '甲',
              query_stem: '乙',
              name: '劫财',
              basis: 'reference_day_stem_and_query_day_stem',
              status: 'observed',
              interpretation_status: 'not_adjudicated',
            },
            relations: [],
            shen_sha_details: [],
            basis: 'exact_start_time_and_query_time',
            status: 'observed',
            interpretation_status: 'not_adjudicated',
          },
          liunian: {} as never,
          liuyue: {} as never,
          xiaoyun: {} as never,
          inter_layer_relations: [
            {
              rule_id: 'fortune.layer-relation.stem-v2.keWo',
              source: '流年天干',
              source_value: '庚',
              target: '大运天干',
              target_value: '甲',
              type: 'keWo',
              name: '流年克大运',
              basis: 'period_layer_stem_pair',
              status: 'observed',
              interpretation_status: 'not_adjudicated',
            },
          ],
        },
      },
      global: { stubs: { Teleport: true } },
    })

    const tabLabels = () => wrapper.findAll('.df-tab-btn').map((tab) => tab.text())

    expect(wrapper.text()).toContain('简明')
    expect(tabLabels()).not.toContain('结构分析')
    for (const removedLabel of ['开运指南', '幸运色', '幸运数', '财位', '吉时', '宜忌']) {
      expect(wrapper.text()).not.toContain(removedLabel)
    }
    expect(wrapper.find('[data-testid="professional-meta"]').exists()).toBe(false)

    await wrapper.get('[data-level="advanced"]').trigger('click')
    expect(tabLabels()).toContain('结构分析')
    expect(wrapper.text()).toContain('月令季节')
    expect(wrapper.text()).not.toContain('季节建议')
    expect(wrapper.text()).not.toContain('日课推算')
    expect(wrapper.text()).not.toContain('九维刻度')

    await wrapper.get('[data-level="professional"]').trigger('click')
    expect(tabLabels()).toContain('日课推算')
    expect(wrapper.text()).toContain('今日十神')
    expect(wrapper.text()).toContain('月令状态')
    expect(wrapper.text()).toContain('天干五合')
    expect(wrapper.text()).toContain('周期层结构')
    expect(wrapper.text()).toContain('岁运月联动结构')
    expect(wrapper.text()).toContain('依据出生时间与查询日期定位')
    expect(wrapper.text()).toContain('这里只展示周期之间的干支关系')
    expect(wrapper.text()).not.toContain('exact_start_time_and_query_time')
    expect(wrapper.text()).not.toContain('解释未裁决')
    expect(wrapper.text()).not.toContain('not_validated')
    expect(wrapper.text()).not.toContain('not_adjudicated')
    for (const removedLabel of ['格局喜忌', '进退气', '用神影响', '偏有利', '需留意']) {
      expect(wrapper.text()).not.toContain(removedLabel)
    }
    expect(wrapper.get('[data-testid="professional-meta"]').attributes('data-testid')).toBe(
      'professional-meta',
    )
    expect(localStorage.getItem('fortune-interpretation-level')).toBe('professional')
  })
})
