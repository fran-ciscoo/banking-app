<template>
  <div class="min-h-screen bg-gray-950 text-white">

    <nav class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="max-w-6xl mx-auto flex items-center justify-between">
        <div class="flex items-center gap-4">
          <router-link to="/dashboard" class="flex items-center gap-2 bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors border border-gray-700">
            ← Volver al inicio
          </router-link>
          <h1 class="text-xl font-bold">Transacciones</h1>
        </div>
      </div>
    </nav>

    <div class="max-w-2xl mx-auto p-6 space-y-6">

      <div class="flex bg-gray-900 rounded-xl p-1 border border-gray-800">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="activeTab === tab.id ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-white'"
          class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          {{ tab.label }}
        </button>
      </div>

      <div v-if="activeTab === 'deposit'" class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
        <h2 class="text-lg font-semibold">Depositar fondos</h2>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Monto (USD)</label>
          <input v-model="amount" type="number" min="1" placeholder="0.00"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Descripción</label>
          <input v-model="description" type="text" placeholder="Descripción opcional"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <button @click="handleDeposit" :disabled="loading"
          class="w-full bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors">
          {{ loading ? 'Procesando...' : 'Depositar' }}
        </button>
      </div>

      <div v-if="activeTab === 'withdraw'" class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
        <h2 class="text-lg font-semibold">Retirar fondos</h2>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Monto (USD)</label>
          <input v-model="amount" type="number" min="1" placeholder="0.00"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Descripción</label>
          <input v-model="description" type="text" placeholder="Descripción opcional"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <button @click="handleWithdraw" :disabled="loading"
          class="w-full bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors">
          {{ loading ? 'Procesando...' : 'Retirar' }}
        </button>
      </div>

      <div v-if="activeTab === 'transfer'" class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
        <h2 class="text-lg font-semibold">Transferir dinero</h2>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Cuenta destino</label>
          <input v-model="toAccount" type="text" placeholder="4001-XXXX-XXXX-XXXX"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Monto (USD)</label>
          <input v-model="amount" type="number" min="1" placeholder="0.00"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Descripción</label>
          <input v-model="description" type="text" placeholder="Descripción opcional"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"/>
        </div>
        <button @click="handleTransfer" :disabled="loading"
          class="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors">
          {{ loading ? 'Procesando...' : 'Transferir' }}
        </button>
      </div>

      <div v-if="error" class="bg-red-900/30 border border-red-500 rounded-lg px-4 py-3 text-red-400 text-sm">
        {{ error }}
      </div>
      <div v-if="success" class="bg-green-900/30 border border-green-500 rounded-lg px-4 py-3 text-green-400 text-sm">
        {{ success }}
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAccountStore } from '../stores/account'

const router = useRouter()
const route = useRoute()

const accountStore = useAccountStore()

const activeTab = ref(route.query.tab || 'deposit')
const amount = ref('')
const description = ref('')
const toAccount = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

const tabs = [
  { id: 'deposit', label: 'Depositar' },
  { id: 'withdraw', label: 'Retirar' },
  { id: 'transfer', label: 'Transferir' }
]

function resetForm() {
  amount.value = ''
  description.value = ''
  toAccount.value = ''
  error.value = ''
  success.value = ''
}

async function handleDeposit() {
  if (!amount.value || amount.value <= 0) {
    error.value = 'Ingresa un monto válido'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
  await accountStore.deposit(parseFloat(amount.value), description.value)
  success.value = `Depósito de $${amount.value} realizado correctamente`
  setTimeout(() => router.push('/dashboard'), 1500)
  resetForm()
  } catch (e) {
    error.value = e.response?.data?.error || 'Error al realizar el depósito'
  } finally {
    loading.value = false
  }
}

async function handleWithdraw() {
  if (!amount.value || amount.value <= 0) {
    error.value = 'Ingresa un monto válido'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
  await accountStore.withdraw(parseFloat(amount.value), description.value)
  success.value = `Retiro de $${amount.value} realizado correctamente`
  setTimeout(() => router.push('/dashboard'), 1500)
  resetForm()
  } catch (e) {
    error.value = e.response?.data?.error || 'Error al realizar el retiro'
  } finally {
    loading.value = false
  }
}

async function handleTransfer() {
  if (!amount.value || amount.value <= 0) {
    error.value = 'Ingresa un monto válido'
    return
  }
  if (!toAccount.value) {
    error.value = 'Ingresa la cuenta destino'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
  await accountStore.transfer(toAccount.value, parseFloat(amount.value), description.value)
  success.value = `Transferencia de $${amount.value} realizada correctamente`
  setTimeout(() => router.push('/dashboard'), 1500)
  resetForm()
  } catch (e) {
     error.value = e.response?.data?.error || 'Error al realizar la transferencia'
  } finally {
    loading.value = false
  }
}
</script>