import api from './index'

export interface Photo {
  id: number
  customer_id: number
  treatment_id?: number
  photo_type: 'before' | 'after'
  body_part: string
  file_name: string
  file_size: number
  created_at: string
}

export async function getPhotos(customerId?: number, photoType?: string) {
  const params: any = {}
  if (customerId) params.customer_id = customerId
  if (photoType) params.photo_type = photoType
  const res = await api.get<Photo[]>('/photos', { params })
  return res.data
}

export async function uploadPhoto(formData: FormData) {
  const res = await api.post<Photo>('/photos', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  return res.data
}

export async function deletePhoto(id: number) {
  await api.delete(`/photos/${id}`)
}

export function getPhotoUrl(id: number) {
  return `/api/v1/photos/${id}/download`
}
