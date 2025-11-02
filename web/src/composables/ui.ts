import { onBeforeUnmount, onMounted, ref, type Ref } from "vue";

/**
 * Handles click outside event for element
 */
export function useClickOutside(elementRef: Ref<HTMLElement | null>, callback: () => void): void {
  const handleClick = (event: MouseEvent) => {
    if (elementRef.value && !elementRef.value.contains(event.target as Node)) {
      callback();
    }
  };

  onMounted(() => {
    document.addEventListener("click", handleClick);
  });

  onBeforeUnmount(() => {
    document.removeEventListener("click", handleClick);
  });
}

/**
 * Handles escape key press
 */
export function useEscapeKey(callback: () => void): void {
  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      callback();
    }
  };

  onMounted(() => {
    document.addEventListener("keydown", handleKeyDown);
  });

  onBeforeUnmount(() => {
    document.removeEventListener("keydown", handleKeyDown);
  });
}

/**
 * Traps focus within container element
 */
export function useFocusTrap(containerRef: Ref<HTMLElement | null>, isActive: Ref<boolean>): void {
  let firstFocusable: HTMLElement | null = null;
  let lastFocusable: HTMLElement | null = null;

  const getFocusableElements = () => {
    if (!containerRef.value) return [];

    const focusableSelectors = [
      "a[href]",
      "button:not([disabled])",
      "textarea:not([disabled])",
      "input:not([disabled])",
      "select:not([disabled])",
      '[tabindex]:not([tabindex="-1"])'
    ].join(", ");

    return Array.from(containerRef.value.querySelectorAll(focusableSelectors)) as HTMLElement[];
  };

  const trapFocus = (event: KeyboardEvent) => {
    if (!isActive.value || !containerRef.value) return;

    if (event.key !== "Tab") return;

    const focusableElements = getFocusableElements();
    if (focusableElements.length === 0) return;

    firstFocusable = focusableElements[0];
    lastFocusable = focusableElements[focusableElements.length - 1];

    if (event.shiftKey) {
      if (document.activeElement === firstFocusable) {
        event.preventDefault();
        lastFocusable?.focus();
      }
    } else {
      if (document.activeElement === lastFocusable) {
        event.preventDefault();
        firstFocusable?.focus();
      }
    }
  };

  onMounted(() => {
    if (isActive.value && containerRef.value) {
      containerRef.value.addEventListener("keydown", trapFocus);
    }
  });

  onBeforeUnmount(() => {
    if (containerRef.value) {
      containerRef.value.removeEventListener("keydown", trapFocus);
    }
  });
}

/**
 * Manages dropdown open/close state with click outside handling
 */
export function useDropdown(dropdownRef: Ref<HTMLElement | null>) {
  const isOpen = ref(false);

  useClickOutside(dropdownRef, () => {
    if (isOpen.value) {
      isOpen.value = false;
    }
  });

  const open = () => {
    isOpen.value = true;
  };

  const close = () => {
    isOpen.value = false;
  };

  const toggle = () => {
    isOpen.value = !isOpen.value;
  };

  return {
    isOpen,
    open,
    close,
    toggle
  };
}
