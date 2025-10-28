import { ref, type Ref } from "vue";
import { useClickOutside } from "./useClickOutside";

export function useDropdown(dropdownRef: Ref<HTMLElement | null>) {
  const isOpen = ref(false);

  // Close dropdown when clicking outside
  useClickOutside(dropdownRef, () => {
    if (isOpen.value) {
      close();
    }
  });

  function open(): void {
    isOpen.value = true;
  }

  function close(): void {
    isOpen.value = false;
    const activeElement = document.activeElement as HTMLElement;
    if (activeElement) {
      activeElement.blur();
    }
  }

  function toggle(state?: boolean): void {
    isOpen.value = state !== undefined ? state : !isOpen.value;
  }

  return {
    isOpen,
    open,
    close,
    toggle
  };
}
