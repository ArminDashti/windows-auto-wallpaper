<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { login } from '@/lib/api'
import { useAuth } from '@/composables/useAuth'

const router = useRouter()
const { refresh, setUsername } = useAuth()

const username = ref('armin')
const password = ref('')
const error = ref('')
const submitting = ref(false)

onMounted(async () => {
  if (await refresh()) {
    await router.replace('/lock-screen')
  }
})

async function onSubmit() {
  error.value = ''
  submitting.value = true
  try {
    const data = await login(username.value.trim(), password.value)
    setUsername(data.username)
    await router.replace('/lock-screen')
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Login failed'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="mx-auto flex w-full max-w-sm flex-col gap-6 py-10">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">Auto Wallpaper</h1>
      <p class="mt-1 text-sm text-muted-foreground">Sign in to manage lock and home wallpapers.</p>
    </div>

    <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium" for="user">Username</label>
        <input
          id="user"
          v-model="username"
          required
          autocomplete="username"
          class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </div>
      <div class="flex flex-col gap-1.5">
        <label class="text-sm font-medium" for="pass">Password</label>
        <input
          id="pass"
          v-model="password"
          type="password"
          required
          autocomplete="current-password"
          class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        />
      </div>
      <p v-if="error" class="text-sm text-destructive" role="alert">{{ error }}</p>
      <Button type="submit" :disabled="submitting">
        {{ submitting ? 'Signing in…' : 'Sign in' }}
      </Button>
    </form>
  </div>
</template>
