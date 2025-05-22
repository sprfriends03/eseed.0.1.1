// Plugins
import { createVueI18nAdapter } from 'vuetify/locale/adapters/vue-i18n'
import { createI18n, useI18n } from 'vue-i18n'
import { en, vi } from 'vuetify/locale'

const messages = {
  en: {
    $vuetify: { ...en },
    sidebar: {
      dashboard: 'Dashboard',
      system: 'System',
      role: 'Role',
      user: 'User',
      tenant: 'Tenant',
      client: 'Client Key',
      auditlog: 'Audit Log',
      setting: 'Setting',
    },
    errcode: {
      success: 'Success',
      forbidden: 'Forbidden',
      bad_request: 'Bad Request',
      unauthorized: 'Unauthorized',
      internal_server_error: 'Internal Server Error',
      old_password_incorrect: 'Old Password Incorrect',
      user_or_password_incorrect: 'User Or Password Incorrect',
      client_key_not_found: 'Client Key Not Found',
      client_key_conflict: 'Client Key Conflict',
      tenant_not_found: 'Tenant Not Found',
      tenant_conflict: 'Tenant Conflict',
      role_not_found: 'Role Not Found',
      role_conflict: 'Role Conflict',
      user_not_found: 'User Not Found',
      user_conflict: 'User Conflict',
    },
    permission: {
      system_setting: 'Setting System',
      system_audit_log: 'View Audit Log',

      client_view: 'View Client',
      client_create: 'Add Client',
      client_delete: 'Delete Client',

      role_view: 'View Role',
      role_create: 'Add Role',
      role_update: 'Edit Role',

      user_view: 'View User',
      user_create: 'Add User',
      user_update: 'Edit User',

      tenant_view: 'View Tenant',
      tenant_create: 'Add Tenant',
      tenant_update: 'Edit Tenant',
    },
    data_status: {
      enable: 'Enable',
      disable: 'Disable',
    },
    tab: {
      webhook: 'Webhook',
      email: 'Email',
      openai: 'Open AI',
      tiktok: 'Tiktok',
      shopee: 'Shopee',
      facebook: 'Facebook',
      twitter: 'Twitter',
      threads: 'Threads',
      instagram: 'Instagram',
      youtube: 'Youtube',
      telegram: 'Telegram',
    },
    text: {
      name: 'Name',
      updated_by: 'Updated By',
      updated_at: 'Updated At',
      actions: 'Actions',
      fullNm: 'Full Name',
      username: 'Username',
      password: 'Password',
      old_password: 'Old Password',
      new_password: 'New Password',
      permission: 'Permission',
      phone: 'Phone',
      email: 'Email',
      role: 'Role',
      status: 'Status',
      url: 'URL',
      method: 'Method',
      data: 'Data',
      keycode: 'Keycode',
      address: 'Address',
      scope_types: 'Scope Type',
      client_id: 'Client ID',
      client_secret: 'Client Secret',
      secure_key: 'Secure Key',
      delete: 'Do you want to delete',
      reset: 'Do you want to reset for',
      change_password: 'Change Password',
      copy_password: 'Please copy and backup your password:',
      copied: 'Copied',
      theme: 'Theme',
      all: 'All',
      role: 'Role',
      status: 'Status',
      filter: 'Filter',
      search: 'Search',
      search_name: 'Search name',
      search_user: 'Search name, username',
      search_auditlog: 'Search name, method, url',
    },
    action: {
      login: 'Login',
      logout: 'Logout',
      create: 'Create',
      update: 'Update',
      delete: 'Delete',
      cancel: 'Cancel',
      reset: 'Reset',
      edit: 'Edit',
      submit: 'Submit',
      close: 'Close',
      copy: 'Copy',
      detail: 'Detail',
      refresh: 'Refresh',
    },
    rule: {
      required: 'Field is required',
      email: 'Invalid email',
    },
  },
  vi: {
    $vuetify: { ...vi },
  },
}

const i18n = createI18n({
  legacy: false, // Vuetify does not support the legacy mode of vue-i18n
  locale: 'en',
  fallbackLocale: 'en',
  messages,
})

export const i18nAdapter = createVueI18nAdapter({ i18n, useI18n })

export default i18n