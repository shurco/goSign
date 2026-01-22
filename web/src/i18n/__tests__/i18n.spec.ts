import { describe, it, expect, beforeEach } from 'vitest'
import { createI18n } from 'vue-i18n'
import { SUPPORTED_LOCALES, SIGNING_LOCALES } from '../index'

describe('i18n Translation Loading and Fallback', () => {
  let i18n: ReturnType<typeof createI18n>

  beforeEach(() => {
    // Reset localStorage before each test
    localStorage.clear()
    
    // Reset navigator.language
    Object.defineProperty(navigator, 'language', {
      writable: true,
      value: 'en-US',
    })
  })

  describe('Locale Detection', () => {
    it('should detect locale from localStorage if available', () => {
      localStorage.setItem('locale', 'ru')
      
      // Create a new i18n instance to test detection
      const messages: Record<string, any> = {
        en: { common: { save: 'Save' } },
        ru: { common: { save: 'Сохранить' } },
      }
      
      i18n = createI18n({
        legacy: false,
        locale: localStorage.getItem('locale') || 'en',
        fallbackLocale: 'en',
        messages,
      })

      expect(i18n.global.locale.value).toBe('ru')
    })

    it('should detect locale from browser language if localStorage is empty', () => {
      localStorage.clear()
      Object.defineProperty(navigator, 'language', {
        writable: true,
        value: 'es-ES',
      })

      const messages: Record<string, any> = {
        en: { common: { save: 'Save' } },
        es: { common: { save: 'Guardar' } },
      }

      // Simulate detection logic
      const browser = navigator.language.split('-')[0]
      const detectedLocale = browser in SUPPORTED_LOCALES ? browser : 'en'

      i18n = createI18n({
        legacy: false,
        locale: detectedLocale,
        fallbackLocale: 'en',
        messages,
      })

      expect(i18n.global.locale.value).toBe('es')
    })

    it('should fallback to English if browser language is not supported', () => {
      localStorage.clear()
      Object.defineProperty(navigator, 'language', {
        writable: true,
        value: 'zh-CN',
      })

      const messages: Record<string, any> = {
        en: { common: { save: 'Save' } },
      }

      const browser = navigator.language.split('-')[0]
      const detectedLocale = browser in SUPPORTED_LOCALES ? browser : 'en'

      i18n = createI18n({
        legacy: false,
        locale: detectedLocale,
        fallbackLocale: 'en',
        messages,
      })

      expect(i18n.global.locale.value).toBe('en')
    })

    it('should fallback to English if localStorage contains unsupported locale', () => {
      localStorage.setItem('locale', 'unsupported')

      const messages: Record<string, any> = {
        en: { common: { save: 'Save' } },
      }

      const stored = localStorage.getItem('locale')
      const detectedLocale =
        stored && stored in SUPPORTED_LOCALES ? stored : 'en'

      i18n = createI18n({
        legacy: false,
        locale: detectedLocale,
        fallbackLocale: 'en',
        messages,
      })

      expect(i18n.global.locale.value).toBe('en')
    })
  })

  describe('Translation Loading', () => {
    beforeEach(() => {
      const messages: Record<string, any> = {
        en: {
          common: {
            save: 'Save',
            cancel: 'Cancel',
          },
          auth: {
            signin: 'Sign In',
          },
        },
        ru: {
          common: {
            save: 'Сохранить',
            cancel: 'Отмена',
          },
          auth: {
            signin: 'Войти',
          },
        },
        es: {
          common: {
            save: 'Guardar',
            cancel: 'Cancelar',
          },
        },
      }

      i18n = createI18n({
        legacy: false,
        locale: 'en',
        fallbackLocale: 'en',
        messages,
      })
    })

    it('should load translations for the current locale', () => {
      i18n.global.locale.value = 'en'
      expect(i18n.global.t('common.save')).toBe('Save')
      expect(i18n.global.t('common.cancel')).toBe('Cancel')
      expect(i18n.global.t('auth.signin')).toBe('Sign In')
    })

    it('should load translations when locale is changed', () => {
      i18n.global.locale.value = 'ru'
      expect(i18n.global.t('common.save')).toBe('Сохранить')
      expect(i18n.global.t('common.cancel')).toBe('Отмена')
      expect(i18n.global.t('auth.signin')).toBe('Войти')
    })

    it('should load translations for Spanish locale', () => {
      i18n.global.locale.value = 'es'
      expect(i18n.global.t('common.save')).toBe('Guardar')
      expect(i18n.global.t('common.cancel')).toBe('Cancelar')
    })

    it('should handle nested translation keys', () => {
      i18n.global.locale.value = 'en'
      expect(i18n.global.t('common.save')).toBe('Save')
      expect(i18n.global.t('auth.signin')).toBe('Sign In')
    })
  })

  describe('Fallback Behavior', () => {
    beforeEach(() => {
      const messages: Record<string, any> = {
        en: {
          common: {
            save: 'Save',
            cancel: 'Cancel',
            delete: 'Delete',
          },
        },
        ru: {
          common: {
            save: 'Сохранить',
            cancel: 'Отмена',
            // Missing 'delete' key
          },
        },
      }

      i18n = createI18n({
        legacy: false,
        locale: 'ru',
        fallbackLocale: 'en',
        messages,
      })
    })

    it('should fallback to English when translation key is missing', () => {
      i18n.global.locale.value = 'ru'
      // 'delete' is missing in Russian, should fallback to English
      expect(i18n.global.t('common.delete')).toBe('Delete')
    })

    it('should use current locale when translation exists', () => {
      i18n.global.locale.value = 'ru'
      expect(i18n.global.t('common.save')).toBe('Сохранить')
      expect(i18n.global.t('common.cancel')).toBe('Отмена')
    })

    it('should fallback to English for completely missing locale', () => {
      const messages: Record<string, any> = {
        en: {
          common: {
            save: 'Save',
          },
        },
      }

      i18n = createI18n({
        legacy: false,
        locale: 'fr', // French not in messages
        fallbackLocale: 'en',
        messages,
      })

      expect(i18n.global.t('common.save')).toBe('Save')
    })
  })

  describe('Supported Locales', () => {
    it('should include all 7 UI locales in SUPPORTED_LOCALES', () => {
      const expectedLocales = ['en', 'ru', 'es', 'fr', 'de', 'it', 'pt']
      expect(Object.keys(SUPPORTED_LOCALES).sort()).toEqual(
        expectedLocales.sort()
      )
    })

    it('should include all 14 signing locales in SIGNING_LOCALES', () => {
      const expectedLocales = [
        'en',
        'ru',
        'es',
        'fr',
        'de',
        'it',
        'pt',
        'zh',
        'ja',
        'ko',
        'ar',
        'hi',
        'pl',
        'nl',
      ]
      expect(Object.keys(SIGNING_LOCALES).sort()).toEqual(
        expectedLocales.sort()
      )
    })

    it('should have SIGNING_LOCALES include all SUPPORTED_LOCALES', () => {
      const supportedKeys = Object.keys(SUPPORTED_LOCALES)
      const signingKeys = Object.keys(SIGNING_LOCALES)

      supportedKeys.forEach((key) => {
        expect(signingKeys).toContain(key)
      })
    })
  })

  describe('Number and Date Formatting', () => {
    beforeEach(() => {
      const messages: Record<string, any> = {
        en: {},
        ru: {},
      }

      i18n = createI18n({
        legacy: false,
        locale: 'en',
        fallbackLocale: 'en',
        messages,
        numberFormats: {
          en: {
            currency: {
              style: 'currency',
              currency: 'USD',
            },
          },
          ru: {
            currency: {
              style: 'currency',
              currency: 'RUB',
            },
          },
        },
        datetimeFormats: {
          en: {
            short: {
              year: 'numeric',
              month: 'short',
              day: 'numeric',
            },
          },
          ru: {
            short: {
              year: 'numeric',
              month: 'short',
              day: 'numeric',
            },
          },
        },
      })
    })

    it('should format numbers according to locale', () => {
      i18n.global.locale.value = 'en'
      const formatted = i18n.global.n(1000, 'currency')
      expect(formatted).toContain('1,000')
      expect(formatted).toContain('$')
    })

    it('should format dates according to locale', () => {
      i18n.global.locale.value = 'en'
      const date = new Date(2024, 0, 15)
      const formatted = i18n.global.d(date, 'short')
      expect(formatted).toBeTruthy()
      expect(typeof formatted).toBe('string')
    })
  })

  describe('Missing Translation Keys', () => {
    beforeEach(() => {
      const messages: Record<string, any> = {
        en: {
          common: {
            save: 'Save',
          },
        },
      }

      i18n = createI18n({
        legacy: false,
        locale: 'en',
        fallbackLocale: 'en',
        messages,
        missingWarn: false, // Suppress warnings in tests
      })
    })

    it('should return the key path when translation is missing', () => {
      const result = i18n.global.t('common.nonexistent')
      expect(result).toBe('common.nonexistent')
    })

    it('should handle deeply nested missing keys', () => {
      const result = i18n.global.t('deeply.nested.missing.key')
      expect(result).toBe('deeply.nested.missing.key')
    })
  })
})
