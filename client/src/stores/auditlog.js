// Utilities
import { defineStore } from 'pinia'
import * as api from '@/stores/api'
import { useLocale } from 'vuetify'

export const useAuditlogStore = defineStore('auditlog', {
  state: () => ({
    page: 1, limit: 10, total: 0, items: [], index: -1, search: '',
    pages: [10, 25, 50, 100],
    valid: false, loading: false, dialog: false,
    headers: () => {
      const { t } = useLocale()
      const h = [
        { title: t('text.method'), key: 'method', width: '10%' },
        { title: t('text.url'), key: 'url', width: '10%' },
        { title: t('text.name'), key: 'name', width: '10%' },
        { title: t('text.updated_by'), key: 'updated_by', width: '5%' },
        { title: t('text.updated_at'), key: 'updated_at', width: '5%' },
        { title: t('text.data'), key: 'data' },
      ]
      return h
    },
    filter: { search: null },
  }),
  actions: {
    async fnSearch() {
      this.search = Math.random().toString()
    },
    async fnList({ page = 1, itemsPerPage: limit = 10, sortBy = [] }) {
      this.loading = true
      const params = { page, limit, sorts: sortBy.length ? `${sortBy[0].key}.${sortBy[0].order}` : '' }
      for (const key of Object.keys(this.filter)) { if (this.filter[key]) params[key] = this.filter[key] }
      const { data: { items, total } } = await api.get({ url: `/cms/v1/auditlogs`, params })
      this.items = items
      this.total = total
      this.loading = false
    },
  }
})