import { useState, useEffect } from 'react'

type Theme = 'dark' | 'light'

const THEME_KEY = 'hawkward-theme'

export function useTheme() {
  const [theme, setThemeState] = useState<Theme>(() => {
    return (localStorage.getItem(THEME_KEY) as Theme) || 'dark'
  })

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem(THEME_KEY, theme)
  }, [theme])

  const toggle = () => {
    setThemeState(t => t === 'dark' ? 'light' : 'dark')
  }

  return { theme, toggle }
}
