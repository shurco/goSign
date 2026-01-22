import { createI18n } from 'vue-i18n'
import type { I18nOptions } from 'vue-i18n'

// Supported UI languages (7 total)
export const SUPPORTED_LOCALES = {
  en: 'English',
  ru: 'Русский',
  es: 'Español',
  fr: 'Français',
  de: 'Deutsch',
  it: 'Italiano',
  pt: 'Português'
} as const

// Supported signing portal languages (14 total, includes UI + extra)
export const SIGNING_LOCALES = {
  ...SUPPORTED_LOCALES,
  zh: '中文',
  ja: '日本語',
  ko: '한국어',
  ar: 'العربية',
  hi: 'हिन्दी',
  pl: 'Polski',
  nl: 'Nederlands'
} as const

export type Locale = keyof typeof SUPPORTED_LOCALES
export type SigningLocale = keyof typeof SIGNING_LOCALES

// Auto-detect user locale
function detectLocale(): Locale {
  const stored = localStorage.getItem('locale')
  if (stored && stored in SUPPORTED_LOCALES) {
    return stored as Locale
  }
  
  const browser = navigator.language.split('-')[0]
  if (browser in SUPPORTED_LOCALES) {
    return browser as Locale
  }
  
  return 'en'
}

// Load messages dynamically
const messages: Record<string, any> = {}

// Import all locale files
const localeModules = import.meta.glob('./locales/*.json', { eager: true })
for (const path in localeModules) {
  const locale = path.match(/\/([^/]+)\.json$/)?.[1]
  if (locale) {
    messages[locale] = (localeModules[path] as any).default || localeModules[path]
  }
}

const i18n = createI18n<I18nOptions>({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages,
  datetimeFormats: {
    en: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    ru: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    es: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    fr: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    de: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    it: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    },
    pt: {
      short: {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      },
      long: {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        weekday: 'long',
        hour: 'numeric',
        minute: 'numeric'
      }
    }
  },
  numberFormats: {
    en: {
      currency: {
        style: 'currency',
        currency: 'USD'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    },
    ru: {
      currency: {
        style: 'currency',
        currency: 'RUB'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    },
    es: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    },
    fr: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    },
    de: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    },
    it: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    },
    pt: {
      currency: {
        style: 'currency',
        currency: 'EUR'
      },
      decimal: {
        style: 'decimal',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }
    }
  }
})

export default i18n
