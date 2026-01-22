# Multilingual Support (i18n)

**Last Updated**: 2026-01-21 00:00 UTC

## Overview

goSign supports 7 UI languages and 14 signing portal languages with automatic locale detection and RTL support.

## Supported Languages

### UI Languages (7)
- English (en)
- Russian (ru)
- Spanish (es)
- French (fr)
- German (de)
- Italian (it)
- Portuguese (pt)

### Signing Portal Languages (14)
Includes all UI languages plus:
- Chinese (zh)
- Japanese (ja)
- Korean (ko)
- Arabic (ar)
- Hindi (hi)
- Polish (pl)
- Dutch (nl)

## Features

### Automatic Locale Detection
- Detects browser language on first visit
- Falls back to English if browser language not supported
- Remembers user preference in localStorage

### RTL Support
- Automatic RTL layout for Arabic and Hebrew
- CSS adjustments for proper text direction
- Component-level RTL styling

### Field-Level Translations
- Each field can have translations for different locales
- Template-level default locale
- Submission-level locale override

## Usage

### Changing Language

**In UI:**
```vue
<LanguageSwitcher />
```

**Programmatically:**
```typescript
import { useI18n } from 'vue-i18n'

const { locale } = useI18n()
locale.value = 'ru'
```

### Using Translations

**In Templates:**
```vue
<template>
  <h1>{{ $t('dashboard.title') }}</h1>
  <Button>{{ $t('common.save') }}</Button>
</template>
```

**In Scripts:**
```typescript
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const title = t('dashboard.title')
```

### Field Labels

Fields support translations:
```typescript
const field: Field = {
  id: 'field_1',
  name: 'Full Name',
  label: 'Full Name',
  translations: {
    'ru': 'Полное имя',
    'es': 'Nombre completo'
  }
}
```

## API Endpoints

### Get Available Locales
```bash
GET /api/v1/i18n/locales
```

### Update User Locale
```bash
PUT /api/v1/user/locale
Content-Type: application/json

{
  "locale": "ru"
}
```

### Update Account Locale
```bash
PUT /api/v1/account/locale
Content-Type: application/json

{
  "locale": "en"
}
```

## Adding New Languages

1. Create translation file: `web/src/i18n/locales/{locale}.json`
2. Add locale to `SUPPORTED_LOCALES` in `web/src/i18n/index.ts`
3. Add datetime and number formats
4. Update all translation keys

## Best Practices

- Always use translation keys, never hardcoded strings
- Provide fallback translations for missing keys
- Test with different locales during development
- Consider text expansion (some languages need more space)
- Test RTL layouts for Arabic/Hebrew

## Translation Keys Structure

```
{
  "common": { ... },
  "auth": { ... },
  "template": { ... },
  "fields": { ... },
  "signing": { ... },
  "submissions": { ... },
  "dashboard": { ... },
  "settings": { ... },
  "errors": { ... },
  "success": { ... }
}
```
