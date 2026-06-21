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
          :class="activeTab === tab.id ? 'bg-brand-violet text-white' : 'text-gray-400 hover:text-white'"
          class="flex-1 py-2 rounded-lg text-sm font-medium transition-colors"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Depositar -->
      <div v-if="activeTab === 'deposit'" class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
        <h2 class="text-lg font-semibold">Depositar fondos</h2>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Cuenta destino</label>
          <select v-model="selectedAccountId"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-brand-blue">
            <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
              {{ accountLabel(acc) }} — {{ acc.id }}
            </option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Monto (USD)</label>
          <input v-model="amount" type="number" min="1" placeholder="0.00"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Descripción</label>
          <input v-model="description" type="text" placeholder="Descripción opcional"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
        </div>
        <button @click="handleDeposit" :disabled="loading"
          class="w-full bg-green-600 hover:bg-green-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors">
          {{ loading ? 'Procesando...' : 'Depositar' }}
        </button>
      </div>

      <!-- Retirar -->
      <div v-if="activeTab === 'withdraw'" class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
        <h2 class="text-lg font-semibold">Retirar fondos</h2>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Cuenta origen</label>
          <select v-model="selectedAccountId"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-brand-blue">
            <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
              {{ accountLabel(acc) }} — {{ formatCurrency(acc.balance) }}
            </option>
          </select>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Monto (USD)</label>
          <input v-model="amount" type="number" min="1" placeholder="0.00"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Descripción</label>
          <input v-model="description" type="text" placeholder="Descripción opcional"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
        </div>
        <button @click="handleWithdraw" :disabled="loading"
          class="w-full bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors">
          {{ loading ? 'Procesando...' : 'Retirar' }}
        </button>
      </div>

      <!-- Transferir -->
      <div v-if="activeTab === 'transfer'" class="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4">
        <h2 class="text-lg font-semibold">Transferir dinero</h2>

        <!-- Modo: propia o terceros -->
        <div class="flex bg-gray-800 rounded-lg p-1">
          <button
            @click="transferMode = 'own'"
            :class="transferMode === 'own' ? 'bg-brand-violet text-white' : 'text-gray-400 hover:text-white'"
            class="flex-1 py-2 rounded-md text-sm font-medium transition-colors"
          >
            Entre mis cuentas
          </button>
          <button
            @click="transferMode = 'third_party'"
            :class="transferMode === 'third_party' ? 'bg-brand-violet text-white' : 'text-gray-400 hover:text-white'"
            class="flex-1 py-2 rounded-md text-sm font-medium transition-colors"
          >
            A un tercero
          </button>
        </div>

        <div>
          <label class="block text-sm text-gray-400 mb-1">Cuenta origen</label>
          <select v-model="fromAccountId"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-brand-blue">
            <option v-for="acc in accounts" :key="acc.id" :value="acc.id">
              {{ accountLabel(acc) }} — {{ formatCurrency(acc.balance) }}
            </option>
          </select>
        </div>

        <div v-if="transferMode === 'own'">
          <label class="block text-sm text-gray-400 mb-1">Cuenta destino</label>
          <select v-model="toOwnAccountId"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-brand-blue">
            <option value="" disabled>Selecciona una cuenta</option>
            <option v-for="acc in accounts.filter(a => a.id !== fromAccountId)" :key="acc.id" :value="acc.id">
              {{ accountLabel(acc) }} — {{ acc.id }}
            </option>
          </select>
        </div>

        <div v-else>
          <label class="block text-sm text-gray-400 mb-1">Número de cuenta destino</label>
          <input v-model="toThirdPartyAccount" type="text" placeholder="4001-XXXX-XXXX-XXXX"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
        </div>

        <div>
          <label class="block text-sm text-gray-400 mb-1">Monto (USD)</label>
          <input v-model="amount" type="number" min="1" placeholder="0.00"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
        </div>
        <div>
          <label class="block text-sm text-gray-400 mb-1">Descripción</label>
          <input v-model="description" type="text" placeholder="Descripción opcional"
            class="w-full bg-gray-800 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-brand-blue"/>
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
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAccountStore } from '../stores/account'

const router = useRouter()
const route = useRoute()

const accountStore = useAccountStore()
const accounts = computed(() => accountStore.accounts)

const activeTab = ref(route.query.tab || 'deposit')
const amount = ref('')
const description = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

const selectedAccountId = ref('')

const transferMode = ref('own')
const fromAccountId = ref('')
const toOwnAccountId = ref('')
const toThirdPartyAccount = ref('')

const tabs = [
  { id: 'deposit', label: 'Depositar' },
  { id: 'withdraw', label: 'Retirar' },
  { id: 'transfer', label: 'Transferir' }
]

onMounted(async () => {
  if (accounts.value.length === 0) {
    await accountStore.fetchAccount()
  }
  if (accounts.value.length > 0) {
    selectedAccountId.value = accounts.value[0].id
    fromAccountId.value = accounts.value[0].id
  }
})

function accountLabel(account) {
  return account.nickname || (account.type === 'checking' ? 'Cuenta corriente' : 'Cuenta de ahorros')
}

function formatCurrency(amount) {
  return new Intl.NumberFormat('es-HN', {
    style: 'currency',
    currency: 'USD'
  }).format(amount || 0)
}

function resetForm() {
  amount.value = ''
  description.value = ''
  toThirdPartyAccount.value = ''
  toOwnAccountId.value = ''
  error.value = ''
  success.value = ''
}

async function handleDeposit() {
  if (!selectedAccountId.value) {
    error.value = 'Selecciona una cuenta'
    return
  }
  if (!amount.value || amount.value <= 0) {
    error.value = 'Ingresa un monto válido'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await accountStore.deposit(selectedAccountId.value, parseFloat(amount.value), description.value)
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
  if (!selectedAccountId.value) {
    error.value = 'Selecciona una cuenta'
    return
  }
  if (!amount.value || amount.value <= 0) {
    error.value = 'Ingresa un monto válido'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await accountStore.withdraw(selectedAccountId.value, parseFloat(amount.value), description.value)
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
  if (!fromAccountId.value) {
    error.value = 'Selecciona la cuenta de origen'
    return
  }

  const destination = transferMode.value === 'own' ? toOwnAccountId.value : toThirdPartyAccount.value

  if (!destination) {
    error.value = 'Selecciona o ingresa la cuenta destino'
    return
  }
  if (!amount.value || amount.value <= 0) {
    error.value = 'Ingresa un monto válido'
    return
  }
  loading.value = true
  error.value = ''
  success.value = ''
  try {
    await accountStore.transfer(fromAccountId.value, destination, parseFloat(amount.value), description.value)
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