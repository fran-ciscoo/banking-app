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

  async function login(email, password) {
    const response = await axios.post('http://localhost:8080/api/auth/login', {
      email,
      password
    })
    const { token: tokenValue, ...userData } = response.data
    setAuth(userData, tokenValue)
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
    logout
  }
})
