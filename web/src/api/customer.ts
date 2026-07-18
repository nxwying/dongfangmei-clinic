import api from '../api'

export interface CustomerListResponse {
  data: any[]
  total: number
  page: number
  page_size: number
}

export async function getCustomers(params?: any) {
  const res = await api.get<CustomerListResponse>('/customers', { params })
  return res.data
}

export async function createCustomer(data: any) {
  const res = await api.post('/customers', data)
  return res.data
}

export async function getCustomer(id: number) {
  const res = await api.get(`/customers/${id}`)
  return res.data
}

export async function updateCustomer(id: number, data: any) {
  const res = await api.put(`/customers/${id}`, data)
  return res.data
}

export async function deleteCustomer(id: number) {
  const res = await api.delete(`/customers/${id}`)
  return res.data
}

export async function getFollowUps(customerId: number) {
  const res = await api.get(`/customers/${customerId}/followups`)
  return res.data
}

export async function createFollowUp(customerId: number, data: any) {
  const res = await api.post(`/customers/${customerId}/followups`, data)
  return res.data
}

export async function openMembership(customerId: number, data: any) {
  const res = await api.post(`/customers/${customerId}/membership`, data)
  return res.data
}

export async function recharge(customerId: number, data: any) {
  const res = await api.post(`/customers/${customerId}/recharge`, data)
  return res.data
}

export async function getCustomerPackages(customerId: number) {
  const res = await api.get(`/customers/${customerId}/packages`)
  return res.data
}

export async function createPackage(customerId: number, data: any) {
  const res = await api.post(`/customers/${customerId}/packages`, data)
  return res.data
}
