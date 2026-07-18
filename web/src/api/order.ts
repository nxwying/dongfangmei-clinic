import api from './index'

export async function createOrder(data: any) {
  const res = await api.post('/orders', data)
  return res.data
}

export async function getOrders(params?: any) {
  const res = await api.get('/orders', { params })
  return res.data
}

export async function getOrder(id: number) {
  const res = await api.get(`/orders/${id}`)
  return res.data
}

export async function payOrder(id: number, data: any) {
  const res = await api.post(`/orders/${id}/pay`, data)
  return res.data
}

export async function refundOrder(id: number) {
  const res = await api.post(`/orders/${id}/refund`)
  return res.data
}
