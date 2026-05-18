<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const savedChartId = ref<number | null>(null)

onMounted(async () => {
  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => {})
  }
  const saved = localStorage.getItem('bazi_last_birth')
  if (saved) {
    try { savedChartId.value = JSON.parse(saved).chartId || null } catch {}
  }
})

function startChart() {
  router.push('/chart/new')
}

function continueChart() {
  if (savedChartId.value) router.push(`/chart/${savedChartId.value}`)
}
</script>

<template>
  <div class="relative min-h-[calc(100vh-56px)] flex items-center justify-center overflow-hidden">
    <div class="absolute inset-0 z-0" style="background:url('https://images.unsplash.com/photo-1617962902293-6caa0a2c68bc?w=1920&q=80') center/cover;filter:brightness(0.25)"></div>
    <div class="absolute inset-0 z-0" style="background:linear-gradient(180deg,rgba(10,8,21,0.4) 0%,rgba(10,8,21,0.95) 100%)"></div>
    <div class="relative z-10 text-center px-4 max-w-2xl">
      <div class="text-8xl mb-6 hero-symbol" style="color:#D4A84B">☯</div>
      <h1 class="text-6xl font-bold mb-4 animate-in delay-1 tracking-wide" style="color:#F0EDE4;font-family:serif">
        八字<span class="text-gold">命理</span>
      </h1>
      <p class="text-lg mb-10 animate-in delay-2" style="color:#8B8378">探索命运密码，洞察人生轨迹</p>
      <button @click="startChart" class="btn-gold text-lg animate-in delay-3">开始排盘</button>
      <button v-if="savedChartId" @click="continueChart" class="btn-ghost text-lg animate-in delay-3 mt-4">继续上次排盘</button>
      <div class="mt-16 flex gap-8 justify-center text-sm animate-in delay-3" style="color:rgba(255,255,255,0.2)">
        <span class="feature-item">命盘分析</span><span style="color:rgba(255,255,255,0.08)">·</span>
        <span class="feature-item">运势解读</span><span style="color:rgba(255,255,255,0.08)">·</span>
        <span class="feature-item">紫微斗数</span>
      </div>
    </div>
  </div>
  </template>

<style scoped>
@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-12px); }
}

@keyframes pulseGlow {
  0%, 100% { filter: drop-shadow(0 0 40px rgba(212,168,75,0.1)); }
  50% { filter: drop-shadow(0 0 80px rgba(212,168,75,0.2)); }
}

.hero-symbol {
  animation: float 4s ease-in-out infinite, pulseGlow 3s ease-in-out infinite;
}

.feature-item {
  opacity: 0;
  animation: fadeUp 0.6s ease both;
}
.feature-item:nth-child(1) { animation-delay: 0.8s; }
.feature-item:nth-child(3) { animation-delay: 1.0s; }
.feature-item:nth-child(5) { animation-delay: 1.2s; }

.btn-gold:hover {
  transform: translateY(-3px) scale(1.02);
  box-shadow: 0 12px 32px rgba(212,168,75,0.35);
}
</style>