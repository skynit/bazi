<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface ChartData {
  day_pillar?: { gan: string; zhi: string }
  year_pillar?: { gan: string; zhi: string }
  month_pillar?: { gan: string; zhi: string }
  hour_pillar?: { gan: string; zhi: string }
  wuxing_distribution?: Record<string, number>
  yongshen?: string
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

  const wuxing = chart?.wuxing_distribution || { 金: 15, 木: 20, 水: 10, 火: 40, 土: 15 }

  const lines = [
    `▶ 初始化天干地支五行矩阵... 成功`,
    `▶ 正在检索生辰星历...`,
    `▶ 正在分析日元强度：日主【${dayGan}${dayElement}】，生于当令之月`,
    `▶ 正在计算五行损益比例：金 (${wuxing.金 || 15}%) | 木 (${wuxing.木 || 20}%) | 水 (${wuxing.水 || 10}%) | 火 (${wuxing.火 || 40}%) | 土 (${wuxing.土 || 15}%)`,
    `✔ 命运解构完成。正在生成专属于您的因果律报告。`
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
      <span class="log-dot animate-ping bg-cyan-400"></span>
      <span class="log-title">命运引擎 · 推演算法</span>
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
      <span class="complete-dot bg-emerald-400"></span>
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
  font-size: 0.72rem;
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
  font-size: 0.78rem;
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
  font-size: 0.72rem;
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
