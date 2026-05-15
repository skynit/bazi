<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

onMounted(async () => {
  if (authStore.isLoggedIn() && !authStore.user) {
    await authStore.fetchMe().catch(() => {})
  }
})

function startChart() {
  router.push('/chart/new')
}
</script>

<template>
  <div class="relative min-h-[calc(100vh-56px)] flex items-center justify-center overflow-hidden">
    <div class="absolute inset-0 z-0" style="background:url('https://images.unsplash.com/photo-1617962902293-6caa0a2c68bc?w=1920&q=80') center/cover;filter:brightness(0.25)"></div>
    <div class="absolute inset-0 z-0" style="background:linear-gradient(180deg,rgba(10,8,21,0.4) 0%,rgba(10,8,21,0.95) 100%)"></div>
    <div class="relative z-10 text-center px-4 max-w-2xl">
      <div class="text-8xl mb-6 animate-in" style="color:#D4A84B">☯</div>
      <h1 class="text-6xl font-bold mb-4 animate-in delay-1 tracking-wide" style="color:#F0EDE4;font-family:serif">
        八字<span class="text-gold">命理</span>
      </h1>
      <p class="text-lg mb-10 animate-in delay-2" style="color:#8B8378">探索命运密码，洞察人生轨迹</p>
      <button @click="startChart" class="btn-gold text-lg animate-in delay-3">开始排盘</button>
      <div class="mt-16 flex gap-8 justify-center text-sm animate-in delay-3" style="color:rgba(255,255,255,0.2)">
        <span>命盘分析</span><span style="color:rgba(255,255,255,0.08)">·</span>
        <span>运势解读</span><span style="color:rgba(255,255,255,0.08)">·</span>
        <span>紫微斗数</span>
      </div>
    </div>
  </div>
</template>