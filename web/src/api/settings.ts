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
