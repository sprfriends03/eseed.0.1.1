<template>
  <v-container>
    <v-chip color="primary" variant="flat" class="mb-4 px-5">
      <v-icon icon="mdi-wrench-cog-outline" size="20" start class="mb-1"></v-icon>
      <h3>{{ t('sidebar.setting') }}</h3>
    </v-chip>
    <v-card class="elevation-3">
      <v-tabs v-model="tab" slider-color="success">
        <v-tab v-for="e in tabs" variant="flat" color="primary" base-color="background-custom" class="mr-1" :key="e.key"
          :value="e.key" :text="e.text">
          <template #prepend>
            <v-icon v-if="e.icon.startsWith('mdi-')">{{ e.icon }}</v-icon>
            <my-icon v-else :icon="e.icon"></my-icon>
          </template>
        </v-tab>
      </v-tabs>
      <v-divider></v-divider>
      <v-tabs-window v-model="tab">
        <v-tabs-window-item v-for="e in tabs" :key="e.key" :value="e.key">
          <v-card flat>
            <v-card-text>{{ e.text }} coming soon !!!</v-card-text>
          </v-card>
        </v-tabs-window-item>
      </v-tabs-window>
    </v-card>
  </v-container>
</template>

<script setup>
import { useAppStore } from '@/stores/app'
import { useTenantStore } from '@/stores/tenant'
import { reactive, ref } from 'vue';
import { useLocale, useDate } from 'vuetify'

const store = useAppStore()
const tenant = useTenantStore()
const { t } = useLocale()
const date = useDate()

const tabs = reactive({
  webhook: { icon: 'mdi-webhook', key: 'webhook', text: t('tab.webhook') },
  email: { icon: 'mdi-email-outline', key: 'email', text: t('tab.email') },
  openai: { icon: 'my-mdi-openai', key: 'openai', text: t('tab.openai') },
  tiktok: { icon: 'my-mdi-tiktok', key: 'tiktok', text: t('tab.tiktok') },
  shopee: { icon: 'my-mdi-shopee', key: 'shopee', text: t('tab.shopee') },
  facebook: { icon: 'my-mdi-facebook', key: 'facebook', text: t('tab.facebook') },
  twitter: { icon: 'my-mdi-twitter', key: 'twitter', text: t('tab.twitter') },
  threads: { icon: 'my-mdi-threads', key: 'threads', text: t('tab.threads') },
  instagram: { icon: 'my-mdi-instagram', key: 'instagram', text: t('tab.instagram') },
  youtube: { icon: 'my-mdi-youtube', key: 'youtube', text: t('tab.youtube') },
  telegram: { icon: 'my-mdi-telegram', key: 'telegram', text: t('tab.telegram') },
})
const tab = ref(tabs.webhook.key)
</script>
