<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import client from '../api/client'
import DailyFortune from '../components/DailyFortune.vue'

interface ElementImage {
  element: string
  image_url: string
  description: string
}

interface FortuneData {
  solar_date: string
  day_gan_zhi: string
  week_day?: string
  lunar_date?: string
  sheng_xiao?: string
  yi_ji?: string
  chong_sha?: string
  element_images?: ElementImage[]
  score?: number
  analysis?: {
    overall?: { summary?: string; key_tip?: string }
    categories?: { name: string; stars: string }[]
    lucky_guide?: { colors?: string; numbers?: string; actions?: string; outfit?: string }
  }
  lucky_color?: string
  lucky_number?: number
  wealth_direction?: string
  clash_zodiac?: string
  auspicious_hours?: string[]
  yi?: string[]
  ji?: string[]
}

const route = useRoute()

const fortune = ref<FortuneData | null>(null)
const loading = ref(true)
const error = ref('')

function todayStr(): string {
  const d = new Date()
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

async function fetchFortune() {
  const chartId = route.query.chart_id
  if (!chartId) {
    error.value = '请提供 chart_id 参数'
    loading.value = false
    return
  }

  try {
    const { data } = await client.post('/fortune', {
      chart_id: Number(chartId),
      query_date: todayStr(),
    })
    fortune.value = data
  } catch (e: any) {
    error.value = e.response?.data?.error || '加载运势失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchFortune()
})
</script>

<template>
  <div class="fortune-page">
    <!-- Loading -->
    <div v-if="loading" class="state-box">
      <el-skeleton animated>
        <template #template>
          <el-skeleton-item variant="rect" style="width: 100%; height: 100px; margin-bottom: 12px; border-radius: 12px;" />
          <el-skeleton-item variant="rect" style="width: 100%; height: 100px; margin-bottom: 12px; border-radius: 12px;" />
          <el-skeleton-item variant="rect" style="width: 100%; height: 100px; border-radius: 12px;" />
        </template>
      </el-skeleton>
    </div>

    <!-- Error -->
    <div v-else-if="error" class="state-box">
      <el-result icon="error" title="加载失败" sub-title="请检查网络连接后重试">
        <template #extra>
          <el-button type="primary" @click="fetchFortune">重试</el-button>
          <router-link to="/history" style="margin-left: 12px;">
            <el-button>返回历史记录</el-button>
          </router-link>
        </template>
      </el-result>
    </div>

    <!-- Empty -->
    <div v-else-if="!fortune" class="state-box">
      <el-empty description="请先创建命盘">
        <template #default>
          <router-link to="/chart/new">
            <el-button type="primary">去排盘</el-button>
          </router-link>
        </template>
      </el-empty>
    </div>

    <!-- Fortune data -->
    <template v-else>
      <div style="background:linear-gradient(135deg,#D4A84B,#B8860B);color:#0A0815;text-align:center;padding:24px;border-radius:14px;margin-bottom:24px">
        <div class="text-5xl font-bold">{{ fortune.score }}</div>
        <div class="text-sm opacity-80 mt-1">今日运势评分</div>
      </div>

      <!-- Analysis text -->
      <div v-if="fortune.analysis" class="mt-4 glass-card rounded-lg shadow p-6">
        <h3 class="text-lg font-bold text-bazi-blue mb-4">运势详解</h3>
        <p class="text-sm text-bazi-ink/80 leading-relaxed mb-4">{{ fortune.analysis.overall?.summary }}</p>
        <p class="text-sm text-bazi-red font-medium">{{ fortune.analysis.overall?.key_tip }}</p>
        <div class="grid grid-cols-5 gap-3 mt-4">
          <div v-for="c in fortune.analysis.categories" :key="c.name" class="text-center p-2  rounded">
            <div class="text-xs text-bazi-blue/60">{{ c.name }}</div>
            <div class="text-sm font-bold mt-1">{{ c.stars }}</div>
          </div>
        </div>
        <h4 class="text-sm font-bold text-bazi-blue mt-6 mb-2">开运指南</h4>
        <div class="text-xs text-bazi-ink/70 space-y-1">
          <p>幸运色：{{ fortune.analysis.lucky_guide?.colors }}</p>
          <p>幸运数字：{{ fortune.analysis.lucky_guide?.numbers }}</p>
          <p>开运动作：{{ fortune.analysis.lucky_guide?.actions }}</p>
          <p>穿搭建议：{{ fortune.analysis.lucky_guide?.outfit }}</p>
        </div>
      </div>

      <DailyFortune
        :solar-date="fortune.solar_date"
        :day-gan-zhi="fortune.day_gan_zhi"
        :lucky-color="fortune.lucky_color"
        :lucky-number="fortune.lucky_number"
        :wealth-dir="fortune.wealth_direction"
        :chong-sha="fortune.clash_zodiac"
        :auspicious-hours="fortune.auspicious_hours"
        :yi-ji="`宜: ${fortune.yi?.join('、')} 忌: ${fortune.ji?.join('、')}`"
        :element-images="fortune.element_images"
      />
    </template>
  </div>
</template>

<style scoped>
.fortune-page {
  min-height: 100vh;
  background: #FAF8F3;
}

.state-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 1rem;
}

.state-box.error {
  color: var(--color-bazi-red);
}

.state-text {
  font-size: 1rem;
  color: var(--color-bazi-ink);
  margin: 0;
}

.back-link {
  color: var(--color-bazi-red);
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
}

.back-link:hover {
  text-decoration: underline;
}
</style>
