import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from '@/lib/api'
import LoginView from '@/views/LoginView.vue'
import LockScreenView from '@/views/LockScreenView.vue'
import HomeScreenView from '@/views/HomeScreenView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
    { path: '/lock-screen', name: 'lock', component: LockScreenView },
    { path: '/home-screen', name: 'home', component: HomeScreenView },
    { path: '/', redirect: '/lock-screen' },
    { path: '/:pathMatch(.*)*', redirect: '/lock-screen' },
  ],
})

router.beforeEach((to) => {
  if (to.meta.public) return true
  if (!getToken()) return { path: '/login', query: { redirect: to.fullPath } }
  return true
})

export default router
