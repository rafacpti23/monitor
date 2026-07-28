import axios from 'axios'
import router from '../router'

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api/v1',
  headers: { 'Content-Type': 'application/json' }
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('p-mon-token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('p-mon-token')
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export function useApi() {
  return api
}

export default api
