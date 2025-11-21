import { ref, watch, computed } from 'vue'

export type Theme = 'light' | 'dark'

const THEME_KEY = 'calibre-theme'

// 检测系统主题偏好
const getSystemTheme = (): Theme => {
    if (typeof window !== 'undefined' && window.matchMedia) {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return 'light'
}

// 全局主题状态
const currentTheme = ref<Theme>(
    (localStorage.getItem(THEME_KEY) as Theme) || getSystemTheme()
)

export function useTheme() {
    const toggleTheme = () => {
        currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
    }

    const setTheme = (theme: Theme) => {
        currentTheme.value = theme
    }

    // 监听主题变化，更新DOM和localStorage
    watch(
        currentTheme,
        (newTheme) => {
            localStorage.setItem(THEME_KEY, newTheme)
            document.documentElement.setAttribute('data-theme', newTheme)
            if (newTheme === 'dark') {
                document.documentElement.classList.add('dark')
            } else {
                document.documentElement.classList.remove('dark')
            }
        },
        { immediate: true }
    )

    // 监听系统主题变化
    if (typeof window !== 'undefined' && window.matchMedia) {
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem(THEME_KEY)) {
                currentTheme.value = e.matches ? 'dark' : 'light'
            }
        })
    }

    return {
        theme: currentTheme,
        toggleTheme,
        setTheme,
        isDark: computed(() => currentTheme.value === 'dark'),
        isLight: computed(() => currentTheme.value === 'light')
    }
}
