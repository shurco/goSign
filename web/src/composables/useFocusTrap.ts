import { onMounted, onUnmounted, Ref, watch } from "vue";

export function useFocusTrap(containerRef: Ref<HTMLElement | null>, isActive: Ref<boolean>): void {
  const focusableSelectors = 'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])';

  const handleTabKey = (event: KeyboardEvent): void => {
    if (!isActive.value || event.key !== "Tab" || !containerRef.value) {
      return;
    }

    const focusableElements = containerRef.value.querySelectorAll<HTMLElement>(focusableSelectors);
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    if (event.shiftKey) {
      if (document.activeElement === firstElement) {
        lastElement?.focus();
        event.preventDefault();
      }
    } else {
      if (document.activeElement === lastElement) {
        firstElement?.focus();
        event.preventDefault();
      }
    }
  };

  watch(isActive, (active) => {
    if (active && containerRef.value) {
      const focusableElements = containerRef.value.querySelectorAll<HTMLElement>(focusableSelectors);
      focusableElements[0]?.focus();
    }
  });

  onMounted(() => {
    document.addEventListener("keydown", handleTabKey);
  });

  onUnmounted(() => {
    document.removeEventListener("keydown", handleTabKey);
  });
}
