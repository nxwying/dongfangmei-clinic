import api from '../api'

export async function getDailySales(date?: string) {
  const res = await api.get('/reports/daily-sales', { params: { date } })
  return res.data
}

export async function getMemberSummary() {
  const res = await api.get('/reports/member-summary')
  return res.data
}
