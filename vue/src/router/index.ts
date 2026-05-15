import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'Home', component: () => import('../views/HomeView.vue') },
    { path: '/login', name: 'Login', component: () => import('../views/LoginView.vue') },
    { path: '/register', name: 'Register', component: () => import('../views/RegisterView.vue') },
    { path: '/chart/:id', name: 'Chart', component: () => import('../views/ChartView.vue'), meta: { requiresAuth: true } },
    { path: '/fortune', name: 'Fortune', component: () => import('../views/FortuneView.vue'), meta: { requiresAuth: true } },
    { path: '/fortune/weekly', name: 'FortuneWeekly', component: () => import('../views/WeeklyFortuneView.vue'), meta: { requiresAuth: true } },
    { path: '/fortune/monthly', name: 'FortuneMonthly', component: () => import('../views/MonthlyFortuneView.vue'), meta: { requiresAuth: true } },
    { path: '/history', name: 'History', component: () => import('../views/HistoryView.vue'), meta: { requiresAuth: true } },
    { path: '/ziwei/:chartId', name: 'ZiWei', component: () => import('../views/ZiWeiView.vue'), meta: { requiresAuth: true } },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) {
    next({ path: '/login', query: { redirect: to.fullPath } })
  } else {
    next()
  }
})

export default router
