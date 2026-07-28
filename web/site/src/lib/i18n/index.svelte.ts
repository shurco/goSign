// Lightweight i18n replacement for vue-i18n, driven by Svelte 5 runes.
// Message format is identical to the previous setup: nested JSON with
// `{name}` named interpolation (no plurals / linked messages are used).

// Supported UI languages (7 total)
export const SUPPORTED_LOCALES = {
  en: "English",
  ru: "Русский",
  es: "Español",
  fr: "Français",
  de: "Deutsch",
  it: "Italiano",
  pt: "Português"
} as const;

// Supported signing portal languages (14 total, includes UI + extra)
export const SIGNING_LOCALES = {
  ...SUPPORTED_LOCALES,
  zh: "中文",
  ja: "日本語",
  ko: "한국어",
  ar: "العربية",
  hi: "हिन्दी",
  pl: "Polski",
  nl: "Nederlands"
} as const;

export type Locale = keyof typeof SUPPORTED_LOCALES;
export type SigningLocale = keyof typeof SIGNING_LOCALES;

type MessageTree = { [key: string]: string | MessageTree };

// Auto-detect user locale
function detectLocale(): Locale {
  if (typeof localStorage !== "undefined") {
    const stored = localStorage.getItem("locale");
    if (stored && stored in SUPPORTED_LOCALES) {
      return stored as Locale;
    }
  }

  if (typeof navigator !== "undefined") {
    const browser = navigator.language.split("-")[0];
    if (browser in SUPPORTED_LOCALES) {
      return browser as Locale;
    }
  }

  return "en";
}

// Load messages eagerly (same as previous import.meta.glob setup)
export const messages: Record<string, MessageTree> = {};

const localeModules = import.meta.glob("./locales/*.json", { eager: true });
for (const path in localeModules) {
  const locale = path.match(/\/([^/]+)\.json$/)?.[1];
  if (locale) {
    messages[locale] = ((localeModules[path] as { default?: MessageTree }).default ||
      localeModules[path]) as MessageTree;
  }
}

const i18nState = $state({ locale: detectLocale() });

export function getLocale(): Locale {
  return i18nState.locale;
}

export function setLocale(locale: Locale): void {
  i18nState.locale = locale;
  if (typeof localStorage !== "undefined") {
    localStorage.setItem("locale", locale);
  }
  if (typeof document !== "undefined") {
    document.documentElement.setAttribute("lang", locale);
  }
}

function resolveKey(tree: MessageTree | undefined, key: string): string | undefined {
  if (!tree) {
    return undefined;
  }
  let node: string | MessageTree | undefined = tree;
  for (const part of key.split(".")) {
    if (typeof node !== "object" || node === null) {
      return undefined;
    }
    node = node[part];
  }
  return typeof node === "string" ? node : undefined;
}

function interpolate(message: string, params?: Record<string, unknown>): string {
  if (!params) {
    return message;
  }
  return message.replace(/\{(\w+)\}/g, (match, name: string) =>
    params[name] !== undefined ? String(params[name]) : match
  );
}

/**
 * Translate a key with optional named params, e.g. t("common.save") or
 * t("templates.deleteConfirm", { name }). Falls back to English, then to the key.
 */
export function t(key: string, params?: Record<string, unknown>): string {
  const message = resolveKey(messages[i18nState.locale], key) ?? resolveKey(messages.en, key);
  if (message === undefined) {
    return key;
  }
  return interpolate(message, params);
}

/** Check whether a translation key exists in the active (or fallback) locale. */
export function te(key: string): boolean {
  return resolveKey(messages[i18nState.locale], key) !== undefined || resolveKey(messages.en, key) !== undefined;
}

const datetimeFormats: Record<string, Intl.DateTimeFormatOptions> = {
  short: { year: "numeric", month: "short", day: "numeric" },
  long: {
    year: "numeric",
    month: "long",
    day: "numeric",
    weekday: "long",
    hour: "numeric",
    minute: "numeric"
  }
};

/** Format a date using the active locale ("short" | "long"). */
export function d(value: Date | number | string, format: "short" | "long" = "short"): string {
  // eslint-disable-next-line svelte/prefer-svelte-reactivity -- transient local value, never stored in reactive state
  const date = value instanceof Date ? value : new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "";
  }
  return new Intl.DateTimeFormat(i18nState.locale, datetimeFormats[format]).format(date);
}

const currencyByLocale: Record<Locale, string> = {
  en: "USD",
  ru: "RUB",
  es: "EUR",
  fr: "EUR",
  de: "EUR",
  it: "EUR",
  pt: "EUR"
};

/** Format a number using the active locale ("currency" | "decimal"). */
export function n(value: number, format: "currency" | "decimal" = "decimal"): string {
  const options: Intl.NumberFormatOptions =
    format === "currency"
      ? { style: "currency", currency: currencyByLocale[i18nState.locale] ?? "USD" }
      : { style: "decimal", minimumFractionDigits: 2, maximumFractionDigits: 2 };
  return new Intl.NumberFormat(i18nState.locale, options).format(value);
}
