import { ref, watch } from 'vue'

type Theme = 'light' | 'dark'

const THEME_STORAGE_KEY = 'theme'

// Initialize theme from localStorage or default to dark
const theme = ref<Theme>((localStorage.getItem(THEME_STORAGE_KEY) as Theme) || 'dark')

// Watch for theme changes and persist to localStorage
watch(theme, (newTheme) => {
    localStorage.setItem(THEME_STORAGE_KEY, newTheme)
})

export function useTheme() {
    const isDark = ref(theme.value === 'dark')

    const toggleTheme = () => {
        theme.value = theme.value === 'dark' ? 'light' : 'dark'
        isDark.value = theme.value === 'dark'
    }

    const setTheme = (newTheme: Theme) => {
        theme.value = newTheme
        isDark.value = newTheme === 'dark'
    }

    // Sync isDark with theme changes from other components
    watch(theme, (newTheme) => {
        isDark.value = newTheme === 'dark'
    })

    return {
        isDark,
        theme,
        toggleTheme,
        setTheme
    }
}
