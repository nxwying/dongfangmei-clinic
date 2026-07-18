import api from '../api'

export interface User {
  id: number
  username: string
  real_name: string
  phone: string
  role_id: number
  role: Role
  status: string
  last_login_at: number
}

export interface Role {
  id: number
  name: string
  description: string
  permissions: string[]
}

export async function getUsers() {
  const res = await api.get<User[]>('/users')
  return res.data
}

export async function createUser(data: {
  username: string
  password: string
  real_name: string
  phone?: string
  role_id: number
}) {
  const res = await api.post('/users', data)
  return res.data
}

export async function updateUser(id: number, data: {
  real_name?: string
  phone?: string
  role_id?: number
  password?: string
}) {
  const res = await api.put(`/users/${id}`, data)
  return res.data
}

export async function updateUserStatus(id: number, status: string) {
  const res = await api.put(`/users/${id}/status`, { status })
  return res.data
}

export async function getRoles() {
  const res = await api.get<Role[]>('/roles')
  return res.data
}

export async function createRole(data: {
  name: string
  description?: string
  permissions: string[]
}) {
  const res = await api.post('/roles', data)
  return res.data
}

export async function updateRole(id: number, data: {
  name?: string
  description?: string
  permissions?: string[]
}) {
  const res = await api.put(`/roles/${id}`, data)
  return res.data
}
