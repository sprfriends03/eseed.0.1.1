// Utilities
import axios, { HttpStatusCode } from 'axios'
import { useAppStore } from '@/stores/app'

const api = axios.create({ baseURL: import.meta.env.VITE_API })

api.interceptors.request.use(config => {
  const token = JSON.parse(localStorage.getItem('token')) || {}
  config.headers['Authorization'] = `Bearer ${token.access_token}`
  config.headers['Content-Type'] = config.url.includes('storage') ? 'multipart/form-data' : 'application/json'
  return config
})

api.interceptors.response.use((res) => {
  if (['get'].includes(res.config.method) && Array.isArray(res.data)) {
    res.data = { items: res.data, total: (res.headers['x-pagination-total'] || res.data.length) }
  }
  if (['post', 'put', 'delete'].includes(res.config.method)) {
    useAppStore().fnNotification({ message: `errcode.success`, type: 'success' })
  }
  return Promise.resolve(res)
}, async ({ response }) => {
  const { status, data, config } = response

  if (status == HttpStatusCode.Unauthorized) {
    const { success, access_token } = await useAppStore().fnRefreshToken()
    if (!success) {
      localStorage.removeItem('token')
      return window.location = `${import.meta.env.VITE_BASE_URL}/login`
    }
    config.headers['Authorization'] = `Bearer ${access_token}`
    return api(config)
  }

  useAppStore().fnNotification({ message: `errcode.${data.error}`, description: data.error_description, type: 'error' })
  return Promise.reject(data)
})

export const upload = async ({ url, files }) => {
  try {
    const body = new FormData()
    for (const f of files) body.append('files', f)
    const { data } = await api.post(url, body)
    return { success: true, data }
  } catch (err) { return { success: false, data: err } }
}

export const post = async ({ url, body }) => {
  try {
    const { data } = await api.post(url, body)
    return { success: true, data }
  } catch (err) { return { success: false, data: err } }
}

export const put = async ({ url, body }) => {
  try {
    const { data } = await api.put(url, body)
    return { success: true, data }
  } catch (err) { return { success: false, data: err } }
}

export const get = async ({ url, params }) => {
  try {
    const { data } = await api.get(url, { params })
    return { success: true, data }
  } catch (err) { return { success: false, data: err } }
}

export const del = async ({ url }) => {
  try {
    const { data } = await api.delete(url)
    return { success: true, data }
  } catch (err) { return { success: false, data: err } }
}