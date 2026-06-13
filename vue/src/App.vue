<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = useRouter()
const authStore = useAuthStore()
const scrolled = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))

function applyTheme(nextDark: boolean) {
  isDark.value = nextDark
  document.documentElement.classList.toggle('dark', nextDark)
  document.documentElement.style.colorScheme = nextDark ? 'dark' : 'light'
  localStorage.setItem('theme', nextDark ? 'dark' : 'light')
}

function onScroll() {
  scrolled.value = window.scrollY > 20
}

function toggleTheme() {
  applyTheme(!isDark.value)
}

onMounted(() => {
  document.documentElement.style.colorScheme = isDark.value ? 'dark' : 'light'
  window.addEventListener('scroll', onScroll, { passive: true })
})
onUnmounted(() => window.removeEventListener('scroll', onScroll))
</script>
<template>
  <div class="app-root">
    <header class="fixed inset-x-0 top-4 z-50">
      <div class="mx-auto w-full max-w-[82rem] px-4 sm:px-8">
        <nav
          class="flex items-center justify-between h-14 rounded-full border px-4 sm:px-5 lg:px-6 transition-all duration-300"
          :class="scrolled
            ? 'bg-[var(--nav-bg)] backdrop-blur-xl border-[var(--nav-border)] shadow-[var(--shadow-sm)]'
            : 'bg-[var(--nav-bg-idle)] backdrop-blur-xl border-transparent'"
        >
          <!-- Logo (left) -->
          <router-link to="/" class="flex items-center gap-3 text-decoration-none">
            <div class="relative flex items-center justify-center w-8 h-8">
              <div class="absolute -inset-[3px] border border-[var(--line-focus)] rounded-full animate-[spin_12s_linear_infinite]"></div>
              <span class="text-[1.2rem] text-[var(--accent)] [text-shadow:0_0_12px_var(--brand-glow)]">☯</span>
            </div>
            <div class="hidden sm:flex flex-col gap-0">
              <span class="font-[family-name:var(--font-serif)] text-[0.95rem] font-bold text-[var(--text)] tracking-[3px] leading-[1.1]">八字命理</span>
              <span class="text-[8px] tracking-[2px] text-[var(--text-soft)] uppercase">BaZi Fortune</span>
            </div>
          </router-link>

          <!-- Nav links (center) -->
          <nav class="hidden md:flex items-center gap-4 lg:gap-6 text-[13px] font-medium text-[var(--text-muted)]">
            <router-link to="/" class="transition-colors hover:text-[var(--accent)] aria-[current=page]:text-[var(--accent)]" exact-active-class="text-[var(--accent)]">首页</router-link>
            <template v-if="authStore.isLoggedIn()">
              <router-link to="/history" class="transition-colors hover:text-[var(--accent)]">历史</router-link>
              <div class="relative group">
                <router-link to="/fortune" class="transition-colors hover:text-[var(--accent)] flex items-center gap-1">
                  运势 <span class="text-[10px] opacity-50">▾</span>
                </router-link>
                <div class="hidden group-hover:block absolute top-full left-0 min-w-[120px] py-1.5 bg-[var(--surface-1)]/95 border border-[var(--line-strong)] rounded-[10px] shadow-[0_12px_40px_rgba(0,0,0,0.18)] backdrop-blur-[20px] z-[100]">
                  <router-link to="/fortune" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">今日运势</router-link>
                  <router-link to="/fortune/weekly" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">本周运势</router-link>
                  <router-link to="/fortune/monthly" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">本月运势</router-link>
                </div>
              </div>
            </template>
          </nav>

          <!-- Right side -->
          <div class="flex items-center gap-3">
            <button
              @click="toggleTheme"
              class="w-8 h-8 flex items-center justify-center rounded-full text-sm transition-colors hover:bg-[var(--glass-bg)]"
              :aria-label="isDark ? '切换到白天模式' : '切换到暗黑模式'"
            >
              <span v-if="isDark" class="text-[0.9rem] text-[var(--icon)]">☀</span>
              <span v-else class="text-[0.9rem] text-[var(--icon)]">☽</span>
            </button>
            <template v-if="authStore.isLoggedIn()">
              <div class="hidden sm:flex items-center gap-2">
                <div class="w-7 h-7 rounded-full bg-[var(--accent)]/10 flex items-center justify-center text-[11px] font-semibold text-[var(--accent)]">
                  {{ authStore.user?.username?.charAt(0).toUpperCase() }}
                </div>
                <span class="text-[12px] font-medium text-[var(--text-muted)] hidden lg:inline">{{ authStore.user?.username }}</span>
              </div>
              <button @click="authStore.logout();router.push('/')" class="text-[12px] text-[var(--text-muted)] hover:text-[var(--accent)] transition-colors hidden sm:inline">退出</button>
              <router-link to="/chart/new" class="inline-flex items-center justify-center gap-1.5 h-8 px-4 text-[12px] font-medium rounded-full bg-[var(--text)] text-[var(--bg)] hover:bg-[var(--text)]/90 transition-colors shadow-[0_1px_2px_rgba(0,0,0,0.06)]">
                排盘
              </router-link>
            </template>
            <template v-else>
              <router-link to="/login" class="text-[13px] font-medium text-[var(--text-muted)] hover:text-[var(--accent)] transition-colors">登录</router-link>
              <router-link to="/register" class="inline-flex items-center justify-center h-8 px-4 text-[12px] font-medium rounded-full bg-[var(--text)] text-[var(--bg)] hover:bg-[var(--text)]/90 transition-colors shadow-[0_1px_2px_rgba(0,0,0,0.06)]">
                注册
              </router-link>
            </template>
          </div>
        </nav>
      </div>
    </header>
    <main class="app-main"><router-view /></main>
  </div>
</template>

<style scoped>
.app-root {
  min-height: 100vh;
  background: var(--bg);
  display: flex;
  flex-direction: column;
}

.app-main {
  flex: 1;
  padding-top: 80px;
}

.router-link-active {
  color: var(--accent) !important;
}
</style>
