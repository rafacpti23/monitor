import { ref, onUnmounted } from 'vue'

const BASE_WS = import.meta.env.VITE_WS_URL || `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}/ws`

export function useSocket() {
  const connected = ref(false)
  const listeners = new Map()
  let ws = null
  let reconnectTimer = null
  let attempt = 0
  let disposed = false

  function connect() {
    if (disposed) return
    const token = localStorage.getItem('p-mon-token')
    if (!token) return

    const url = `${BASE_WS}?token=${encodeURIComponent(token)}`
    ws = new WebSocket(url)

    ws.onopen = () => {
      connected.value = true
      attempt = 0
    }

    ws.onclose = () => {
      connected.value = false
      if (!disposed) scheduleReconnect()
    }

    ws.onerror = () => {
      ws?.close()
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        const handlers = listeners.get(msg.type)
        if (handlers) {
          handlers.forEach((fn) => fn(msg.payload))
        }
      } catch {
        /* ignore non-JSON frames */
      }
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) clearTimeout(reconnectTimer)
    const delay = Math.min(1000 * Math.pow(2, attempt), 30000)
    attempt++
    reconnectTimer = setTimeout(connect, delay)
  }

  function on(type, handler) {
    if (!listeners.has(type)) {
      listeners.set(type, new Set())
    }
    listeners.get(type).add(handler)
  }

  function off(type, handler) {
    listeners.get(type)?.delete(handler)
  }

  function send(type, payload) {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, payload }))
    }
  }

  function disconnect() {
    disposed = true
    if (reconnectTimer) clearTimeout(reconnectTimer)
    ws?.close()
  }

  connect()
  onUnmounted(disconnect)

  return { connected, on, off, send, disconnect, connect }
}
