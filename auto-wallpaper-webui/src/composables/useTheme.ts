import { onMounted, ref, watch } from 'vue'

type Theme = 'light' | 'dark'

const theme = ref<Theme>('light')

function apply(t: Theme) {
  document.documentElement.classList.toggle('dark', t === 'dark')
}

export function useTheme() {
  onMounted(() => {
    const saved = localStorage.getItem('autowall_theme') as Theme | null
    theme.value = saved === 'dark' ? 'dark' : 'light'
    apply(theme.value)
  })

  watch(theme, (t) => {
    localStorage.setItem('autowall_theme', t)
    apply(t)
  })

  function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  return { theme, toggleTheme }
}
