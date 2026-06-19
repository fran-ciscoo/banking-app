import { defineStore } from 'pinia'
import { ref } from 'vue'
import axios from 'axios'

export const useAccountStore = defineStore('account', () => {
  const accounts = ref([])
  const transactions = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function fetchAccount() {
    loading.value = true
    error.value = null
    try {
      const response = await axios.get('http://localhost:8080/api/account')
      accounts.value = response.data.accounts || []
    } catch (e) {
      error.value = 'Error obteniendo cuenta'
    } finally {
      loading.value = false
    }
  }

  async function fetchHistory(limit = 20) {
    loading.value = true
    try {
      const response = await axios.get(`http://localhost:8080/api/transactions/history?limit=${limit}`)
      transactions.value = response.data.transactions || []
    } catch (e) {
      error.value = 'Error obteniendo historial'
    } finally {
      loading.value = false
    }
  }

  async function deposit(amount, description) {
    const response = await axios.post('http://localhost:8080/api/transactions/deposit', {
      amount,
      description
    })
    await fetchAccount()
    return response.data
  }

  async function withdraw(amount, description) {
    const response = await axios.post('http://localhost:8080/api/transactions/withdraw', {
      amount,
      description
    })
    await fetchAccount()
    return response.data
  }

  async function transfer(toAccountId, amount, description) {
    const response = await axios.post('http://localhost:8080/api/transactions/transfer', {
      to_account_id: toAccountId,
      amount,
      description
    })
    await fetchAccount()
    return response.data
  }

  async function createAccount(type) {
    const response = await axios.post('http://localhost:8080/api/account/create', { type })
    await fetchAccount()
    return response.data
  }

  async function updateNickname(accountId, nickname) {
  const response = await axios.put(`http://localhost:8080/api/account/${accountId}/nickname`, { nickname })
  await fetchAccount()
  return response.data
}

async function deleteAccount(accountId) {
  const response = await axios.delete(`http://localhost:8080/api/account/${accountId}`)
  await fetchAccount()
  return response.data
}

  return {
  accounts,
  transactions,
  loading,
  error,
  fetchAccount,
  fetchHistory,
  deposit,
  withdraw,
  transfer,
  createAccount,
  updateNickname,
  deleteAccount
}
})