<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '../api/client'
import DailyFortune from '../components/DailyFortune.vue'
import type { FortuneGuide } from '../api/fortune'

interface HiddenStemGod {
  stem: string
  type: string
  element: string
  ten_god: string
  favorable: boolean
}
interface StemRelation {
  type: string
  target: string
  detail: string
  is_favorable: boolean
  note?: string
}
interface BranchRelation {
  type: string
  target: string
  detail: string
  is_favorable: boolean
}
interface ShenShaActivation {
  name: string
  type: string
  description: string
  activation: string
}
interface DaYunInfluence {
  current_pillar: string
  start_age: number
  end_age: number
  ten_god: string
  favorable: boolean
  relation: string
  score: number
  description: string
}
interface LiuNianInfluence {
  year_pillar: string
  ten_god: string
  favorable: boolean
  relation: string
  tai_sui_relation: string
  score: number
  description: string
}
interface AdvanceRetreat {
  phase: string
  phase_desc: string
  element: string
  score: number
  description: string
}
interface YongShenImpact {
  tiao_hou_element: string
  tiao_hou_hit: boolean
  tong_guan_element: string
  tong_guan_hit: boolean
  fu_yi_elements: string[]
  fu_yi_hit: boolean
  score: number
  description: string
}

interface FortuneOverall {
  score?: number
  base_score?: number
  detail_score?: number
  stars?: string
  level?: string
  summary?: string
  key_tip?: string
}

interface FortuneCategory {
  name: string
  score?: number
  weight?: number
  stars?: string
  level?: string
  trend?: string
  keywords?: string[]
  analysis?: string
  advice?: string
}

interface FortuneData {
  solar_date: string
  day_gan_zhi: string
  score?: number
  analysis?: {
    overall?: FortuneOverall
    categories?: FortuneCategory[]
    lucky_guide?: { colors?: string; numbers?: string; actions?: string; outfit?: string; favorable_elems?: string[]; unfavorable_elems?: string[] }
  }
  lucky_color?: string
  lucky_number?: number
  wealth_direction?: string
  guide?: FortuneGuide
  clash_zodiac?: string
  auspicious_hours?: string[]
  yi?: string[]
  ji?: string[]
  element_images?: { element: string; image_url: string; description: string }[]
  today_elements?: Record<string, number>
  tiao_hou?: string
  week_day?: string
  lunar_date?: string
  sheng_xiao?: string
  ji_shen?: string
  xiong_shen?: string
  tai_shen?: string
  wu_xing?: string
  peng_zu?: string
  gua?: string
  jie_qi?: string
  sheng_ke_analysis?: { day_stem_relation?: string; day_branch_relation?: string; summary?: string }
  flow_impact?: string
  season_element_advice?: string
  // 日课推算
  today_ten_god?: string
  ten_god_favorable?: boolean
  ten_god_desc?: string
  twelve_stage?: string
  stage_favorable?: boolean
  stage_desc?: string
  stage_flexible?: string
  hidden_stems?: HiddenStemGod[]
  stem_relations?: StemRelation[]
  branch_relations?: BranchRelation[]
  activated_shen_sha?: ShenShaActivation[]
  dayun_influence?: DaYunInfluence
  liunian_influence?: LiuNianInfluence
  advance_retreat?: AdvanceRetreat
  yongshen_impact?: YongShenImpact
  overall_verdict?: string
  favor_score?: number
  pattern_name?: string
  pattern_type?: string
  pattern_favorable?: string[]
  pattern_unfavorable?: string[]
}

const route = useRoute()
const fortune = ref<FortuneData | null>(null)
const loading = ref(true)
const error = ref('')
const mounted = ref(false)
const chartId = ref<string | number>('')

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function fetchFortune() {
  let cid: string | number | null = route.query.chart_id as string | null
  if (!cid) {
    try { const s = localStorage.getItem('bazi_last_birth'); if (s) cid = JSON.parse(s).chartId } catch {}
    if (!cid) { error.value = '请先创建命盘'; loading.value = false; return }
  }
  chartId.value = cid
  try {
    const { data } = await client.post('/fortune', { chart_id: Number(chartId.value), query_date: todayStr() })
    fortune.value = data
  } catch (e: any) { error.value = e.response?.data?.error || '加载运势失败' }
  finally { loading.value = false }
}

onMounted(() => {
  fetchFortune(); setTimeout(() => mounted.value = true, 100)
})

function scoreColor(s: number) {
  const t = Math.max(0, Math.min(1, s / 100))
  return `color-mix(in oklab, var(--jade-accent) ${Math.round(54 + t * 46)}%, var(--text) ${Math.round((1 - t) * 18)}%)`
}

function scoreGlow(s: number) {
  const t = Math.max(0, Math.min(1, s / 100))
  return `rgba(var(--jade-accent-rgb), ${(0.14 + t * 0.22).toFixed(3)})`
}

function scoreWord(s: number) {
  if (s >= 85) return '大吉'
  if (s >= 70) return '良好'
  if (s >= 55) return '平稳'
  if (s >= 40) return '欠佳'
  return '低迷'
}

function starCount(stars: string) { return (stars.match(/★/g) || []).length }

function clampScore(score?: number) {
  return Math.max(20, Math.min(100, score ?? 0))
}

function categoryScore(category: FortuneCategory) {
  const fallback = starCount(category.stars || '') * 20 || 60
  return clampScore(category.score ?? fallback)
}

function trendLabel(trend?: string) {
  if (trend === 'up') return '走高'
  if (trend === 'down') return '承压'
  return '平稳'
}

function categoryBg(score?: number) {
  const s = clampScore(score)
  return `linear-gradient(135deg, color-mix(in oklab, var(--surface-1) 88%, transparent), color-mix(in oklab, var(--accent) ${Math.max(8, Math.min(32, s / 3))}%, transparent))`
}

function legacyGuide(fortuneData: FortuneData): FortuneGuide | undefined {
  const old = fortuneData.analysis?.lucky_guide
  if (!old) return undefined
  return {
    precision_level: 'legacy',
    confidence: 50,
    primary_element: old.favorable_elems?.[0] ?? '',
    secondary_element: old.favorable_elems?.[1] ?? '',
    avoid_element: old.unfavorable_elems?.[0] ?? '',
    lucky_colors: old.colors ? [{ label: '幸运色', value: old.colors, reason: '按旧版喜用五行规则生成。' }] : [],
    lucky_numbers: old.numbers ? [{ label: '幸运数', value: old.numbers, reason: '按旧版喜用五行规则生成。' }] : [],
    face_direction: { label: '朝向', value: '', reason: '' },
    wealth_direction: { label: '财位', value: fortuneData.wealth_direction ?? '', reason: '按流日天干财位生成。' },
    avoid_direction: { label: '避开', value: '', reason: '' },
    recommended_actions: old.actions ? [{ label: '动作', value: old.actions, reason: '按旧版开运动作生成。' }] : [],
    cautions: [],
    best_hours: (fortuneData.auspicious_hours ?? []).slice(0, 3).map(h => ({ label: '吉时', value: h, reason: '按流日地支取吉时。' })),
    analysis: '',
    strategy: old.outfit ? `穿搭建议：${old.outfit}` : '',
  }
}

function activeGuide(fortuneData: FortuneData): FortuneGuide | undefined {
  return fortuneData.guide ?? legacyGuide(fortuneData)
}

function guideColorValue(guide: FortuneGuide | undefined): string {
  return guide?.lucky_colors?.[0]?.value ?? ''
}

function guideNumberValue(guide: FortuneGuide | undefined): string {
  return guide?.lucky_numbers?.[0]?.value ?? ''
}

function guideConfidenceLabel(guide: FortuneGuide | undefined): string {
  if (!guide?.confidence) return ''
  if (guide.confidence >= 80) return '高置信'
  if (guide.confidence >= 65) return '中高置信'
  return '参考'
}
</script>

<template>
  <div class="fortune-page">

    <!-- ── Loading ── -->
    <div v-if="loading" class="loading-state">
      <div class="loading-orbs">
        <div class="l-orb l-orb-1"></div>
        <div class="l-orb l-orb-2"></div>
        <div class="l-orb l-orb-3"></div>
        <div class="l-core"></div>
      </div>
      <p class="loading-text">星象推演中</p>
    </div>

    <!-- ── Error ── -->
    <div v-else-if="error" class="error-state">
      <div class="error-sigil">✕</div>
      <p class="error-text">{{ error }}</p>
      <button class="retry-btn" @click="fetchFortune">重新加载</button>
    </div>

    <!-- ── Empty ── -->
    <div v-else-if="!fortune" class="empty-state">
      <div class="empty-sigil">◈</div>
      <p class="empty-title">请先创建命盘</p>
      <router-link to="/chart/new" class="go-chart-btn">去排盘 →</router-link>
    </div>

    <!-- ── Main ── -->
    <main v-else class="fortune-main" :class="{ visible: mounted }">

      <!-- Score + Info Hero -->
      <header class="hero-panel">
        <div class="hero-inner">
          <!-- Score orb - massive and dramatic -->
          <div class="score-sphere" :style="{ '--sc': scoreColor(fortune.score || 0), '--sg': scoreGlow(fortune.score || 0) }">
            <div class="sphere-ring sphere-ring-1"></div>
            <div class="sphere-ring sphere-ring-2"></div>
            <div class="sphere-ring sphere-ring-3"></div>
            <div class="sphere-glow-a"></div>
            <div class="sphere-glow-b"></div>
            <div class="sphere-value">{{ fortune.score }}</div>
            <div class="sphere-label">{{ scoreWord(fortune.score || 0) }}</div>
          </div>

          <!-- Info block -->
          <div class="hero-text">
            <div class="hero-date-row">
              <span class="date-label">{{ fortune.solar_date }}</span>
            </div>
            <div class="hero-pillar-display">
              <span class="pillar-prefix">日柱</span>
              <span class="pillar-value">{{ fortune.day_gan_zhi }}</span>
            </div>
            <p class="hero-summary" v-if="fortune.analysis?.overall?.summary">{{ fortune.analysis.overall.summary }}</p>
            <p class="hero-tip" v-if="fortune.analysis?.overall?.key_tip">{{ fortune.analysis.overall.key_tip }}</p>
          </div>
        </div>
      </header>

      <!-- Dimension matrix -->
      <section v-if="fortune.analysis?.categories?.length" class="dimension-panel" aria-label="分项运势">
        <div class="dimension-head">
          <div>
            <span class="dimension-eyebrow">分项运势矩阵</span>
            <h2 class="dimension-title">把今天拆成可行动的九个维度</h2>
          </div>
          <div class="score-breakdown" v-if="fortune.analysis?.overall">
            <span>细项 {{ fortune.analysis.overall.detail_score ?? fortune.score }}</span>
            <span>基础 {{ fortune.analysis.overall.base_score ?? fortune.score }}</span>
            <strong>{{ fortune.analysis.overall.level || scoreWord(fortune.score || 0) }}</strong>
          </div>
        </div>

        <div class="dimension-grid">
          <article
            v-for="(c, ci) in fortune.analysis.categories"
            :key="c.name"
            class="dimension-card"
            :class="`trend-${c.trend || 'flat'}`"
            :style="{ animationDelay: (ci * 70) + 'ms', background: categoryBg(categoryScore(c)) }"
          >
            <div class="dimension-top">
              <div>
                <span class="dimension-name">{{ c.name }}</span>
                <span class="dimension-stars">{{ c.stars }}</span>
              </div>
              <strong class="dimension-score">{{ categoryScore(c) }}</strong>
            </div>

            <div class="dimension-meter" aria-hidden="true">
              <span :style="{ width: `${categoryScore(c)}%` }"></span>
            </div>

            <div class="dimension-meta-row">
              <span>{{ c.level || scoreWord(categoryScore(c)) }}</span>
              <span>{{ trendLabel(c.trend) }}</span>
              <span>权重 {{ c.weight ?? 0 }}%</span>
            </div>

            <p class="dimension-analysis" v-if="c.analysis">{{ c.analysis }}</p>
            <p class="dimension-advice" v-if="c.advice">{{ c.advice }}</p>

            <div v-if="c.keywords?.length" class="dimension-tags">
              <span v-for="kw in c.keywords.slice(0, 4)" :key="`${c.name}-${kw}`">{{ kw }}</span>
            </div>
          </article>
        </div>
      </section>

      <!-- Content -->
      <div class="main-grid">
        <DailyFortune
          :solar-date="fortune.solar_date"
          :day-gan-zhi="fortune.day_gan_zhi"
          :week-day="fortune.week_day"
          :lunar-date="fortune.lunar_date"
          :sheng-xiao="fortune.sheng_xiao"
          :lucky-color="fortune.lucky_color"
          :lucky-number="fortune.lucky_number"
          :wealth-dir="fortune.wealth_direction"
          :chong-sha="fortune.clash_zodiac"
          :auspicious-hours="fortune.auspicious_hours"
          :yi-ji="`宜: ${fortune.yi?.join('、')} 忌: ${fortune.ji?.join('、')}`"
          :element-images="fortune.element_images"
          :today-elements="fortune.today_elements"
          :tiao-hou="fortune.tiao_hou"
          :ji-shen="fortune.ji_shen"
          :xiong-shen="fortune.xiong_shen"
          :tai-shen="fortune.tai_shen"
          :peng-zu="fortune.peng_zu"
          :gua="fortune.gua"
          :jie-qi="fortune.jie_qi"
          :sheng-ke-analysis="fortune.sheng_ke_analysis"
          :flow-impact="fortune.flow_impact"
          :season-element-advice="fortune.season_element_advice"
          :today-ten-god="fortune.today_ten_god"
          :ten-god-favorable="fortune.ten_god_favorable"
          :ten-god-desc="fortune.ten_god_desc"
          :twelve-stage="fortune.twelve_stage"
          :stage-favorable="fortune.stage_favorable"
          :stage-desc="fortune.stage_desc"
          :stage-flexible="fortune.stage_flexible"
          :hidden-stems="fortune.hidden_stems"
          :stem-relations="fortune.stem_relations"
          :branch-relations="fortune.branch_relations"
          :activated-shen-sha="fortune.activated_shen_sha"
          :dayun-influence="fortune.dayun_influence"
          :liunian-influence="fortune.liunian_influence"
          :advance-retreat="fortune.advance_retreat"
          :yongshen-impact="fortune.yongshen_impact"
          :overall-verdict="fortune.overall_verdict"
          :favor-score="fortune.favor_score"
          :pattern-name="fortune.pattern_name"
          :pattern-type="fortune.pattern_type"
          :pattern-favorable="fortune.pattern_favorable"
          :pattern-unfavorable="fortune.pattern_unfavorable"
        />

        <aside v-if="activeGuide(fortune)" class="side-panels">
          <div class="panel-card panel-accent-gold">
            <div class="panel-header">
              <span class="panel-icon">✦</span>
              <h3 class="panel-title">精准开运指南</h3>
              <span v-if="guideConfidenceLabel(activeGuide(fortune))" class="guide-confidence">{{ guideConfidenceLabel(activeGuide(fortune)) }} · {{ activeGuide(fortune)?.confidence }}%</span>
            </div>
            <div class="panel-body">
              <div class="guide-item" v-if="guideColorValue(activeGuide(fortune))">
                <span class="guide-label">幸运色</span>
                <span class="guide-value">
                  <span class="color-swatch" :style="{ background: guideColorValue(activeGuide(fortune)) }"></span>
                  {{ guideColorValue(activeGuide(fortune)) }}
                </span>
                <small class="guide-reason">{{ activeGuide(fortune)?.lucky_colors?.[0]?.reason }}</small>
              </div>
              <div class="guide-item" v-if="guideNumberValue(activeGuide(fortune))">
                <span class="guide-label">幸运数字</span>
                <span class="guide-value guide-value-xl">{{ guideNumberValue(activeGuide(fortune)) }}</span>
                <small class="guide-reason">{{ activeGuide(fortune)?.lucky_numbers?.[0]?.reason }}</small>
              </div>
              <div class="guide-item" v-if="activeGuide(fortune)?.wealth_direction?.value">
                <span class="guide-label">财位</span>
                <span class="guide-value">{{ activeGuide(fortune)?.wealth_direction.value }}</span>
                <small class="guide-reason">{{ activeGuide(fortune)?.wealth_direction.reason }}</small>
              </div>
              <div class="guide-item" v-if="activeGuide(fortune)?.face_direction?.value">
                <span class="guide-label">朝向</span>
                <span class="guide-value">{{ activeGuide(fortune)?.face_direction.value }}</span>
                <small class="guide-reason">{{ activeGuide(fortune)?.face_direction.reason }}</small>
              </div>
              <p v-if="activeGuide(fortune)?.strategy" class="guide-strategy">{{ activeGuide(fortune)?.strategy }}</p>
              <p v-if="activeGuide(fortune)?.analysis" class="guide-analysis">{{ activeGuide(fortune)?.analysis }}</p>
            </div>
          </div>

          <div v-if="activeGuide(fortune)?.recommended_actions?.length || activeGuide(fortune)?.cautions?.length" class="panel-card panel-accent-crimson">
            <div class="panel-header">
              <span class="panel-icon">☯</span>
              <h3 class="panel-title">行动与避忌</h3>
            </div>
            <div v-if="activeGuide(fortune)?.recommended_actions?.length" class="guide-list">
              <span class="guide-list-title">宜用</span>
              <div v-for="item in activeGuide(fortune)?.recommended_actions?.slice(0, 4)" :key="`action-${item.label}-${item.value}`" class="guide-mini">
                <strong>{{ item.value }}</strong>
                <small>{{ item.reason }}</small>
              </div>
            </div>
            <div v-if="activeGuide(fortune)?.cautions?.length" class="guide-list">
              <span class="guide-list-title guide-list-warn">避忌</span>
              <div v-for="item in activeGuide(fortune)?.cautions?.slice(0, 4)" :key="`caution-${item.label}-${item.value}`" class="guide-mini warn">
                <strong>{{ item.value }}</strong>
                <small>{{ item.reason }}</small>
              </div>
            </div>
            <div v-if="activeGuide(fortune)?.best_hours?.length" class="guide-hours">
              <span v-for="item in activeGuide(fortune)?.best_hours?.slice(0, 4)" :key="`hour-${item.label}-${item.value}`">{{ item.label }} {{ item.value }}</span>
            </div>
          </div>

          <div v-if="activeGuide(fortune)?.primary_element" class="panel-card panel-accent-crimson">
            <div class="panel-header">
              <span class="panel-icon">☯</span>
              <h3 class="panel-title">喜忌五行</h3>
            </div>
            <div class="five-elements">
              <span
                v-for="el in ['金','木','水','火','土']"
                :key="el"
                class="el-badge"
                :class="{
                  'el-fav': activeGuide(fortune)?.primary_element === el || activeGuide(fortune)?.secondary_element === el,
                  'el-dis': activeGuide(fortune)?.avoid_element === el,
                }"
              >{{ el }}</span>
            </div>
          </div>
        </aside>
      </div>

      <div class="fortune-nav">
        <router-link :to="`/fortune/weekly?chart_id=${chartId}`" class="fortune-nav-link">
          本周运势 →
        </router-link>
        <router-link :to="`/fortune/monthly?chart_id=${chartId}`" class="fortune-nav-link">
          本月运势 →
        </router-link>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* ── Page ── */
.fortune-page {
  min-height: 100vh;
  background: transparent;
  position: relative;
  overflow-x: hidden;
}

/* ── Loading ── */
.loading-state {
  position: relative; z-index: 2;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  min-height: 100vh; gap: 2.5rem;
}
.loading-orbs { position: relative; width: 100px; height: 100px; }
.l-orb {
  position: absolute; inset: 0; border-radius: 50%;
  border: 1px solid var(--line-strong);
  animation: l-spin linear infinite;
}
.l-orb-2 { inset: 15px; border-color: var(--line-subtle); animation-duration: 5s; animation-direction: reverse; }
.l-orb-3 { inset: 30px; border-color: var(--line-subtle); animation-duration: 8s; }
.l-core {
  position: absolute; inset: 40px; border-radius: 50%;
  background: var(--accent-dim);
  animation: l-pulse 2s ease-in-out infinite;
}
@keyframes l-spin { to { transform: rotate(360deg); } }
@keyframes l-pulse { 0%,100%{transform:scale(1);opacity:0.3} 50%{transform:scale(1.6);opacity:0.8} }
.loading-text { color: var(--text-soft); font-size: 11px; letter-spacing: 5px; text-transform: uppercase; }

/* ── Error ── */
.error-state {
  position: relative; z-index: 2;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  min-height: 100vh; gap: 1.5rem;
}
.error-sigil { font-size: 4rem; color: var(--crimson); opacity: 0.5; }
.error-text { color: var(--text-muted); font-size: 0.95rem; }
.retry-btn {
  padding: 0.7rem 2rem;
  background: var(--crimson);
  color: var(--destructive-foreground); border: none; border-radius: 8px;
  font-size: 0.85rem; font-weight: 700; cursor: pointer;
  box-shadow: 0 4px 20px color-mix(in oklab, var(--crimson) 30%, transparent);
  transition: all 0.3s;
}
.retry-btn:hover { transform: translateY(-2px); box-shadow: 0 8px 30px color-mix(in oklab, var(--crimson) 50%, transparent); }

/* ── Empty ── */
.empty-state {
  position: relative; z-index: 2;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  min-height: 100vh; gap: 1.5rem;
}
.empty-sigil { font-size: 4rem; color: var(--text-soft); }
.empty-title { color: var(--text-muted); font-size: 1.1rem; }
.go-chart-btn {
  padding: 0.8rem 2.5rem;
  background: var(--accent);
  color: var(--bg); font-weight: 800; border: none;
  border-radius: 50px; cursor: pointer; text-decoration: none;
  font-size: 0.9rem; letter-spacing: 1px;
  box-shadow: 0 4px 30px color-mix(in oklab, var(--accent) 40%, transparent);
  transition: all 0.3s;
}
.go-chart-btn:hover { transform: translateY(-3px); box-shadow: 0 8px 50px color-mix(in oklab, var(--accent) 60%, transparent); }

/* ── Main ── */
.fortune-main {
  position: relative; z-index: 2;
  max-width: 960px; margin: 0 auto;
  padding: 2.5rem 1.5rem 5rem;
  opacity: 0; transform: translateY(30px);
  transition: opacity 1s ease, transform 1s ease;
}
.fortune-main.visible { opacity: 1; transform: translateY(0); }

/* ── Hero panel ── */
.hero-panel {
  background: color-mix(in oklab, var(--surface-1) 88%, transparent);
  border: 1px solid var(--line-strong);
  border-radius: 20px;
  padding: 2.5rem;
  margin-bottom: 1.5rem;
  position: relative;
  overflow: hidden;
  box-shadow: var(--shadow-lg), inset 0 1px 0 var(--line-subtle);
}
.hero-panel::before {
  content: '';
  position: absolute; left: 0; top: 0; bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, var(--accent), var(--crimson), var(--accent));
  border-radius: 2px;
  opacity: 0.6;
}
.hero-panel::after {
  content: '';
  position: absolute; top: -50%; right: -10%;
  width: 400px; height: 400px;
  background: radial-gradient(circle, var(--accent-dim), transparent 60%);
  pointer-events: none;
}
.hero-inner { display: flex; align-items: center; gap: 3rem; }

/* ── Score sphere ── */
.score-sphere {
  flex-shrink: 0;
  width: 180px; height: 180px;
  border-radius: 50%;
  background: var(--surface-0);
  border: 2px solid var(--line-strong);
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  position: relative;
  box-shadow:
    0 0 0 1px var(--line-subtle),
    0 0 60px var(--sg, var(--line-subtle)),
    0 0 120px var(--sg, var(--line-subtle)),
    inset 0 0 60px rgba(0,0,0,0.15);
}
.sphere-ring {
  position: absolute; border-radius: 50%;
  border: 1px solid var(--line-subtle);
  animation: ring-spin linear infinite;
}
.sphere-ring-1 { inset: -12px; animation-duration: 20s; }
.sphere-ring-2 { inset: -24px; animation-duration: 35s; animation-direction: reverse; border-color: color-mix(in oklab, var(--crimson) 5%, transparent); }
.sphere-ring-3 { inset: -40px; animation-duration: 50s; border-color: var(--line-subtle); }
@keyframes ring-spin { to { transform: rotate(360deg); } }
.sphere-glow-a {
  position: absolute; inset: -30px; border-radius: 50%;
  background: radial-gradient(circle, var(--sg, var(--line-subtle)) 0%, transparent 70%);
  animation: glow-pulse 3s ease-in-out infinite;
}
.sphere-glow-b {
  position: absolute; inset: 10%; border-radius: 50%;
  background: radial-gradient(circle, var(--sc, #cbd5e1) 0%, transparent 70%);
  opacity: 0.06;
  animation: glow-inner 4s ease-in-out infinite;
}
@keyframes glow-pulse { 0%,100%{opacity:0.6;transform:scale(1)} 50%{opacity:1;transform:scale(1.08)} }
@keyframes glow-inner { 0%,100%{opacity:0.04} 50%{opacity:0.1} }
.sphere-value {
  font-family: var(--font-serif);
  font-size: 4rem; font-weight: 900;
  color: var(--sc, #cbd5e1);
  line-height: 1; position: relative; z-index: 2;
  text-shadow: 0 0 60px var(--sg, var(--accent-glow)), 0 0 20px var(--sc, var(--accent));
  transition: color 0.6s, text-shadow 0.6s;
  letter-spacing: -2px;
}
.sphere-label {
  font-size: 0.65rem; font-weight: 700;
  color: var(--text-muted);
  letter-spacing: 4px; text-transform: uppercase;
  position: relative; z-index: 2; margin-top: 2px;
}

/* ── Hero text ── */
.hero-text { flex: 1; }
.hero-date-row { margin-bottom: 0.75rem; }
.date-label {
  font-size: 0.78rem; color: var(--text-muted);
  letter-spacing: 3px; text-transform: uppercase;
}
.hero-pillar-display {
  display: flex; align-items: baseline; gap: 0.75rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--line-subtle);
}
.pillar-prefix {
  font-size: 0.78rem; color: var(--text-soft); letter-spacing: 2px;
}
.pillar-value {
  font-family: var(--font-serif);
  font-size: 2.5rem; font-weight: 900;
  color: var(--accent); letter-spacing: 4px;
  text-shadow: 0 0 40px var(--accent-glow);
}
.hero-summary {
  font-size: 0.88rem; color: var(--text-muted);
  line-height: 1.8; margin: 0 0 0.75rem;
  border-left: 2px solid var(--line-strong);
  padding-left: 0.75rem;
}
.hero-tip {
  font-size: 0.8rem; color: var(--accent); font-weight: 600;
  margin: 0; opacity: 0.85;
}

/* ── Dimension matrix ── */
.dimension-panel {
  margin-bottom: 1.25rem;
  padding: 1.35rem;
  border: 1px solid var(--line-strong);
  border-radius: 16px;
  background:
    linear-gradient(135deg, color-mix(in oklab, var(--surface-1) 92%, transparent), color-mix(in oklab, var(--surface-0) 78%, transparent)),
    radial-gradient(circle at 8% 12%, var(--accent-dim), transparent 34%);
  box-shadow: var(--shadow-lg), inset 0 1px 0 var(--line-subtle);
  position: relative;
  overflow: hidden;
}
.dimension-panel::before {
  content: '';
  position: absolute;
  inset: 0 0 auto;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent), color-mix(in oklab, var(--crimson) 45%, var(--accent)), transparent);
  opacity: 0.75;
}
.dimension-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 1rem;
  margin-bottom: 1rem;
  position: relative;
  z-index: 1;
}
.dimension-eyebrow {
  display: block;
  color: var(--accent);
  font-size: 0.66rem;
  font-weight: 800;
  letter-spacing: 0.18em;
}
.dimension-title {
  margin: 0.25rem 0 0;
  color: var(--text);
  font-family: var(--font-serif);
  font-size: 1.05rem;
  font-weight: 800;
  letter-spacing: 0;
}
.score-breakdown {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.4rem;
  max-width: 320px;
}
.score-breakdown span,
.score-breakdown strong {
  padding: 0.34rem 0.56rem;
  border-radius: 8px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: 0.68rem;
  font-weight: 700;
}
.score-breakdown strong {
  color: var(--accent);
  background: var(--accent-dim);
  border-color: var(--line-focus);
}
.dimension-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
  position: relative;
  z-index: 1;
}
.dimension-card {
  min-height: 246px;
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
  padding: 0.95rem;
  border: 1px solid var(--line-subtle);
  border-radius: 10px;
  box-shadow: inset 0 1px 0 var(--line-subtle);
  animation: dimension-in 0.55s ease both;
  transition: transform 0.25s, border-color 0.25s, box-shadow 0.25s;
}
@keyframes dimension-in { from{opacity:0;transform:translateY(12px)} to{opacity:1;transform:translateY(0)} }
.dimension-card:hover {
  transform: translateY(-2px);
  border-color: var(--line-focus);
  box-shadow: 0 14px 36px rgba(var(--jade-accent-rgb), 0.12), inset 0 1px 0 var(--line-subtle);
}
.dimension-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.75rem;
}
.dimension-name {
  display: block;
  color: var(--text);
  font-size: 0.9rem;
  font-weight: 850;
}
.dimension-stars {
  display: block;
  min-height: 1rem;
  margin-top: 0.18rem;
  color: var(--text-soft);
  font-size: 0.66rem;
  letter-spacing: 0.06em;
}
.dimension-score {
  color: var(--accent);
  font-family: var(--font-serif);
  font-size: 1.8rem;
  line-height: 1;
  text-shadow: 0 0 24px var(--accent-glow);
}
.dimension-meter {
  height: 6px;
  border-radius: 999px;
  background: color-mix(in oklab, var(--surface-0) 70%, transparent);
  border: 1px solid var(--line-subtle);
  overflow: hidden;
}
.dimension-meter span {
  display: block;
  height: 100%;
  min-width: 6%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--accent), color-mix(in oklab, var(--accent) 72%, var(--crimson)));
  box-shadow: 0 0 18px var(--accent-glow);
}
.dimension-meta-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.dimension-meta-row span,
.dimension-tags span {
  padding: 0.22rem 0.45rem;
  border-radius: 8px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: 0.62rem;
  font-weight: 700;
}
.dimension-analysis {
  margin: 0;
  color: var(--text-muted);
  font-size: 0.73rem;
  line-height: 1.62;
}
.dimension-advice {
  margin: auto 0 0;
  padding-left: 0.55rem;
  border-left: 2px solid var(--line-focus);
  color: var(--accent);
  font-size: 0.72rem;
  line-height: 1.55;
}
.dimension-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.dimension-card.trend-up .dimension-meta-row span:nth-child(2) {
  color: var(--accent);
  border-color: var(--line-focus);
}
.dimension-card.trend-down .dimension-meta-row span:nth-child(2) {
  color: var(--crimson);
  border-color: color-mix(in oklab, var(--crimson) 22%, transparent);
}

/* ── Main grid ── */
.main-grid {
  display: grid;
  grid-template-columns: 1fr 230px;
  gap: 1rem;
  align-items: start;
}

/* ── Side panels ── */
.side-panels { display: flex; flex-direction: column; gap: 0.75rem; }
.panel-card {
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  border-radius: 14px;
  padding: 1.25rem;
  position: relative; overflow: hidden;
}
.panel-accent-gold { border-color: var(--line-strong); }
.panel-accent-gold::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
}
.panel-accent-crimson { border-color: color-mix(in oklab, var(--crimson) 12%, transparent); }
.panel-accent-crimson::before {
  content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px;
  background: linear-gradient(90deg, transparent, var(--crimson), transparent);
}
.panel-header {
  display: flex; align-items: center; gap: 0.5rem;
  margin-bottom: 1rem;
  padding-bottom: 0.6rem;
  border-bottom: 1px solid var(--line-subtle);
}
.panel-icon { font-size: 0.85rem; color: var(--text-muted); }
.panel-title {
  font-family: var(--font-serif);
  font-size: 0.78rem; font-weight: 700;
  color: var(--accent); margin: 0; letter-spacing: 2px;
}
.guide-confidence {
  margin-left: auto;
  font-size: 0.58rem;
  color: var(--text-soft);
  letter-spacing: 0.08em;
}
.guide-item {
  display: flex; flex-direction: column; gap: 0.2rem;
  padding: 0.5rem 0.6rem;
  background: var(--glass-bg);
  border-radius: 8px; margin-bottom: 0.4rem;
  border: 1px solid var(--line-subtle);
  transition: background 0.25s;
}
.guide-item:last-child { margin-bottom: 0; }
.guide-item:hover { background: var(--glass-bg-hover); }
.guide-label {
  font-size: 0.58rem; color: var(--text-soft);
  text-transform: uppercase; letter-spacing: 0.1em;
}
.guide-value {
  font-size: 0.82rem; color: var(--text);
  font-weight: 500; display: flex; align-items: center; gap: 0.4rem;
}
.guide-value-xl { color: var(--accent); font-size: 1.4rem; font-weight: 900; letter-spacing: 2px; text-shadow: 0 0 20px var(--accent-glow); }
.color-swatch { display: inline-block; width: 16px; height: 16px; border-radius: 50%; border: 1px solid var(--line-strong); flex-shrink: 0; }
.guide-reason {
  font-size: 0.64rem;
  line-height: 1.5;
  color: var(--text-soft);
}
.guide-strategy,
.guide-analysis {
  margin: 0.7rem 0 0;
  font-size: 0.72rem;
  line-height: 1.65;
  color: var(--text-muted);
}
.guide-strategy {
  color: var(--accent);
  border-left: 2px solid var(--line-focus);
  padding-left: 0.55rem;
}
.guide-list {
  display: flex;
  flex-direction: column;
  gap: 0.42rem;
  margin-bottom: 0.75rem;
}
.guide-list:last-child { margin-bottom: 0; }
.guide-list-title {
  font-size: 0.58rem;
  color: rgba(var(--jade-accent-rgb), 1);
  letter-spacing: 0.12em;
}
.guide-list-warn { color: var(--crimson); }
.guide-mini {
  display: flex;
  flex-direction: column;
  gap: 0.16rem;
  padding: 0.48rem 0.55rem;
  border-radius: 8px;
  border: 1px solid rgba(var(--jade-accent-rgb), 0.16);
  background: rgba(var(--jade-accent-rgb), 0.05);
}
.guide-mini.warn {
  border-color: color-mix(in oklab, var(--crimson) 16%, transparent);
  background: color-mix(in oklab, var(--crimson) 5%, transparent);
}
.guide-mini strong {
  color: var(--text);
  font-size: 0.76rem;
}
.guide-mini small {
  color: var(--text-soft);
  line-height: 1.45;
  font-size: 0.63rem;
}
.guide-hours {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}
.guide-hours span {
  padding: 0.24rem 0.55rem;
  border-radius: 999px;
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  color: var(--text-muted);
  font-size: 0.64rem;
}
.five-elements { display: flex; gap: 0.35rem; }
.el-badge {
  flex: 1; aspect-ratio: 1;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px;
  font-size: 0.8rem; font-weight: 800;
  color: var(--text-soft);
  background: var(--glass-bg);
  border: 1px solid var(--line-subtle);
  transition: all 0.3s;
}
.el-badge.el-fav {
  color: var(--accent);
  background: var(--accent-dim);
  border-color: var(--line-focus);
  box-shadow: 0 0 20px var(--accent-glow), inset 0 0 10px var(--accent-dim);
  text-shadow: 0 0 10px var(--accent-glow);
}
.el-badge.el-dis {
  color: var(--crimson);
  background: color-mix(in oklab, var(--crimson) 6%, transparent);
  border-color: color-mix(in oklab, var(--crimson) 20%, transparent);
}

/* ── Responsive ── */
@media (max-width: 768px) {
  .hero-inner { flex-direction: column; align-items: center; text-align: center; gap: 1.5rem; }
  .hero-panel { padding: 1.75rem 1.5rem; }
  .score-sphere { width: 140px; height: 140px; }
  .sphere-value { font-size: 3rem; }
  .pillar-value { font-size: 2rem; }
  .dimension-panel { padding: 1rem; border-radius: 14px; }
  .dimension-head { align-items: flex-start; flex-direction: column; }
  .score-breakdown { justify-content: flex-start; max-width: none; }
  .dimension-grid { grid-template-columns: 1fr; }
  .dimension-card { min-height: 0; }
  .main-grid { grid-template-columns: 1fr; }
  .fortune-main { padding: 1.5rem 1rem 4rem; }
}

/* ── Fortune nav ── */
.fortune-nav {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
  padding: 1.5rem 0 0;
}
.fortune-nav-link {
  color: var(--accent);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s;
}
.fortune-nav-link:hover {
  text-shadow: 0 0 12px var(--accent-glow);
}

/* ── Dark mode overrides ── */
:global(.dark) .score-sphere {
  background: radial-gradient(circle at 35% 35%, #1a1530 0%, #050308 100%);
  box-shadow:
    0 0 0 1px rgba(203, 213, 225,0.05),
    0 0 60px var(--sg, rgba(203, 213, 225,0.2)),
    0 0 120px var(--sg, rgba(203, 213, 225,0.1)),
    inset 0 0 60px rgba(0,0,0,0.6);
}
:global(.dark) .sphere-ring-2 {
  border-color: rgba(251, 113, 133, 0.05);
}
</style>
