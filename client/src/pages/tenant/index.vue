<template>
  <v-container>
    <v-chip color="primary" variant="flat" class="mb-4 px-5">
      <v-icon icon="mdi-hoop-house" size="20" start class="mb-1"></v-icon>
      <h3>{{ t('sidebar.tenant') }}</h3>
    </v-chip>
    <v-card class="elevation-3 px-4">
      <v-data-table-server must-sort v-model:items-per-page="tenant.limit" :items-per-page-options="tenant.pages"
        :headers="tenant.headers()" :items="tenant.items" :items-length="tenant.total" :loading="tenant.loading"
        show-current-page :search="tenant.search" item-value="tenant_id" density="compact" class="text-no-wrap"
        fixed-header @update:options="tenant.fnList">

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
                <v-btn v-bind="props" v-if="store.hasPermissions(['tenant_view'])" color="teal" icon="mdi-eye"
                  size="x-small" rounded="0" @click="tenant.fnDetailItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.edit')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['tenant_update'])" color="warning"
                  icon="mdi-square-edit-outline" size="x-small" rounded="0" @click="tenant.fnEditItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.delete')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['tenant_delete'])" color="red"
                  icon="mdi-trash-can-outline" size="x-small" rounded="0" @click="tenant.fnDelItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.reset') + ' ' + t('text.password')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['tenant_update'])" color="red" icon="mdi-lock-reset"
                  size="x-small" rounded="0" @click="tenant.fnResetItem(item)"></v-btn>
              </template>
            </v-tooltip>
          </div>
        </template>

        <template #top>
          <v-dialog v-model="tenant.dialog" max-width="500" transition="false" persistent scrollable no-click-animation>
            <template #activator="{ props }">
              <v-row class="justify-space-between ga-2 my-4 mx-0">
                <v-responsive max-width="350">
                  <v-text-field v-model="tenant.filter.search" density="compact" variant="outlined" color="primary"
                    :placeholder="t('text.search_name')" append-inner-icon="mdi-magnify" clearable hide-details
                    @keyup.enter="tenant.fnSearch()">
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
                        <v-select v-model="tenant.filter.data_status" density="compact" variant="outlined"
                          color="primary" :prefix="t('text.status') + ':'" item-title="name" item-value="id"
                          :items="[{ id: null, name: t('text.all') }, ...statuses]" hide-selected
                          @update:modelValue="tenant.fnSearch()">
                        </v-select>
                      </v-col>
                    </v-card-text>
                  </v-card>
                </v-menu>

                <v-btn color="success" height="38" @click="tenant.fnList()">
                  <v-icon start>mdi-refresh</v-icon> {{ t('action.refresh') }}
                </v-btn>
                <v-btn color="primary" height="38" v-bind="props">
                  <v-icon start>mdi-plus-thick</v-icon> {{ t('action.create') }}
                </v-btn>
              </v-row>
            </template>
            <v-form v-model="tenant.valid" @submit.prevent.stop="tenant.fnSave()">
              <v-card>
                <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                  <span>
                    {{ tenant.index == -1 ? t('action.create') : t('action.update') }}
                    {{ t('sidebar.tenant') }}
                  </span>
                  <v-btn icon="mdi-close" variant="text" :loading="tenant.loading" @click="tenant.fnCancel()"></v-btn>
                </v-card-title>
                <v-card-text class="pb-0">
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="tenant.model.name" density="compact" variant="outlined" color="primary"
                      :label="t('text.name')" :rules="[tenant.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="tenant.model.username" density="compact" variant="outlined" color="primary"
                      :label="t('text.username')" :rules="[tenant.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="tenant.model.phone" density="compact" variant="outlined" color="primary"
                      :label="t('text.phone')" :rules="[tenant.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="tenant.model.email" density="compact" variant="outlined" color="primary"
                      :label="t('text.email')" :rules="[tenant.rules().email]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="tenant.model.keycode" density="compact" variant="outlined" color="primary"
                      :label="t('text.keycode')" :rules="[tenant.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="tenant.model.address" density="compact" variant="outlined" color="primary"
                      :label="t('text.address')" :rules="[tenant.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-select v-model="tenant.model.data_status" density="compact" variant="outlined" color="primary"
                      :label="t('text.status')" :rules="[tenant.rules().required]" :items="statuses" chips
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
                </v-card-text>
                <v-divider></v-divider>
                <v-card-actions class="mx-5">
                  <v-btn color="primary" variant="flat" type="submit" :loading="tenant.loading">
                    {{ tenant.index == -1 ? t('action.create') : t('action.update') }}
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-form>
          </v-dialog>
          <v-dialog v-model="tenant.dialogDetail" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.detail') }} {{ t('sidebar.tenant') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="tenant.loading" @click="tenant.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text class="font-weight-bold py-2">
                <v-list density="compact">
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.name') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ tenant.model.name }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.username') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ tenant.model.username }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.keycode') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ tenant.model.keycode }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.status') }}
                    </v-list-item-subtitle>
                    <v-chip :class="fnStatusColor(tenant.model.data_status)" label density="compact">
                      {{ t(`data_status.${tenant.model.data_status}`) }}
                    </v-chip>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_by') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ tenant.model.updated_by }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_at') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ date.format(tenant.model.updated_at, 'fullDate') }}</span>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>
          </v-dialog>
          <v-dialog v-model="tenant.dialogReset" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.reset') }} {{ t('text.password') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="tenant.loading" @click="tenant.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text v-if="tenant.model.password" class="font-weight-bold px-7 pb-5">
                {{ t('text.copy_password') }}
                <span class="text-primary"> {{ fnFormatPassword(tenant.model.password) }} </span>
                <v-tooltip :text="t('action.copy')">
                  <template #activator="{ props }">
                    <v-icon v-bind="props" size="18" class="ml-1"
                      @click="store.fnCopy(tenant.model.password)">mdi-content-copy</v-icon>
                  </template>
                </v-tooltip>
              </v-card-text>
              <v-card-text v-else class="font-weight-bold px-7 pb-5">
                {{ t('text.reset') }} <span class="text-primary">{{ tenant.model.name }}</span>
              </v-card-text>
              <v-divider></v-divider>
              <v-card-actions class="mx-5">
                <v-btn v-if="!tenant.model.password" color="red" variant="flat" :loading="tenant.loading"
                  @click="tenant.fnReset()">
                  {{ t('action.reset') }}
                </v-btn>
              </v-card-actions>
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
import { useTenantStore } from '@/stores/tenant'
import { useLocale, useDate, useDisplay } from 'vuetify'

const store = useAppStore()
const display = useDisplay()
const tenant = useTenantStore()
const { t } = useLocale()
const date = useDate()
const statuses = reactive([])

onMounted(async () => {
  const { data_status } = await store.fnGetMeta()
  statuses.push(...data_status.map(e => ({ id: e, name: t(`data_status.${e}`) })))
})

const fnFormatPassword = (data = '') => {
  return `${data.substring(0, 6)}......${data.substring(data.length - 6)}`
}

const fnStatusColor = (status) => {
  switch (status) {
    case 'enable':
      return 'text-green'
    case 'disable':
      return 'text-red'
  }
}
</script>
