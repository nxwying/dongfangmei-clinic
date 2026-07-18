import api from './index'

export interface Document {
  id: number
  doc_type: string
  title: string
  file_name: string
  file_size: number
  file_type: string
  product_name: string
  supplier: string
  serial_no: string
  issue_date: string
  expiry_date: string
  amount: number
  remark: string
  created_by: number
  created_at: string
  status?: string
}

export interface DocumentListParams {
  doc_type?: string
  product_name?: string
  supplier?: string
  keyword?: string
  start_date?: string
  end_date?: string
  expiring_soon?: boolean
  expired?: boolean
  page?: number
  page_size?: number
}

export interface DocumentListResponse {
  total: number
  page: number
  page_size: number
  items: Document[]
}

export async function getDocuments(params: DocumentListParams = {}) {
  const res = await api.get<DocumentListResponse>('/documents', { params })
  return res.data
}

export async function getDocument(id: number) {
  const res = await api.get<Document>(`/documents/${id}`)
  return res.data
}

export async function createDocument(formData: FormData) {
  const res = await api.post<Document>('/documents', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}

export async function deleteDocument(id: number) {
  await api.delete(`/documents/${id}`)
}

export function getDocumentDownloadUrl(id: number, download = false) {
  const token = localStorage.getItem('token') || ''
  let url = `/api/v1/documents/${id}/download`
  const params: string[] = []
  if (download) params.push('download=1')
  if (token) params.push('token=' + token)
  if (params.length) url += '?' + params.join('&')
  return url
}

export async function getExpiringDocuments() {
  const res = await api.get<Document[]>('/documents/expiring')
  return res.data
}
