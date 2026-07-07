<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CircleDotIcon, RefreshCwIcon, SparklesIcon, TriangleAlertIcon } from '@lucide/vue'
import { drawBuyiToday, fetchBuyiToday, type BuyiTodayResponse } from '../api/buyi'

const data = ref<BuyiTodayResponse | null>(null)
const loading = ref(true)
const drawing = ref(false)
const error = ref('')

const record = computed(() => data.value?.record ?? null)
const scoreWidth = computed(() => `${Math.max(0, Math.min(100, record.value?.score ?? 0))}%`)
const statusText = computed(() => (record.value ? '今日已卜' : '今日可卜'))

async function loadToday() {
  loading.value = true
  error.value = ''
  try {
    data.value = await fetchBuyiToday()
  } catch (e: any) {
    error.value = e?.response?.data?.error || '加载卜易失败'
  } finally {
    loading.value = false
  }
}

async function drawToday() {
  drawing.value = true
  error.value = ''
  try {
    data.value = await drawBuyiToday()
  } catch (e: any) {
    error.value = e?.response?.data?.error || '卜易失败'
  } finally {
    drawing.value = false
  }
}

onMounted(loadToday)
</script>

<template>
  <main class="buyi-page">
    <section class="buyi-shell">
      <header class="buyi-head">
        <div>
          <p class="eyebrow">BUYI · {{ data?.date || '今日' }}</p>
          <h1>卜易</h1>
        </div>
        <span class="status-pill" :class="{ drawn: record }">
          <CircleDotIcon class="status-icon" />
          {{ statusText }}
        </span>
      </header>

      <div v-if="loading" class="state-panel">
        <RefreshCwIcon class="state-icon spinning" />
        <p>正在观象</p>
      </div>

      <div v-else-if="error" class="state-panel error-panel">
        <TriangleAlertIcon class="state-icon" />
        <p>{{ error }}</p>
        <button class="ghost-btn" type="button" @click="loadToday">
          <RefreshCwIcon class="btn-icon" />
          重试
        </button>
      </div>

      <div v-else-if="!record" class="draw-panel">
        <div class="gua-seal" aria-hidden="true">
          <span></span><span></span><span></span><span></span><span></span><span></span>
        </div>
        <div class="draw-copy">
          <p class="draw-title">今日尚未卜易</p>
          <p class="draw-date">{{ data?.date }}</p>
        </div>
        <button class="draw-btn" type="button" :disabled="drawing" @click="drawToday">
          <SparklesIcon class="btn-icon" />
          {{ drawing ? '起卦中' : '起卦' }}
        </button>
      </div>

      <article v-else class="result-panel">
        <div class="result-hero">
          <div class="hexagram-meta">
            <span class="hexagram-no">第 {{ record.hexagram_number }} 卦</span>
            <h2>{{ record.hexagram_name }}</h2>
            <p>{{ record.summary }}</p>
          </div>
          <div class="score-box">
            <span class="score-value">{{ record.score }}</span>
            <span class="score-level">{{ record.level }}</span>
          </div>
        </div>

        <div class="score-track" aria-label="吉凶评分">
          <div class="score-fill" :style="{ width: scoreWidth }"></div>
        </div>

        <div class="reading-grid">
          <section class="reading-card">
            <span>卦义</span>
            <p>{{ record.human_way }}</p>
          </section>
          <section class="reading-card">
            <span>象解</span>
            <p>{{ record.image_reading }}</p>
          </section>
          <section class="reading-card reading-wide">
            <span>今日建议</span>
            <p>{{ record.advice }}</p>
          </section>
        </div>

        <footer class="result-foot">
          <span>{{ record.source }}</span>
          <time>{{ record.created_at }}</time>
        </footer>
      </article>
    </section>
  </main>
</template>

<style scoped>
.buyi-page {
  min-height: calc(100vh - 80px);
  padding: 40px 20px 72px;
  color: var(--text);
  background:
    linear-gradient(180deg, transparent, var(--surface-1)),
    var(--bg);
}

.buyi-shell {
  width: min(980px, 100%);
  margin: 0 auto;
}

.buyi-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  margin-bottom: 24px;
}

.eyebrow {
  margin: 0 0 8px;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
  letter-spacing: 0.18em;
}

.buyi-head h1 {
  margin: 0;
  font-family: var(--font-serif);
  font-size: var(--fs-5xl);
  line-height: 1.15;
  letter-spacing: 0;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 0 14px;
  border: 1px solid var(--line-strong);
  border-radius: 999px;
  color: var(--text-muted);
  background: var(--glass-bg);
  font-size: var(--fs-xs);
  white-space: nowrap;
}

.status-pill.drawn {
  color: var(--accent);
  border-color: var(--line-focus);
  background: var(--accent-dim);
}

.status-icon,
.btn-icon {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
}

.state-panel,
.draw-panel,
.result-panel {
  border: 1px solid var(--line-strong);
  border-radius: 8px;
  background: color-mix(in oklab, var(--surface-0) 84%, transparent);
  box-shadow: var(--shadow-md);
}

.state-panel {
  min-height: 320px;
  display: grid;
  place-items: center;
  align-content: center;
  gap: 14px;
  color: var(--text-muted);
}

.state-panel p {
  margin: 0;
  font-size: var(--fs-sm);
}

.state-icon {
  width: 28px;
  height: 28px;
  color: var(--accent);
}

.error-panel .state-icon {
  color: var(--danger);
}

.spinning {
  animation: spin 1.2s linear infinite;
}

.draw-panel {
  min-height: 420px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 24px;
  padding: 48px 24px;
  text-align: center;
}

.gua-seal {
  width: 132px;
  height: 132px;
  display: grid;
  place-content: center;
  gap: 9px;
  border: 1px solid var(--line-focus);
  border-radius: 50%;
  background: var(--accent-dim);
}

.gua-seal span {
  display: block;
  width: 68px;
  height: 5px;
  border-radius: 999px;
  background: var(--accent);
  box-shadow: 0 0 16px var(--accent-glow);
}

.gua-seal span:nth-child(2),
.gua-seal span:nth-child(5) {
  background: linear-gradient(90deg, var(--accent) 0 38%, transparent 38% 62%, var(--accent) 62% 100%);
}

.draw-title {
  margin: 0;
  font-family: var(--font-serif);
  font-size: var(--fs-3xl);
  line-height: 1.3;
}

.draw-date {
  margin: 8px 0 0;
  color: var(--text-soft);
  font-size: var(--fs-sm);
}

.draw-btn,
.ghost-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  min-height: 42px;
  padding: 0 18px;
  border: 1px solid transparent;
  border-radius: 999px;
  cursor: pointer;
  font-size: var(--fs-sm);
  font-weight: 600;
  transition: transform 0.18s ease, border-color 0.18s ease, background 0.18s ease;
}

.draw-btn {
  color: var(--bg);
  background: var(--text);
}

.draw-btn:disabled {
  opacity: 0.66;
  cursor: wait;
}

.draw-btn:not(:disabled):hover,
.ghost-btn:hover {
  transform: translateY(-1px);
}

.ghost-btn {
  color: var(--text);
  border-color: var(--line-strong);
  background: var(--glass-bg);
}

.result-panel {
  padding: 28px;
}

.result-hero {
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  gap: 24px;
}

.hexagram-meta {
  min-width: 0;
}

.hexagram-no {
  display: inline-flex;
  margin-bottom: 10px;
  color: var(--accent);
  font-size: var(--fs-xs);
  font-weight: 600;
}

.hexagram-meta h2 {
  margin: 0;
  font-family: var(--font-serif);
  font-size: var(--fs-5xl);
  line-height: 1.18;
  letter-spacing: 0;
}

.hexagram-meta p {
  max-width: 680px;
  margin: 14px 0 0;
  color: var(--text-muted);
  font-size: var(--fs-md);
  line-height: 1.75;
  word-break: break-word;
}

.score-box {
  width: 112px;
  min-width: 112px;
  min-height: 112px;
  display: grid;
  place-items: center;
  align-content: center;
  border: 1px solid var(--line-focus);
  border-radius: 8px;
  background: var(--accent-dim);
}

.score-value {
  font-family: var(--font-serif);
  font-size: var(--fs-stat-lg);
  line-height: 1;
  color: var(--accent);
}

.score-level {
  margin-top: 8px;
  color: var(--text-muted);
  font-size: var(--fs-xs);
  font-weight: 600;
}

.score-track {
  height: 8px;
  margin: 26px 0;
  overflow: hidden;
  border-radius: 999px;
  background: var(--surface-2);
}

.score-fill {
  height: 100%;
  border-radius: inherit;
  background: var(--accent);
  box-shadow: 0 0 18px var(--accent-glow);
}

.reading-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
}

.reading-card {
  min-height: 148px;
  padding: 18px;
  border: 1px solid var(--line-subtle);
  border-radius: 8px;
  background: var(--surface-1);
}

.reading-wide {
  grid-column: 1 / -1;
  min-height: 112px;
}

.reading-card span {
  display: block;
  margin-bottom: 10px;
  color: var(--accent);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.reading-card p {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  line-height: 1.8;
  word-break: break-word;
}

.result-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-top: 20px;
  color: var(--text-soft);
  font-size: var(--fs-2xs);
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 720px) {
  .buyi-page {
    padding: 28px 14px 56px;
  }

  .buyi-head,
  .result-hero,
  .result-foot {
    flex-direction: column;
    align-items: flex-start;
  }

  .result-panel {
    padding: 20px;
  }

  .score-box {
    width: 100%;
    min-width: 0;
    min-height: 92px;
  }

  .reading-grid {
    grid-template-columns: 1fr;
  }
}
</style>
