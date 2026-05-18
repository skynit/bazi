<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import client from '../api/client'
import BaziChart from '../components/BaziChart.vue'
import BirthInputForm from '../components/BirthInputForm.vue'

const route = useRoute()
const router = useRouter()
const chartId = route.params.id as string
const isNew = chartId === 'new'
const chartData = ref<any>(null)
const loading = ref(false)
const error = ref('')

onMounted(async () => {
  tryLoadChart()
})

watch(() => route.fullPath, () => {
  tryLoadChart()
})

function tryLoadChart() {
  if (isNew) {
    const raw = sessionStorage.getItem('lastChart')
    if (raw) {
      chartData.value = JSON.parse(raw)
      sessionStorage.removeItem('lastChart')
    }
    // Don't set error for new charts — just show the form
  } else {
    loadChart()
  }
}

async function loadChart() {
  loading.value = true
  error.value = ''
  try {
    const res = await client.get(`/charts/${chartId}`)
    chartData.value = res.data.chart || res.data
  } catch (err: any) {
    error.value = err.response?.data?.error || err.message || '加载命盘失败'
  } finally {
    loading.value = false
  }
}

function goFortune() {
  router.push(`/fortune?chart_id=${chartData.value.id}`)
}

function goZiWei() { router.push(`/ziwei/${chartData.value.id}`) }
</script>
<template>
  <div class="min-h-screen ">
    <!-- Header -->
    <header style="background:rgba(255,255,255,0.03);color:#EAE6DD" class="px-6 py-4 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <router-link to="/" class="text-bazi-paper/70 hover:text-white transition-colors">
          ← 返回
        </router-link>
        <h1 class="text-lg font-bold tracking-wider">八字命盘</h1>
      </div>
    </header>

    <main class="max-w-3xl mx-auto px-4 py-8">
      <!-- Loading state -->
      <div v-if="loading" class="p-8">
        <div class="skeleton h-8 w-48 mb-4"></div>
        <div class="grid grid-cols-2 gap-4">
          <div class="skeleton h-32 rounded-xl" v-for="i in 4" :key="i"></div>
        </div>
      </div>

      <!-- Error state -->
      <div v-else-if="error" class="py-20">
        <el-result icon="error" title="加载失败" sub-title="请检查网络连接后重试">
          <template #extra>
            <el-button type="primary" @click="loadChart">重试</el-button>
          </template>
        </el-result>
      </div>

      <!-- New chart: show input form -->
      <div v-else-if="isNew && !chartData">
        <h2 class="text-2xl font-bold text-bazi-ink text-center mb-8">新建命盘</h2>
        <BirthInputForm />
      </div>

      <!-- Chart display -->
      <div v-else-if="chartData">
        <BaziChart :chart="chartData" />

        <div class="mt-8 flex items-center justify-center gap-4">
          <button
            class="btn-gold"
            @click="goFortune"
          >
            查看运势
          </button>
          <el-button type="success" @click.stop="goZiWei">紫微斗数</el-button>
        </div>

        <div class="mt-8 text-center">
          <router-link
            to="/chart/new"
            class="text-bazi-blue hover: transition-colors text-sm"
          >
            重新排盘
          </router-link>
        </div>
      </div>

      <!-- Not found -->
      <div v-else class="text-center py-20">
        <p class="text-bazi-blue/60 text-lg">未找到命盘</p>
        <router-link
          to="/chart/new"
          class=" hover:underline mt-4 inline-block"
        >
          创建新的命盘
        </router-link>
      </div>
    </main>
  </div>
</template>
