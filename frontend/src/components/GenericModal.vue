<script setup lang="ts">
defineProps({
  isOpen: {
    type: Boolean,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  message: {
    type: String,
    required: true,
  },
  type: {
    type: String as () => 'alert' | 'confirm',
    default: 'alert',
  },
  confirmText: {
    type: String,
    default: 'Confirmar',
  },
  cancelText: {
    type: String,
    default: 'Cancelar',
  },
  confirmVariant: {
    type: String as () => 'primary' | 'danger',
    default: 'primary',
  },
});

const emit = defineEmits(['close', 'confirm']);
</script>

<template>
  <div v-if="isOpen" class="fixed inset-0 bg-black/60 z-50 flex items-center justify-center" @click.self="emit('close')">
    <div class="bg-surface rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
      <h3 class="text-lg font-bold mb-4">{{ title }}</h3>
      <p class="text-sm text-subtle mb-6 whitespace-pre-wrap">{{ message }}</p>

      <div class="flex justify-end gap-3 mt-6">
        <button v-if="type === 'confirm'" @click="emit('close')" class="px-4 py-2 text-xs text-gray-400 hover:text-white rounded-md hover:bg-overlay">
          {{ cancelText }}
        </button>
        <button @click="emit('confirm')" class="px-4 py-2 text-xs font-bold text-white rounded-md transition-colors" :class="{
          'bg-accent hover:bg-accent/80': confirmVariant === 'primary',
          'bg-red-500 hover:bg-red-600': confirmVariant === 'danger',
        }">
          {{ type === 'alert' ? 'OK' : confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>