// Utilities
import { defineStore } from 'pinia'
import * as api from '@/stores/api'

export const useAppStore = defineStore('app', {
  state: () => ({
    logged: !!localStorage.getItem('token'),
    notifications: [],
    session: { user_id: '', tenant_id: '', username: '', phone: '', email: '', name: '', permissions: [] },
    meta: {},
  }),
  actions: {
    hasPermissions(permissions = []) {
      return permissions.some(e => this.session?.permissions?.includes(e))
    },
    async fnNotification({ message, description, type = 'error' }) {
      const noti = { message, description, type, key: Math.random() }
      this.notifications.push(noti)
      setTimeout(() => { this.notifications = this.notifications.filter(e => e.key != noti.key) }, 3000)
    },
    async fnLogin({ username, password, keycode }) {
      const { success, data } = await api.post({ url: `/auth/v1/login`, body: { username, password, keycode } })
      if (success) {
        const token = { username, keycode, access_token: data.access_token, refresh_token: data.refresh_token }
        localStorage.setItem('token', JSON.stringify(token))
      }
      return { success }
    },
    async fnRefreshToken() {
      const token = JSON.parse(localStorage.getItem('token')) || {}
      const body = { username: token.username, keycode: token.keycode, refresh_token: token.refresh_token }
      const { success, data } = await api.post({ url: `/auth/v1/refresh-token`, body })
      if (success) {
        token.access_token = data.access_token
        token.refresh_token = data.refresh_token
        localStorage.setItem('token', JSON.stringify(token))
      }
      return { success, access_token: token.access_token }
    },
    async fnLogout() {
      const { success } = await api.post({ url: `/auth/v1/logout`, body: {} })
      if (success) localStorage.removeItem('token')
      return { success }
    },
    async fnChangePass({ old_password, new_password }) {
      const { success } = await api.post({ url: `/auth/v1/change-password`, body: { old_password, new_password } })
      if (success) localStorage.removeItem('token')
      return { success }
    },
    async fnGetMe() {
      if (this.logged && this.session.user_id) return this.session
      const { success, data } = await api.get({ url: `/auth/v1/me` })
      if (success) this.session = data
      return this.session
    },
    async fnGetMeta() {
      if (this.logged && Object.keys(this.meta).length) return this.meta
      const { success, data } = await api.get({ url: `/rest/v1/metas` })
      if (success) this.meta = data
      return this.meta
    },
    async fnCopy(data) {
      navigator.clipboard.writeText(data)
      this.fnNotification({ message: 'text.copied', type: 'success' })
    }
  }
})