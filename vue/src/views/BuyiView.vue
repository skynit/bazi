<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { CircleDotIcon, RefreshCwIcon, SparklesIcon, TriangleAlertIcon } from '@lucide/vue'
import { drawBuyiToday, fetchBuyiToday, type BuyiTodayResponse } from '../api/buyi'
import { getApiErrorMessage } from '../api/client'

const data = ref<BuyiTodayResponse | null>(null)
const loading = ref(true)
const drawing = ref(false)
const error = ref('')
const failedAction = ref<'load' | 'draw'>('load')

const record = computed(() => data.value?.record ?? null)
const statusText = computed(() => (record.value ? '今日已抽取' : '今日可抽取'))

async function loadToday() {
  failedAction.value = 'load'
  loading.value = true
  error.value = ''
  try {
    data.value = await fetchBuyiToday()
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '今日卜易加载失败，请稍后重试。')
  } finally {
    loading.value = false
  }
}

async function drawToday() {
  failedAction.value = 'draw'
  drawing.value = true
  error.value = ''
  try {
    data.value = await drawBuyiToday()
  } catch (reason: unknown) {
    error.value = getApiErrorMessage(reason, '起卦失败，请稍后重试。')
  } finally {
    drawing.value = false
  }
}

function retryLastAction() {
  if (failedAction.value === 'draw') return drawToday()
  return loadToday()
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
        <p>正在加载今日记录</p>
      </div>

      <div v-else-if="error" class="state-panel error-panel">
        <TriangleAlertIcon class="state-icon" />
        <p>{{ error }}</p>
        <button class="ghost-btn" type="button" @click="retryLastAction">
          <RefreshCwIcon class="btn-icon" />
          重试
        </button>
      </div>

      <div v-else-if="!record" class="draw-panel">
        <div class="gua-seal" aria-hidden="true">
          <span></span><span></span><span></span><span></span><span></span><span></span>
        </div>
        <div class="draw-copy">
          <p class="draw-title">今日尚未抽取</p>
          <p class="draw-date">{{ data?.date }}</p>
          <p class="draw-boundary">结果来自随机抽取，仅作传统文本反思，不代表吉凶或现实结果。</p>
        </div>
        <button class="draw-btn" type="button" :disabled="drawing" @click="drawToday">
          <SparklesIcon class="btn-icon" />
          {{ drawing ? '抽取中' : '随机抽取一卦' }}
        </button>
      </div>

      <article v-else class="result-panel">
        <div class="result-hero">
          <div class="hexagram-meta">
            <span class="hexagram-no">第 {{ record.hexagram_number }} 卦</span>
            <h2>{{ record.hexagram_name }}</h2>
            <span class="summary-label">卦象说明</span>
            <p>{{ record.summary }}</p>
          </div>
        </div>

        <div class="reading-grid">
          <section class="reading-card">
            <h3>对照当前问题</h3>
            <p>{{ record.human_way }}</p>
          </section>
          <section class="reading-card">
            <h3>观察重点</h3>
            <p>{{ record.image_reading }}</p>
          </section>
          <section class="reading-card reading-wide">
            <h3>可以怎么做</h3>
            <p>{{ record.advice }}</p>
          </section>
        </div>

        <p class="result-boundary">
          本次结果来自随机抽取，只提供传统卦象的反思角度，不代表吉凶、概率或现实结果。涉及健康、法律、财务或人身安全时，请依据专业意见。
        </p>

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
  background: linear-gradient(180deg, transparent, var(--surface-1)), var(--bg);
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
  background: linear-gradient(
    90deg,
    var(--accent) 0 38%,
    transparent 38% 62%,
    var(--accent) 62% 100%
  );
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

.draw-boundary {
  max-width: 38rem;
  margin: 0;
  color: var(--text-muted);
  font-size: var(--fs-sm);
  line-height: 1.65;
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
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease;
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

.summary-label {
  display: block;
  margin-top: 24px;
  color: var(--accent);
  font-size: var(--fs-xs);
  font-weight: 700;
}

.hexagram-meta p {
  max-width: 680px;
  margin: 8px 0 0;
  color: var(--text-muted);
  font-size: var(--fs-md);
  line-height: 1.75;
  word-break: break-word;
}

.reading-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 28px;
  border-top: 1px solid var(--line-subtle);
  border-bottom: 1px solid var(--line-subtle);
}

.reading-card {
  min-height: 132px;
  padding: 20px 22px 20px 0;
}

.reading-card:nth-child(2) {
  padding-right: 0;
  padding-left: 22px;
  border-left: 1px solid var(--line-subtle);
}

.reading-wide {
  grid-column: 1 / -1;
  min-height: 104px;
  padding-right: 0;
  border-top: 1px solid var(--line-subtle);
}

.reading-card h3 {
  margin: 0 0 10px;
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

.result-boundary {
  max-width: 72ch;
  margin: 18px 0 0;
  color: var(--text-soft);
  font-size: var(--fs-xs);
  line-height: 1.7;
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

  .reading-grid {
    grid-template-columns: 1fr;
  }

  .reading-card,
  .reading-card:nth-child(2) {
    min-height: auto;
    padding: 18px 0;
    border-left: 0;
  }

  .reading-card + .reading-card {
    border-top: 1px solid var(--line-subtle);
  }
}
</style>
