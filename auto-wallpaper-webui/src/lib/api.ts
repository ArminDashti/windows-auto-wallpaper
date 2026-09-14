export type Screen = 'lock' | 'home'

export type Wallpaper = {
  id: number
  screen: Screen
  filename: string
  enabled: boolean
  sortOrder: number
  createdAt: string
  fileUrl: string
}

export type Schedule = {
  screen: Screen
  mode: 'daily' | 'hourly'
  intervalHours: number
  startHour: number
  enabled: boolean
}

const TOKEN_KEY = 'autowall_token'

export function getApiBase(): string {
  const base = import.meta.env.VITE_API_BASE_URL
  if (base && base.trim()) {
    return base.replace(/\/$/, '')
  }
  return ''
}

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export function setToken(token: string | null): void {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token)
    else localStorage.removeItem(TOKEN_KEY)
  } catch {
    /* ignore */
  }
}

async function apiFetch<T>(path: string, init: RequestInit = {}, auth = false): Promise<T> {
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth) {
    const token = getToken()
    if (token) headers.set('Authorization', `Bearer ${token}`)
  }
  const res = await fetch(`${getApiBase()}${path}`, { ...init, headers })
  if (!res.ok) {
    let message = `HTTP ${res.status}`
    try {
      const body = (await res.json()) as { error?: string }
      if (body.error) message = body.error
    } catch {
      /* ignore */
    }
    throw new Error(message)
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export async function login(username: string, password: string) {
  const data = await apiFetch<{ token: string; username: string }>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  })
  setToken(data.token)
  return data
}

export async function logout(): Promise<void> {
  try {
    await apiFetch('/api/auth/logout', { method: 'POST' }, true)
  } finally {
    setToken(null)
  }
}

export async function fetchMe(): Promise<{ id: number; username: string }> {
  return apiFetch('/api/me', {}, true)
}

export async function listWallpapers(screen: Screen): Promise<Wallpaper[]> {
  const data = await apiFetch<{ wallpapers: Wallpaper[] }>(
    `/api/screens/${screen}/wallpapers`,
    {},
    true,
  )
  return data.wallpapers ?? []
}

export async function uploadWallpaper(screen: Screen, file: File): Promise<Wallpaper> {
  const body = new FormData()
  body.append('file', file)
  return apiFetch(`/api/screens/${screen}/wallpapers`, { method: 'POST', body }, true)
}

export async function patchWallpaper(
  id: number,
  patch: { enabled?: boolean; sortOrder?: number },
): Promise<Wallpaper> {
  return apiFetch(`/api/wallpapers/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(patch),
  }, true)
}

export async function deleteWallpaper(id: number): Promise<void> {
  await apiFetch(`/api/wallpapers/${id}`, { method: 'DELETE' }, true)
}

export function wallpaperFileUrl(id: number): string {
  return `${getApiBase()}/api/wallpapers/${id}/file`
}

export async function getSchedule(screen: Screen): Promise<Schedule> {
  return apiFetch(`/api/schedules/${screen}`, {}, true)
}

export async function putSchedule(screen: Screen, body: Partial<Schedule>): Promise<Schedule> {
  return apiFetch(`/api/schedules/${screen}`, {
    method: 'PUT',
    body: JSON.stringify(body),
  }, true)
}

/** Authenticated image fetch → blob URL for <img>. */
export async function fetchWallpaperBlobUrl(id: number): Promise<string> {
  const token = getToken()
  const res = await fetch(wallpaperFileUrl(id), {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  })
  if (!res.ok) throw new Error('image load failed')
  const blob = await res.blob()
  return URL.createObjectURL(blob)
}
