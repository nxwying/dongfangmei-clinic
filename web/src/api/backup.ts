import api from '../api'

export async function listBackups() {
  const res = await api.get('/backup/list')
  return res.data
}

export async function createBackup() {
  const res = await api.post('/backup/create')
  return res.data
}

export async function deleteBackup(id: number) {
  const res = await api.delete(`/backup/${id}`)
  return res.data
}

export async function downloadBackupFile(id: number, filename: string) {
  try {
    const res = await api.get(`/backup/${id}/download`, { responseType: 'blob', timeout: 60000 })
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (e: any) {
    throw e
  }
}

export async function uploadToCloud(id: number) {
  const res = await api.post(`/backup/${id}/upload-cloud`)
  return res.data
}

export async function getBackupSettings() {
  const res = await api.get('/backup/settings')
  return res.data
}

export async function saveBackupSettings(settings: any) {
  const res = await api.put('/backup/settings', settings)
  return res.data
}

export async function exportBackup() {
  try {
    const res = await api.get('/backup/export', { responseType: 'blob', timeout: 30000 })
    const disposition = res.headers['content-disposition'] || ''
    const match = disposition.match(/filename="?(.+?)"?$/)
    const filename = match ? match[1] : `clinic_backup_${new Date().toISOString().slice(0, 10)}.sql`
    const url = window.URL.createObjectURL(new Blob([res.data]))
    const link = document.createElement('a')
    link.href = url
    link.setAttribute('download', filename)
    document.body.appendChild(link)
    link.click()
    link.remove()
    window.URL.revokeObjectURL(url)
  } catch (e: any) {
    throw e
  }
}

export async function importBackup(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  const res = await api.post('/backup/import', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    timeout: 120000,
  })
  return res.data
}

export async function resetSystem() {
  const res = await api.post("/backup/reset")
  return res.data
}
