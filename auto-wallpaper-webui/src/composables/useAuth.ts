import { computed, ref } from 'vue'
import { fetchMe, getToken, logout as apiLogout, setToken } from '@/lib/api'

const username = ref<string | null>(null)
const ready = ref(false)

export function useAuth() {
  const isAuthenticated = computed(() => Boolean(getToken()) && Boolean(username.value))

  async function refresh(): Promise<boolean> {
    const token = getToken()
    if (!token) {
      username.value = null
      ready.value = true
      return false
    }
    try {
      const me = await fetchMe()
      username.value = me.username
      ready.value = true
      return true
    } catch {
      setToken(null)
      username.value = null
      ready.value = true
      return false
    }
  }

  async function signOut(): Promise<void> {
    await apiLogout()
    username.value = null
  }

  function setUsername(name: string) {
    username.value = name
  }

  return { username, ready, isAuthenticated, refresh, signOut, setUsername }
}
