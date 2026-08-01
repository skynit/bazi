<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface ChartData {
  day_pillar?: { gan: string; zhi: string }
  year_pillar?: { gan: string; zhi: string }
  month_pillar?: { gan: string; zhi: string }
  hour_pillar?: { gan: string; zhi: string }
  five_elements?: Record<string, number>
  body_strength?: { score_band_candidate?: string }
  [key: string]: any
}

const props = defineProps<{
  chartData?: ChartData
}>()

const emit = defineEmits<{
  complete: []
}>()

const visibleLines = ref(0)
const completed = ref(false)

const ganElement: Record<string, string> = {
  甲: '木', 乙: '木', 丙: '火', 丁: '火',
  戊: '土', 己: '土', 庚: '金', 辛: '金',
  壬: '水', 癸: '水'
}

const logLines = computed(() => {
  const chart = props.chartData
  const dayGan = chart?.day_pillar?.gan || '庚'
  const dayElement = ganElement[dayGan] || '金'

  const wuxing = chart?.five_elements || {}
  const bodyBand = chart?.body_strength?.score_band_candidate || '暂无'
  const score = (element: string) =>
    typeof wuxing[element] === 'number' ? String(wuxing[element]) : '—'

  const lines = [
    `▶ 初始化干支与五行规则...`,
    `▶ 正在计算四柱结构...`,
    `▶ 日主【${dayGan}${dayElement}】；身强本地分段候选【${bodyBand}】`,
    `▶ 原始五行计分：金 ${score('金')} | 木 ${score('木')} | 水 ${score('水')} | 火 ${score('火')} | 土 ${score('土')}`,
    `✔ 命盘结构计算完成，正在载入证据明细。`
  ]
  return lines
})

onMounted(() => {
  const totalLines = logLines.value.length
  let current = 0

  const interval = setInterval(() => {
    current++
    visibleLines.value = current

    if (current >= totalLines) {
      clearInterval(interval)
      completed.value = true
      setTimeout(() => {
        emit('complete')
      }, 600)
    }
  }, 700)
})
</script>

<template>
  <div class="computation-log">
    <div class="log-header">
      <span class="log-dot animate-ping" style="background: var(--wuxing-shui)"></span>
      <span class="log-title">命盘计算 · 结构证据</span>
    </div>

    <div class="log-terminal">
      <div
        v-for="(line, idx) in logLines"
        :key="idx"
        class="log-line"
        :class="{ 'log-visible': idx < visibleLines, 'log-success': line.startsWith('✔') }"
      >
        <span class="log-text">{{ line }}</span>
      </div>

      <div v-if="!completed" class="log-line log-typing">
        <span class="log-cursor"></span>
      </div>
    </div>

    <div v-if="completed" class="log-complete-badge">
      <span class="complete-dot" style="background: var(--wuxing-mu)"></span>
      解构完成
    </div>
  </div>
</template>

<style scoped>
.computation-log {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 2rem;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.log-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.log-title {
  font-family: var(--font-mono);
  font-size: var(--fs-xs);
  font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 3px;
  text-transform: uppercase;
}

.log-terminal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  padding: 1.25rem;
  font-family: var(--font-mono);
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.log-line {
  opacity: 0;
  transform: translateY(8px);
  transition: opacity 0.4s, transform 0.4s;
}

.log-visible {
  opacity: 1;
  transform: translateY(0);
}

.log-success .log-text {
  color: var(--wuxing-mu);
}

.log-text {
  font-size: var(--fs-xs);
  color: var(--text-muted);
  line-height: 1.6;
  white-space: pre-wrap;
}

.log-typing {
  opacity: 1;
  transform: translateY(0);
}

.log-cursor {
  display: inline-block;
  width: 8px;
  height: 14px;
  background: var(--wuxing-shui);
  animation: cursor-blink 1s step-end infinite;
}

@keyframes cursor-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

.log-complete-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 1rem;
  background: rgba(22, 163, 74, 0.08);
  border: 1px solid rgba(22, 163, 74, 0.2);
  border-radius: 20px;
  font-size: var(--fs-xs);
  color: var(--wuxing-mu);
  font-weight: 600;
  letter-spacing: 1px;
  animation: fadeUp 0.5s ease both;
}

.complete-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

@keyframes fadeUp {
  from { opacity: 0; transform: translateY(12px); }
  to { opacity: 1; transform: translateY(0); }
}

:global(.dark) .log-complete-badge {
  background: rgba(52, 211, 153, 0.06);
  border-color: rgba(52, 211, 153, 0.2);
}
</style>
