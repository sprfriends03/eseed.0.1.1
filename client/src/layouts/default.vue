<template>
  <v-responsive>
    <v-app :theme="theme">
      <v-app-bar v-if="store.logged" flat color="background-custom">
        <v-app-bar-nav-icon @click.stop="drawer = !drawer"></v-app-bar-nav-icon>
        <v-toolbar-title class="ml-0">
          <v-avatar size="28" class="mb-1 mr-1" tile> <v-img src="@/assets/favicon.ico"></v-img> </v-avatar>
          <span v-if="!display.mobile.value">CMS</span>
        </v-toolbar-title>
        <v-spacer></v-spacer>
        <!-- <v-btn @click="fnLocale()">
          <v-img v-if="locale.current.value == 'vi'" width="24" height="24" src="@/assets/flag-vi.png" />
          <v-img v-else width="24" height="24" src="@/assets/flag-en.png" />
        </v-btn> -->
        <v-tooltip :text="t('text.theme')">
          <template #activator="{ props }">
            <v-btn @click="fnTheme()" v-bind="props">
              <v-icon v-if="theme == 'light'">mdi-weather-sunny</v-icon>
              <v-icon v-else>mdi-weather-night</v-icon>
            </v-btn>
          </template>
        </v-tooltip>
        <v-tooltip :text="t('text.change_password')">
          <template #activator="{ props }">
            <v-btn v-bind="props" @click="changePass.dialog = true"><v-icon>mdi-shield-key-outline</v-icon></v-btn>
            <v-dialog v-model="changePass.dialog" max-width="500" transition="false" persistent scrollable
              no-click-animation>
              <v-form v-model="changePass.valid" @submit.prevent.stop="fnChangePass()">
                <v-card>
                  <v-card-title class="d-flex justify-space-between align-center font-weight-bold bg-primary pl-7">
                    <span>{{ t('text.change_password') }}</span>
                    <v-btn icon="mdi-close" variant="text" :loading="changePass.loading"
                      @click="fnCancelPass()"></v-btn>
                  </v-card-title>
                  <v-card-text class="pb-0">
                    <v-col cols="12" class="pa-1">
                      <v-text-field v-model="changePass.model.old_password" density="compact" variant="outlined"
                        color="primary" :label="t('text.old_password')" :rules="[changePass.rules.required]"
                        :type="changePass.visible ? 'text' : 'password'"
                        :append-inner-icon="changePass.visible ? 'mdi-eye-off' : 'mdi-eye'"
                        @click:append-inner="changePass.visible = !changePass.visible">
                      </v-text-field>
                    </v-col>
                    <v-col cols="12" class="pa-1">
                      <v-text-field v-model="changePass.model.new_password" density="compact" variant="outlined"
                        color="primary" :label="t('text.new_password')" :rules="[changePass.rules.required]"
                        :type="changePass.visible ? 'text' : 'password'"
                        :append-inner-icon="changePass.visible ? 'mdi-eye-off' : 'mdi-eye'"
                        @click:append-inner="changePass.visible = !changePass.visible">
                      </v-text-field>
                    </v-col>
                  </v-card-text>
                  <v-divider></v-divider>
                  <v-card-actions class="mx-5">
                    <v-btn color="primary" variant="flat" type="submit" :loading="changePass.loading">
                      {{ t('action.submit') }}
                    </v-btn>
                  </v-card-actions>
                </v-card>
              </v-form>
            </v-dialog>
          </template>
        </v-tooltip>
        <v-tooltip :text="t('action.logout')">
          <template #activator="{ props }">
            <v-btn v-bind="props" @click="fnLogout()"><v-icon>mdi-logout</v-icon></v-btn>
          </template>
        </v-tooltip>
      </v-app-bar>

      <v-navigation-drawer v-if="store.logged" v-model="drawer" color="background-custom">
        <v-list density="compact">
          <div v-for="(e1, i1) in menus" :key="i1">
            <v-list-group v-if="e1.children.length">
              <template #activator="{ props }">
                <v-list-item class="v-list-item__custom-spacer mx-4 rounded" v-bind="props" :prepend-icon="e1.icon"
                  :title="e1.text" active-class="bg-surface-light elevation-1">
                </v-list-item>
              </template>
              <v-list-item class="v-list-item__custom-spacer mx-4 rounded" v-for="(e2, i2) in e1.children" :key="i2"
                :to="e2.link" :prepend-icon="e2.icon" :title="e2.text" active-class="bg-primary elevation-1">
              </v-list-item>
            </v-list-group>
            <v-list-item class="v-list-item__custom-spacer mx-4 rounded" v-else :to="e1.link" :prepend-icon="e1.icon"
              :title="e1.text" active-class="bg-primary elevation-1">
            </v-list-item>
          </div>
        </v-list>
      </v-navigation-drawer>

      <v-main>
        <router-view />
      </v-main>
    </v-app>

    <v-snackbar v-model="snackbar" v-for="e in store.notifications" :key="e.key" :color="e.type" location="top">
      {{ [t(e.message), e.description].filter(Boolean).join(' - ') }}
    </v-snackbar>
  </v-responsive>
</template>

<script setup>
import { useAppStore } from '@/stores/app'
import { onMounted, ref, computed, reactive } from 'vue'
import { useDisplay, useLocale } from 'vuetify'

const store = useAppStore()
const display = useDisplay()
const locale = useLocale()
const { t } = useLocale()

const drawer = ref(true)
const snackbar = ref(true)
const theme = ref('light')

onMounted(() => {
  // window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', event => {
  //   theme.value = event.matches ? 'dark' : 'light'
  //   localStorage.removeItem('theme')
  // })

  drawer.value = !display.mobile.value
  theme.value = localStorage.getItem('theme') || theme.value
  locale.current.value = localStorage.getItem('locale') || locale.current.value
})

const menus = computed(() => {
  const main = [{ link: '/', text: t('sidebar.dashboard'), icon: 'mdi-view-dashboard', children: [] }]

  const system = { link: '/', text: t('sidebar.system'), icon: 'mdi-cog-outline', children: [] }
  if (store.hasPermissions(['role_view'])) system.children.push({ link: '/role', text: t('sidebar.role'), icon: 'mdi-security' })
  if (store.hasPermissions(['user_view'])) system.children.push({ link: '/user', text: t('sidebar.user'), icon: 'mdi-account-multiple' })
  if (store.hasPermissions(['tenant_view'])) system.children.push({ link: '/tenant', text: t('sidebar.tenant'), icon: 'mdi-hoop-house' })
  if (store.hasPermissions(['client_view'])) system.children.push({ link: '/client', text: t('sidebar.client'), icon: 'mdi-key' })
  if (store.hasPermissions(['system_audit_log'])) system.children.push({ link: '/auditlog', text: t('sidebar.auditlog'), icon: 'mdi-post-outline' })
  if (store.hasPermissions(['system_setting'])) system.children.push({ link: '/setting', text: t('sidebar.setting'), icon: 'mdi-wrench-cog-outline' })
  if (system.children.length) main.push(system)

  return main
})

async function fnLocale() {
  locale.current.value = locale.current.value == 'vi' ? 'en' : 'vi'
  localStorage.setItem('locale', locale.current.value)
}

async function fnTheme() {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  localStorage.setItem('theme', theme.value)
}

async function fnLogout() {
  const { success } = await store.fnLogout()
  if (success) return window.location = `${import.meta.env.VITE_BASE_URL}/login`
}

const changePass = reactive({
  loading: false, valid: false, visible: false, dialog: false,
  model: { old_password: '', new_password: '' },
  rules: { required: v => !!v || t('rule.required') },
})

async function fnCancelPass() {
  changePass.valid = false
  changePass.visible = false
  changePass.dialog = false
  changePass.model.old_password = ''
  changePass.model.new_password = ''
}

async function fnChangePass() {
  if (changePass.valid) {
    changePass.loading = true
    const { old_password, new_password } = changePass.model
    const { success } = await store.fnChangePass({ old_password, new_password })
    if (success) return window.location = `${import.meta.env.VITE_BASE_URL}/login`
    changePass.loading = false
  }
}
</script>

<style scoped>
.v-list-group {
  --prepend-width: 12px;
}

.v-list-item__custom-spacer :deep(.v-list-item__prepend .v-list-item__spacer) {
  width: 10px;
}

.bg-primary {
  --v-theme-overlay-multiplier: 0;
}
</style>