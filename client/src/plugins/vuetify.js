/**
 * plugins/vuetify.js
 *
 * Framework documentation: https://vuetifyjs.com`
 */

// Styles
import '@mdi/font/css/materialdesignicons.css'
import 'vuetify/styles'

// Colors
import colors from 'vuetify/util/colors'

// Composables
import { createVuetify } from 'vuetify'
import { i18nAdapter } from './i18n'

// https://vuetifyjs.com/en/introduction/why-vuetify/#feature-guides
export default createVuetify({
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          background: colors.grey.lighten4,
          'background-custom': '#FFFFFF',
        },
      },
      dark: {
        colors: {
          background: colors.blueGrey.darken4,
          'background-custom': colors.blueGrey.darken4,
        },
      },
    },
  },
  locale: {
    adapter: i18nAdapter
  }
})
