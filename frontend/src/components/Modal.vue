<template>
  <div v-if="visible" class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-box">
      <div class="modal-header">
        <h3 class="modal-title">{{ title }}</h3>
        <button class="modal-close" @click="$emit('close')">&times;</button>
      </div>
      <div class="modal-body">
        <slot></slot>
      </div>
      <div v-if="$slots.footer" class="modal-footer">
        <slot name="footer"></slot>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  title: { type: String, required: true },
  visible: { type: Boolean, required: true }
})
defineEmits(['close'])
</script>
