<template>
  <v-container class="align-center justify-center fill-height">
    <v-col cols :sm="4">
      <v-img class="mx-auto my-6" max-width="250" src="@/assets/logo.svg"></v-img>
      <v-form v-model="valid" @submit.prevent.stop="fnSubmit">
        <v-card class="elevation-5 pa-4">
          <v-card-text>
            <v-col cols="12" class="pa-1">
              <v-text-field v-model="model.username" density="comfortable" variant="outlined" color="primary"
                :label="t('text.username')" :rules="[rules.required]" type="text" prepend-inner-icon="mdi-account">
              </v-text-field>
            </v-col>
            <v-col cols="12" class="pa-1">
              <v-text-field v-model="model.password" density="comfortable" variant="outlined" color="primary"
                :label="t('text.password')" :rules="[rules.required]" :type="visible ? 'text' : 'password'"
                prepend-inner-icon="mdi-lock" :append-inner-icon="visible ? 'mdi-eye-off' : 'mdi-eye'"
                @click:append-inner="visible = !visible">
              </v-text-field>
            </v-col>
            <v-col cols="12" class="pa-1">
              <v-text-field v-model="model.keycode" density="comfortable" variant="outlined" color="primary"
                :label="t('text.keycode')" type="text" prepend-inner-icon="mdi-hoop-house">
              </v-text-field>
            </v-col>
          </v-card-text>
          <v-card-actions>
            <v-btn :loading="loading" type="submit" color="primary" block variant="flat" size="x-large">
              {{ t('action.login') }}
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-form>
    </v-col>
  </v-container>
</template>

<script setup>
import { useAppStore } from '@/stores/app'
import { reactive, ref } from 'vue'
import { useLocale } from 'vuetify'

const store = useAppStore()
const { t } = useLocale()
const valid = ref(false)
const loading = ref(false)
const visible = ref(false)

const rules = reactive({ required: v => !!v || t('rule.required') })
const model = reactive({ username: '', password: '', keycode: '' })

async function fnSubmit() {
  if (valid.value) {
    loading.value = true
    const { username, password, keycode } = model
    const { success } = await store.fnLogin({ username, password, keycode })
    if (success) return window.location = `${import.meta.env.VITE_BASE_URL}`
    loading.value = false
  }
}
</script>
