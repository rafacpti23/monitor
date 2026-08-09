<template>
  <div class="electric-wrapper" ref="wrapperRef">
    <canvas ref="canvasRef" class="electric-canvas" aria-hidden="true" />
    <div class="electric-content">
      <slot />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'

const props = defineProps({
  // cor principal do raio
  color: { type: String, default: '#00e676' },
  // velocidade da animação (1 = normal)
  speed: { type: Number, default: 1 },
  // caos dos raios (0 = suave, 1 = extremo)
  chaos: { type: Number, default: 0.1 },
  // espessura do raio
  thickness: { type: Number, default: 1.5 },
  // intensidade geral (0 = invisível, 1 = máximo)
  intensity: { type: Number, default: 0.35 },
  // raio de borda
  borderRadius: { type: Number, default: 8 },
  // ativo ou não
  active: { type: Boolean, default: true },
})

const wrapperRef = ref(null)
const canvasRef = ref(null)

let animId = null
let ctx = null
let W = 0, H = 0
let t = 0

// Segmentos de raio ao longo do perímetro
const NUM_BOLTS = 3

function hexToRgb(hex) {
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return { r, g, b }
}

// Versão mais clara do color (mais sutil)
function lightenColor(hex) {
  const rgb = hexToRgb(hex)
  const r = Math.min(255, Math.floor(rgb.r + (255 - rgb.r) * 0.5))
  const g = Math.min(255, Math.floor(rgb.g + (255 - rgb.g) * 0.5))
  const b = Math.min(255, Math.floor(rgb.b + (255 - rgb.b) * 0.5))
  return { r, g, b }
}

// Gera pontos ao longo do retângulo arredondado
function perimeterPoint(frac, w, h, r) {
  // perímetro total aproximado (4 lados + arcos)
  const straight = 2 * (w - 2 * r) + 2 * (h - 2 * r)
  const arcs = 2 * Math.PI * r
  const total = straight + arcs

  const pos = ((frac % 1) + 1) % 1
  let d = pos * total

  // Segmentos: top, arc TR, right, arc BR, bottom, arc BL, left, arc TL
  const segments = [
    { len: w - 2 * r, fn: (s) => ({ x: r + s, y: 0 }) },
    { len: Math.PI * 0.5 * r, fn: (s) => ({ x: w - r + Math.sin(s / r) * r, y: r - Math.cos(s / r) * r }) },
    { len: h - 2 * r, fn: (s) => ({ x: w, y: r + s }) },
    { len: Math.PI * 0.5 * r, fn: (s) => ({ x: w - r + Math.cos(s / r) * r, y: h - r + Math.sin(s / r) * r }) },
    { len: w - 2 * r, fn: (s) => ({ x: w - r - s, y: h }) },
    { len: Math.PI * 0.5 * r, fn: (s) => ({ x: r - Math.sin(s / r) * r, y: h - r + Math.cos(s / r) * r }) },
    { len: h - 2 * r, fn: (s) => ({ x: 0, y: h - r - s }) },
    { len: Math.PI * 0.5 * r, fn: (s) => ({ x: r - Math.cos(s / r) * r, y: r - Math.sin(s / r) * r }) },
  ]

  for (const seg of segments) {
    if (d <= seg.len) return seg.fn(d)
    d -= seg.len
  }
  return { x: r, y: 0 }
}

// Ruído suave usando seno com fases aleatórias
function noise(x, seed) {
  return (
    Math.sin(x * 1.7 + seed * 13.1) * 0.5 +
    Math.sin(x * 3.3 + seed * 7.7) * 0.3 +
    Math.sin(x * 6.1 + seed * 3.3) * 0.2
  )
}

function drawBolt(ctx, startFrac, endFrac, seed, alpha) {
  const rgbFull = hexToRgb(props.color)
  const rgbLight = lightenColor(props.color)
  
  const steps = 28 + Math.floor(props.chaos * 20)
  const points = []

  for (let i = 0; i <= steps; i++) {
    const frac = startFrac + (endFrac - startFrac) * (i / steps)
    const base = perimeterPoint(frac, W, H, props.borderRadius)

    // deslocamento perpendicular ao bordo — ruído caótico
    const n = noise(frac * 8 + t * 0.8, seed) * props.chaos * 22
    // normal aproximada: afastar do centro
    const cx = W / 2, cy = H / 2
    const dx = base.x - cx, dy = base.y - cy
    const len = Math.sqrt(dx * dx + dy * dy) || 1
    points.push({
      x: base.x + (dx / len) * n,
      y: base.y + (dy / len) * n,
    })
  }

  // trilha principal — mais clara e sutil
  ctx.beginPath()
  ctx.moveTo(points[0].x, points[0].y)
  for (let i = 1; i < points.length; i++) {
    const mx = (points[i - 1].x + points[i].x) / 2
    const my = (points[i - 1].y + points[i].y) / 2
    ctx.quadraticCurveTo(points[i - 1].x, points[i - 1].y, mx, my)
  }
  ctx.strokeStyle = `rgba(${rgbLight.r},${rgbLight.g},${rgbLight.b},${alpha * props.intensity * 0.7})`
  ctx.lineWidth = props.thickness
  ctx.shadowColor = `rgba(${rgbLight.r},${rgbLight.g},${rgbLight.b},${alpha * props.intensity * 0.5})`
  ctx.shadowBlur = 6 + props.intensity * 8
  ctx.stroke()

  // brilho fino por cima — bem sutil
  ctx.beginPath()
  ctx.moveTo(points[0].x, points[0].y)
  for (let i = 1; i < points.length; i++) {
    const mx = (points[i - 1].x + points[i].x) / 2
    const my = (points[i - 1].y + points[i].y) / 2
    ctx.quadraticCurveTo(points[i - 1].x, points[i - 1].y, mx, my)
  }
  ctx.strokeStyle = `rgba(255,255,255,${alpha * props.intensity * 0.25})`
  ctx.lineWidth = props.thickness * 0.35
  ctx.shadowBlur = 0
  ctx.stroke()
}

function render() {
  if (!ctx || !props.active || props.intensity === 0) {
    if (ctx) ctx.clearRect(0, 0, W, H)
    animId = requestAnimationFrame(render)
    return
  }

  ctx.clearRect(0, 0, W, H)
  ctx.shadowBlur = 0

  t += 0.012 * props.speed

  for (let i = 0; i < NUM_BOLTS; i++) {
    const seed = i * 137.5
    // cada raio cobre ~30-60% do perímetro e se move
    const span = 0.2 + props.chaos * 0.25 + Math.abs(noise(t * 0.3, seed + 1)) * 0.15
    const offset = ((t * 0.12 * (1 + i * 0.3)) % 1 + seed * 0.37) % 1
    const start = offset
    const end = offset + span

    // pulsação de alpha
    const alpha = 0.5 + 0.5 * Math.sin(t * 2.1 + seed)
    drawBolt(ctx, start, end, seed, alpha * 0.85 + 0.15)
  }

  animId = requestAnimationFrame(render)
}

function resize() {
  if (!wrapperRef.value || !canvasRef.value) return
  const rect = wrapperRef.value.getBoundingClientRect()
  W = rect.width
  H = rect.height
  canvasRef.value.width = W
  canvasRef.value.height = H
}

onMounted(() => {
  ctx = canvasRef.value.getContext('2d')
  resize()
  const ro = new ResizeObserver(resize)
  ro.observe(wrapperRef.value)
  render()
  wrapperRef._ro = ro
})

onBeforeUnmount(() => {
  cancelAnimationFrame(animId)
  if (wrapperRef.value?._ro) wrapperRef.value._ro.disconnect()
})

watch(() => props.intensity, () => {
  // reativa render quando intensidade muda
})
</script>

<style scoped>
.electric-wrapper {
  position: relative;
  display: block;
}

.electric-canvas {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 2;
  border-radius: inherit;
}

.electric-content {
  position: relative;
  z-index: 1;
  height: 100%;
}
</style>
