<template>
  <div class="min-h-screen bg-gray-950 text-white">

    <!-- Navbar -->
    <nav class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="max-w-6xl mx-auto flex items-center gap-4">
        <router-link to="/dashboard" class="flex items-center gap-2 bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors border border-gray-700">
            ← Volver al inicio
          </router-link>
        <h1 class="text-xl font-bold">Historial de transacciones</h1>
      </div>
    </nav>

    <div class="max-w-3xl mx-auto p-6 space-y-4">

      <!-- Loading -->
      <div v-if="loading" class="text-center text-gray-500 py-12">
        Cargando historial...
      </div>

      <!-- Vacío -->
      <div v-else-if="transactions.length === 0" class="text-center text-gray-500 py-12">
        No hay transacciones registradas
      </div>

      <!-- Lista -->
      <div v-else class="bg-gray-900 border border-gray-800 rounded-xl divide-y divide-gray-800">
        <div
          v-for="tx in transactions"
          :key="tx.id"
          class="flex items-center justify-between px-6 py-4"
        >
          <div class="flex items-center gap-4">
            <!-- Icono tipo -->
            <div :class="iconClass(tx)" class="w-10 h-10 rounded-full flex items-center justify-center text-lg font-bold">
              {{ typeIcon(tx) }}
            </div>
            <div>
              <p class="text-sm font-medium text-white">{{ tx.description || typeLabel(tx.type) }}</p>
              <p class="text-xs text-gray-500">{{ formatDate(tx.timestamp) }}</p>
              <p class="text-xs text-gray-600">{{ tx.from_account }} → {{ tx.to_account }}</p>
            </div>
          </div>
          <p :class="amountClass(tx)" class="font-semibold">
            {{ amountSign(tx) }}{{ formatCurrency(tx.amount) }}
          </p>
        </div>
      </div>

      <!-- Cargar más -->
      <button
        v-if="transactions.length >= limit"
        @click="loadMore"
        class="w-full bg-gray-900 border border-gray-800 hover:border-gray-600 text-gray-400 py-3 rounded-xl text-sm transition-colors"
      >
        Cargar más
      </button>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAccountStore } from '../stores/account'

const accountStore = useAccountStore()
const loading = ref(true)
const limit = ref(20)

const transactions = computed(() => accountStore.transactions)
const accounts = computed(() => accountStore.accounts)

onMounted(async () => {
  await accountStore.fetchAccount()
  await accountStore.fetchHistory(limit.value)
  loading.value = false
})

async function loadMore() {
  limit.value += 20
  await accountStore.fetchHistory(limit.value)
}

function typeIcon(tx) {
  if (tx.type === 'deposit') return '↓'
  if (tx.type === 'withdrawal') return '↑'
  return '→'
}

function iconClass(tx) {
  if (tx.type === 'deposit') return 'bg-green-900/50 text-green-400'
  if (tx.type === 'withdrawal') return 'bg-red-900/50 text-red-400'
  return 'bg-blue-900/50 text-blue-400'
}

function typeLabel(type) {
  const labels = {
    deposit: 'Depósito',
    withdrawal: 'Retiro',
    transfer: 'Transferencia',
    internal_transfer: 'Transferencia interna'
  }
  return labels[type] || type
}

function amountSign(tx) {
  const myAccountIds = accounts.value.map(a => a.id)
  return myAccountIds.includes(tx.to_account) ? '+' : '-'
}

function amountClass(tx) {
  const myAccountIds = accounts.value.map(a => a.id)
  return myAccountIds.includes(tx.to_account) ? 'text-green-400' : 'text-red-400'
}

function formatCurrency(amount) {
  return new Intl.NumberFormat('es-HN', {
    style: 'currency',
    currency: 'USD'
  }).format(amount || 0)
}

function formatDate(date) {
  return new Date(date).toLocaleDateString('es-HN', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>