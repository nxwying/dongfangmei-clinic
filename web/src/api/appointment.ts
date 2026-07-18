import api from '.'

export async function getAppointments(params?: any) {
  const res = await api.get('/appointments', { params })
  return res.data
}

export async function createAppointment(data: any) {
  const res = await api.post('/appointments', data)
  return res.data
}

export async function checkIn(id: number) {
  const res = await api.put(`/appointments/${id}/checkin`)
  return res.data
}

export async function completeAppt(id: number) {
  const res = await api.put(`/appointments/${id}/complete`)
  return res.data
}

export async function cancelAppt(id: number) {
  const res = await api.put(`/appointments/${id}/cancel`)
  return res.data
}
