<template>
  <div class="chart-container" :style="{ height: height + 'px' }">
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" style="width: 100%; height: 100%; overflow: visible;" v-if="points.length > 0">
      <defs>
        <linearGradient :id="'grad-' + _uid" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" :stop-color="color" stop-opacity="0.3"/>
          <stop offset="100%" :stop-color="color" stop-opacity="0"/>
        </linearGradient>
      </defs>
      <path :d="areaPath" :fill="'url(#grad-' + _uid + ')'" />
      <path :d="linePath" fill="none" :stroke="color" stroke-width="2" vector-effect="non-scaling-stroke" />
    </svg>
    <div v-else style="display: flex; align-items: center; justify-content: center; height: 100%; color: var(--text-muted); font-size: 11px;">Sem dados</div>
  </div>
</template>

<script setup>
import { computed, getCurrentInstance } from 'vue'

const props = defineProps({
  data: { type: Array, default: () => [] },
  color: { type: String, default: 'var(--accent)' },
  height: { type: Number, default: 200 }
})

const _uid = getCurrentInstance()?.uid || Math.random().toString(36).slice(2)

const paths = computed(() => {
  if (!props.data || props.data.length < 2) return { line: '', area: '' }
  
  const width = 100
  const height = 100
  
  // Normalize X and Y
  const xs = props.data.map(d => d.x)
  const ys = props.data.map(d => d.y)
  
  const minX = Math.min(...xs)
  const maxX = Math.max(...xs)
  const rangeX = maxX - minX || 1
  
  const minY = Math.min(...ys, 0)
  const maxY = Math.max(...ys, 1)
  const rangeY = maxY - minY || 1
  
  const pts = props.data.map(d => {
    const nx = ((d.x - minX) / rangeX) * width
    const ny = height - ((d.y - minY) / rangeY) * height
    return { x: nx, y: ny }
  })
  
  const line = `M ${pts.map(p => `${p.x} ${p.y}`).join(' L ')}`
  const area = `${line} L ${pts[pts.length - 1].x} ${height} L ${pts[0].x} ${height} Z`
  
  return { line, area }
})

const linePath = computed(() => paths.value.line)
const areaPath = computed(() => paths.value.area)
</script>
