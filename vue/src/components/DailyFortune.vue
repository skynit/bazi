<script setup lang="ts">
import { ref, computed } from 'vue'

interface ElementImage {
  element: string
  image_url: string
  description: string
}

interface Props {
  solarDate: string
  dayGanZhi: string
  weekDay?: string
  lunarDate?: string
  shengXiao?: string
  yiJi?: string
  chongSha?: string
  elementImages?: ElementImage[]
  luckyColor?: string
  luckyNumber?: number
  wealthDir?: string
  auspiciousHours?: string[]
}

const props = withDefaults(defineProps<Props>(), {
  weekDay: '',
  lunarDate: '',
  shengXiao: '',
  yiJi: '',
  chongSha: '',
  elementImages: () => [],
  luckyColor: '',
  luckyNumber: 0,
  wealthDir: '',
  auspiciousHours: () => [],
})

const showAiModal = ref(false)

const elementNameMap: Record<string, string> = {
  metal: '金',
  wood: '木',
  water: '水',
  fire: '火',
  earth: '土',
}

const elementColorMap: Record<string, string> = {
  metal: '#FFD700',
  wood: '#228B22',
  water: '#4169E1',
  fire: '#DC143C',
  earth: '#DAA520',
}

const allElements = ['metal', 'wood', 'water', 'fire', 'earth'] as const

const elementImagesMap = computed(() => {
  const map: Record<string, ElementImage | undefined> = {}
  for (const img of props.elementImages) {
    map[img.element] = img
    // Also map English names to Chinese for backward compat
    const enKey = Object.entries(elementNameMap).find(([,v]) => v === img.element)?.[0]
    if (enKey) map[enKey] = img
  }
  return map
})

const yiItems = computed(() => {
  if (!props.yiJi) return []
  const match = props.yiJi.match(/宜[:：]?\s*(.+?)(?:忌|$)/)
  if (!match) return []
  return match[1].split(/[、，,]/).filter(Boolean).map(s => s.trim())
})

const jiItems = computed(() => {
  if (!props.yiJi) return []
  const match = props.yiJi.match(/忌[:：]?\s*(.+)/)
  if (!match) return []
  return match[1].split(/[、，,]/).filter(Boolean).map(s => s.trim())
})

const aiButtonLabel = 'AI 深度解析'

function openAiModal() {
  showAiModal.value = true
}
</script>

<template>
  <div class="daily-fortune">
    <!-- Header: Date + Day Pillar -->
    <div class="fortune-header">
      <div class="date-block">
        <p class="solar-date">{{ solarDate }}</p>
        <p v-if="lunarDate" class="lunar-date">{{ lunarDate }}</p>
        <p v-if="weekDay" class="week-day">{{ weekDay }}</p>
      </div>
      <div class="pillar-block">
        <span class="day-pillar">{{ dayGanZhi }}</span>
        <p v-if="shengXiao" class="sheng-xiao">属{{ shengXiao }}</p>
      </div>
    </div>

    <!-- Lucky Info Cards -->
    <div class="lucky-grid">
      <div class="lucky-card color-swatch">
        <span class="lucky-label">幸运色</span>
        <div class="color-dot" :style="{ background: luckyColor || '#C41E3A' }"></div>
        <span class="lucky-value">{{ luckyColor || '—' }}</span>
      </div>
      <div class="lucky-card lucky-number">
        <span class="lucky-label">幸运数字</span>
        <span class="big-number">{{ luckyNumber || '—' }}</span>
      </div>
      <div class="lucky-card wealth-dir">
        <span class="lucky-label">财神方位</span>
        <span class="direction-icon">🧭</span>
        <span class="lucky-value">{{ wealthDir || '—' }}</span>
      </div>
      <div class="lucky-card clash-warn">
        <span class="lucky-label">冲煞</span>
        <span class="clash-icon">⚠️</span>
        <span class="lucky-value clash">{{ chongSha || '—' }}</span>
      </div>
    </div>

    <!-- Yi Ji Columns -->
    <div class="yiji-section">
      <div class="yiji-col yi-col">
        <h3 class="yiji-title yi-title">宜</h3>
        <ul v-if="yiItems.length" class="yiji-list">
          <li v-for="item in yiItems" :key="item" class="yiji-item">{{ item }}</li>
        </ul>
        <p v-else class="yiji-empty">—</p>
      </div>
      <div class="yiji-divider"></div>
      <div class="yiji-col ji-col">
        <h3 class="yiji-title ji-title">忌</h3>
        <ul v-if="jiItems.length" class="yiji-list">
          <li v-for="item in jiItems" :key="item" class="yiji-item">{{ item }}</li>
        </ul>
        <p v-else class="yiji-empty">—</p>
      </div>
    </div>

    <!-- Auspicious Hours -->
    <div v-if="auspiciousHours.length" class="auspicious-section">
      <h3 class="section-title">吉时</h3>
      <div class="hours-list">
        <span v-for="hour in auspiciousHours" :key="hour" class="hour-tag">{{ hour }}</span>
      </div>
    </div>

    <!-- Element Images -->
    <div class="elements-section">
      <h3 class="section-title">五行元素</h3>
      <div class="elements-grid">
        <div
          v-for="el in allElements"
          :key="el"
          class="element-card"
          :style="{ borderColor: elementColorMap[el] }"
        >
          <div class="element-icon" :style="{ background: elementColorMap[el] }">
            {{ elementNameMap[el] }}
          </div>
          <span class="element-name">{{ elementNameMap[el] }}</span>
          <p v-if="elementImagesMap[el]" class="element-desc">
            {{ elementImagesMap[el]!.description }}
          </p>
          <p v-else class="element-desc placeholder">—</p>
        </div>
      </div>
    </div>

    <!-- AI Analysis Button -->
    <div class="ai-section">
      <button class="ai-button" @click="openAiModal">
        <span class="ai-icon">🤖</span>
        {{ aiButtonLabel }}
      </button>
    </div>

    <!-- AI Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showAiModal" class="modal-overlay" @click.self="showAiModal = false">
          <div class="modal-content">
            <div class="modal-header">
              <h2>AI 深度解析</h2>
              <button class="modal-close" @click="showAiModal = false">✕</button>
            </div>
            <div class="modal-body">
              <div class="coming-soon">
                <span class="coming-soon-icon">🚀</span>
                <p>AI分析功能即将上线</p>
                <p class="coming-soon-sub">敬请期待智能运势解读</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.daily-fortune {
  max-width: 480px;
  margin: 0 auto;
  padding: 1rem;
  font-family: 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

/* Header */
.fortune-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.25rem;
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, var(--color-bazi-paper) 0%, #EDE8D8 100%);
  border-radius: 0.75rem;
  border: 1px solid rgba(196, 30, 58, 0.12);
}

.date-block {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.solar-date {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-bazi-ink);
  margin: 0;
}

.lunar-date {
  font-size: 0.8rem;
  color: #888;
  margin: 0;
}

.week-day {
  font-size: 0.8rem;
  color: #666;
  margin: 0;
}

.pillar-block {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.day-pillar {
  font-size: 2rem;
  font-weight: 800;
  color: var(--color-bazi-red);
  letter-spacing: 0.1em;
  line-height: 1.1;
}

.sheng-xiao {
  font-size: 0.75rem;
  color: #999;
  margin: 0.25rem 0 0 0;
}

/* Lucky Grid */
.lucky-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.625rem;
  margin-bottom: 1rem;
}

.lucky-card {
  background: white;
  border: 1px solid #E8E3D5;
  border-radius: 0.625rem;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
}

.lucky-label {
  font-size: 0.7rem;
  color: #999;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.color-dot {
  width: 2rem;
  height: 2rem;
  border-radius: 50%;
  border: 2px solid white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
}

.lucky-value {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-bazi-ink);
}

.lucky-value.clash {
  color: var(--color-bazi-red);
  font-size: 0.78rem;
}

.big-number {
  font-size: 2rem;
  font-weight: 800;
  color: var(--color-bazi-red);
  line-height: 1;
}

.direction-icon {
  font-size: 1.5rem;
}

.clash-icon {
  font-size: 1.3rem;
}

/* Yi Ji Section */
.yiji-section {
  display: flex;
  background: white;
  border: 1px solid #E8E3D5;
  border-radius: 0.75rem;
  overflow: hidden;
  margin-bottom: 1rem;
}

.yiji-col {
  flex: 1;
  padding: 0.875rem;
}

.yi-title {
  color: #228B22;
}

.ji-title {
  color: var(--color-bazi-red);
}

.yiji-title {
  font-size: 1.1rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
}

.yiji-divider {
  width: 1px;
  background: #E8E3D5;
  margin: 0.75rem 0;
}

.yiji-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.yiji-item {
  font-size: 0.8rem;
  color: var(--color-bazi-ink);
  padding: 0.15rem 0;
  border-bottom: 1px dashed #F0ECE0;
  line-height: 1.35;
}

.yiji-item:last-child {
  border-bottom: none;
}

.yiji-empty {
  font-size: 0.85rem;
  color: #bbb;
  margin: 0.5rem 0;
}

/* Auspicious Hours */
.auspicious-section {
  margin-bottom: 1rem;
  padding: 0.875rem;
  background: white;
  border: 1px solid #E8E3D5;
  border-radius: 0.75rem;
}

.section-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--color-bazi-ink);
  margin: 0 0 0.625rem 0;
}

.hours-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.hour-tag {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  background: #F5F0E8;
  color: var(--color-bazi-ink);
  border-radius: 0.375rem;
  font-size: 0.78rem;
  border: 1px solid #E8E3D5;
}

/* Elements Section */
.elements-section {
  margin-bottom: 1rem;
}

.elements-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 0.4rem;
}

.element-card {
  background: white;
  border: 1.5px solid;
  border-radius: 0.625rem;
  padding: 0.6rem 0.4rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
}

.element-icon {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.85rem;
  font-weight: 700;
  color: white;
}

.element-name {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-bazi-ink);
}

.element-desc {
  font-size: 0.6rem;
  color: #888;
  text-align: center;
  margin: 0;
  line-height: 1.2;
}

.element-desc.placeholder {
  color: #ccc;
}

/* AI Section */
.ai-section {
  text-align: center;
  margin-top: 0.5rem;
}

.ai-button {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.625rem 1.75rem;
  background: linear-gradient(135deg, var(--color-bazi-red), #A01830);
  color: white;
  border: none;
  border-radius: 2rem;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
  box-shadow: 0 4px 14px rgba(196, 30, 58, 0.3);
}

.ai-button:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(196, 30, 58, 0.4);
}

.ai-icon {
  font-size: 1.1rem;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-content {
  background: white;
  border-radius: 1rem;
  width: 100%;
  max-width: 380px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid #E8E3D5;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--color-bazi-ink);
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.2rem;
  color: #999;
  cursor: pointer;
  padding: 0.25rem;
  line-height: 1;
}

.modal-body {
  padding: 2rem 1.25rem;
}

.coming-soon {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.coming-soon-icon {
  font-size: 3rem;
}

.coming-soon p {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-bazi-ink);
}

.coming-soon-sub {
  font-size: 0.8rem !important;
  color: #999 !important;
  font-weight: 400 !important;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .modal-content,
.modal-leave-active .modal-content {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content {
  transform: scale(0.95);
}

.modal-leave-to .modal-content {
  transform: scale(0.95);
}
</style>
