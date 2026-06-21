import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || null)
  const user = ref(JSON.parse(localStorage.getItem('user') || 'null'))

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(userData, tokenValue) {
    token.value = tokenValue
    user.value = userData
    localStorage.setItem('token', tokenValue)
    localStorage.setItem('user', JSON.stringify(userData))
    axios.defaults.headers.common['Authorization'] = `Bearer ${tokenValue}`
  }

  function clearAuth() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    delete axios.defaults.headers.common['Authorization']
  }

  async function login(email, password, code = '') {
    const response = await axios.post('http://localhost:8080/api/auth/login', {
      email,
      password,
      code
    })

    // Si el backend pide 2FA, no hay token todavía
    if (response.data.requires_2fa) {
      return response.data
    }

    const { token: tokenValue, ...userData } = response.data
    setAuth(userData, tokenValue)
    return response.data
  }

  async function setup2FA() {
    const response = await axios.post('http://localhost:8080/api/2fa/setup')
    return response.data
  }

  async function confirm2FA(code) {
    const response = await axios.post('http://localhost:8080/api/2fa/confirm', { code })
    return response.data
  }

  async function disable2FA() {
    const response = await axios.post('http://localhost:8080/api/2fa/disable')
    return response.data
  }

  async function register(email, password, fullName) {
    const response = await axios.post('http://localhost:8080/api/auth/register', {
      email,
      password,
      full_name: fullName
    })
    return response.data
  }

  async function logout() {
    try {
      await axios.post('http://localhost:8080/api/auth/logout')
    } finally {
      clearAuth()
    }
  }

  // Restaurar token al cargar la app
  if (token.value) {
    axios.defaults.headers.common['Authorization'] = `Bearer ${token.value}`
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    register,
    logout, 
    setup2FA,
    confirm2FA,
    disable2FA
  }
})
