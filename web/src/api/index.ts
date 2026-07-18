import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// Request interceptor: attach JWT
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: handle errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
    } else if (error.response?.status === 403) {
      ElMessage.error('权限不足')
    }
    return Promise.reject(error)
  }
)

export default api

// ---- Auth API ----
export interface LoginResponse {
  token: string
  real_name: string
  role_id: number
}

export async function loginAPI(username: string, password: string) {
  const res = await api.post<LoginResponse>('/auth/login', { username, password })
  return res.data
}

export async function getProfileAPI() {
  const res = await api.get('/auth/profile')
  return res.data
}
