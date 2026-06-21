<template>
  <div class="min-h-screen bg-gray-950 text-white">

    <nav class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="max-w-2xl mx-auto flex items-center gap-4">
        <router-link to="/dashboard" class="flex items-center gap-2 bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors border border-gray-700">
          ← Volver al inicio
        </router-link>
        <h1 class="text-xl font-bold">Seguridad</h1>
      </div>
    </nav>

    <div class="max-w-2xl mx-auto p-6 space-y-6">

      <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
        <h2 class="text-lg font-semibold mb-2">Autenticación de dos factores (2FA)</h2>
        <p class="text-gray-400 text-sm mb-6">
          Agrega una capa extra de seguridad. Además de tu contraseña, necesitarás un código de tu app de autenticación para iniciar sesión.
        </p>

        <!-- Estado: 2FA ya activado -->
        <div v-if="is2FAEnabled && !setupMode" class="space-y-4">
          <div class="bg-green-900/30 border border-green-500 rounded-lg px-4 py-3 text-green-400 text-sm flex items-center gap-2">
            ✓ La autenticación de dos factores está activada en tu cuenta
          </div>
          <button
            @click="handleDisable"
            :disabled="loading"
            class="bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white px-4 py-2 rounded-lg text-sm transition-colors"
          >
            {{ loading ? 'Desactivando...' : 'Desactivar 2FA' }}
          </button>
        </div>

        <!-- Estado: no activado, no en proceso de setup -->
        <div v-else-if="!setupMode">
          <button
            @click="startSetup"
            :disabled="loading"
            class="bg-brand-violet hover:bg-brand-darkest disabled:opacity-50 text-white px-4 py-2 rounded-lg text-sm transition-colors"
          >
            {{ loading ? 'Generando...' : 'Activar 2FA' }}
          </button>
        </div>

        <!-- Proceso de configuración -->
        <div v-else class="space-y-4">
          <div class="bg-gray-800 rounded-lg p-4 text-center">
            <p class="text-sm text-gray-400 mb-3">
              Escanea este código con Google Authenticator, Authy, o tu app de autenticación favorita.
            </p>
            <img :src="qrImage" alt="Código QR 2FA" class="mx-auto rounded-lg" />
            <p class="text-xs text-gray-500 mt-3">
              ¿No puedes escanear? Ingresa este código manualmente:
            </p>
            <code class="text-xs text-brand-blue break-all">{{ secret }}</code>
          </div>

          <div>
            <label class="block text-sm text-gray-400 mb-1">Confirma con el código generado</label>
            <input
              v-model="confirmCode"
              type="text"
              inputmode="numeric"
              maxlength="6"
              placeholder="123456"
              class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue text-center text-2xl tracking-widest"
            />
          </div>

          <div v-if="error" class="bg-red-900/30 border border-red-500 rounded-lg px-4 py-3 text-red-400 text-sm">
            {{ error }}
          </div>
          <div v-if="success" class="bg-green-900/30 border border-green-500 rounded-lg px-4 py-3 text-green-400 text-sm">
            {{ success }}
          </div>

          <div class="flex gap-3">
            <button
              @click="setupMode = false"
              class="flex-1 bg-gray-800 hover:bg-gray-700 text-gray-300 py-3 rounded-lg transition-colors"
            >
              Cancelar
            </button>
            <button
              @click="handleConfirm"
              :disabled="loading || confirmCode.length !== 6"
              class="flex-1 bg-brand-violet hover:bg-brand-darkest disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
            >
              {{ loading ? 'Verificando...' : 'Confirmar y activar' }}
            </button>
          </div>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()

const is2FAEnabled = ref(false)
const setupMode = ref(false)
const qrImage = ref('')
const secret = ref('')
const confirmCode = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

onMounted(() => {
  is2FAEnabled.value = authStore.user?.totp_enabled || false
})

async function startSetup() {
  loading.value = true
  error.value = ''
  try {
    const data = await authStore.setup2FA()
    qrImage.value = data.qr_image
    secret.value = data.secret
    setupMode.value = true
  } catch (e) {
    error.value = e.response?.data?.error || 'Error generando configuración 2FA'
  } finally {
    loading.value = false
  }
}

async function handleConfirm() {
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await authStore.confirm2FA(confirmCode.value)
    success.value = 'Autenticación de dos factores activada correctamente'
    is2FAEnabled.value = true
    setTimeout(() => {
      setupMode.value = false
      confirmCode.value = ''
    }, 1500)
  } catch (e) {
    error.value = e.response?.data?.error || 'Código incorrecto'
  } finally {
    loading.value = false
  }
}

async function handleDisable() {
  loading.value = true
  try {
    await authStore.disable2FA()
    is2FAEnabled.value = false
  } catch (e) {
    error.value = e.response?.data?.error || 'Error desactivando 2FA'
  } finally {
    loading.value = false
  }
}
</script>