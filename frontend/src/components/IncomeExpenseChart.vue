<template>
  <div class="bg-gray-900 border border-gray-800 rounded-xl p-6">
    <h3 class="text-lg font-semibold mb-4">Ingresos vs gastos</h3>
    <div v-if="!hasData" class="text-gray-500 text-sm">
        No hay suficientes movimientos para mostrar el gráfico
    </div>
    <div v-else class="h-64">
      <Bar :data="chartData" :options="chartOptions" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Bar } from 'vue-chartjs'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale
} from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, BarElement, CategoryScale, LinearScale)

const props = defineProps({
  transactions: {
    type: Array,
    default: () => []
  },
  accountIds: {
    type: Array,
    default: () => []
  }
})

const hasData = computed(() => props.transactions.length > 0)

// Agrupa las transacciones por mes (últimos 6 meses con datos)
const monthlyData = computed(() => {
  const groups = {}

  for (const tx of props.transactions) {
    const date = new Date(tx.timestamp)
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
    const label = date.toLocaleDateString('es-HN', { month: 'short', year: '2-digit' })

    if (!groups[key]) {
      groups[key] = { label, income: 0, expense: 0 }
    }

    const isIncoming = props.accountIds.includes(tx.to_account)
    if (isIncoming) {
      groups[key].income += tx.amount
    } else {
      groups[key].expense += tx.amount
    }
  }

  return Object.entries(groups)
    .sort(([a], [b]) => a.localeCompare(b))
    .slice(-6)
    .map(([, value]) => value)
})

const chartData = computed(() => ({
  labels: monthlyData.value.map(m => m.label),
  datasets: [
    {
      label: 'Ingresos',
      backgroundColor: '#26BFBF',
      data: monthlyData.value.map(m => m.income)
    },
    {
      label: 'Gastos',
      backgroundColor: '#E24B4A',
      data: monthlyData.value.map(m => m.expense)
    }
  ]
}))

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      labels: { color: '#d1d5db' }
    }
  },
  scales: {
    x: {
      ticks: { color: '#9ca3af' },
      grid: { color: 'rgba(255,255,255,0.05)' }
    },
    y: {
      ticks: { color: '#9ca3af' },
      grid: { color: 'rgba(255,255,255,0.05)' }
    }
  }
}
</script>