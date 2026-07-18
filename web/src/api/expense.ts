import api from '../api'

export interface Expense {
  id: number
  type: string   // 'commission' | 'cost'
  category: string
  amount: number
  note: string
  date: string
  created_at: string
}

export async function getExpenses(params?: any) {
  const res = await api.get<Expense[]>('/expenses', { params })
  return res.data
}

export async function createExpense(data: any) {
  const res = await api.post('/expenses', data)
  return res.data
}

export async function updateExpense(id: number, data: any) {
  const res = await api.put(`/expenses/${id}`, data)
  return res.data
}

export async function deleteExpense(id: number) {
  const res = await api.delete(`/expenses/${id}`)
  return res.data
}

export async function getProfitReport(params?: any) {
  const res = await api.get('/reports/profit', { params })
  return res.data
}
