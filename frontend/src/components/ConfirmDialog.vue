<template>
  <Modal :title="title" :visible="visible" @close="$emit('close')">
    <p style="margin-bottom: 24px; font-size: 13.5px; color: var(--text-secondary);">
      {{ message }}
    </p>
    <template #footer>
      <button class="btn btn-secondary" @click="$emit('close')">Cancelar</button>
      <button class="btn btn-danger" @click="$emit('confirm')">{{ confirmText }}</button>
    </template>
  </Modal>
</template>

<script setup>
import Modal from './Modal.vue'

defineProps({
  visible: { type: Boolean, required: true },
  title: { type: String, default: 'Confirmar' },
  message: { type: String, required: true },
  confirmText: { type: String, default: 'Confirmar' }
})
defineEmits(['close', 'confirm'])
</script>
