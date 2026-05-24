<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from './stores/auth'

const router = useRouter()
const authStore = useAuthStore()
</script>
<template>
  <div class="app-root">
    <header class="app-header">
      <div class="header-inner">
        <!-- Logo -->
        <router-link to="/" class="logo-link">
          <div class="logo-symbol">
            <div class="logo-ring"></div>
            <span class="logo-icon">☯</span>
          </div>
          <div class="logo-text">
            <span class="logo-title">八字命理</span>
            <span class="logo-sub">BaZi Fortune</span>
          </div>
        </router-link>

        <!-- Nav -->
        <nav class="app-nav">
          <router-link to="/" class="nav-link">
            <span class="nav-dot"></span>
            首页
          </router-link>
          <template v-if="authStore.isLoggedIn()">
            <router-link to="/history" class="nav-link">
              <span class="nav-dot"></span>
              历史
            </router-link>
            <router-link to="/fortune" class="nav-link">
              <span class="nav-dot"></span>
              运势
            </router-link>
            <div class="nav-divider"></div>
            <div class="user-chip">
              <div class="user-avatar">{{ authStore.user?.username?.charAt(0).toUpperCase() }}</div>
              <span class="user-name">{{ authStore.user?.username }}</span>
            </div>
            <button @click="authStore.logout();router.push('/')" class="logout-btn">退出</button>
          </template>
          <template v-else>
            <router-link to="/login" class="nav-link">登录</router-link>
            <router-link to="/register" class="btn-gold nav-register">注册</router-link>
          </template>
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

/* ── Header ── */
.app-header {
  position: sticky;
  top: 0;
  z-index: 50;
  background: rgba(10, 8, 21, 0.85);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-bottom: 1px solid rgba(212, 168, 75, 0.08);
  transition: border-color 0.3s ease;
}

.app-header:hover {
  border-bottom-color: rgba(212, 168, 75, 0.15);
}

.header-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 32px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
}

/* ── Logo ── */
.logo-link {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
}

.logo-symbol {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
}

.logo-ring {
  position: absolute;
  inset: -3px;
  border: 1px solid rgba(212, 168, 75, 0.2);
  border-radius: 50%;
  animation: logo-spin 12s linear infinite;
}

@keyframes logo-spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }

.logo-icon {
  font-size: 1.4rem;
  color: var(--gold);
  text-shadow: 0 0 12px rgba(212, 168, 75, 0.4);
}

.logo-text {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.logo-title {
  font-family: var(--font-serif), serif;
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--text);
  letter-spacing: 3px;
  line-height: 1.1;
}

.logo-sub {
  font-size: 9px;
  letter-spacing: 2px;
  color: rgba(212, 168, 75, 0.4);
  text-transform: uppercase;
}

/* ── Nav ── */
.app-nav {
  display: flex;
  align-items: center;
  gap: 8px;
}

.nav-link {
  position: relative;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 500;
  color: var(--muted);
  text-decoration: none;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.nav-dot {
  display: block;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(212, 168, 75, 0.3);
  transition: all 0.2s ease;
}

.nav-link:hover {
  color: var(--gold);
  background: rgba(212, 168, 75, 0.05);
}

.nav-link:hover .nav-dot {
  background: var(--gold);
  box-shadow: 0 0 6px rgba(212, 168, 75, 0.5);
}

.nav-link.router-link-active {
  color: var(--gold);
}

.nav-link.router-link-active .nav-dot {
  background: var(--gold);
}

.nav-divider {
  width: 1px;
  height: 20px;
  background: rgba(255,255,255,0.08);
  margin: 0 4px;
}

/* ── User chip ── */
.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px 4px 4px;
  border-radius: 20px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(212,168,75,0.08);
}

.user-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--gold), #B8860B);
  color: #0A0815;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-name {
  font-size: 12px;
  color: rgba(212,168,75,0.6);
  font-weight: 500;
}

.logout-btn {
  padding: 6px 12px;
  font-size: 12px;
  color: var(--muted);
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.logout-btn:hover {
  color: var(--crimson);
  background: rgba(196,30,58,0.08);
}

.nav-register {
  padding: 7px 18px;
  font-size: 13px;
  border-radius: 8px;
}

/* ── Main ── */
.app-main {
  flex: 1;
}

/* ── Responsive ── */
@media (max-width: 640px) {
  .header-inner { padding: 0 16px; }
  .logo-text { display: none; }
  .user-name { display: none; }
  .nav-link { padding: 6px 10px; }
}
</style>