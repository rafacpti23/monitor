<template>
  <div class="metric-ring">
    <div class="ring-chart">
      <svg viewBox="0 0 100 100" style="width: 100%; height: 100%; transform: rotate(-90deg);">
        <circle cx="50" cy="50" r="40" fill="none" class="track" :stroke="color" stroke-width="8" stroke-dasharray="251.2" stroke-opacity="0.1" />
        <circle cx="50" cy="50" r="40" fill="none" :stroke="color" stroke-width="8" stroke-dasharray="251.2" :stroke-dashoffset="dashOffset" stroke-linecap="round" style="transition: stroke-dashoffset 0.5s ease" />
      </svg>
      <div class="ring-value">
        <span class="num" :style="{ color }">{{ value }}</span>
        <span class="unit">{{ unit }}</span>
      </div>
    </div>
    <div class="ring-label">{{ label }}</div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: Number, required: true },
  label: { type: String, required: true },
  unit: { type: String, default: '%' },
  color: { type: String, default: 'var(--accent)' }
})

const dashOffset = computed(() => {
  const c = Math.PI * (40 * 2)
  const val = Math.max(0, Math.min(100, props.value))
  return ((100 - val) / 100) * c
})
</script>
