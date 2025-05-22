<template>
  <v-container>
    <v-chip color="primary" variant="flat" class="mb-4 px-5">
      <v-icon icon="mdi-key" size="20" start class="mb-1"></v-icon>
      <h3>{{ t('sidebar.client') }}</h3>
    </v-chip>
    <v-card class="elevation-3 px-4">
      <v-data-table-server must-sort v-model:items-per-page="client.limit" :items-per-page-options="client.pages"
        :headers="client.headers()" :items="client.items" :items-length="client.total" :loading="client.loading"
        show-current-page :search="client.search" item-value="client_id" density="compact" class="text-no-wrap"
        fixed-header @update:options="client.fnList">

        <template #[`item.client_id`]="{ item }">
          <span class="d-flex align-center">
            <span class="font-weight-regular">{{ fnFormatSecret(item.client_id) }}</span>
            <v-tooltip :text="t('action.copy')">
              <template #activator="{ props }">
                <v-icon v-bind="props" size="18" class="ml-1"
                  @click="store.fnCopy(item.client_id)">mdi-content-copy</v-icon>
              </template>
            </v-tooltip>
          </span>
        </template>

        <template #[`item.client_secret`]="{ item }">
          <span class="d-flex align-center">
            <span class="font-weight-regular">{{ fnFormatSecret(item.client_secret) }}</span>
            <v-tooltip :text="t('action.copy')">
              <template #activator="{ props }">
                <v-icon v-bind="props" size="18" class="ml-1"
                  @click="store.fnCopy(item.client_secret)">mdi-content-copy</v-icon>
              </template>
            </v-tooltip>
          </span>
        </template>

        <template #[`item.secure_key`]="{ item }">
          <span class="d-flex align-center">
            <span class="font-weight-regular">{{ fnFormatSecret(item.secure_key) }}</span>
            <v-tooltip :text="t('action.copy')">
              <template #activator="{ props }">
                <v-icon v-bind="props" size="18" class="ml-1"
                  @click="store.fnCopy(item.secure_key)">mdi-content-copy</v-icon>
              </template>
            </v-tooltip>
          </span>
        </template>

        <template #[`item.updated_at`]="{ item }">
          {{ date.format(item.updated_at, 'fullDate') }}
        </template>

        <template #[`item.actions`]="{ item }">
          <div class="d-flex flex-row justify-center align-center ga-2 my-2">
            <v-tooltip :text="t('action.detail')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['client_view'])" color="teal" icon="mdi-eye"
                  size="x-small" rounded="0" @click="client.fnDetailItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.edit')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['client_update'])" color="warning"
                  icon="mdi-square-edit-outline" size="x-small" rounded="0" @click="client.fnEditItem(item)"></v-btn>
              </template>
            </v-tooltip>
            <v-tooltip :text="t('action.delete')">
              <template #activator="{ props }">
                <v-btn v-bind="props" v-if="store.hasPermissions(['client_delete'])" color="red"
                  icon="mdi-trash-can-outline" size="x-small" rounded="0" @click="client.fnDelItem(item)"></v-btn>
              </template>
            </v-tooltip>
          </div>
        </template>

        <template #top>
          <v-dialog v-model="client.dialog" max-width="500" transition="false" persistent scrollable no-click-animation>
            <template #activator="{ props }">
              <v-row class="justify-space-between ga-2 my-4 mx-0">
                <v-responsive max-width="350">
                  <v-text-field v-model="client.filter.search" density="compact" variant="outlined" color="primary"
                    :placeholder="t('text.search_name')" append-inner-icon="mdi-magnify" clearable hide-details
                    @keyup.enter="client.fnSearch()">
                  </v-text-field>
                </v-responsive>

                <v-spacer v-if="!display.mobile.value"></v-spacer>

                <v-btn color="success" height="38" @click="client.fnList()">
                  <v-icon start>mdi-refresh</v-icon> {{ t('action.refresh') }}
                </v-btn>
                <v-btn color="primary" height="38" v-bind="props">
                  <v-icon start>mdi-plus-thick</v-icon> {{ t('action.create') }}
                </v-btn>
              </v-row>
            </template>
            <v-form v-model="client.valid" @submit.prevent.stop="client.fnSave()">
              <v-card>
                <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                  <span>
                    {{ client.index == -1 ? t('action.create') : t('action.update') }}
                    {{ t('sidebar.client') }}
                  </span>
                  <v-btn icon="mdi-close" variant="text" :loading="client.loading" @click="client.fnCancel()"></v-btn>
                </v-card-title>
                <v-card-text class="pb-0">
                  <v-col cols="12" class="pa-1">
                    <v-text-field v-model="client.model.name" density="compact" variant="outlined" color="primary"
                      :label="t('text.name')" :rules="[client.rules().required]">
                    </v-text-field>
                  </v-col>
                </v-card-text>
                <v-divider></v-divider>
                <v-card-actions class="mx-5">
                  <v-btn color="primary" variant="flat" type="submit" :loading="client.loading">
                    {{ client.index == -1 ? t('action.create') : t('action.update') }}
                  </v-btn>
                </v-card-actions>
              </v-card>
            </v-form>
          </v-dialog>
          <v-dialog v-model="client.dialogDetail" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.detail') }} {{ t('sidebar.client') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="client.loading" @click="client.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text class="font-weight-bold py-2">
                <v-list density="compact">
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.name') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ client.model.name }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.client_id') }}
                    </v-list-item-subtitle>
                    <span class="d-flex align-center">
                      <span class="font-weight-regular">{{ fnFormatSecret(client.model.client_id) }}</span>
                      <v-tooltip :text="t('action.copy')">
                        <template #activator="{ props }">
                          <v-icon v-bind="props" size="18" class="ml-1"
                            @click="store.fnCopy(client.model.client_id)">mdi-content-copy</v-icon>
                        </template>
                      </v-tooltip>
                    </span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.client_secret') }}
                    </v-list-item-subtitle>
                    <span class="d-flex align-center">
                      <span class="font-weight-regular">{{ fnFormatSecret(client.model.client_secret) }}</span>
                      <v-tooltip :text="t('action.copy')">
                        <template #activator="{ props }">
                          <v-icon v-bind="props" size="18" class="ml-1"
                            @click="store.fnCopy(client.model.client_secret)">mdi-content-copy</v-icon>
                        </template>
                      </v-tooltip>
                    </span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.secure_key') }}
                    </v-list-item-subtitle>
                    <span class="d-flex align-center">
                      <span class="font-weight-regular">{{ fnFormatSecret(client.model.secure_key) }}</span>
                      <v-tooltip :text="t('action.copy')">
                        <template #activator="{ props }">
                          <v-icon v-bind="props" size="18" class="ml-1"
                            @click="store.fnCopy(client.model.secure_key)">mdi-content-copy</v-icon>
                        </template>
                      </v-tooltip>
                    </span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_by') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ client.model.updated_by }}</span>
                  </v-list-item>
                  <v-list-item class="px-1 my-2 bg-background">
                    <v-list-item-subtitle class="font-weight-bold">
                      {{ t('text.updated_at') }}
                    </v-list-item-subtitle>
                    <span class="font-weight-regular">{{ date.format(client.model.updated_at, 'fullDate') }}</span>
                  </v-list-item>
                </v-list>
              </v-card-text>
            </v-card>
          </v-dialog>
          <v-dialog v-model="client.dialogDel" max-width="500" transition="false" persistent scrollable
            no-click-animation>
            <v-card>
              <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                <span> {{ t('action.delete') }} {{ t('sidebar.client') }} </span>
                <v-btn icon="mdi-close" variant="text" :loading="client.loading" @click="client.fnCancel()"></v-btn>
              </v-card-title>
              <v-card-text class="font-weight-bold px-7">
                {{ t('text.delete') }} <span class="text-primary">{{ client.model.name }}</span>
              </v-card-text>
              <v-divider></v-divider>
              <v-card-actions class="mx-5">
                <v-btn color="red" variant="flat" :loading="client.loading" @click="client.fnDel()">
                  {{ t('action.delete') }}
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
import { useClientStore } from '@/stores/client'
import { useLocale, useDate, useDisplay } from 'vuetify'

const store = useAppStore()
const display = useDisplay()
const client = useClientStore()
const { t } = useLocale()
const date = useDate()

const fnFormatSecret = (data = '') => {
  return `${data.substring(0, 6)}......${data.substring(data.length - 6)}`
}
</script>
