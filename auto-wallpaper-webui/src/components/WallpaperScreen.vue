<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Button } from '@/components/ui/button'
import {
  type Schedule,
  type Screen,
  type Wallpaper,
  deleteWallpaper,
  fetchWallpaperBlobUrl,
  getSchedule,
  listWallpapers,
  patchWallpaper,
  putSchedule,
  uploadWallpaper,
} from '@/lib/api'

const props = defineProps<{ screen: Screen; title: string }>()

const wallpapers = ref<Wallpaper[]>([])
const schedule = ref<Schedule | null>(null)
const loading = ref(true)
const error = ref('')
const uploading = ref(false)
const savingSchedule = ref(false)
const blobUrls = ref<Record<number, string>>({})
const fileInput = ref<HTMLInputElement | null>(null)

const hours = computed(() => Array.from({ length: 24 }, (_, i) => i))

async function load() {
  loading.value = true
  error.value = ''
  try {
    const [w, s] = await Promise.all([listWallpapers(props.screen), getSchedule(props.screen)])
    wallpapers.value = w
    schedule.value = s
    await loadBlobs(w)
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'load failed'
  } finally {
    loading.value = false
  }
}

async function loadBlobs(list: Wallpaper[]) {
  for (const url of Object.values(blobUrls.value)) URL.revokeObjectURL(url)
  blobUrls.value = {}
  await Promise.all(
    list.map(async (w) => {
      try {
        blobUrls.value[w.id] = await fetchWallpaperBlobUrl(w.id)
      } catch {
        /* skip */
      }
    }),
  )
}

onMounted(load)
watch(() => props.screen, load)
onBeforeUnmount(() => {
  for (const url of Object.values(blobUrls.value)) URL.revokeObjectURL(url)
})

async function onUpload(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploading.value = true
  error.value = ''
  try {
    await uploadWallpaper(props.screen, file)
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'upload failed'
  } finally {
    uploading.value = false
    input.value = ''
  }
}

async function toggleEnabled(w: Wallpaper) {
  try {
    await patchWallpaper(w.id, { enabled: !w.enabled })
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'update failed'
  }
}

async function remove(w: Wallpaper) {
  if (!confirm(`Delete ${w.filename}?`)) return
  try {
    await deleteWallpaper(w.id)
    await load()
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'delete failed'
  }
}

async function saveSchedule() {
  if (!schedule.value) return
  savingSchedule.value = true
  error.value = ''
  try {
    schedule.value = await putSchedule(props.screen, {
      mode: schedule.value.mode,
      intervalHours: schedule.value.intervalHours,
      startHour: schedule.value.startHour,
      enabled: schedule.value.enabled,
    })
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'schedule save failed'
  } finally {
    savingSchedule.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-8">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{{ title }}</h1>
      <p class="mt-1 text-sm text-muted-foreground">
        Browse, upload, enable/disable, delete wallpapers. Set rotation schedule.
      </p>
    </div>

    <p v-if="error" class="text-sm text-destructive" role="alert">{{ error }}</p>
    <p v-if="loading" class="text-sm text-muted-foreground">Loading…</p>

    <section v-if="schedule && !loading" class="flex flex-col gap-4 rounded-lg border bg-card p-4">
      <h2 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Schedule</h2>
      <div class="flex flex-wrap items-end gap-4">
        <label class="flex flex-col gap-1.5 text-sm">
          <span class="font-medium">Mode</span>
          <select
            v-model="schedule.mode"
            class="h-9 rounded-md border border-input bg-background px-3 text-sm"
          >
            <option value="daily">Daily</option>
            <option value="hourly">Hourly</option>
          </select>
        </label>
        <label v-if="schedule.mode === 'hourly'" class="flex flex-col gap-1.5 text-sm">
          <span class="font-medium">Every N hours</span>
          <select
            v-model.number="schedule.intervalHours"
            class="h-9 rounded-md border border-input bg-background px-3 text-sm"
          >
            <option :value="1">1</option>
            <option :value="2">2</option>
            <option :value="4">4</option>
            <option :value="8">8</option>
          </select>
        </label>
        <label class="flex flex-col gap-1.5 text-sm">
          <span class="font-medium">Start at HH</span>
          <select
            v-model.number="schedule.startHour"
            class="h-9 rounded-md border border-input bg-background px-3 text-sm"
          >
            <option v-for="h in hours" :key="h" :value="h">
              {{ String(h).padStart(2, '0') }}:00
            </option>
          </select>
        </label>
        <label class="flex items-center gap-2 text-sm">
          <input v-model="schedule.enabled" type="checkbox" class="size-4" />
          Schedule enabled
        </label>
        <Button :disabled="savingSchedule" @click="saveSchedule">
          {{ savingSchedule ? 'Saving…' : 'Save schedule' }}
        </Button>
      </div>
    </section>

    <section class="flex flex-col gap-4">
      <div class="flex flex-wrap items-center gap-3">
        <h2 class="text-sm font-semibold uppercase tracking-wide text-muted-foreground">Wallpapers</h2>
        <input ref="fileInput" type="file" accept="image/*" class="hidden" @change="onUpload" />
        <Button size="sm" :disabled="uploading" @click="fileInput?.click()">
          {{ uploading ? 'Uploading…' : 'Upload' }}
        </Button>
      </div>

      <div v-if="!loading && wallpapers.length === 0" class="text-sm text-muted-foreground">
        No wallpapers yet. Upload one.
      </div>

      <ul class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <li
          v-for="w in wallpapers"
          :key="w.id"
          class="overflow-hidden rounded-lg border bg-card"
          :class="{ 'opacity-50': !w.enabled }"
        >
          <div class="aspect-video bg-muted">
            <img
              v-if="blobUrls[w.id]"
              :src="blobUrls[w.id]"
              :alt="w.filename"
              class="size-full object-cover"
            />
          </div>
          <div class="flex flex-col gap-2 p-3">
            <p class="truncate text-sm font-medium" :title="w.filename">{{ w.filename }}</p>
            <div class="flex flex-wrap gap-2">
              <Button size="sm" variant="outline" @click="toggleEnabled(w)">
                {{ w.enabled ? 'Disable' : 'Enable' }}
              </Button>
              <Button size="sm" variant="destructive" @click="remove(w)">Delete</Button>
            </div>
          </div>
        </li>
      </ul>
    </section>
  </div>
</template>
