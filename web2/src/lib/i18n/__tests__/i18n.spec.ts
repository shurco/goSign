import { afterEach, beforeEach, describe, expect, it } from "vitest";
import {
  SIGNING_LOCALES,
  SUPPORTED_LOCALES,
  d,
  getLocale,
  messages,
  n,
  setLocale,
  t,
  te,
  type Locale
} from "../index.svelte";

describe("i18n Translation Loading and Fallback", () => {
  beforeEach(() => {
    setLocale("en");
  });

  afterEach(() => {
    setLocale("en");
    delete messages.zz;
  });

  describe("Locale State", () => {
    it("should default to a supported locale", () => {
      expect(Object.keys(SUPPORTED_LOCALES)).toContain(getLocale());
    });

    it("should switch locale via setLocale", () => {
      setLocale("ru");
      expect(getLocale()).toBe("ru");
    });
  });

  describe("Translation Loading", () => {
    it("should load translations for the current locale", () => {
      setLocale("en");
      expect(t("common.save")).toBe("Save");
      expect(t("common.cancel")).toBe("Cancel");
      expect(t("auth.signin")).toBe("Sign In");
    });

    it("should load translations when locale is changed", () => {
      setLocale("ru");
      expect(t("common.save")).toBe("Сохранить");
      expect(t("common.cancel")).toBe("Отмена");
    });

    it("should load translations for Spanish locale", () => {
      setLocale("es");
      expect(t("common.save")).toBe("Guardar");
      expect(t("common.cancel")).toBe("Cancelar");
    });

    it("should handle nested translation keys", () => {
      setLocale("en");
      expect(t("common.save")).toBe("Save");
      expect(t("auth.signin")).toBe("Sign In");
    });
  });

  describe("Fallback Behavior", () => {
    it("should fallback to English when translation key is missing in current locale", () => {
      // Inject a synthetic partial locale to guarantee a missing key
      messages.zz = { common: { save: "ZZ Save" } };
      setLocale("zz" as Locale);

      expect(t("common.save")).toBe("ZZ Save");
      // Missing in zz, falls back to English
      expect(t("common.delete")).toBe("Delete");
    });

    it("should use current locale when translation exists", () => {
      setLocale("ru");
      expect(t("common.save")).toBe("Сохранить");
      expect(t("common.cancel")).toBe("Отмена");
    });

    it("should fallback to English for completely missing locale", () => {
      setLocale("zz" as Locale);
      expect(t("common.save")).toBe("Save");
    });
  });

  describe("Interpolation", () => {
    it("should interpolate named params", () => {
      messages.zz = { greeting: "Hello, {name}!" };
      setLocale("zz" as Locale);
      expect(t("greeting", { name: "World" })).toBe("Hello, World!");
    });

    it("should keep placeholder when param is not provided", () => {
      messages.zz = { greeting: "Hello, {name}!" };
      setLocale("zz" as Locale);
      expect(t("greeting")).toBe("Hello, {name}!");
    });
  });

  describe("Supported Locales", () => {
    it("should include all 7 UI locales in SUPPORTED_LOCALES", () => {
      const expectedLocales = ["en", "ru", "es", "fr", "de", "it", "pt"];
      expect(Object.keys(SUPPORTED_LOCALES).sort()).toEqual(expectedLocales.sort());
    });

    it("should include all 14 signing locales in SIGNING_LOCALES", () => {
      const expectedLocales = ["en", "ru", "es", "fr", "de", "it", "pt", "zh", "ja", "ko", "ar", "hi", "pl", "nl"];
      expect(Object.keys(SIGNING_LOCALES).sort()).toEqual(expectedLocales.sort());
    });

    it("should have SIGNING_LOCALES include all SUPPORTED_LOCALES", () => {
      const supportedKeys = Object.keys(SUPPORTED_LOCALES);
      const signingKeys = Object.keys(SIGNING_LOCALES);

      supportedKeys.forEach((key) => {
        expect(signingKeys).toContain(key);
      });
    });

    it("should load messages for every supported UI locale", () => {
      for (const locale of Object.keys(SUPPORTED_LOCALES)) {
        expect(messages[locale]).toBeTruthy();
      }
    });
  });

  describe("Number and Date Formatting", () => {
    it("should format numbers according to locale", () => {
      setLocale("en");
      const formatted = n(1000, "currency");
      expect(formatted).toContain("1,000");
      expect(formatted).toContain("$");
    });

    it("should format dates according to locale", () => {
      setLocale("en");
      const date = new Date(2024, 0, 15);
      const formatted = d(date, "short");
      expect(formatted).toBeTruthy();
      expect(typeof formatted).toBe("string");
    });
  });

  describe("Missing Translation Keys", () => {
    it("should return the key path when translation is missing", () => {
      const result = t("common.nonexistent");
      expect(result).toBe("common.nonexistent");
    });

    it("should handle deeply nested missing keys", () => {
      const result = t("deeply.nested.missing.key");
      expect(result).toBe("deeply.nested.missing.key");
    });

    it("should report key existence via te()", () => {
      expect(te("common.save")).toBe(true);
      expect(te("common.nonexistent")).toBe(false);
    });
  });
});
