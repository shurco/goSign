import { beforeEach } from "vitest";

// Mock localStorage
const localStorageMock = (() => {
  const store = new Map<string, string>();

  return {
    getItem: (key: string) => store.get(key) ?? null,
    setItem: (key: string, value: string) => {
      store.set(key, value.toString());
    },
    removeItem: (key: string) => {
      store.delete(key);
    },
    clear: () => {
      store.clear();
    }
  };
})();

Object.defineProperty(window, "localStorage", {
  value: localStorageMock
});

// Mock navigator.language
Object.defineProperty(navigator, "language", {
  writable: true,
  value: "en-US"
});

// Clean up localStorage before each test
beforeEach(() => {
  localStorage.clear();
});
