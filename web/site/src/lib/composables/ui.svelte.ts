import type { Action } from "svelte/action";

/**
 * Svelte action: invoke callback when clicking outside the element.
 * Usage: <div use:clickOutside={() => close()}>
 */
export const clickOutside: Action<HTMLElement, () => void> = (node, callback) => {
  let cb = callback;
  const handleClick = (event: MouseEvent) => {
    if (!node.contains(event.target as Node)) {
      cb();
    }
  };
  document.addEventListener("click", handleClick);
  return {
    update(newCallback: () => void) {
      cb = newCallback;
    },
    destroy() {
      document.removeEventListener("click", handleClick);
    }
  };
};

/**
 * Svelte action: invoke callback when Escape is pressed anywhere.
 * Usage: <div use:escapeKey={() => close()}>
 */
export const escapeKey: Action<HTMLElement, () => void> = (_node, callback) => {
  let cb = callback;
  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      cb();
    }
  };
  document.addEventListener("keydown", handleKeyDown);
  return {
    update(newCallback: () => void) {
      cb = newCallback;
    },
    destroy() {
      document.removeEventListener("keydown", handleKeyDown);
    }
  };
};

/**
 * Svelte action: trap Tab focus within the element while active.
 * Usage: <div use:focusTrap={isActive}>
 */
export const focusTrap: Action<HTMLElement, boolean> = (node, active) => {
  let isActive = active;

  const getFocusableElements = () => {
    const focusableSelectors = [
      "a[href]",
      "button:not([disabled])",
      "textarea:not([disabled])",
      "input:not([disabled])",
      "select:not([disabled])",
      '[tabindex]:not([tabindex="-1"])'
    ].join(", ");
    return Array.from(node.querySelectorAll(focusableSelectors)) as HTMLElement[];
  };

  const trapFocus = (event: KeyboardEvent) => {
    if (!isActive || event.key !== "Tab") {
      return;
    }

    const focusableElements = getFocusableElements();
    if (focusableElements.length === 0) {
      return;
    }

    const firstFocusable = focusableElements[0];
    const lastFocusable = focusableElements[focusableElements.length - 1];

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

  node.addEventListener("keydown", trapFocus);
  return {
    update(newActive: boolean) {
      isActive = newActive;
    },
    destroy() {
      node.removeEventListener("keydown", trapFocus);
    }
  };
};

/**
 * Svelte action: move the element to document.body (Vue Teleport equivalent).
 * Usage: <div use:portal>
 */
export const portal: Action<HTMLElement> = (node) => {
  document.body.appendChild(node);
  return {
    destroy() {
      node.remove();
    }
  };
};

/**
 * Dropdown open/close state. Combine with the clickOutside action:
 *   const dropdown = createDropdown();
 *   <div use:clickOutside={() => dropdown.close()}>
 */
export function createDropdown() {
  let isOpen = $state(false);

  return {
    get isOpen() {
      return isOpen;
    },
    open() {
      isOpen = true;
    },
    close() {
      isOpen = false;
    },
    toggle() {
      isOpen = !isOpen;
    }
  };
}
