import {
  createRouter,
  createWebHistory,
  type NavigationGuardNext,
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
    // {
    //   path: '/auth',
    //   name: 'auth',
    //   component: () => import('../pages/AuthPage.vue'),
    // },
    // TODO: add authenticated guard
  ],
})

router.beforeEach(
  (to: RouteLocationNormalized, from: RouteLocationNormalizedLoaded, next: NavigationGuardNext) => {
    // protect auth
    if (to.meta.authRequired && !userStore.isAuthenticated()) return { name: 'auth' }

    next()
  },
)

export default router
