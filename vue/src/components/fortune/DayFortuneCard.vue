<script setup lang="ts">
/**
 * DayFortuneCard — compact daily fortune card surfacing fields the older
 * WeeklyView discarded: score, lucky color, lucky number, wealth direction,
 * yi/ji preview, optional ten-god label.
 */
import { computed } from 'vue'

interface Props {
  date: string
  dayPillar: string
  score?: number
  luckyColor?: string
  luckyNumber?: number
  wealthDir?: string
  guideStrategy?: string
  guideElement?: string
  yiItems?: string[]
  jiItems?: string[]
  todayTenGod?: string
  isBest?: boolean
  isWorst?: boolean
  weekday?: string
}
const props = defineProps<Props>()

const tone = computed(() => {
  const s = Math.max(0, Math.min(100, props.score ?? 0))
  const t = s / 100
  return `color-mix(in oklab, var(--jade-accent) ${Math.round(52 + t * 48)}%, var(--text) ${Math.round((1 - t) * 18)}%)`
})

// luckyColor may come back as a Chinese name; map common ones to a CSS color
function colorSwatch(name: string | undefined): string {
  if (!name) return 'transparent'
  const m: Record<string, string> = {
    '红色': '#e84057', '红': '#e84057', '朱红': '#dc2626',
    '橙色': '#f97316', '黄色': '#fcd34d', '金色': '#d4a017', '金': '#d4a017',
    '绿色': '#22c55e', '青色': '#10b981', '翠': '#34d399',
    '蓝色': '#3b82f6', '青蓝': '#0ea5e9',
    '紫色': '#a855f7', '紫': '#a855f7',
    '黑色': '#1f2937', '黑': '#1f2937',
    '白色': '#f8fafc', '白': '#f8fafc',
    '灰色': '#94a3b8',
  }
  return m[name] ?? '#94a3b8'
}
const swatch = computed(() => colorSwatch(props.luckyColor))

const yiPreview = computed(() => (props.yiItems ?? []).slice(0, 3).join('·'))
const jiPreview = computed(() => (props.jiItems ?? []).slice(0, 2).join('·'))
</script>

<template>
  <article
    class="card"
    :class="{ best: isBest, worst: isWorst }"
    :title="`${date} ${dayPillar}`"
  >
    <header class="head">
      <div class="left">
        <span v-if="weekday" class="weekday">{{ weekday }}</span>
        <span class="date tabular-nums">{{ date.slice(5) }}</span>
      </div>
      <span class="pillar">{{ dayPillar }}</span>
    </header>

    <div class="score-line">
      <span class="score-num tabular-nums" :style="{ color: tone }">{{ score ?? '—' }}</span>
      <span v-if="todayTenGod" class="ten-god">{{ todayTenGod }}</span>
    </div>

    <dl class="lucky">
      <div class="row" v-if="guideElement">
        <dt>主气</dt>
        <dd>{{ guideElement }}</dd>
      </div>
      <div class="row" v-if="luckyColor">
        <dt>幸运色</dt>
        <dd><span class="swatch" :style="{ background: swatch }"></span>{{ luckyColor }}</dd>
      </div>
      <div class="row" v-if="typeof luckyNumber === 'number'">
        <dt>幸运数</dt>
        <dd class="tabular-nums">{{ luckyNumber }}</dd>
      </div>
      <div class="row" v-if="wealthDir">
        <dt>财位</dt>
        <dd>{{ wealthDir }}</dd>
      </div>
    </dl>

    <p v-if="guideStrategy" class="guide"><span class="tag guide-tag">策</span>{{ guideStrategy }}</p>
    <p v-if="yiPreview" class="yi"><span class="tag yi-tag">宜</span>{{ yiPreview }}</p>
    <p v-if="jiPreview" class="ji"><span class="tag ji-tag">忌</span>{{ jiPreview }}</p>
  </article>
</template>

<style scoped>
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 14px 16px;
  border-radius: 16px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  backdrop-filter: blur(14px) saturate(140%);
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.25s ease, box-shadow 0.25s ease;
  min-width: 0;
}
.card:hover {
  transform: translateY(-2px);
  border-color: rgba(var(--jade-accent-rgb), 0.42);
  box-shadow: 0 14px 36px rgba(0,0,0,0.06);
}
.card.best {
  border-color: rgba(var(--jade-accent-rgb), 0.6);
  box-shadow: 0 0 0 1px rgba(var(--jade-accent-rgb), 0.25) inset;
}
.card.worst {
  border-color: rgba(232,64,87,0.5);
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}
.left { display: flex; align-items: baseline; gap: 8px; }
.weekday {
  font-size: 0.7rem;
  letter-spacing: 0.18em;
  color: var(--text-muted);
  text-transform: uppercase;
}
.date { font-size: 0.95rem; font-weight: 700; color: var(--text); letter-spacing: 0.02em; }
.pillar {
  font-family: var(--font-serif), serif;
  font-size: 0.9rem;
  font-weight: 700;
  color: rgba(var(--jade-accent-rgb), 1);
  letter-spacing: 0.06em;
}

.score-line {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}
.score-num {
  font-family: var(--font-serif), serif;
  font-size: 2rem;
  font-weight: 800;
  line-height: 1;
}
.ten-god {
  font-size: 0.7rem;
  color: var(--text-muted);
  letter-spacing: 0.06em;
  padding: 2px 8px;
  border-radius: 999px;
  border: 1px solid var(--line-subtle);
}

.lucky {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px 12px;
  margin: 0;
  font-size: 0.72rem;
}
.lucky .row { display: flex; gap: 6px; align-items: center; min-width: 0; }
.lucky dt { color: var(--text-soft); margin: 0; flex-shrink: 0; }
.lucky dd { margin: 0; color: var(--text); display: inline-flex; align-items: center; gap: 4px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.swatch {
  display: inline-block;
  width: 10px; height: 10px;
  border-radius: 50%;
  border: 1px solid var(--line-strong);
}

.yi, .ji, .guide {
  margin: 0;
  font-size: 0.72rem;
  color: var(--text-muted);
  line-height: 1.4;
  display: flex; gap: 6px; align-items: baseline;
}
.tag {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 0.6rem;
  font-weight: 700;
  letter-spacing: 0.05em;
}
.yi-tag { background: rgba(var(--jade-accent-rgb), 0.14); color: rgba(var(--jade-accent-rgb), 1); }
.ji-tag { background: rgba(232,64,87,0.14); color: var(--crimson); }
.guide-tag { background: var(--accent-dim); color: var(--accent); }

.tabular-nums { font-variant-numeric: tabular-nums; }
</style>
