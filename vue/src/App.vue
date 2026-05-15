<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
const router = useRouter()
const authStore = useAuthStore()
</script>
<template>
  <div class="min-h-screen" style="background:#0A0815">
    <header class="sticky top-0 z-50 border-b" style="background:rgba(10,8,21,0.88);backdrop-filter:blur(20px);border-color:rgba(212,168,75,0.1)">
      <div class="max-w-6xl mx-auto px-6 py-3 flex items-center justify-between">
        <router-link to="/" class="flex items-center gap-2">
          <span class="text-2xl" style="color:#D4A84B">☯</span>
          <span class="text-xl font-bold tracking-widest" style="color:#F0EDE4">八字命理</span>
        </router-link>
        <nav class="flex items-center gap-5 text-sm" style="color:#8B8378">
          <router-link to="/" class="hover:text-gold transition-colors" style="color:#8B8378">首页</router-link>
          <template v-if="authStore.isLoggedIn()">
            <router-link to="/history" class="hover:text-gold transition-colors" style="color:#8B8378">历史</router-link>
            <router-link to="/fortune" class="hover:text-gold transition-colors" style="color:#8B8378">运势</router-link>
            <span style="color:rgba(255,255,255,0.08)">|</span>
            <span class="text-xs" style="color:rgba(255,255,255,0.25)">{{ authStore.user?.username }}</span>
            <button @click="authStore.logout();router.push('/')" class="hover:text-gold transition-colors cursor-pointer" style="color:#8B8378;background:none;border:none">退出</button>
          </template>
          <template v-else>
            <router-link to="/login" class="hover:text-gold transition-colors" style="color:#8B8378">登录</router-link>
            <router-link to="/register" class="btn-gold text-sm px-4 py-1.5">注册</router-link>
          </template>
        </nav>
      </div>
    </header>
    <main><router-view /></main>
  </div>
</template>
