import api from './index'

export interface LicenseStatus {
  activated: boolean
  machine_code: string
  customer?: string
  expires_at?: string
  features?: string[]
  is_expired?: boolean
  days_left?: number
}

export async function getLicenseStatus() {
  const res = await api.get<LicenseStatus>('/license/status')
  return res.data
}

export async function activateLicense(file: File) {
  const fd = new FormData()
  fd.append('license', file)
  const res = await api.post('/license/activate', fd, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}
