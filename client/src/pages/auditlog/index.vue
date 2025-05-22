<template>
  <v-container>
    <v-chip color="primary" variant="flat" class="mb-4 px-5">
      <v-icon icon="mdi-post-outline" size="20" start class="mb-0"></v-icon>
      <h3>{{ t('sidebar.auditlog') }}</h3>
    </v-chip>
    <v-card class="elevation-3 px-4">
      <v-data-table-server must-sort v-model:items-per-page="auditlog.limit" :items-per-page-options="auditlog.pages"
        :headers="auditlog.headers()" :items="auditlog.items" :items-length="auditlog.total" :loading="auditlog.loading"
        show-current-page :search="auditlog.search" item-value="log_id" density="compact" class="text-no-wrap"
        fixed-header @update:options="auditlog.fnList">

        <template #[`item.method`]="{ item }">
          <span :class="fnMethodColor(item.method)">{{ item.method }}</span>
        </template>

        <template #[`item.name`]="{ item }">
          <span class="text-capitalize">{{ item.name }}</span>
        </template>

        <template #[`item.updated_at`]="{ item }">
          {{ date.format(item.updated_at, 'fullDate') }}
        </template>

        <template #[`item.data`]="{ item }">
          <code>{{ item.data }}</code>
        </template>

        <template #top>
          <v-row class="justify-space-between ga-2 my-4 mx-0">
            <v-responsive max-width="350">
              <v-text-field v-model="auditlog.filter.search" density="compact" variant="outlined" color="primary"
                :placeholder="t('text.search_auditlog')" append-inner-icon="mdi-magnify" clearable hide-details
                @keyup.enter="auditlog.fnSearch()">
              </v-text-field>
            </v-responsive>
          </v-row>
          <v-divider></v-divider>
        </template>
      </v-data-table-server>
    </v-card>
  </v-container>
</template>

<script setup>
import { useAuditlogStore } from '@/stores/auditlog'
import { useLocale, useDate } from 'vuetify'

const auditlog = useAuditlogStore()
const { t } = useLocale()
const date = useDate()

const fnMethodColor = (method = '') => {
  const colors = { 'POST': 'text-success', 'PUT': 'text-warning', 'DELETE': 'text-error', 'GET': 'text-info' }
  return colors[method]
}
</script>
