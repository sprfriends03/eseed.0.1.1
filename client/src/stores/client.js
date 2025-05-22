// Utilities
import { defineStore } from 'pinia'
import * as api from '@/stores/api'
import { useLocale } from 'vuetify'
import { useAppStore } from './app'

const store = useAppStore()

export const useClientStore = defineStore('client', {
  state: () => ({
    page: 1, limit: 10, total: 0, items: [], index: -1, search: '',
    pages: [10, 25, 50, 100],
    valid: false, loading: false, dialog: false, dialogDetail: false, dialogDel: false,
    headers: () => {
      const { t } = useLocale()
      const h = [
        { title: t('text.name'), key: 'name', width: '10%' },
        { title: t('text.client_id'), key: 'client_id', width: '10%' },
        { title: t('text.client_secret'), key: 'client_secret', width: '10%' },
        { title: t('text.secure_key'), key: 'secure_key' },
        { title: t('text.updated_by'), key: 'updated_by', width: '5%' },
        { title: t('text.updated_at'), key: 'updated_at', width: '5%' },
      ]
      if (store.hasPermissions(['client_update', 'client_delete'])) {
        h.push({ title: t('text.actions'), key: 'actions', width: '5%', align: 'center', sortable: false })
      }
      return h
    },
    rules: () => {
      const { t } = useLocale()
      return {
        required: v => !!v || t('rule.required'),
        min1: v => (!!v && v.length > 0) || t('rule.required'),
      }
    },
    filter: { search: null },
    model: { client_id: '', name: '' },
    default: { client_id: '', name: '' },
  }),
  actions: {
    async fnSearch() {
      this.search = Math.random().toString()
    },
    async fnList({ page = 1, itemsPerPage: limit = 10, sortBy = [] }) {
      this.loading = true
      const params = { page, limit, sorts: sortBy.length ? `${sortBy[0].key}.${sortBy[0].order}` : '' }
      for (const key of Object.keys(this.filter)) { if (this.filter[key]) params[key] = this.filter[key] }
      const { data: { items, total } } = await api.get({ url: `/cms/v1/clients`, params })
      this.items = items
      this.total = total
      this.loading = false
    },
    async fnCancel() {
      this.index = -1
      this.dialog = false
      this.dialogDetail = false
      this.dialogDel = false
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
    async fnDelItem(item) {
      this.dialogDel = true
      this.model = Object.assign({}, item)
      this.index = this.items.indexOf(item)
    },
    async fnSave() {
      this.loading = true
      if (this.valid) {
        const { success, data } = this.index == -1
          ? await api.post({ url: `/cms/v1/clients`, body: this.model })
          : await api.put({ url: `/cms/v1/clients/${this.model.client_id}`, body: this.model })
        if (success) {
          if (this.index == -1) this.items.unshift(data)
          else this.items[this.index] = data
          await this.fnCancel()
        }
      }
      this.loading = false
    },
    async fnDel() {
      this.loading = true
      const { success } = await api.del({ url: `/cms/v1/clients/${this.model.client_id}` })
      if (success) {
        await this.fnCancel()
        this.fnSearch()
      }
      this.loading = false
    }
  }
})