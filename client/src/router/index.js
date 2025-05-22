/**
 * router/index.ts
 *
 * Automatic routes for `./src/pages/*.vue`
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router/auto'
import { setupLayouts } from 'virtual:generated-layouts'
// import { routes } from 'vue-router/auto-routes'
import { useAppStore } from '@/stores/app'

const routes = [
  { path: '', component: () => import('@/pages/dashboard'), meta: { verify: true } },
  { path: '/role', component: () => import('@/pages/role'), meta: { verify: true, permissions: ['role_view'] } },
  { path: '/user', component: () => import('@/pages/user'), meta: { verify: true, permissions: ['user_view'] } },
  { path: '/client', component: () => import('@/pages/client'), meta: { verify: true, permissions: ['client_view'] } },
  { path: '/tenant', component: () => import('@/pages/tenant'), meta: { verify: true, permissions: ['tenant_view'] } },
  { path: '/auditlog', component: () => import('@/pages/auditlog'), meta: { verify: true, permissions: ['system_audit_log'] } },
  { path: '/setting', component: () => import('@/pages/setting'), meta: { verify: true, permissions: ['system_setting'] } },
  { path: '/login', component: () => import('@/pages/login'), meta: { verify: false } },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.VITE_BASE_URL),
  routes: setupLayouts(routes),
})

// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (!localStorage.getItem('vuetify:dynamic-reload')) {
      console.log('Reloading page to fix dynamic import error')
      localStorage.setItem('vuetify:dynamic-reload', 'true')
      location.assign(to.fullPath)
    } else {
      console.error('Dynamic import error, reloading page did not fix it', err)
    }
  } else {
    console.error(err)
  }
})

router.isReady().then(() => {
  localStorage.removeItem('vuetify:dynamic-reload')
})

router.beforeEach(async (to, from, next) => {
  const { verify, permissions } = to.meta
  const store = useAppStore()

  if (verify) {
    if (!store.logged) return next('/login')

    await store.fnGetMe()
    if (permissions && !store.hasPermissions(permissions)) {
      return next('')
    }
  } else if (store.logged) {
    return next('')
  }

  return next()
})

export default router
