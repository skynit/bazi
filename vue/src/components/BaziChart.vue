<script setup lang="ts">
import { computed } from 'vue'
import { use } from 'echarts/core'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'

const props = defineProps<{
  chart: {
    id?: number
    year_pillar: { gan: string; zhi: string }
    month_pillar: { gan: string; zhi: string }
    day_pillar: { gan: string; zhi: string }
    hour_pillar: { gan: string; zhi: string }
    [key: string]: any
  }
}>()

const ganElement: Record<string, { name: string; elemColor: string }> = {
  甲: { name: '木', elemColor: '#228B22' },
  乙: { name: '木', elemColor: '#228B22' },
  丙: { name: '火', elemColor: '#DC143C' },
  丁: { name: '火', elemColor: '#DC143C' },
  戊: { name: '土', elemColor: '#DAA520' },
  己: { name: '土', elemColor: '#DAA520' },
  庚: { name: '金', elemColor: '#FFD700' },
  辛: { name: '金', elemColor: '#FFD700' },
  壬: { name: '水', elemColor: '#4169E1' },
  癸: { name: '水', elemColor: '#4169E1' },
}

const zhiElement: Record<string, { name: string; elemColor: string }> = {
  寅: { name: '木', elemColor: '#228B22' },
  卯: { name: '木', elemColor: '#228B22' },
  巳: { name: '火', elemColor: '#DC143C' },
  午: { name: '火', elemColor: '#DC143C' },
  辰: { name: '土', elemColor: '#DAA520' },
  戌: { name: '土', elemColor: '#DAA520' },
  丑: { name: '土', elemColor: '#DAA520' },
  未: { name: '土', elemColor: '#DAA520' },
  申: { name: '金', elemColor: '#FFD700' },
  酉: { name: '金', elemColor: '#FFD700' },
  亥: { name: '水', elemColor: '#4169E1' },
  子: { name: '水', elemColor: '#4169E1' },
}

const pillars = computed(() => [
  { label: '年柱', key: 'year' as const, idx: 0, ...props.chart.year_pillar },
  { label: '月柱', key: 'month' as const, idx: 1, ...props.chart.month_pillar },
  { label: '日柱', key: 'day' as const, idx: 2, ...props.chart.day_pillar },
  { label: '时柱', key: 'hour' as const, idx: 3, ...props.chart.hour_pillar },
])

// --- 天干地支分析（从 API 数据读取）---
const ganZhi = computed(() => props.chart.gan_zhi_analysis)

function ganRelClass(type: string): string {
  if (type === '五合') return 'rel-he'
  if (type === '相克') return 'rel-ke'
  if (type === '相生') return 'rel-sheng'
  return ''
}

function zhiRelClass(type: string): string {
  if (type === '六冲') return 'rel-chong'
  if (type === '六合') return 'rel-he'
  if (type === '六害') return 'rel-hai'
  if (type === '相刑') return 'rel-xing'
  if (type === '三会') return 'rel-hui'
  return ''
}

const elemColor = (e: string) =>
  ({ 金: '#FFD700', 木: '#228B22', 水: '#4169E1', 火: '#DC143C', 土: '#DAA520' })[e] || '#999'

const pillarLabel = (k: string) => ({ year: '年柱', month: '月柱', day: '日柱', hour: '时柱' }[k] || k)

function parseShenSha(raw: string) {
  const [head, desc = ''] = raw.split('｜')
  const colonIndex = head.indexOf('：')
  if (colonIndex === -1) return { name: head, target: '', desc }

  const name = head.slice(0, colonIndex)
  const tail = head.slice(colonIndex + 1)
  const target = desc ? tail : ''
  return { name, target, desc: desc || tail }
}

const parsedDayShenSha = computed(() => (props.chart.day_shen_sha || []).map(parseShenSha))

const groupedShenSha = computed(() => {
  const raw = props.chart.shen_sha_by_pillar
  if (!raw || !raw.length) return null
  return raw.map((g: any) => ({ ...g, items: (g.items || []).map(parseShenSha) }))
})

const globalShenSha = computed(() => (props.chart.global_shen_sha || []).map(parseShenSha))

const showSummary = computed(() => !!props.chart.shen_sha_summary)

const pillarShenShaColor = (p: string) =>
  ({ day: 'var(--gold)', year: '#5BA4CF', month: '#60B89A', hour: '#A182CF' }[p] || '#888')

const pillarShenShaBg = (p: string) =>
  ({ day: 'rgba(212,168,75,0.07)', year: 'rgba(91,164,207,0.06)', month: 'rgba(96,184,154,0.06)', hour: 'rgba(161,130,207,0.06)' }[p] || 'rgba(255,255,255,0.02)')

function strengthLevel(total: number): string {
  if (total <= 0) return 'none'
  if (total <= 5) return 'weak'
  if (total <= 15) return 'medium'
  if (total <= 25) return 'strong'
  return 'very-strong'
}

const fiveElementsOption = computed(() => {
  const fe = props.chart.five_elements
  if (!fe) return null
  const total = Object.values(fe as Record<string, number>).reduce((s, v) => s + v, 0)
  if (total === 0) return null

  // 五行配色 — 珠宝色调，呼应项目金色/深色基调
  const barColors = ['#2ECC71', '#E74C3C', '#D4A84B', '#FFD700', '#3498DB']
  const labels = ['木', '火', '土', '金', '水']

  return {
    backgroundColor: 'transparent',
    grid: { left: 8, right: 8, bottom: 28, top: 12, containLabel: true },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: { lineStyle: { color: 'rgba(212,168,75,0.2)' } },
      axisTick: { show: false },
      axisLabel: {
        color: 'rgba(240,237,228,0.75)',
        fontSize: 12,
        fontWeight: '600',
        fontFamily: 'Noto Serif SC, Songti SC, serif',
      },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      max: 30,
      axisLabel: { color: 'rgba(240,237,228,0.35)', fontSize: 10, formatter: '{value}' },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.04)', type: 'dashed' } },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    series: [{
      type: 'bar',
      data: labels.map((l, i) => ({
        value: fe[l] || 0,
        itemStyle: {
          color: {
            type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
            colorStops: [
              { offset: 0, color: barColors[i] },
              { offset: 1, color: barColors[i] + '66' },
            ],
          },
          borderRadius: [4, 4, 0, 0],
        },
        emphasis: {
          itemStyle: {
            color: {
              type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: barColors[i] + 'EE' },
                { offset: 1, color: barColors[i] + '99' },
              ],
            },
          },
        },
      })),
      barMaxWidth: 38,
      barCategoryGap: '30%',
      label: {
        show: true,
        position: 'top',
        formatter: '{c}',
        fontSize: 11,
        fontWeight: '600',
        color: 'rgba(240,237,228,0.8)',
        fontFamily: 'DM Mono, Fira Code, monospace',
      },
    }],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      backgroundColor: 'rgba(10,8,21,0.95)',
      borderColor: 'rgba(212,168,75,0.3)',
      borderWidth: 1,
      padding: [8, 14],
      textStyle: {
        color: 'rgba(240,237,228,0.9)',
        fontSize: 13,
        fontFamily: 'Noto Serif SC, Songti SC, serif',
        fontWeight: '600',
      },
      formatter: (params: any[]) => {
        const p = params[0]
        return `<span style="color:${p.color};font-weight:700">${p.name}</span>：<span style="color:#D4A84B;font-weight:700">${p.value}</span> 分`
      },
    },
    animationDuration: 900,
    animationEasing: 'cubicOut' as const,
  }
})

const pillarDetails = computed(() => props.chart.pillar_details || [])

const birthMonthLabel = computed(() => {
  const m = props.chart.birth_month
  if (!m) return ''
  const labels: Record<number, string> = { 5: '五月', 6: '六月' }
  return labels[m] || ''
})

use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

const tenGodChartOptions = computed(() => {
  const data = props.chart.ten_god_proportion || []
  // 10 ten gods — warm gold/crimson palette, matching BaZi aesthetic
  const barColors = [
    '#D4A84B', // 比肩 - gold
    '#E8C97A', // 劫财 - light gold
    '#9B72CF', // 食神 - purple
    '#C85FCF', // 伤官 - magenta
    '#4A7FBF', // 正财 - steel blue
    '#2E5A8F', // 偏财 - deep blue
    '#7B9E87', // 正官 - sage green
    '#5A7A62', // 七杀 - dark sage
    '#C93C3C', // 正印 - crimson
    '#A02020', // 偏印 - deep red
  ]
  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'none' },
      formatter: (params: any[]) => {
        const p = params[0]
        return `<span style="color:${p.color};font-weight:700">${p.name}</span>：${p.value}%`
      },
      backgroundColor: 'rgba(15,12,8,0.92)',
      borderColor: 'rgba(212,168,75,0.25)',
      borderWidth: 1,
      padding: [6, 10],
      textStyle: { color: 'rgba(255,255,255,0.75)', fontSize: 12 },
    },
    grid: {
      left: 8,
      right: 8,
      bottom: data.length > 0 ? 32 : 8,
      top: data.length > 0 ? 20 : 8,
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: data.map((d: any) => d.name),
      axisLine: { lineStyle: { color: 'rgba(212,168,75,0.15)' } },
      axisTick: { show: false },
      axisLabel: {
        color: 'rgba(255,255,255,0.5)',
        fontSize: 10,
        fontWeight: '500',
        interval: 0,
        rotate: data.length > 6 ? 30 : 0,
      },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      max: 100,
      axisLabel: {
        formatter: '{value}%',
        color: 'rgba(255,255,255,0.3)',
        fontSize: 9,
      },
      splitLine: {
        lineStyle: { color: 'rgba(255,255,255,0.04)', type: 'dashed' },
      },
      axisLine: { show: false },
      axisTick: { show: false },
    },
    series: [
      {
        type: 'bar',
        data: data.map((d: any, i: number) => ({
          value: d.percent,
          itemStyle: {
            color: {
              type: 'linear',
              x: 0, y: 0, x2: 0, y2: 1,
              colorStops: [
                { offset: 0, color: barColors[i % barColors.length] },
                { offset: 1, color: barColors[i % barColors.length] + '88' },
              ],
            },
            borderRadius: [6, 6, 0, 0],
            shadowBlur: 12,
            shadowColor: barColors[i % barColors.length] + '55',
          },
        })),
        barMaxWidth: data.length > 6 ? 18 : 28,
        barGap: '6px',
        label: {
          show: true,
          position: 'top',
          formatter: '{c}%',
          fontSize: 10,
          fontWeight: '600',
          color: 'rgba(255,255,255,0.65)',
          distance: 6,
        },
        emphasis: {
          itemStyle: {
            shadowBlur: 20,
            shadowColor: '#D4A84B66',
          },
        },
      },
    ],
    animationDuration: 1000,
    animationEasing: 'cubicOut',
    animationDelay: (idx: number) => idx * 80,
  }
})
</script>

<template>
  <div class="bazi-chart">
    <!-- Constellation decoration -->
    <div class="chart-bg" aria-hidden="true">
      <svg viewBox="0 0 600 200" preserveAspectRatio="xMidYMid slice" class="bg-svg">
        <circle cx="50" cy="30" r="1" fill="#D4A84B" opacity="0.2" />
        <circle cx="550" cy="40" r="1.2" fill="#D4A84B" opacity="0.25" />
        <circle cx="300" cy="100" r="1.5" fill="#D4A84B" opacity="0.15" />
        <line
          x1="50"
          y1="30"
          x2="300"
          y2="100"
          stroke="#D4A84B"
          stroke-width="0.3"
          opacity="0.05"
        />
        <line
          x1="550"
          y1="40"
          x2="300"
          y2="100"
          stroke="#D4A84B"
          stroke-width="0.3"
          opacity="0.05"
        />
      </svg>
    </div>

    <div class="chart-card glass-card overflow-hidden">
      <!-- Title -->
      <div class="chart-header">
        <div class="header-eyebrow">BaZi Fortune</div>
        <h2 class="chart-title">八字命盘</h2>
      </div>

      <!-- Four pillars grid -->
      <div class="pillars-grid">
        <div v-for="pillar in pillars" :key="pillar.key" class="pillar-col">
          <!-- Pillar label -->
          <div class="pillar-label">{{ pillar.label }}</div>

          <!-- Gan -->
          <div class="pillar-gan" :style="{ color: ganElement[pillar.gan]?.elemColor }">
            <div class="gan-char">{{ pillar.gan }}</div>
            <div
              class="elem-tag"
              :style="{
                background: ganElement[pillar.gan]?.elemColor + '22',
                color: ganElement[pillar.gan]?.elemColor,
                borderColor: ganElement[pillar.gan]?.elemColor + '44',
              }"
            >
              {{ ganElement[pillar.gan]?.name }}
            </div>
          </div>

          <!-- Zhi -->
          <div class="pillar-zhi" :style="{ color: zhiElement[pillar.zhi]?.elemColor }">
            <div class="zhi-char">{{ pillar.zhi }}</div>
            <div
              class="elem-tag"
              :style="{
                background: zhiElement[pillar.zhi]?.elemColor + '22',
                color: zhiElement[pillar.zhi]?.elemColor,
                borderColor: zhiElement[pillar.zhi]?.elemColor + '44',
              }"
            >
              {{ zhiElement[pillar.zhi]?.name }}
            </div>
          </div>

          <!-- ShengXiao + Empties from pillar_details -->
          <div v-if="pillarDetails[pillar.idx]" class="pillar-sub">
            <span class="sheng-xiao-tag">{{ pillarDetails[pillar.idx].sheng_xiao }}</span>
            <span v-if="pillarDetails[pillar.idx].empties[0]" class="empties-tag">
              空{{ pillarDetails[pillar.idx].empties[0] }}{{ pillarDetails[pillar.idx].empties[1] }}
            </span>
          </div>
        </div>
      </div>

      <!-- 天干地支综合分析 -->
      <div v-if="ganZhi" class="ganzhi-analysis">
        <!-- 天干关系 -->
        <div v-if="ganZhi.gan_relations?.length > 0" class="relations-section">
          <div class="relations-title">天干关系</div>
          <div class="relations-list">
            <div
              v-for="(rel, ri) in ganZhi.gan_relations"
              :key="'g'+ri"
              class="rel-chip"
              :class="ganRelClass(rel.type)"
            >
              <span>{{ rel.pillar1 }}</span>
              <span class="rel-symbol">{{ rel.type === '五合' ? '合' : rel.type === '相克' ? '克' : '生' }}</span>
              <span>{{ rel.pillar2 }}</span>
              <span class="rel-detail">{{ rel.detail }}</span>
            </div>
          </div>
        </div>
        <div v-else class="no-relations">天干无特殊关系</div>

        <!-- 地支关系 -->
        <div v-if="ganZhi.zhi_relations?.length > 0" class="relations-section">
          <div class="relations-title">地支关系</div>
          <div class="relations-list">
            <div
              v-for="(rel, ri) in ganZhi.zhi_relations"
              :key="'z'+ri"
              class="rel-chip"
              :class="zhiRelClass(rel.type)"
            >
              <span>{{ rel.pillar1 }}</span>
              <span class="rel-symbol">{{ rel.type === '六冲' ? '冲' : rel.type === '六合' ? '合' : rel.type === '六害' ? '害' : rel.type === '相刑' ? '刑' : '会' }}</span>
              <span>{{ rel.pillar2 }}</span>
              <span class="rel-detail">{{ rel.detail }}</span>
            </div>
          </div>
        </div>
        <div v-else class="no-relations">地支无特殊关系</div>
      </div>

      <!-- Analysis sections -->
      <div class="analysis-section">
        <!-- Five Elements bar chart -->
        <div v-if="fiveElementsOption" class="analysis-block">
          <div class="block-title">五行分布</div>
          <v-chart class="five-elem-chart" :option="fiveElementsOption" autoresize />
        </div>

        <!-- Element Detail -->
        <div v-if="chart.element_detail && chart.element_detail.length" class="analysis-block">
          <div class="block-title">五行力量与藏干分析</div>
          <div class="element-detail-table">
            <div class="ed-header">
              <span>五行</span>
              <span>天干</span>
              <span>地支藏干</span>
              <span>总力量</span>
            </div>
            <div
              v-for="ed in chart.element_detail"
              :key="ed.element"
              class="ed-row"
              :class="'level-' + strengthLevel(ed.total)"
            >
              <span class="ed-elem" :style="{ color: elemColor(ed.element) }">{{ ed.element }}</span>
              <span class="ed-tg">{{ ed.tian_gan }}</span>
              <span class="ed-zc">{{ ed.cang_gan_list ? ed.cang_gan_list.join('、') : '—' }}</span>
              <span class="ed-total">{{ ed.total }}</span>
            </div>
          </div>
        </div>

        <!-- Body Strength -->
        <div v-if="chart.body_strength" class="analysis-block">
          <div class="block-title">身旺喜忌</div>
          <div class="body-strength">
            <div class="bs-verdict">{{ chart.body_strength.verdict }}</div>
            <div class="bs-tags">
              <span class="bs-like-label">喜</span>
              <span v-for="l in chart.body_strength.like" :key="l" class="bs-like">{{ l }}</span>
              <span class="bs-dislike-label">忌</span>
              <span v-for="d in chart.body_strength.dislike" :key="d" class="bs-dislike">{{ d }}</span>
            </div>
          </div>
        </div>

        <!-- Ten Gods -->
        <div v-if="chart.ten_gods" class="analysis-block">
          <div class="block-title">十神</div>
          <div class="ten-gods-grid">
            <div v-for="(god, pillar) in chart.ten_gods" :key="pillar" class="god-item">
              <span class="god-pillar">{{ pillar }}</span>
              <span class="god-name">{{ god }}</span>
            </div>
          </div>
        </div>

        <!-- TenGodProportion -->
        <div v-if="chart.ten_god_proportion && chart.ten_god_proportion.length" class="analysis-block">
          <div class="block-title">十神占比</div>
          <div class="ten-god-chart-wrap">
            <v-chart class="ten-god-chart" :option="(tenGodChartOptions as any)" autoresize />
          </div>
        </div>

        <!-- NaYin -->
        <div v-if="chart.na_yin" class="analysis-block">
          <div class="block-title">纳音</div>
          <div class="nayin-list">
            <el-popover
              v-for="(info, key) in chart.na_yin"
              :key="key"
              placement="bottom"
              :width="300"
              trigger="click"
              popper-class="nayin-popover"
            >
              <template #reference>
                <span class="nayin-tag" :style="{ borderColor: elemColor(info.element) }">
                  <span class="nayin-pillar">{{ pillarLabel(String(key)) }}</span>
                  <span class="nayin-name" :style="{ color: elemColor(info.element) }">{{ info.name }}</span>
                </span>
              </template>
              <div class="nayin-detail">
                <div class="nayin-detail-header">
                  <span class="nayin-detail-name">{{ info.name }}</span>
                  <span class="nayin-detail-elem" :style="{ background: elemColor(info.element) }">{{ info.element }}</span>
                </div>
                <div class="nayin-detail-section">
                  <div class="nayin-detail-label">取象释义</div>
                  <div class="nayin-detail-value">{{ info.image_desc }}</div>
                </div>
                <div class="nayin-detail-section">
                  <div class="nayin-detail-label">性格命运</div>
                  <div class="nayin-detail-value">{{ info.personality }}</div>
                </div>
                <div class="nayin-detail-section">
                  <div class="nayin-detail-label">能量阶段</div>
                  <div class="nayin-detail-value nayin-energy">{{ info.energy_stage }}</div>
                </div>
                <div class="nayin-detail-section">
                  <div class="nayin-detail-label">现代延伸</div>
                  <div class="nayin-detail-value">{{ info.modern_ext }}</div>
                </div>
                <div v-if="info.judgments && info.judgments.length" class="nayin-detail-section">
                  <div class="nayin-detail-label">特质断语</div>
                  <div class="nayin-detail-tags">
                    <span v-for="j in info.judgments" :key="j" class="nayin-judgment-tag">{{ j }}</span>
                  </div>
                </div>
              </div>
            </el-popover>
          </div>
        </div>

        <!-- RiZhuDesc (legacy fallback) -->
        <div v-if="chart.ri_zhu_desc && !chart.ri_zhu_poem" class="analysis-block">
          <div class="block-title">日主坐命</div>
          <p class="ri-zhu-text">{{ chart.ri_zhu_desc }}</p>
        </div>

        <!-- RiZhuPoem -->
        <div v-if="chart.ri_zhu_poem" class="analysis-block">
          <div class="block-title">日主诗意</div>
          <p class="ri-zhu-text">{{ chart.ri_zhu_poem }}</p>
        </div>

        <!-- RiZhuSource -->
        <div v-if="chart.ri_zhu_source" class="analysis-block">
          <div class="block-title">古籍出处</div>
          <p class="ri-zhu-text">{{ chart.ri_zhu_source }}</p>
        </div>

        <!-- RiZhuComment -->
        <div v-if="chart.ri_zhu_comment" class="analysis-block">
          <div class="block-title">补充判断</div>
          <p class="ri-zhu-text">{{ chart.ri_zhu_comment }}</p>
        </div>

        <!-- RiZhuHourDetail -->
        <div v-if="chart.ri_zhu_hour_detail" class="analysis-block">
          <div class="block-title">时辰详批</div>
          <p class="ri-zhu-text">{{ chart.ri_zhu_hour_detail }}</p>
        </div>

        <!-- DaYun -->
        <div v-if="chart.da_yun && chart.da_yun.start_age" class="analysis-block">
          <div class="block-title">
            大运 ({{ chart.da_yun.direction }} · {{ chart.da_yun.start_age }}岁起运)
          </div>
          <div class="dayun-list">
            <span v-for="(p, i) in chart.da_yun.pillars" :key="i" class="dayun-tag">
              {{ (Number(chart.da_yun.start_age) || 0) + (Number(i) || 0) * 10 }}岁 {{ p.gan }}{{ p.zhi }}
            </span>
          </div>
        </div>

        <!-- MingGong -->
        <div v-if="chart.ming_gong" class="analysis-block">
          <div class="block-title">命宫</div>
          <p class="ming-gong-text">{{ chart.ming_gong }}</p>
        </div>

        <!-- TiaoHou -->
        <div v-if="chart.tiao_hou" class="analysis-block">
          <div class="block-title">调候用神</div>
          <p class="tiao-hou-text">{{ chart.tiao_hou }}</p>
        </div>

        <!-- JinBuHuan -->
        <div v-if="chart.jin_bu_huan" class="analysis-block">
          <div class="block-title">金不换</div>
          <p class="jin-bu-huan-text">{{ chart.jin_bu_huan }}</p>
        </div>

        <!-- ShenSha (grouped by pillar when available) -->
        <template v-if="groupedShenSha">
          <template v-for="group in groupedShenSha" :key="group.pillar">
            <div v-if="group.items && group.items.length" class="analysis-block">
              <div class="shen-sha-group-title">
                <span class="shen-sha-group-dot" :style="{ background: pillarShenShaColor(group.pillar) }"></span>
                {{ group.label }}神煞
                <span class="shen-sha-group-role">· {{ group.role }}</span>
              </div>
              <div class="shen-sha-list">
                <article
                  v-for="sha in group.items"
                  :key="sha.name + sha.target + sha.desc"
                  class="shen-sha-row"
                  :style="{ background: pillarShenShaBg(group.pillar) }"
                >
                  <span class="shen-sha-name" :style="{ color: pillarShenShaColor(group.pillar) }">{{ sha.name }}</span>
                  <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                  <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
                </article>
              </div>
            </div>
          </template>
          <div v-if="globalShenSha.length" class="analysis-block">
            <div class="shen-sha-group-title">
              <span class="shen-sha-group-dot" style="background: var(--gold)"></span>
              全局组合神煞
              <span class="shen-sha-group-role">· 多柱配合</span>
            </div>
            <div class="shen-sha-list">
              <article
                v-for="sha in globalShenSha"
                :key="sha.name + sha.target + sha.desc"
                class="shen-sha-row"
                :style="{ background: 'rgba(212,168,75,0.07)' }"
              >
                <span class="shen-sha-name" :style="{ color: 'var(--gold)' }">{{ sha.name }}</span>
                <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
                <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
              </article>
            </div>
          </div>
          <div v-if="showSummary" class="analysis-block shen-sha-summary-block">
            <div class="shen-sha-summary-title">{{ props.chart.shen_sha_summary.title }}</div>
            <ul class="shen-sha-summary-list">
              <li v-for="line in props.chart.shen_sha_summary.description" :key="line.slice(0, 16)">
                {{ line }}
              </li>
            </ul>
          </div>
        </template>
        <div v-else-if="parsedDayShenSha.length" class="analysis-block">
          <div class="block-title">日柱神煞</div>
          <div class="shen-sha-list">
            <article v-for="sha in parsedDayShenSha" :key="sha.name + sha.target + sha.desc" class="shen-sha-row">
              <span class="shen-sha-name">{{ sha.name }}</span>
              <span v-if="sha.target" class="shen-sha-target">{{ sha.target }}</span>
              <span v-if="sha.desc" class="shen-sha-desc">{{ sha.desc }}</span>
            </article>
          </div>
        </div>

        <!-- SeasonText -->
        <div v-if="chart.season_text" class="analysis-block">
          <div class="block-title">季节解读</div>
          <p class="season-text">{{ chart.season_text }}</p>
        </div>

        <!-- SeasonTextMonth (五月/六月 specific) -->
        <div v-if="chart.season_text_month" class="analysis-block">
          <div class="block-title">{{ birthMonthLabel }}解读</div>
          <p class="season-text">{{ chart.season_text_month }}</p>
        </div>

      </div>
    </div>
  </div>
</template>

<style scoped>
.bazi-chart {
  position: relative;
}

.chart-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: hidden;
}

.bg-svg {
  width: 100%;
  height: 100%;
  position: absolute;
  top: 0;
  left: 0;
}

.chart-card {
  position: relative;
  z-index: 1;
}

/* Header */
.chart-header {
  background: linear-gradient(180deg, rgba(212, 168, 75, 0.06), transparent);
  border-bottom: 1px solid rgba(212, 168, 75, 0.08);
  padding: 1rem 1.25rem;
  text-align: center;
}

.header-eyebrow {
  font-size: 10px;
  letter-spacing: 3px;
  color: rgba(212, 168, 75, 0.35);
  text-transform: uppercase;
  margin-bottom: 4px;
}

.chart-title {
  font-family: var(--font-serif), serif;
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text);
  margin: 0;
  letter-spacing: 3px;
}

/* Pillars grid */
.pillars-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  border-bottom: 1px solid rgba(212, 168, 75, 0.06);
}

.pillar-col {
  display: flex;
  flex-direction: column;
  text-align: center;
  border-right: 1px solid rgba(212, 168, 75, 0.06);
}

.pillar-col:last-child {
  border-right: none;
}

.pillar-label {
  background: rgba(255, 255, 255, 0.02);
  padding: 0.5rem 0;
  font-size: 0.72rem;
  color: var(--muted);
  letter-spacing: 1px;
  border-bottom: 1px solid rgba(212, 168, 75, 0.06);
}

.pillar-gan,
.pillar-zhi {
  padding: 1rem 0.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  border-bottom: 1px solid rgba(212, 168, 75, 0.04);
}

.gan-char,
.zhi-char {
  font-family: var(--font-serif), serif;
  font-size: 2.2rem;
  font-weight: 700;
  line-height: 1;
}

.elem-tag {
  display: inline-block;
  font-size: 0.65rem;
  padding: 0.1rem 0.4rem;
  border-radius: 4px;
  border: 1px solid;
}

/* Relations */
.relations-section {
  padding: 0.875rem 1.25rem;
  border-bottom: 1px solid rgba(212, 168, 75, 0.06);
}

.relations-title {
  font-size: 0.75rem;
  color: var(--muted);
  letter-spacing: 1px;
  margin-bottom: 0.5rem;
  text-transform: uppercase;
}

.relations-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.rel-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.35rem 0.75rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 500;
}

/* 天干地支分析 */
.ganzhi-analysis {
  margin-top: 4px;
}

/* 关系颜色 */
.rel-he {
  background: rgba(74, 222, 128, 0.08);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.15);
}

.rel-ke {
  background: rgba(196, 30, 58, 0.08);
  color: var(--crimson);
  border: 1px solid rgba(196, 30, 58, 0.15);
}

.rel-sheng {
  background: rgba(65, 105, 225, 0.08);
  color: #6495ed;
  border: 1px solid rgba(65, 105, 225, 0.15);
}

.rel-chong {
  background: rgba(196, 30, 58, 0.1);
  color: var(--crimson);
  border: 1px solid rgba(196, 30, 58, 0.2);
}

.rel-hai {
  background: rgba(255, 140, 0, 0.08);
  color: #ff8c00;
  border: 1px solid rgba(255, 140, 0, 0.15);
}

.rel-xing {
  background: rgba(138, 43, 226, 0.08);
  color: #ba55d3;
  border: 1px solid rgba(138, 43, 226, 0.15);
}

.rel-hui {
  background: rgba(65, 105, 225, 0.08);
  color: #6495ed;
  border: 1px solid rgba(65, 105, 225, 0.15);
}

.rel-symbol {
  font-weight: 700;
}

.rel-detail {
  font-size: 0.7rem;
  opacity: 0.7;
  margin-left: 2px;
}

.no-relations {
  padding: 0.875rem 1.25rem;
  text-align: center;
  font-size: 0.78rem;
  color: rgba(255, 255, 255, 0.15);
  border-bottom: 1px solid rgba(212, 168, 75, 0.06);
}

/* Analysis sections */
.analysis-section {
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.analysis-block {
}

.block-title {
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--gold);
  letter-spacing: 1px;
  margin-bottom: 0.5rem;
}

/* Five elements chart */
.five-elem-chart {
  height: 148px;
  width: 100%;
}

/* Ten gods */
.ten-gods-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.4rem;
}

.god-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.4rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  border: 1px solid rgba(212, 168, 75, 0.06);
}

.god-pillar {
  font-size: 0.65rem;
  color: var(--muted);
  margin-bottom: 0.2rem;
}

.god-name {
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--crimson);
}

/* NaYin */
.nayin-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.nayin-tag {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.72rem;
  padding: 0.25rem 0.6rem;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid;
  border-radius: 4px;
  color: var(--text);
  cursor: pointer;
  transition: background 0.2s;
}

.nayin-tag:hover {
  background: rgba(255, 255, 255, 0.08);
}

.nayin-pillar {
  opacity: 0.6;
  font-size: 0.68rem;
}

.nayin-name {
  font-weight: 600;
}

/* NaYin Detail Popover */
.nayin-detail {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.25rem 0;
}

.nayin-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(212, 168, 75, 0.15);
}

.nayin-detail-name {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--gold);
}

.nayin-detail-elem {
  font-size: 0.68rem;
  padding: 0.15rem 0.5rem;
  border-radius: 3px;
  color: #1a1a1a;
  font-weight: 600;
}

.nayin-detail-section {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.nayin-detail-label {
  font-size: 0.65rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: rgba(255, 255, 255, 0.4);
}

.nayin-detail-value {
  font-size: 0.82rem;
  line-height: 1.5;
  color: var(--text);
}

.nayin-energy {
  color: var(--gold);
  font-weight: 600;
}

.nayin-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin-top: 0.15rem;
}

.nayin-judgment-tag {
  font-size: 0.7rem;
  padding: 0.15rem 0.5rem;
  background: rgba(212, 168, 75, 0.08);
  border: 1px solid rgba(212, 168, 75, 0.15);
  border-radius: 3px;
  color: var(--gold);
}

/* DaYun */
.dayun-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.dayun-tag {
  font-size: 0.72rem;
  padding: 0.25rem 0.625rem;
  background: rgba(196, 30, 58, 0.06);
  border: 1px solid rgba(196, 30, 58, 0.12);
  border-radius: 4px;
  color: var(--crimson);
}

/* Element Detail Table */
.element-detail-table {
  border: 1px solid rgba(212, 168, 75, 0.1);
  border-radius: 8px;
  overflow: hidden;
}

.ed-header {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  background: rgba(255, 255, 255, 0.03);
  font-size: 0.65rem;
  color: var(--muted);
  letter-spacing: 0.5px;
  border-bottom: 1px solid rgba(212, 168, 75, 0.08);
}

.ed-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr;
  padding: 0.4rem 0.75rem;
  font-size: 0.78rem;
  border-bottom: 1px solid rgba(212, 168, 75, 0.04);
}

.ed-row:last-child {
  border-bottom: none;
}

.ed-elem {
  font-weight: 700;
}

.ed-tg,
.ed-zc,
.ed-total {
  color: var(--text);
}

.level-none { background: rgba(255, 255, 255, 0.01); }
.level-weak .ed-total { color: rgba(255, 255, 255, 0.4); }
.level-medium .ed-total { color: #9ca3af; }
.level-strong .ed-total { color: #fbbf24; font-weight: 700; }
.level-very-strong .ed-total { color: #f97316; font-weight: 800; }

/* Body Strength */
.body-strength {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.bs-verdict {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--gold);
}

.bs-tags {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.bs-like-label,
.bs-dislike-label {
  font-size: 0.65rem;
  padding: 0.15rem 0.5rem;
  border-radius: 4px;
}

.bs-like-label {
  background: rgba(74, 222, 128, 0.1);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.2);
}

.bs-dislike-label {
  background: rgba(196, 30, 58, 0.1);
  color: var(--crimson);
  border: 1px solid rgba(196, 30, 58, 0.2);
}

.bs-like {
  font-size: 0.72rem;
  padding: 0.15rem 0.5rem;
  background: rgba(74, 222, 128, 0.06);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.12);
  border-radius: 4px;
}

.bs-dislike {
  font-size: 0.72rem;
  padding: 0.15rem 0.5rem;
  background: rgba(196, 30, 58, 0.06);
  color: var(--crimson);
  border: 1px solid rgba(196, 30, 58, 0.12);
  border-radius: 4px;
}

.ming-gong-text,
.ri-zhu-text,
.tiao-hou-text,
.jin-bu-huan-text,
.season-text {
  font-size: 0.82rem;
  color: rgba(255, 255, 255, 0.65);
  line-height: 1.7;
  white-space: pre-wrap;
  margin: 0;
}

/* Pillar sub info (shengxiao + empties) */
.pillar-sub {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
  padding: 0.4rem 0.25rem;
  border-top: 1px solid rgba(212, 168, 75, 0.04);
}

.sheng-xiao-tag {
  font-size: 0.65rem;
  color: rgba(212, 168, 75, 0.5);
  letter-spacing: 1px;
}

.empties-tag {
  font-size: 0.6rem;
  color: rgba(196, 30, 58, 0.5);
}

/* ShenSha list */
.shen-sha-list {
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
}

.shen-sha-row {
  display: grid;
  grid-template-columns: 4.6rem auto 1fr;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.1rem;
  padding: 0.35rem 0.65rem;
  background: rgba(255, 255, 255, 0.028);
  border: 1px solid rgba(212, 168, 75, 0.1);
  border-radius: 7px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.035);
}

.shen-sha-name {
  font-family: var(--font-serif), serif;
  font-size: 0.8rem;
  font-weight: 700;
  color: var(--gold);
  letter-spacing: 0.06em;
  white-space: nowrap;
}

.shen-sha-target {
  min-width: 1.35rem;
  height: 1.35rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgba(212, 168, 75, 0.1);
  border: 1px solid rgba(212, 168, 75, 0.18);
  color: rgba(240, 237, 228, 0.86);
  font-size: 0.68rem;
  font-weight: 700;
}

.shen-sha-desc {
  color: rgba(240, 237, 228, 0.58);
  font-size: 0.72rem;
  line-height: 1.45;
}

/* ShenSha group title */
.shen-sha-group-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-serif), serif;
  font-size: 0.86rem;
  font-weight: 700;
  color: var(--text);
  margin-bottom: 0.55rem;
  letter-spacing: 0.06em;
}

.shen-sha-group-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 50%;
  flex-shrink: 0;
}

.shen-sha-group-role {
  font-family: var(--font-sans), sans-serif;
  font-size: 0.66rem;
  font-weight: 400;
  color: rgba(240, 237, 228, 0.4);
  letter-spacing: 0.03em;
}

/* ShenSha summary */
.shen-sha-summary-block {
  border-top: 1px solid rgba(212, 168, 75, 0.08);
  margin-top: 0.5rem;
  padding-top: 0.75rem;
}

.shen-sha-summary-title {
  font-size: 0.78rem;
  font-weight: 600;
  color: rgba(240, 237, 228, 0.65);
  margin-bottom: 0.45rem;
}

.shen-sha-summary-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.shen-sha-summary-list li {
  font-size: 0.7rem;
  color: rgba(240, 237, 228, 0.48);
  line-height: 1.5;
  padding-left: 0.65rem;
  position: relative;
}

.shen-sha-summary-list li::before {
  content: '–';
  position: absolute;
  left: 0;
  color: rgba(212, 168, 75, 0.3);
}

.ten-god-chart-wrap {
  border: 1px solid rgba(212, 168, 75, 0.12);
  border-radius: 10px;
  overflow: hidden;
  background:
    linear-gradient(180deg, rgba(212,168,75,0.03) 0%, transparent 100%),
    rgba(255, 255, 255, 0.015);
  padding: 0.75rem 0.5rem 0.5rem;
  box-shadow: inset 0 1px 0 rgba(212,168,75,0.08);
}

.ten-god-chart {
  width: 100%;
  height: 220px;
}
</style>

<style>
/* NaYin popover — dark theme to match app */
.nayin-popover {
  background: #0a0815 !important;
  border: 1px solid rgba(212, 168, 75, 0.15) !important;
  border-radius: 10px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6), 0 0 24px rgba(212, 168, 75, 0.08) !important;
  padding: 14px 16px !important;
  color: #f0ede4;
}
.nayin-popover .el-popover__title {
  color: #f0ede4;
}
</style>
