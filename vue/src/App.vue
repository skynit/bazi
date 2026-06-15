<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useThemeStore } from './stores/theme'
import type { WuxingKey } from './composables/useWuxingThemes'

const router = useRouter()
const authStore = useAuthStore()
const themeStore = useThemeStore()
const scrolled = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))

const wuxingCycle: Array<{ key: WuxingKey; label: string }> = [
  { key: 'mu', label: '木' },
  { key: 'huo', label: '火' },
  { key: 'tu', label: '土' },
  { key: 'jin', label: '金' },
  { key: 'shui', label: '水' }
]

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

// /** 紫微斗数入口：有最近排盘则跳到对应 chart，否则去新建 */
// function goZiwei() {
//   const saved = localStorage.getItem('bazi_last_birth')
//   if (saved) {
//     try {
//       const id = JSON.parse(saved).chartId
//       if (id) {
//         router.push(`/ziwei/${id}`)
//         return
//       }
//     } catch { /* fallthrough */ }
//   }
//   router.push('/chart/new')
// }

/**
 * 紫微斗数路由需要 chartId。
 * 优先使用最近一次排盘 (bazi_last_birth)，否则引导用户先排盘。
 */
function goZiwei() {
  try {
    const raw = localStorage.getItem('bazi_last_birth')
    const chartId = raw ? JSON.parse(raw).chartId : null
    if (chartId) {
      router.push(`/ziwei/${chartId}`)
      return
    }
  } catch { /* ignore */ }
  router.push('/chart/new')
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
          class="relative flex items-center h-14 rounded-full border px-4 sm:px-5 lg:px-6 transition-all duration-300"
          :class="scrolled
            ? 'bg-[var(--nav-bg)] backdrop-blur-md border-[var(--nav-border)] shadow-[var(--shadow-sm)]'
            : 'bg-[var(--nav-bg-idle)] backdrop-blur-md border-transparent'"
        >
          <!-- Logo (left) -->
          <router-link to="/" class="flex items-center gap-3 text-decoration-none shrink-0 mr-auto">
            <div class="relative flex items-center justify-center w-8 h-8">
              <div class="absolute -inset-[3px] border border-[var(--line-focus)] rounded-full animate-[spin_12s_linear_infinite]"></div>
              <span class="text-[1.2rem] text-[var(--accent)] [text-shadow:0_0_12px_var(--brand-glow)]">☯</span>
            </div>
            <div class="hidden sm:flex flex-col gap-0">
              <span class="font-[family-name:var(--font-serif)] text-[0.95rem] font-bold text-[var(--text)] tracking-[3px] leading-[1.1]">八字命理</span>
              <span class="text-[8px] tracking-[2px] text-[var(--text-soft)] uppercase">BaZi Fortune</span>
            </div>
          </router-link>

          <!-- Nav links (absolutely centered, optical balance) -->
          <nav class="hidden md:flex absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 items-center gap-5 lg:gap-7 text-[13px] font-medium text-[var(--text-muted)] whitespace-nowrap">
            <router-link to="/" class="transition-colors hover:text-[var(--accent)] aria-[current=page]:text-[var(--accent)]" exact-active-class="text-[var(--accent)]">首页</router-link>
            <template v-if="authStore.isLoggedIn()">
              <router-link to="/history" class="transition-colors hover:text-[var(--accent)]">历史</router-link>
              <router-link
                to="/chart/new"
                class="transition-colors hover:text-[var(--accent)]"
                :class="{ 'text-[var(--accent)]': $route.name === 'Chart' }"
              >
                八字命盘
              </router-link>
              <div class="relative group">
                <router-link to="/fortune" class="transition-colors hover:text-[var(--accent)] flex items-center gap-1">
                  运势 <span class="text-[10px] opacity-50">▾</span>
                </router-link>
                <div class="hidden group-hover:block absolute top-full left-1/2 -translate-x-1/2 min-w-[128px] py-1.5 mt-1 bg-[var(--surface-1)]/95 border border-[var(--line-strong)] rounded-[10px] shadow-[0_12px_40px_rgba(0,0,0,0.18)] backdrop-blur-[20px] z-[100]">
                  <router-link to="/fortune" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">今日运势</router-link>
                  <router-link to="/fortune/weekly" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">本周运势</router-link>
                  <router-link to="/fortune/monthly" class="block px-4 py-2 text-xs text-[var(--text-muted)] hover:text-[var(--accent)] hover:bg-[var(--menu-hover)] transition-colors">本月运势</router-link>
                </div>
              </div>
              <button
                @click="goZiwei"
                class="transition-colors hover:text-[var(--accent)] flex items-center gap-1 bg-transparent border-0 p-0 cursor-pointer font-medium text-[13px] text-[var(--text-muted)]"
                :class="{ 'text-[var(--accent)]': $route.name === 'ZiWei' }"
              >
                <span class="text-[10px] opacity-60">✦</span>
                紫微斗数
              </button>
            </template>
          </nav>

          <!-- Right side -->
          <div class="flex items-center gap-3 shrink-0 ml-auto">
            <!-- Five-element switcher -->
            <div class="hidden md:flex items-center gap-1 px-1.5 py-1 rounded-full border border-[var(--line-subtle)] bg-[var(--glass-bg)] backdrop-blur-md">
              <button
                v-for="el in wuxingCycle"
                :key="el.key"
                type="button"
                @click="themeStore.setYongshen(el.key)"
                :class="[
                  'w-6 h-6 flex items-center justify-center rounded-full text-[11px] font-[family-name:var(--font-serif)] transition-all',
                  themeStore.yongshen === el.key
                    ? 'bg-[var(--accent)] text-[var(--bg)] shadow-[0_0_10px_var(--brand-glow)]'
                    : 'text-[var(--text-muted)] hover:text-[var(--accent)]'
                ]"
                :aria-label="`切换至${el.label}主题`"
                :title="`${el.label}行主题`"
              >{{ el.label }}</button>
            </div>
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
