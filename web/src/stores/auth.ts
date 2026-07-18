import { defineStore } from 'pinia'
import { ref } from 'vue'
import { loginAPI, getProfileAPI } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref<any>(null)

  async function login(username: string, password: string) {
    const data = await loginAPI(username, password)
    token.value = data.token
    localStorage.setItem('token', data.token)
    user.value = data
    return data
  }

  async function fetchProfile() {
    const data = await getProfileAPI()
    user.value = data
    return data
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  function isLoggedIn() {
    return !!token.value
  }

  return { token, user, login, fetchProfile, logout, isLoggedIn }
})
