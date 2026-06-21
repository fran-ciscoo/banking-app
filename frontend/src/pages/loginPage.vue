<template>
  <div class="min-h-screen bg-gray-950 flex items-center justify-center p-4">
    <div class="bg-gray-900 rounded-2xl p-8 w-full max-w-md border border-gray-800">

      <!-- Logo -->
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-white">BankingApp</h1>
        <p class="text-gray-400 mt-2">
          {{ needsCode ? 'Verifica tu identidad' : 'Inicia sesión en tu cuenta' }}
        </p>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <template v-if="!needsCode">
          <div>
            <label class="block text-sm text-gray-400 mb-1">Email</label>
            <input
              v-model="email"
              type="email"
              placeholder="tu@email.com"
              class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
              required
            />
          </div>

          <div>
            <label class="block text-sm text-gray-400 mb-1">Contraseña</label>
            <input
              v-model="password"
              type="password"
              placeholder="••••••••"
              class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
              required
            />
          </div>
        </template>

        <template v-else>
          <div>
            <label class="block text-sm text-gray-400 mb-1">Código de verificación</label>
            <input
              v-model="code"
              type="text"
              inputmode="numeric"
              maxlength="6"
              placeholder="123456"
              class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500 text-center text-2xl tracking-widest"
              autofocus
            />
            <p class="text-xs text-gray-500 mt-2">Abre tu app de autenticación e ingresa el código de 6 dígitos.</p>
          </div>
        </template>

        <!-- Error -->
        <div v-if="error" class="bg-red-900/30 border border-red-500 rounded-lg px-4 py-3 text-red-400 text-sm">
          {{ error }}
        </div>

        <button
          type="submit"
          :disabled="loading"
          class="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
        >
          {{ loading ? 'Verificando...' : (needsCode ? 'Verificar' : 'Iniciar sesión') }}
        </button>

        <button
          v-if="needsCode"
          type="button"
          @click="needsCode = false; code = ''"
          class="w-full text-gray-500 hover:text-gray-300 text-sm"
        >
          ← Volver
        </button>
      </form>

      <p v-if="!needsCode" class="text-center text-gray-500 mt-6 text-sm">
        ¿No tienes cuenta?
        <router-link to="/register" class="text-blue-400 hover:text-blue-300">Regístrate</router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const code = ref('')
const needsCode = ref(false)
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    const result = await authStore.login(email.value, password.value, code.value)

    if (result.requires_2fa) {
      needsCode.value = true
      return
    }

    router.push('/dashboard')
  } catch (e) {
    error.value = e.response?.data?.error || 'Error al iniciar sesión'
  } finally {
    loading.value = false
  }
}
</script>