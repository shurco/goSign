import { ref } from 'vue'
import { apiGet } from '@/services/api'
import type { BrandingSettings } from '@/models/account'

export interface ThemeConfig {
  primaryColor: string
  secondaryColor: string
  accentColor: string
  backgroundColor: string
  textColor: string
  fontFamily: string
  fontURL?: string
  customCSS?: string
}

export function useTheme() {
  const theme = ref<ThemeConfig | null>(null)
  
  async function loadTheme() {
    try {
      const response = await apiGet('/api/v1/branding')
      if (response && response.data?.branding) {
        const branding = response.data.branding as BrandingSettings
        const themeConfig: ThemeConfig = {
          primaryColor: branding.primary_color || '#4F46E5',
          secondaryColor: branding.secondary_color || '#6366F1',
          accentColor: branding.accent_color || '#10B981',
          backgroundColor: branding.background_color || '#FFFFFF',
          textColor: branding.text_color || '#111827',
          fontFamily: branding.font_family || 'Inter',
          fontURL: branding.font_url,
          customCSS: branding.custom_css,
        }
        theme.value = themeConfig
        applyTheme(themeConfig)
      }
    } catch (error) {
      console.error('Failed to load theme:', error)
    }
  }
  
  function applyTheme(config: ThemeConfig) {
    const root = document.documentElement
    
    // Apply CSS variables
    root.style.setProperty('--color-primary', config.primaryColor)
    root.style.setProperty('--color-secondary', config.secondaryColor)
    root.style.setProperty('--color-accent', config.accentColor)
    root.style.setProperty('--color-background', config.backgroundColor)
    root.style.setProperty('--color-text', config.textColor)
    
    // Apply font
    if (config.fontFamily) {
      root.style.setProperty('--font-family', config.fontFamily)
    }
    
    // Load custom font
    if (config.fontURL) {
      const existingLink = document.querySelector(`link[href="${config.fontURL}"]`)
      if (!existingLink) {
        const link = document.createElement('link')
        link.href = config.fontURL
        link.rel = 'stylesheet'
        document.head.appendChild(link)
      }
    }
    
    // Apply custom CSS
    if (config.customCSS) {
      let styleElement = document.getElementById('custom-branding-css')
      if (!styleElement) {
        styleElement = document.createElement('style')
        styleElement.id = 'custom-branding-css'
        document.head.appendChild(styleElement)
      }
      styleElement.textContent = config.customCSS
    }
  }
  
  function resetTheme() {
    const root = document.documentElement
    root.style.removeProperty('--color-primary')
    root.style.removeProperty('--color-secondary')
    root.style.removeProperty('--color-accent')
    root.style.removeProperty('--color-background')
    root.style.removeProperty('--color-text')
    root.style.removeProperty('--font-family')
    
    // Remove custom CSS
    const styleElement = document.getElementById('custom-branding-css')
    if (styleElement) {
      styleElement.remove()
    }
    
    // Remove font link
    const fontLinks = document.querySelectorAll('link[rel="stylesheet"]')
    fontLinks.forEach((link) => {
      if (link.getAttribute('href')?.includes('fonts.googleapis.com')) {
        link.remove()
      }
    })
    
    theme.value = null
  }
  
  return {
    theme,
    loadTheme,
    applyTheme,
    resetTheme
  }
}
