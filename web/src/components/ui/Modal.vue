<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4" @click="handleBackdropClick">
        <div class="fixed inset-0 bg-black/50 transition-opacity" />
        <div
          ref="modalRef"
          :class="modalClasses"
          class="relative z-10 w-full rounded-lg border border-gray-200 bg-white transition-colors"
          @click.stop
        >
          <div v-if="$slots.header || title" class="border-b border-gray-200 px-6 py-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">
                <slot name="header">{{ title }}</slot>
              </h3>
              <button v-if="showClose" class="text-gray-400 transition-colors hover:text-gray-600" @click="handleClose">
                <SvgIcon name="x" class="h-5 w-5" />
              </button>
            </div>
          </div>
          <div class="px-6 py-4">
            <slot />
          </div>
          <div v-if="$slots.footer" class="border-t border-gray-200 px-6 py-4">
            <slot name="footer" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useEscapeKey, useFocusTrap } from "@/composables/ui";
import SvgIcon from "@/components/SvgIcon.vue";

interface Props {
  modelValue: boolean;
  title?: string;
  size?: "sm" | "md" | "lg" | "xl";
  showClose?: boolean;
  closeOnBackdrop?: boolean;
  closeOnEscape?: boolean;
}

interface Emits {
  (e: "update:modelValue", value: boolean): void;
  (e: "close"): void;
}

const props = withDefaults(defineProps<Props>(), {
  title: "",
  size: "md",
  showClose: true,
  closeOnBackdrop: true,
  closeOnEscape: true
});

const emit = defineEmits<Emits>();

const modalRef = ref<HTMLElement | null>(null);
const isOpen = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value)
});

const modalClasses = computed(() => {
  const sizes = {
    sm: "max-w-sm",
    md: "max-w-2xl",
    lg: "max-w-4xl",
    xl: "max-w-6xl"
  };
  return sizes[props.size];
});

watch(isOpen, (open) => {
  if (open) {
    document.body.style.overflow = "hidden";
  } else {
    document.body.style.overflow = "";
  }
});

function handleClose(): void {
  isOpen.value = false;
  emit("close");
}

function handleBackdropClick(): void {
  if (props.closeOnBackdrop) {
    handleClose();
  }
}

if (props.closeOnEscape) {
  useEscapeKey(handleClose);
}

useFocusTrap(modalRef, isOpen);
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
