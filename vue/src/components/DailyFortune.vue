<script setup lang="ts">
import { ref, computed } from 'vue'

interface ElementImage { element: string; image_url: string; description: string }
interface Props {
  solarDate: string; dayGanZhi: string; weekDay?: string; lunarDate?: string
  shengXiao?: string; yiJi?: string; chongSha?: string; elementImages?: ElementImage[]
  luckyColor?: string; luckyNumber?: number; wealthDir?: string
  auspiciousHours?: string[]; todayElements?: Record<string, number>
  tiaoHou?: string
}
const props = withDefaults(defineProps<Props>(), {
  weekDay: '', lunarDate: '', shengXiao: '', yiJi: '',
  chongSha: '', elementImages: () => [],
  luckyColor: '', luckyNumber: 0, wealthDir: '', auspiciousHours: () => [], todayElements: () => ({}),
  tiaoHou: '',
})
const showAiModal = ref(false)
const elementEntries = [['金','#FFD700'],['木','#3CB371'],['水','#4169E1'],['火','#DC143C'],['土','#DAA520']] as [string,string][]
function elPct(el: string) {
  const n = props.todayElements || {}, t = Object.values(n).reduce((s,v) => s + v, 0)
  return t ? Math.round(((n[el]||0)/t)*100) : 0
}
const yiItems = computed(() => {
  if (!props.yiJi) return []
  const m = props.yiJi.match(/宜[:：]?\s*(.+?)(?:忌|$)/)
  return m ? m[1].split(/[、，,]/).filter(Boolean).map(s => s.trim()) : []
})
const jiItems = computed(() => {
  if (!props.yiJi) return []
  const m = props.yiJi.match(/忌[:：]?\s*(.+)/)
  return m ? m[1].split(/[、，,]/).filter(Boolean).map(s => s.trim()) : []
})
</script>

<template>
  <div class="daily-fortune">

    <!-- Date + Pillar -->
    <div class="df-header glass-card">
      <div class="df-date-col">
        <p class="df-solar">{{ solarDate }}</p>
        <p v-if="lunarDate" class="df-lunar">{{ lunarDate }}</p>
      </div>
      <div class="df-pillar-col">
        <div class="df-pillar-glow"></div>
        <span class="df-pillar-val">{{ dayGanZhi }}</span>
        <span v-if="shengXiao" class="df-sx">属{{ shengXiao }}</span>
      </div>
    </div>

    <!-- Lucky 4-grid -->
    <div class="df-lucky-grid">
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <div class="lc-color-dot" :style="{ background: luckyColor||'#C41E3A', boxShadow: `0 0 18px ${luckyColor||'#C41E3A'}88` }"></div>
        </div>
        <span class="lc-lbl">幸运色</span>
        <span class="lc-val">{{ luckyColor || '—' }}</span>
      </div>
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <rect x="3" y="3" width="26" height="26" rx="5" stroke="#D4A84B" stroke-width="1.5" opacity="0.35" />
            <text x="16" y="22" text-anchor="middle" font-size="13" font-weight="900" fill="#D4A84B" opacity="0.7">{{ luckyNumber || '?' }}</text>
          </svg>
        </div>
        <span class="lc-lbl">幸运数字</span>
        <span class="lc-val lc-val-gold">{{ luckyNumber || '—' }}</span>
      </div>
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <circle cx="16" cy="16" r="13" stroke="#D4A84B" stroke-width="1.2" opacity="0.3" />
            <line x1="16" y1="3" x2="16" y2="16" stroke="#D4A84B" stroke-width="2" stroke-linecap="round" />
            <line x1="16" y1="16" x2="24" y2="22" stroke="#D4A84B" stroke-width="2" stroke-linecap="round" />
          </svg>
        </div>
        <span class="lc-lbl">财神方位</span>
        <span class="lc-val">{{ wealthDir || '—' }}</span>
      </div>
      <div class="df-lucky-cell glass-card">
        <div class="lc-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <path d="M16 2L21 12H30L23 18L26 28L16 22L6 28L9 18L2 12H11L16 2Z" stroke="#C41E3A" stroke-width="1.3" opacity="0.45" />
          </svg>
        </div>
        <span class="lc-lbl">冲煞</span>
        <span class="lc-val lc-val-red">{{ chongSha || '—' }}</span>
      </div>
    </div>

    <!-- Yi Ji -->
    <div class="df-yiji glass-card">
      <div class="yj-col yj-yi">
        <div class="yj-header">
          <span class="yj-tag yj-tag-yi">宜</span>
        </div>
        <div v-if="yiItems.length" class="yj-tags">
          <span v-for="it in yiItems" :key="it" class="yj-tag-item yj-tag-yi-item">{{ it }}</span>
        </div>
        <p v-else class="yj-empty">—</p>
      </div>
      <div class="yj-divider"></div>
      <div class="yj-col yj-ji">
        <div class="yj-header">
          <span class="yj-tag yj-tag-ji">忌</span>
        </div>
        <div v-if="jiItems.length" class="yj-tags">
          <span v-for="it in jiItems" :key="it" class="yj-tag-item yj-tag-ji-item">{{ it }}</span>
        </div>
        <p v-else class="yj-empty">—</p>
      </div>
    </div>

    <!-- TiaoHou -->
    <div v-if="tiaoHou" class="df-tiaohou glass-card">
      <div class="df-sec-header">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M7 1L9 5H13L9.5 7.5L11 12L7 9.5L3 12L4.5 7.5L1 5H5L7 1Z" stroke="#D4A84B" stroke-width="1" opacity="0.5"/>
        </svg>
        <span class="df-sec-title">调候吉言</span>
      </div>
      <p class="tiaohou-text">{{ tiaoHou }}</p>
    </div>

    <!-- Hours + Elements -->
    <div class="df-bottom-row">
      <div v-if="auspiciousHours.length" class="df-hours glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="6" stroke="#D4A84B" stroke-width="1" opacity="0.4"/>
            <line x1="7" y1="3" x2="7" y2="7" stroke="#D4A84B" stroke-width="1.5" stroke-linecap="round"/>
            <line x1="7" y1="7" x2="10" y2="9" stroke="#D4A84B" stroke-width="1.5" stroke-linecap="round"/>
          </svg>
          <span class="df-sec-title">吉时</span>
        </div>
        <div class="df-hours-list">
          <span v-for="h in auspiciousHours" :key="h" class="df-hour-chip">
            <span class="df-hour-dot"></span>{{ h }}
          </span>
        </div>
      </div>
      <div class="df-elems glass-card">
        <div class="df-sec-header">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <circle cx="7" cy="7" r="6" stroke="#D4A84B" stroke-width="1" opacity="0.25"/>
            <circle cx="7" cy="7" r="2.5" fill="#D4A84B" opacity="0.35"/>
          </svg>
          <span class="df-sec-title">今日五行</span>
        </div>
        <div class="df-el-bars">
          <div v-for="[el, clr] in elementEntries" :key="el" class="df-el-row">
            <span class="df-el-name">{{ el }}</span>
            <div class="df-el-track">
              <div class="df-el-fill" :style="{ width: elPct(el)+'%', background: clr, boxShadow: `0 0 8px ${clr}66` }"></div>
            </div>
            <span class="df-el-num">{{ todayElements[el] ?? 0 }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- AI button -->
    <button class="df-ai-btn" @click="showAiModal = true">
      <span class="df-ai-btn-icon">◈</span>
      AI 深度解析
    </button>

    <!-- Modal -->
    <Teleport to="body">
      <Transition name="df-modal">
        <div v-if="showAiModal" class="df-modal-overlay" @click.self="showAiModal=false">
          <div class="df-modal-box glass-panel">
            <div class="df-modal-hdr">
              <div class="df-modal-title-group">
                <span class="df-modal-orb">☯</span>
                <h2>AI 深度解析</h2>
              </div>
              <button class="df-modal-close" @click="showAiModal=false">✕</button>
            </div>
            <div class="df-modal-body">
              <div class="df-ai-coming">
                <svg width="90" height="90" viewBox="0 0 90 90" fill="none" class="df-ai-svg">
                  <circle cx="45" cy="45" r="42" stroke="#D4A84B" stroke-width="0.6" stroke-dasharray="2 4" opacity="0.2" />
                  <circle cx="45" cy="45" r="28" stroke="#D4A84B" stroke-width="0.6" stroke-dasharray="1 5" opacity="0.15" />
                  <circle cx="45" cy="45" r="8" fill="#D4A84B" opacity="0.2" />
                  <circle cx="45" cy="45" r="13" fill="none" stroke="#C41E3A" stroke-width="0.5" opacity="0.3" />
                  <circle cx="22" cy="24" r="2.5" fill="#D4A84B" opacity="0.45" class="df-star-pulse" style="animation-delay:0s" />
                  <circle cx="68" cy="22" r="2" fill="#D4A84B" opacity="0.35" class="df-star-pulse" style="animation-delay:0.6s" />
                  <circle cx="70" cy="66" r="2.5" fill="#D4A84B" opacity="0.4" class="df-star-pulse" style="animation-delay:1.2s" />
                </svg>
                <p class="df-ai-title">AI分析功能即将上线</p>
                <p class="df-ai-sub">智能运势深度解读</p>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.daily-fortune { display: flex; flex-direction: column; gap: 0.75rem; }

/* Header */
.df-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem;
  position: relative; overflow: hidden;
}
.df-header::after {
  content: ''; position: absolute; top: -20px; right: -20px;
  width: 100px; height: 100px;
  background: radial-gradient(circle, rgba(196,30,58,0.07), transparent 70%);
  pointer-events: none;
}
.df-date-col { display: flex; flex-direction: column; gap: 0.15rem; }
.df-solar { font-size: 1.05rem; font-weight: 700; color: rgba(255,255,255,0.9); margin: 0; letter-spacing: 1px; }
.df-lunar { font-size: 0.72rem; color: rgba(255,255,255,0.28); margin: 0; }
.df-pillar-col { display: flex; flex-direction: column; align-items: flex-end; position: relative; gap: 0.2rem; }
.df-pillar-glow { position: absolute; top: -15px; right: -15px; width: 80px; height: 80px; background: radial-gradient(circle, rgba(196,30,58,0.08), transparent 70%); pointer-events: none; }
.df-pillar-val {
  font-size: 2.75rem; font-weight: 950;
  color: #C41E3A; letter-spacing: 0.05em; line-height: 1;
  text-shadow: 0 0 30px rgba(196,30,58,0.4);
}
.df-sx { font-size: 0.7rem; color: rgba(255,255,255,0.25); margin: 0; text-align: right; }

/* Lucky grid */
.df-lucky-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.df-lucky-cell {
  display: flex; flex-direction: column; align-items: center; gap: 0.35rem;
  padding: 1rem 0.5rem;
  transition: border-color 0.3s, box-shadow 0.3s;
}
.df-lucky-cell:hover { border-color: rgba(212,168,75,0.2); box-shadow: 0 4px 20px rgba(0,0,0,0.2); }
.lc-icon { display: flex; align-items: center; justify-content: center; height: 40px; }
.lc-color-dot { width: 30px; height: 30px; border-radius: 50%; border: 1.5px solid rgba(255,255,255,0.12); }
.lc-lbl { font-size: 0.58rem; color: rgba(255,255,255,0.22); text-transform: uppercase; letter-spacing: 0.1em; }
.lc-val { font-size: 0.85rem; font-weight: 700; color: rgba(255,255,255,0.75); }
.lc-val-gold { color: #D4A84B; font-size: 1.4rem; font-weight: 900; letter-spacing: 1px; text-shadow: 0 0 20px rgba(212,168,75,0.3); }
.lc-val-red { color: #C41E3A; font-size: 0.78rem; }

/* Yi Ji */
.df-yiji { display: flex; overflow: hidden; }
.yj-col { flex: 1; padding: 0.9rem; }
.yj-header { margin-bottom: 0.5rem; }
.yj-tag {
  display: inline-block; font-size: 0.72rem; font-weight: 800;
  padding: 0.15rem 0.7rem; border-radius: 4px; letter-spacing: 1px;
}
.yj-tag-yi { background: rgba(74,222,128,0.1); color: #4ade80; border: 1px solid rgba(74,222,128,0.2); }
.yj-tag-ji { background: rgba(196,30,58,0.1); color: #C41E3A; border: 1px solid rgba(196,30,58,0.2); }
.yj-tags { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.yj-tag-item {
  display: inline-block; font-size: 0.72rem; padding: 0.2rem 0.5rem;
  border-radius: 4px; transition: all 0.2s; cursor: default;
}
.yj-tag-yi-item { background: rgba(74,222,128,0.04); border: 1px solid rgba(74,222,128,0.1); color: rgba(74,222,128,0.65); }
.yj-tag-yi-item:hover { background: rgba(74,222,128,0.1); color: #4ade80; }
.yj-tag-ji-item { background: rgba(196,30,58,0.04); border: 1px solid rgba(196,30,58,0.1); color: rgba(196,30,58,0.6); }
.yj-tag-ji-item:hover { background: rgba(196,30,58,0.1); color: #C41E3A; }
.yj-divider { width: 1px; background: rgba(212,168,75,0.06); margin: 0.75rem 0; }
.yj-empty { font-size: 0.82rem; color: rgba(139,131,120,0.2); margin: 0.5rem 0; }

/* Hours + Elements */
.df-bottom-row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.5rem; }
.df-hours, .df-elems { padding: 0.9rem; }
.df-sec-header { display: flex; align-items: center; gap: 0.4rem; margin-bottom: 0.6rem; }
.df-sec-title { font-size: 0.72rem; font-weight: 800; color: #D4A84B; margin: 0; letter-spacing: 2px; }
.df-hours-list { display: flex; flex-wrap: wrap; gap: 0.35rem; }
.df-hour-chip {
  display: inline-flex; align-items: center; gap: 5px;
  padding: 0.25rem 0.65rem;
  background: rgba(212,168,75,0.04);
  border: 1px solid rgba(212,168,75,0.1);
  border-radius: 20px; font-size: 0.7rem;
  color: rgba(255,255,255,0.6);
  transition: all 0.25s;
}
.df-hour-chip:hover { background: rgba(212,168,75,0.1); border-color: rgba(212,168,75,0.25); color: #D4A84B; }
.df-hour-dot { width: 4px; height: 4px; border-radius: 50%; background: #D4A84B; box-shadow: 0 0 6px rgba(212,168,75,0.7); flex-shrink: 0; }
.df-el-bars { display: flex; flex-direction: column; gap: 0.35rem; }
.df-el-row { display: flex; align-items: center; gap: 0.4rem; }
.df-el-name { width: 14px; font-size: 0.65rem; font-weight: 800; color: rgba(255,255,255,0.3); flex-shrink: 0; }
.df-el-track { flex: 1; height: 5px; background: rgba(255,255,255,0.04); border-radius: 3px; overflow: hidden; }
.df-el-fill { height: 100%; border-radius: 3px; transition: width 0.8s ease; }
.df-el-num { width: 18px; font-size: 0.6rem; color: rgba(255,255,255,0.2); text-align: right; flex-shrink: 0; }

/* AI btn */
.df-ai-btn {
  display: flex; align-items: center; justify-content: center; gap: 0.5rem;
  width: 100%; padding: 0.7rem 1rem;
  background: rgba(255,255,255,0.02);
  color: rgba(255,255,255,0.28);
  border: 1px solid rgba(212,168,75,0.1);
  border-radius: 10px; font-size: 0.8rem; font-weight: 600;
  cursor: pointer; transition: all 0.3s; letter-spacing: 1.5px;
}
.df-ai-btn:hover { border-color: rgba(212,168,75,0.3); color: #D4A84B; background: rgba(212,168,75,0.05); }
.df-ai-btn-icon { font-size: 1rem; }

/* TiaoHou */
.df-tiaohou { padding: 0.9rem; }
.tiaohou-text {
  font-size: 0.82rem;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.7;
  white-space: pre-wrap;
  margin: 0.4rem 0 0;
  font-style: italic;
}

/* Modal */
.df-modal-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.7);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 1rem; backdrop-filter: blur(8px);
}
.df-modal-box { width: 100%; max-width: 380px; overflow: hidden; }
.df-modal-hdr {
  display: flex; justify-content: space-between; align-items: center;
  padding: 1.25rem 1.5rem; border-bottom: 1px solid rgba(212,168,75,0.07);
}
.df-modal-title-group { display: flex; align-items: center; gap: 0.75rem; }
.df-modal-orb {
  font-size: 1.6rem; color: #D4A84B;
  text-shadow: 0 0 25px rgba(212,168,75,0.5);
  animation: orb-glow 3s ease-in-out infinite;
}
@keyframes orb-glow { 0%,100%{text-shadow:0 0 20px rgba(212,168,75,0.3)} 50%{text-shadow:0 0 40px rgba(212,168,75,0.7)} }
.df-modal-hdr h2 { margin: 0; font-family: var(--font-serif), serif; font-size: 1.05rem; font-weight: 700; color: rgba(255,255,255,0.85); letter-spacing: 2px; }
.df-modal-close { background: none; border: none; font-size: 1.2rem; color: rgba(255,255,255,0.2); cursor: pointer; padding: 0.25rem; transition: color 0.2s; }
.df-modal-close:hover { color: #D4A84B; }
.df-modal-body { padding: 2.5rem 1.5rem; }
.df-ai-coming { display: flex; flex-direction: column; align-items: center; gap: 1rem; text-align: center; }
.df-ai-svg { opacity: 0.75; animation: svg-rot 20s linear infinite; }
@keyframes svg-rot { from{transform:rotate(0deg)} to{transform:rotate(360deg)} }
.df-star-pulse { animation: st-pulse 2.5s ease-in-out infinite; }
@keyframes st-pulse { 0%,100%{opacity:0.2} 50%{opacity:0.7} }
.df-ai-title { font-size: 1rem; font-weight: 700; color: rgba(255,255,255,0.75); margin: 0; }
.df-ai-sub { font-size: 0.78rem; color: rgba(255,255,255,0.25); margin: 0; }

/* Modal transition */
.df-modal-enter-active, .df-modal-leave-active { transition: opacity 0.25s ease; }
.df-modal-enter-from, .df-modal-leave-to { opacity: 0; }
.df-modal-enter-active .df-modal-box, .df-modal-leave-active .df-modal-box { transition: transform 0.25s ease; }
.df-modal-enter-from .df-modal-box { transform: scale(0.9) translateY(12px); }
.df-modal-leave-to .df-modal-box { transform: scale(0.9) translateY(12px); }
</style>