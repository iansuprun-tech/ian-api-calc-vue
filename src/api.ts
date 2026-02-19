import { useAuthStore } from '@/stores/auth'
import router from '@/router'

export async function apiFetch(url: string, options: RequestInit = {}): Promise<Response> {
  const auth = useAuthStore()

  const headers = new Headers(options.headers)
  if (auth.token) {
    headers.set('Authorization', `Bearer ${auth.token}`)
  }
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(url, { ...options, headers })

  if (response.status === 401) {
    auth.logout()
    router.push('/login')
  }

  return response
}
