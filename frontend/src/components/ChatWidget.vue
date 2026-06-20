<template>
  <!-- Botón flotante -->
  <button
    v-if="!isOpen"
    @click="isOpen = true"
    class="fixed bottom-6 right-6 bg-blue-600 hover:bg-blue-700 text-white w-14 h-14 rounded-full shadow-lg flex items-center justify-center text-2xl transition-colors z-40"
  >
    💬
  </button>

  <!-- Overlay invisible para detectar click fuera -->
  <div
    v-if="isOpen"
    class="fixed inset-0 z-40"
    @click="isOpen = false"
  ></div>

  <!-- Ventana del chat -->
  <div
    v-if="isOpen"
    @click.stop
    class="fixed bottom-6 right-6 w-96 h-[500px] bg-gray-900 border border-gray-800 rounded-2xl shadow-2xl flex flex-col z-50"
  >
    <!-- Header -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-800">
      <div class="flex items-center gap-2">
        <div class="w-2 h-2 bg-green-400 rounded-full"></div>
        <h3 class="text-sm font-semibold text-white">Asistente bancario</h3>
      </div>
      <button @click="isOpen = false" class="text-gray-500 hover:text-white text-lg">✕</button>
    </div>

    <!-- Mensajes -->
    <div ref="messagesContainer" class="flex-1 overflow-y-auto p-4 space-y-3">
      <div v-if="messages.length === 0" class="text-gray-500 text-sm text-center mt-8">
        Pregúntame sobre tu saldo, historial, o pídeme realizar una transacción.
      </div>
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        :class="msg.role === 'user' ? 'justify-end' : 'justify-start'"
        class="flex"
      >
        <div
          :class="msg.role === 'user' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-200'"
          class="max-w-[80%] rounded-lg px-3 py-2 text-sm whitespace-pre-wrap"
        >
          {{ msg.content }}
        </div>
      </div>
      <div v-if="loading" class="flex justify-start">
        <div class="bg-gray-800 text-gray-400 rounded-lg px-3 py-2 text-sm">
          Escribiendo...
        </div>
      </div>
    </div>

    <!-- Input -->
    <form @submit.prevent="handleSend" class="p-3 border-t border-gray-800 flex gap-2">
      <input
        v-model="inputText"
        type="text"
        placeholder="Escribe tu mensaje..."
        :disabled="loading"
        class="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-blue-500"
      />
      <button
        type="submit"
        :disabled="loading || !inputText.trim()"
        class="bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white px-4 py-2 rounded-lg text-sm transition-colors"
      >
        Enviar
      </button>
    </form>
  </div>
</template>

<script setup>
import { ref, nextTick } from 'vue'
import { useAccountStore } from '../stores/account'

const accountStore = useAccountStore()

const isOpen = ref(false)
const messages = ref([])
const inputText = ref('')
const loading = ref(false)
const messagesContainer = ref(null)

async function scrollToBottom() {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

async function handleSend() {
  const text = inputText.value.trim()
  if (!text) return

  const history = messages.value.map(m => ({ role: m.role, content: m.content }))

  messages.value.push({ role: 'user', content: text })
  inputText.value = ''
  loading.value = true
  scrollToBottom()

  try {
    const reply = await accountStore.sendChatMessage(text, history)
    messages.value.push({ role: 'assistant', content: reply })
    // Refrescar cuentas e historial por si la IA hizo algún movimiento
    await accountStore.fetchAccount()
    await accountStore.fetchHistory(5)
  } catch (e) {
    messages.value.push({
      role: 'assistant',
      content: 'Lo siento, ocurrió un error procesando tu solicitud. Intenta de nuevo.'
    })
  } finally {
    loading.value = false
    scrollToBottom()
  }
}
</script>