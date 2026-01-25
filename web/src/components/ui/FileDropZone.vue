<template>
  <label
    :for="inputId"
    class="relative block w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50 transition-colors"
    :class="[
      heightClass,
      selectedLabel ? 'bg-gray-50 border-blue-400' : '',
      disabled ? 'opacity-60 cursor-not-allowed' : ''
    ]"
    @dragover.prevent
    @drop.prevent="onDrop"
  >
    <div class="absolute inset-0 flex items-center justify-center p-2">
      <div class="flex flex-col items-center text-center">
        <template v-if="!selectedLabel">
          <SvgIcon name="cloud-upload" class="h-8 w-8 text-gray-400 shrink-0" />
          <div class="mt-2 text-sm font-medium text-gray-700">{{ clickLabel }}</div>
          <div class="text-xs text-gray-500">{{ dragLabel }}</div>
        </template>
        <template v-else>
          <SvgIcon name="document" class="h-8 w-8 text-blue-500 shrink-0" />
          <div class="mt-2 text-sm font-medium text-gray-900 truncate max-w-full">{{ selectedLabel }}</div>
          <button
            v-if="!disabled"
            type="button"
            class="mt-1 text-xs text-red-600 hover:text-red-800"
            @click.stop="clear"
          >
            {{ removeLabel }}
          </button>
        </template>
      </div>
    </div>

    <input
      :id="inputId"
      ref="inputRef"
      type="file"
      class="hidden"
      :accept="accept"
      :disabled="disabled"
      @change="onChange"
    />
  </label>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useI18n } from "vue-i18n";
import SvgIcon from "@/components/SvgIcon.vue";

interface Props {
  accept?: string;
  disabled?: boolean;
  /** Display when a file is selected (e.g. file name). */
  selectedLabel?: string;
  /** Height of the zone, e.g. 'h-32' or '128px'. Default h-32. */
  height?: string;
  /** Override "Click to upload" (otherwise from i18n templates.clickToUpload). */
  clickLabel?: string;
  /** Override "or drag and drop here" (otherwise from i18n templates.dragAndDrop). */
  dragLabel?: string;
  /** Override "Remove file" (otherwise from i18n templates.removeFile). */
  removeLabel?: string;
}

const props = withDefaults(defineProps<Props>(), {
  accept: undefined,
  disabled: false,
  selectedLabel: "",
  height: "h-32",
  clickLabel: "",
  dragLabel: "",
  removeLabel: ""
});

const emit = defineEmits<{
  (e: "change", file: File): void;
  (e: "clear"): void;
}>();

const { t } = useI18n();
const inputRef = ref<HTMLInputElement | null>(null);

const inputId = computed(() => `file-drop-zone-${Math.random().toString(36).slice(2, 9)}`);

const heightClass = computed(() => props.height || "h-32");

const clickLabel = computed(() => props.clickLabel || t("templates.clickToUpload"));
const dragLabel = computed(() => props.dragLabel || t("templates.dragAndDrop"));
const removeLabel = computed(() => props.removeLabel || t("templates.removeFile"));

function onChange(e: Event): void {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (file) emit("change", file);
  input.value = "";
}

function onDrop(e: DragEvent): void {
  if (props.disabled) return;
  const file = e.dataTransfer?.files?.[0];
  if (file) emit("change", file);
}

function clear(): void {
  if (inputRef.value) inputRef.value.value = "";
  emit("clear");
}
</script>
