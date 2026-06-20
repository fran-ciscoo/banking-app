<template>
  <div class="min-h-screen bg-gray-950 text-white">

    <!-- Navbar -->
    <nav class="bg-gray-900 border-b border-gray-800 px-6 py-4">
      <div class="max-w-6xl mx-auto flex items-center justify-between">
        <h1 class="text-xl font-bold text-white">BankingApp</h1>
        <div class="flex items-center gap-4">
          <span class="text-gray-400 text-sm">{{ authStore.user?.full_name }}</span>
          <button
            @click="handleLogout"
            class="bg-gray-800 hover:bg-gray-700 text-gray-300 px-4 py-2 rounded-lg text-sm transition-colors"
          >
            Cerrar sesión
          </button>
        </div>
      </div>
    </nav>

    <div class="max-w-6xl mx-auto p-6 space-y-6">

      <!-- Balance card -->
      <div class="bg-gradient-to-r from-blue-600 to-blue-800 rounded-2xl p-6">
        <p class="text-blue-200 text-sm mb-1">Saldo disponible</p>
        <h2 class="text-4xl font-bold text-white">
          {{ loading ? '...' : formatCurrency(totalBalance) }}
        </h2>
        <p class="text-blue-200 text-sm mt-2">{{ accounts.length }} cuenta(s)</p>
      </div>

      <!-- Acciones rápidas -->
      <div class="grid grid-cols-3 gap-4">
        <router-link
  to="/transactions?tab=deposit"
  class="bg-gray-900 border border-gray-800 rounded-xl p-4 text-center hover:border-blue-500 transition-colors"
>
  <div class="text-2xl mb-2">+</div>
  <p class="text-sm text-gray-300">Depositar</p>
</router-link>
<router-link
  to="/transactions?tab=withdraw"
  class="bg-gray-900 border border-gray-800 rounded-xl p-4 text-center hover:border-blue-500 transition-colors"
>
  <div class="text-2xl mb-2">-</div>
  <p class="text-sm text-gray-300">Retirar</p>
</router-link>
<router-link
  to="/transactions?tab=transfer"
  class="bg-gray-900 border border-gray-800 rounded-xl p-4 text-center hover:border-blue-500 transition-colors"
>
  <div class="text-2xl mb-2">→</div>
  <p class="text-sm text-gray-300">Transferir</p>
</router-link>
      </div>

      <!-- Cuentas -->
      <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold">Mis cuentas</h3>
          <button
            @click="showCreateAccount = true"
            class="bg-blue-600 hover:bg-blue-700 text-white text-sm px-4 py-2 rounded-lg transition-colors"
          >
            + Nueva cuenta
          </button>
        </div>
        <div v-if="loading" class="text-gray-500 text-sm">Cargando...</div>
        <div v-else-if="accounts.length === 0" class="text-gray-500 text-sm">
          No tienes cuentas registradas
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="account in accounts"
            :key="account.id"
            class="flex items-center justify-between bg-gray-800 rounded-lg px-4 py-3"
          >
            <div class="flex-1">
              <div v-if="editingAccountId === account.id" class="flex items-center gap-2">
                <input
                  v-model="editingNickname"
                  @keyup.enter="saveNickname(account.id)"
                  type="text"
                  class="bg-gray-700 border border-gray-600 rounded px-2 py-1 text-sm text-white focus:outline-none focus:border-blue-500"
                  placeholder="Nombre de la cuenta"
                />
                <button @click="saveNickname(account.id)" class="text-green-400 hover:text-green-300 text-sm">✓</button>
                <button @click="cancelEditNickname" class="text-gray-500 hover:text-gray-400 text-sm">✕</button>
              </div>
              <div v-else class="flex items-center gap-2">
                <p class="text-sm font-medium text-white">
                  {{ account.nickname || (account.type === 'checking' ? 'Cuenta corriente' : 'Cuenta de ahorros') }}
                </p>
                <button @click="startEditNickname(account)" class="text-gray-500 hover:text-gray-300 text-xs">✎</button>
                <button @click="handleDeleteAccount(account)" class="text-gray-500 hover:text-red-400 text-xs">🗑</button>
              </div>
              <p class="text-xs text-gray-400">{{ account.id }}</p>
            </div>
            <p class="font-semibold text-white">{{ formatCurrency(account.balance) }}</p>
          </div>
        </div>
      </div>

      <!-- Modal crear cuenta -->
      <div v-if="showCreateAccount" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" @click.self="closeCreateAccount">
        <div class="bg-gray-900 border border-gray-800 rounded-2xl p-6 w-full max-w-sm mx-4">
          <h3 class="text-lg font-semibold mb-4">Crear nueva cuenta</h3>
          <div class="space-y-3 mb-6">
            <button
              @click="selectedAccountType = 'checking'"
              :class="selectedAccountType === 'checking' ? 'border-blue-500 bg-blue-600/10' : 'border-gray-700'"
              class="w-full border rounded-lg px-4 py-3 text-left transition-colors"
            >
              <p class="font-medium text-white">Cuenta corriente</p>
              <p class="text-xs text-gray-400">Para uso diario y transacciones</p>
            </button>
            <button
              @click="selectedAccountType = 'savings'"
              :class="selectedAccountType === 'savings' ? 'border-blue-500 bg-blue-600/10' : 'border-gray-700'"
              class="w-full border rounded-lg px-4 py-3 text-left transition-colors"
            >
              <p class="font-medium text-white">Cuenta de ahorros</p>
              <p class="text-xs text-gray-400">Para guardar y ahorrar dinero</p>
            </button>
          </div>
          <div v-if="createError" class="bg-red-900/30 border border-red-500 rounded-lg px-4 py-3 text-red-400 text-sm mb-4">
            {{ createError }}
          </div>
          <div class="flex gap-3">
            <button @click="closeCreateAccount" class="flex-1 bg-gray-800 hover:bg-gray-700 text-gray-300 py-3 rounded-lg transition-colors">
              Cancelar
            </button>
            <button
              @click="handleCreateAccount"
              :disabled="!selectedAccountType || creating"
              class="flex-1 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
            >
              {{ creating ? 'Creando...' : 'Crear cuenta' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Modal eliminar cuenta -->
      <div v-if="accountToDelete" class="fixed inset-0 bg-black/60 flex items-center justify-center z-50" @click.self="accountToDelete = null">
        <div class="bg-gray-900 border border-gray-800 rounded-2xl p-6 w-full max-w-sm mx-4">
          <h3 class="text-lg font-semibold mb-2">Eliminar cuenta</h3>
          <p class="text-gray-400 text-sm mb-4">
            {{ accountToDelete.nickname || 'Esta cuenta' }} — {{ accountToDelete.id }}
          </p>

          <div v-if="accountToDelete.balance !== 0" class="bg-amber-900/30 border border-amber-500 rounded-lg px-4 py-3 text-amber-400 text-sm mb-4">
            Esta cuenta tiene saldo disponible. Debes acercarte al banco para cerrarla.
          </div>
          <div v-else class="bg-gray-800 rounded-lg px-4 py-3 text-gray-300 text-sm mb-4">
            ¿Estás seguro de eliminar esta cuenta? Esta acción no se puede deshacer.
          </div>

          <div v-if="deleteError" class="bg-red-900/30 border border-red-500 rounded-lg px-4 py-3 text-red-400 text-sm mb-4">
            {{ deleteError }}
          </div>

          <div class="flex gap-3">
            <button @click="accountToDelete = null" class="flex-1 bg-gray-800 hover:bg-gray-700 text-gray-300 py-3 rounded-lg transition-colors">
              Cancelar
            </button>
            <button
              v-if="accountToDelete.balance === 0"
              @click="confirmDeleteAccount"
              :disabled="deleting"
              class="flex-1 bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-semibold py-3 rounded-lg transition-colors"
            >
              {{ deleting ? 'Eliminando...' : 'Eliminar' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Transacciones recientes -->
      <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold">Transacciones recientes</h3>
          <router-link to="/history" class="bg-blue-600 hover:bg-blue-700 text-white text-sm px-4 py-2 rounded-lg transition-colors">
            Ver todas
          </router-link>
        </div>
        <div v-if="transactions.length === 0" class="text-gray-500 text-sm">
          No hay transacciones recientes
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="tx in transactions.slice(0, 5)"
            :key="tx.id"
            class="flex items-center justify-between py-2 border-b border-gray-800 last:border-0"
          >
            <div>
              <p class="text-sm text-white">{{ tx.description || tx.type }}</p>
              <p class="text-xs text-gray-500">{{ formatDate(tx.timestamp) }}</p>
            </div>
            <p :class="isIncoming(tx) ? 'text-green-400' : 'text-red-400'" class="font-medium text-sm">
              {{ isIncoming(tx) ? '+' : '-' }}{{ formatCurrency(tx.amount) }}
            </p>
          </div>
        </div>
      </div>

    </div>
    <ChatWidget />
  </div>
</template>

<script setup>
import ChatWidget from '../components/ChatWidget.vue'
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useAccountStore } from '../stores/account'

const router = useRouter()
const authStore = useAuthStore()
const accountStore = useAccountStore()

const loading = ref(true)
const accounts = computed(() => accountStore.accounts)
const transactions = computed(() => accountStore.transactions)
const totalBalance = computed(() => accounts.value.reduce((sum, a) => sum + a.balance, 0))

const showCreateAccount = ref(false)
const selectedAccountType = ref('')
const creating = ref(false)
const createError = ref('')

const editingAccountId = ref(null)
const editingNickname = ref('')

const accountToDelete = ref(null)
const deleting = ref(false)
const deleteError = ref('')

function closeCreateAccount() {
  showCreateAccount.value = false
  selectedAccountType.value = ''
  createError.value = ''
}

async function handleCreateAccount() {
  creating.value = true
  createError.value = ''
  try {
    await accountStore.createAccount(selectedAccountType.value)
    closeCreateAccount()
  } catch (e) {
    createError.value = e.response?.data?.error || 'Error creando cuenta'
  } finally {
    creating.value = false
  }
}

function startEditNickname(account) {
  editingAccountId.value = account.id
  editingNickname.value = account.nickname || ''
}

function cancelEditNickname() {
  editingAccountId.value = null
  editingNickname.value = ''
}

async function saveNickname(accountId) {
  if (!editingNickname.value.trim()) {
    cancelEditNickname()
    return
  }
  try {
    await accountStore.updateNickname(accountId, editingNickname.value.trim())
    cancelEditNickname()
  } catch (e) {
    console.error('Error actualizando nombre', e)
  }
}

function handleDeleteAccount(account) {
  accountToDelete.value = account
  deleteError.value = ''
}

async function confirmDeleteAccount() {
  deleting.value = true
  deleteError.value = ''
  try {
    await accountStore.deleteAccount(accountToDelete.value.id)
    accountToDelete.value = null
  } catch (e) {
    deleteError.value = e.response?.data?.error || 'Error eliminando la cuenta'
  } finally {
    deleting.value = false
  }
}

onMounted(async () => {
  await accountStore.fetchAccount()
  await accountStore.fetchHistory(5)
  loading.value = false
})

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
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
    year: 'numeric'
  })
}

function isIncoming(tx) {
  const myAccountIds = accounts.value.map(a => a.id)
  return myAccountIds.includes(tx.to_account)
}
</script>