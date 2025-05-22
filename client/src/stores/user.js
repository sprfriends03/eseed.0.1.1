// Utilities
import { defineStore } from 'pinia'
import * as api from '@/stores/api'
import { useLocale } from 'vuetify'
import { useAppStore } from './app'

const store = useAppStore()

export const useUserStore = defineStore('user', {
  state: () => ({
    page: 1, limit: 10, total: 0, items: [], index: -1, search: '',
    pages: [10, 25, 50, 100],
    valid: false, loading: false, dialog: false, dialogDetail: false, dialogReset: false,
    headers: () => {
      const { t } = useLocale()
      const h = [
        { title: t('text.fullNm'), key: 'name', width: '10%' },
        { title: t('text.username'), key: 'username', width: '10%' },
        { title: t('text.role'), key: 'role_ids' },
        { title: t('text.status'), key: 'data_status', width: '5%' },
        { title: t('text.updated_by'), key: 'updated_by', width: '5%' },
        { title: t('text.updated_at'), key: 'updated_at', width: '5%' },
      ]
      if (store.hasPermissions(['user_update', 'user_delete'])) {
        h.push({ title: t('text.actions'), key: 'actions', width: '5%', align: 'center', sortable: false })
      }
      return h
    },
    rules: () => {
      const { t } = useLocale()
      return {
        required: v => !!v || t('rule.required'),
        min1: v => (!!v && v.length > 0) || t('rule.required'),
        email: v => (!!v && /.+@.+/.test(v)) || t('rule.email'),
      }
    },
    filter: { search: null, data_status: null, role_id: null },
    model: { user_id: '', username: '', name: '', phone: '', email: '', data_status: null, role_ids: [] },
    default: { user_id: '', username: '', name: '', phone: '', email: '', data_status: null, role_ids: [] },
  }),
  actions: {
    async fnSearch() {
      this.search = Math.random().toString()
    },
    async fnRoleSelect() {
      return api.get({ url: `/cms/v1/users/roles` })
    },
    async fnList({ page = 1, itemsPerPage: limit = 10, sortBy = [] }) {
      this.loading = true
      const params = { page, limit, sorts: sortBy.length ? `${sortBy[0].key}.${sortBy[0].order}` : '' }
      for (const key of Object.keys(this.filter)) { if (this.filter[key]) params[key] = this.filter[key] }
      const { data: { items, total } } = await api.get({ url: `/cms/v1/users`, params })
      this.items = items
      this.total = total
      this.loading = false
    },
    async fnCancel() {
      this.index = -1
      this.dialog = false
      this.dialogDetail = false
      this.dialogReset = false
      this.model = Object.assign({}, this.default)
    },
    async fnDetailItem(item) {
      this.dialogDetail = true
      this.model = Object.assign({}, item)
      this.index = this.items.indexOf(item)
    },
    async fnEditItem(item) {
      this.dialog = true
      this.model = Object.assign({}, item)
      this.index = this.items.indexOf(item)
    },
    async fnResetItem(item) {
      this.dialogReset = true
      this.model = Object.assign({}, item)
      this.index = this.items.indexOf(item)
    },
    async fnSave() {
      this.loading = true
      if (this.valid) {
        const { success, data } = this.index == -1
          ? await api.post({ url: `/cms/v1/users`, body: this.model })
          : await api.put({ url: `/cms/v1/users/${this.model.user_id}`, body: this.model })
        if (success) {
          if (this.index == -1) this.items.unshift(data)
          else this.items[this.index] = data
          await this.fnCancel()
        }
      }
      this.loading = false
    },
    async fnReset() {
      this.loading = true
      const { success } = await api.post({ url: `/cms/v1/users/${this.model.user_id}/reset-password`, body: {} })
      if (success) await this.fnCancel()
      this.loading = false
    }
  }
})