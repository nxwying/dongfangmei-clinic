import api from '../api'

export interface LicenseStatus {
  activated: boolean
  machine_code: string
  customer?: string
  expires_at?: string
  is_expired: boolean
  days_left: number
}

export async function getLicenseStatus(): Promise<LicenseStatus> {
  const res = await api.get('/license/status')
  return res.data
}

export async function activateLicense(file: File): Promise<{ message: string; customer: string; expires: string }> {
  const form = new FormData()
  form.append('file', file)
  const res = await api.post('/license/activate', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}
