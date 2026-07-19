import api from '../api'

export async function getItems() {
  const res = await api.get('/settings/items')
  return res.data
}

export async function createItem(data: any) {
  const res = await api.post('/settings/items', data)
  return res.data
}

export async function updateItem(id: number, data: any) {
  const res = await api.put(`/settings/items/${id}`, data)
  return res.data
}

export async function getPackageTemplates() {
  const res = await api.get('/settings/package-templates')
  return res.data
}

export async function createPackageTemplate(data: any) {
  const res = await api.post('/settings/package-templates', data)
  return res.data
}

export async function updatePackageTemplate(id: number, data: any) {
  const res = await api.put(`/settings/package-templates/${id}`, data)
  return res.data
}
import api from '../api'

export interface ThemeConfig {
  primary_color: string
  sidebar_bg: string
  sidebar_text: string
  sidebar_active: string
  font_size: string
}

export interface SystemConfig {
  app_name: string
  theme: ThemeConfig
}

export async function getSystemConfig(): Promise<SystemConfig> {
  const res = await api.get('/settings/system-config')
  return res.data
}

export async function updateSystemConfig(data: Partial<SystemConfig>) {
  const res = await api.put('/settings/system-config', data)
  return res.data
}
