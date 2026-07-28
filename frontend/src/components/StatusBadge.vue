<template>
  <span class="badge" :class="badgeClass">
    {{ displayStr }}
  </span>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  status: { type: String, required: true }
})

const map = {
  online: { class: 'badge-success', str: '● online' },
  up: { class: 'badge-success', str: '● online' },
  offline: { class: 'badge-danger', str: '○ offline' },
  down: { class: 'badge-danger', str: '○ down' },
  warning: { class: 'badge-warning', str: '▲ aviso' },
  unknown: { class: 'badge-neutral', str: '— desc' }
}

const badgeClass = computed(() => map[props.status.toLowerCase()]?.class || 'badge-neutral')
const displayStr = computed(() => map[props.status.toLowerCase()]?.str || props.status)
</script>
