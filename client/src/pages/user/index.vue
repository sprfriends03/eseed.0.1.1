<template>
  <v-container>
    <v-chip color="primary" variant="flat" class="mb-4 px-5">
      <v-icon icon="mdi-account-multiple" size="20" start class="mb-0"></v-icon>
      <h3>{{ t('sidebar.user') }}</h3>
    </v-chip>
    <v-card class="elevation-3 px-4">
      <v-data-table-server must-sort v-model:items-per-page="user.limit" :items-per-page-options="user.pages"
        :headers="user.headers()" :items="user.items" :items-length="user.total" :loading="user.loading"
        show-current-page :search="user.search" item-value="user_id" density="compact" class="text-no-wrap" fixed-header
        @update:options="user.fnList">

        <template #[`item.role_ids`]="{ item }">
          <v-chip v-for="role_id in item.role_ids.sort()" class="mr-1" density="compact">
            {{roles.find(e => e.role_id == role_id)?.name}}
          </v-chip>
        </template>

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
                <v-btn v-bind="props" v-if="store.hasPermissions(['user_view'])" color="teal" icon="mdi-eye"
                  size="x-small" rounded="0" @click="user.fnDetailItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.edit')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['user_update'])" color="warning"
                  icon="mdi-square-edit-outline" size="x-small" rounded="0" @click="user.fnEditItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.delete')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['user_delete'])" color="red"
                  icon="mdi-trash-can-outline" size="x-small" rounded="0" @click="user.fnDelItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.reset') + ' ' + t('text.password')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['user_update'])" color="red" icon="mdi-lock-reset"
                  size="x-small" rounded="0" @click="user.fnResetItem(item)"></v-btn>
              </template>
            </v-tooltip>
          </div>
        </template>

        <template #top>
          <v-dialog v-model="user.dialog" max-width="500" transition="false" persistent scrollable no-click-animation>
            <template #activator="{ props }">
              <v-row class="justify-space-between ga-2 my-4 mx-0">
                <v-responsive max-width="350">
                  <v-text-field v-model="user.filter.search" density="compact" variant="outlined" color="primary"
                    :placeholder="t('text.search_user')" append-inner-icon="mdi-magnify" clearable hide-details
                    @keyup.enter="user.fnSearch()">
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
                        <v-select v-model="user.filter.role_id" density="compact" variant="outlined" color="primary"
                          :prefix="t('text.role') + ':'" item-title="name" item-value="role_id"
                          :items="[{ role_id: null, name: t('text.all') }, ...roles]" hide-selected
                          @update:modelValue="user.fnSearch()">
                        </v-select>
                      </v-col>
                      <v-col cols="12" class="pa-1">
                        <v-select v-model="user.filter.data_status" density="compact" variant="outlined" color="primary"
                          :prefix="t('text.status') + ':'" item-title="name" item-value="id"
                          :items="[{ id: null, name: t('text.all') }, ...statuses]" hide-selected
                          @update:modelValue="user.fnSearch()">
                        </v-select>
                      </v-col>
                    </v-card-text>
                  </v-card>
                </v-menu>

                <v-btn color="success" height="38" @click="user.fnList()">
                  <v-icon start>mdi-refresh</v-icon> {{ t('action.refresh') }}
                </v-btn>
                <v-btn color="primary" height="38" v-bind="props">
                  <v-icon start>mdi-plus-thick</v-icon> {{ t('action.create') }}
                </v-btn>
              </v-row>
            </template>
            <v-form v-model="user.valid" @submit.prevent.stop="user.fnSave()">
              <v-card>
                <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                  <span>
                    {{ user.index == -1 ? t('action.create') : t('action.update') }} {{ t('sidebar.user') }}
                  </span>
                  <v-btn icon="mdi-close" variant="text" :loading="user.loading" @click="user.fnCancel()"></v-btn>
                </v-card-title>
                <v-card-text class="pb-0">
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="user.model.name" density="compact" variant="outlined" color="primary"
                      :label="t('text.fullNm')" :rules="[user.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="user.model.username" density="compact" variant="outlined" color="primary"
                      :label="t('text.username')" :rules="[user.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="user.model.phone" density="compact" variant="outlined" color="primary"
                      :label="t('text.phone')" :rules="[user.rules().required]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="user.model.email" density="compact" variant="outlined" color="primary"
                      :label="t('text.email')" :rules="[user.rules().email]">
                    </v-text-field>
                  </v-col>
                  <v-col cols="12" class="pa-1">
                    <v-select v-model="user.model.data_status" density="compact" variant="outlined" color="primary"
                      :label="t('text.status')" :rules="[user.rules().required]" :items="statuses" chips
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
                    <v-select v-model="user.model.role_ids" density="compact" variant="outlined" color="primary"
                      :label="t('text.role') + ` (${user.model.role_ids.length}/${roles.length})`"
                      :rules="[user.rules().min1]" :items="roles" hide-selected multiple clearable chips closable-chips
                      item-title="name" item-value="role_id">
                    </v-select>
                  </v-col>
                </v-card-text>
                <v-divider></v-divider>
                <v-card-actions class="mx-5">
                  <v-btn color="primary" variant="flat" type="submit" :loading="user.loading">
                    {{ user.index == -1 ? t('action.create') : t('action.update') }}
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-form>
          </v-dialog>
          <v-dialog v-model="user.dialogDetail" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.detail') }} {{ t('sidebar.user') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="user.loading" @click="user.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text class="font-weight-bold py-2">
                <v-list density="compact">
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.name') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ user.model.name }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.username') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ user.model.username }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.phone') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ user.model.phone }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.email') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ user.model.email }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.role') }}
                    </v-list-item-subtitle>
                    <v-chip v-for="id in user.model.role_ids.sort()" class="mr-1" density="compact">
                      {{roles.find(e => e.role_id == id)?.name}}
                    </v-chip>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.status') }}
                    </v-list-item-subtitle>
                    <v-chip :class="fnStatusColor(user.model.data_status)" label density="compact">
                      {{ t(`data_status.${user.model.data_status}`) }}
                    </v-chip>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_by') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ user.model.updated_by }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_at') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ date.format(user.model.updated_at, 'fullDate') }}</span>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>
          </v-dialog>
          <v-dialog v-model="user.dialogReset" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.reset') }} {{ t('text.password') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="user.loading" @click="user.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text class="font-weight-bold px-7 pb-5">
                {{ t('text.reset') }} <span class="text-primary">{{ user.model.name }}</span>
              </v-card-text>
              <v-divider></v-divider>
              <v-card-actions class="mx-5">
                <v-btn color="red" variant="flat" :loading="user.loading" @click="user.fnReset()">
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
import { useUserStore } from '@/stores/user'
import { onMounted, reactive } from 'vue'
import { useLocale, useDate, useDisplay } from 'vuetify'

const store = useAppStore()
const display = useDisplay()
const user = useUserStore()
const { t } = useLocale()
const date = useDate()
const roles = reactive([])
const statuses = reactive([])

onMounted(async () => {
  const { data_status } = await store.fnGetMeta()
  statuses.push(...data_status.map(e => ({ id: e, name: t(`data_status.${e}`) })))
  const { data: { items } } = await user.fnRoleSelect()
  roles.push(...items)
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
