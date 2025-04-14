import {
  createRouter,
  createWebHistory,
  type RouteLocationNormalized,
  type RouteLocationNormalizedLoaded,
} from 'vue-router'
import HomeView from '../pages/HomePage.vue'

import { userStore } from '@/store/userStore'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
      meta: { authRequired: true },
    },
    {
      path: '/auth',
      name: 'auth',
      component: () => import('../pages/AuthPage.vue'),
    },
  ],
})

router.beforeEach((to: RouteLocationNormalized, from: RouteLocationNormalizedLoaded) => {
  // protect auth
  if (to.meta.authRequired && !userStore.isAuthenticated()) return { name: 'auth' }

  return true
})

export default router
