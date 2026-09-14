<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { useAuth } from '@/composables/useAuth'
import { useTheme } from '@/composables/useTheme'

const { theme, toggleTheme } = useTheme()
const { username, ready, isAuthenticated, refresh, signOut } = useAuth()
const route = useRoute()
const router = useRouter()

onMounted(async () => {
  await refresh()
})

async function onLogout() {
  await signOut()
  await router.push('/login')
}
</script>

<template>
  <div class="flex min-h-dvh flex-col">
    <header
      v-if="route.path !== '/login'"
      class="sticky top-0 z-20 border-b bg-card/90 backdrop-blur"
    >
      <div class="mx-auto flex w-full flex-wrap items-center gap-3 px-4 py-3">
        <RouterLink to="/lock-screen" class="shrink-0 text-lg font-semibold text-primary">
          Auto Wallpaper
        </RouterLink>
        <nav class="flex min-w-0 flex-1 flex-wrap items-center gap-1 sm:gap-2" aria-label="Primary">
          <RouterLink
            to="/lock-screen"
            class="rounded-md px-2 py-1 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            active-class="!text-foreground bg-accent font-medium"
          >
            Lock screen
          </RouterLink>
          <RouterLink
            to="/home-screen"
            class="rounded-md px-2 py-1 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            active-class="!text-foreground bg-accent font-medium"
          >
            Home screen
          </RouterLink>
        </nav>
        <span v-if="ready && isAuthenticated" class="text-sm text-muted-foreground">{{ username }}</span>
        <Button variant="outline" size="sm" class="shrink-0" @click="toggleTheme">
          {{ theme === 'dark' ? 'Light' : 'Dark' }}
        </Button>
        <Button v-if="isAuthenticated" variant="ghost" size="sm" @click="onLogout">Logout</Button>
      </div>
    </header>

    <main class="mx-auto flex w-full max-w-6xl flex-1 flex-col px-4 py-6">
      <RouterView />
    </main>
  </div>
</template>
