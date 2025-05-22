<template>
  <v-container>
    <v-chip color="primary" variant="flat" class="mb-4 px-5">
      <v-icon icon="mdi-security" size="20" start class="mb-0"></v-icon>
      <h3>{{ t('sidebar.role') }}</h3>
    </v-chip>
    <v-card class="elevation-3 px-4">
      <v-data-table-server must-sort v-model:items-per-page="role.limit" :items-per-page-options="role.pages"
        :headers="role.headers()" :items="role.items" :items-length="role.total" :loading="role.loading"
        show-current-page :search="role.search" item-value="role_id" density="compact" class="text-no-wrap" fixed-header
        @update:options="role.fnList">

        <template #[`item.data_status`]="{ item }">
          <v-chip :class="fnStatusColor(item.data_status)" label density="compact">
            {{ t(`data_status.${item.data_status}`) }}
          </v-chip>
        </template>

        <template #[`item.updated_at`]="{ item }">
          {{ date.format(item.updated_at, 'fullDate') }}
        </template>

        <template #[`item.actions`]="{ item }">
          <div class="d-flex flex-row justify-center align-center ga-2 my-2">
            <v-tooltip :text="t('action.detail')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['role_view'])" color="teal" icon="mdi-eye"
                  size="x-small" rounded="0" @click="role.fnDetailItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.edit')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['role_update'])" color="warning"
                  icon="mdi-square-edit-outline" size="x-small" rounded="0" @click="role.fnEditItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.delete')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['role_delete'])" color="red"
                  icon="mdi-trash-can-outline" size="x-small" rounded="0" @click="role.fnDelItem(item)"></v-btn>
              </template>
            </v-tooltip>
          </div>
        </template>

        <template #top>
          <v-dialog v-model="role.dialog" max-width="500" transition="false" persistent scrollable no-click-animation>
            <template #activator="{ props }">
              <v-row class="justify-space-between ga-2 my-4 mx-0">
                <v-responsive max-width="350">
                  <v-text-field v-model="role.filter.search" density="compact" variant="outlined" color="primary"
                    :placeholder="t('text.search_name')" append-inner-icon="mdi-magnify" clearable hide-details
                    single-line @keyup.enter="role.fnSearch()">
                  </v-text-field>
                </v-responsive>

                <v-spacer v-if="!display.mobile.value"></v-spacer>

                <v-menu :close-on-content-click="false">
                  <template #activator="{ props }">
                    <v-btn color="brown-lighten-2" height="38" v-bind="props">
                      <v-icon start>mdi-tune</v-icon> {{ t('text.filter') }}
                    </v-btn>
                  </template>
                  <v-card min-width="350">
                    <v-card-text class="pb-0">
                      <v-col cols="12" class="pa-1">
                        <v-select v-model="role.filter.data_status" density="compact" variant="outlined" color="primary"
                          :prefix="t('text.status') + ':'" item-title="name" item-value="id"
                          :items="[{ id: null, name: t('text.all') }, ...statuses]" hide-selected
                          @update:modelValue="role.fnSearch()">
                        </v-select>
                      </v-col>
                    </v-card-text>
                  </v-card>
                </v-menu>

                <v-btn color="success" height="38" @click="role.fnList()">
                  <v-icon start>mdi-refresh</v-icon> {{ t('action.refresh') }}
                </v-btn>
                <v-btn color="primary" height="38" v-bind="props">
                  <v-icon start>mdi-plus-thick</v-icon> {{ t('action.create') }}
                </v-btn>
              </v-row>
            </template>
            <v-form v-model="role.valid" @submit.prevent.stop="role.fnSave()">
              <v-card>
                <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                  <span>
                    {{ role.index == -1 ? t('action.create') : t('action.update') }} {{ t('sidebar.role') }}
                  </span>
                  <v-btn icon="mdi-close" variant="text" :loading="role.loading" @click="role.fnCancel()"></v-btn>
                </v-card-title>
                <v-card-text class="pb-0">
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="role.model.name" density="compact" variant="outlined" color="primary"
                      :label="t('text.name')" :rules="[role.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-select v-model="role.model.data_status" density="compact" variant="outlined" color="primary"
                      :label="t('text.status')" :rules="[role.rules().required]" :items="statuses" chips
                      item-title="name" item-value="id" hide-selected>
                      <template v-slot:chip="{ props, item }">
                        <v-chip v-bind="props" :class="fnStatusColor(item.value)" :text="item.title"></v-chip>
                      </template>
                      <template v-slot:item="{ props, item }">
                        <v-list-item v-bind="props" :class="fnStatusColor(item.value)"
                          :title="item.title"></v-list-item>
                      </template>
                    </v-select>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-select v-model="role.model.permissions" density="compact" variant="outlined" color="primary"
                      :label="t('text.permission') + ` (${role.model.permissions.length}/${permissions.length})`"
                      :rules="[role.rules().min1]" :items="permissions" item-title="name" item-value="id" hide-selected
                      chips multiple clearable closable-chips>
                    </v-select>
                  </v-col>
                </v-card-text>
                <v-divider></v-divider>
                <v-card-actions class="mx-5">
                  <v-btn color="primary" variant="flat" type="submit" :loading="role.loading">
                    {{ role.index == -1 ? t('action.create') : t('action.update') }}
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-form>
          </v-dialog>
          <v-dialog v-model="role.dialogDetail" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.detail') }} {{ t('sidebar.role') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="role.loading" @click="role.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text class="font-weight-bold py-2">
                <v-list density="compact">
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.name') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ role.model.name }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.permission') }}
                    </v-list-item-subtitle>
                    <v-chip v-for="e in role.model.permissions.sort()" class="mr-1" density="compact">
                      {{ t(`permission.${e}`) }}
                    </v-chip>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.status') }}
                    </v-list-item-subtitle>
                    <v-chip :class="fnStatusColor(role.model.data_status)" label density="compact">
                      {{ t(`data_status.${role.model.data_status}`) }}
                    </v-chip>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_by') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ role.model.updated_by }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_at') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ date.format(role.model.updated_at, 'fullDate') }}</span>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>
          </v-dialog>
          <v-divider></v-divider>
        </template>

      </v-data-table-server>
    </v-card>
  </v-container>
</template>

<script setup>
import { useAppStore } from '@/stores/app'
import { useRoleStore } from '@/stores/role'
import { onMounted, reactive } from 'vue'
import { useLocale, useDate, useDisplay } from 'vuetify'

const store = useAppStore()
const display = useDisplay()
const role = useRoleStore()
const { t } = useLocale()
const date = useDate()
const statuses = reactive([])
const permissions = reactive([])

onMounted(async () => {
  const { data_status, permission } = await store.fnGetMeta()
  statuses.push(...data_status.map(e => ({ id: e, name: t(`data_status.${e}`) })))
  permission.sort((a, b) => a.localeCompare(b))
  permissions.push(...permission.map(e => ({ id: e, name: t(`permission.${e}`) })))
})

const fnStatusColor = (status) => {
  switch (status) {
    case 'enable':
      return 'text-green'
    case 'disable':
      return 'text-red'
  }
}
</script>