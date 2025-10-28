import { onMounted, onUnmounted } from "vue";

export function useEscapeKey(callback: () => void): void {
  const handleEscape = (event: KeyboardEvent): void => {
    if (event.key === "Escape") {
      callback();
    }
  };

  onMounted(() => {
    document.addEventListener("keydown", handleEscape);
  });

  onUnmounted(() => {
    document.removeEventListener("keydown", handleEscape);
  });
}
