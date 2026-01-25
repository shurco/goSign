<template>
  <div>
    <h3 class="mb-3 text-sm font-medium text-gray-700">{{ $t('signingMode.title') }}</h3>
    <div class="space-y-3">
      <!-- Mode Selection -->
      <ButtonGroup
        :model-value="signingMode"
        :options="signingModes"
        :disabled="!editable"
        @update:model-value="updateSigningMode"
      />

      <!-- Mode Description -->
      <div class="rounded-md bg-blue-50 p-3 text-sm text-blue-800">
        <p class="mb-1 font-medium">{{ currentMode.title }}</p>
        <p>{{ currentMode.description }}</p>
      </div>

      <!-- Sequential Order Management (hidden when order is handled by parent, e.g. draggable signer cards) -->
      <div v-if="!hideOrderList && signingMode === 'sequential' && submitters.length > 1" class="space-y-3">
        <div class="rounded-md bg-amber-50 p-3">
          <div class="flex items-start gap-2">
            <SvgIcon name="info" width="16" height="16" class="mt-0.5 text-amber-600" />
            <div class="text-sm text-amber-800">
              <p class="mb-1 font-medium">{{ $t('signingMode.orderTitle') }}</p>
              <p>{{ $t('signingMode.orderHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Draggable Submitter Order List -->
        <div class="space-y-2">
          <h4 class="text-sm font-medium text-gray-700">{{ $t('signingMode.signingOrder') }}</h4>
          <div ref="submittersList" class="space-y-2" @dragover.prevent="onDragOver" @drop="onDrop">
            <div
              v-for="(submitter, index) in orderedSubmitters"
              :key="submitter.id"
              :data-submitter-id="submitter.id"
              class="flex cursor-move items-center gap-3 rounded-md border border-gray-200 bg-white p-3 transition-colors hover:bg-gray-50"
              :class="{ 'border-blue-300 bg-blue-50': draggedSubmitter === submitter.id }"
              draggable="true"
              @dragstart="onDragStart($event, submitter)"
              @dragend="onDragEnd"
            >
              <div class="flex items-center gap-2">
                <div
                  class="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-700"
                >
                  {{ index + 1 }}
                </div>
                <div class="h-3 w-3 rounded-full" :class="getSubmitterColor(submitter, index)" />
                <span class="text-sm font-medium text-gray-900">{{ submitter.name }}</span>
              </div>
              <div class="ml-auto">
                <SvgIcon name="drag" width="16" height="16" class="text-gray-400" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { SigningMode } from "@/models";
import ButtonGroup from "@/components/ui/ButtonGroup.vue";
import SvgIcon from "@/components/SvgIcon.vue";

const { t } = useI18n();

interface Props {
  signingMode: SigningMode;
  submitters: any[];
  editable: boolean;
  /** When true, hide the "Signing Order" drag list; order is controlled by parent (e.g. by dragging signer fields). */
  hideOrderList?: boolean;
}

interface Emits {
  "update:signing-mode": [value: SigningMode];
  "update:submitter-order": [orderedSubmitters: any[]];
}

const props = withDefaults(defineProps<Props>(), { hideOrderList: false });
const emit = defineEmits<Emits>();

// Drag state
const draggedSubmitter = ref<string | null>(null);
const submittersList = ref<HTMLElement | null>(null);

// Submitter colors for visual distinction
const submitterColors = [
  "bg-red-500",
  "bg-blue-500",
  "bg-green-500",
  "bg-yellow-500",
  "bg-purple-500",
  "bg-pink-500",
  "bg-indigo-500",
  "bg-orange-500"
];

// Ordered submitters based on the order field
const orderedSubmitters = computed(() => {
  return [...props.submitters].sort((a, b) => (a.order || 0) - (b.order || 0));
});

// Parallel first (default), then Sequential — labels/titles/descriptions from i18n
const signingModes = computed(() => [
  {
    value: "parallel" as SigningMode,
    label: t("signingMode.parallel"),
    icon: "arrows-right-left",
    title: t("signingMode.parallelTitle"),
    description: t("signingMode.parallelDescription")
  },
  {
    value: "sequential" as SigningMode,
    label: t("signingMode.sequential"),
    icon: "arrow-right",
    title: t("signingMode.sequentialTitle"),
    description: t("signingMode.sequentialDescription")
  }
]);

const currentMode = computed(() => {
  return signingModes.value.find((mode) => mode.value === props.signingMode) || signingModes.value[0];
});

const updateSigningMode = (mode: string | number) => {
  if (!props.editable) {
    return;
  }
  // Ensure mode is a valid SigningMode string
  const signingModeValue = typeof mode === "string" ? (mode as SigningMode) : (String(mode) as SigningMode);
  emit("update:signing-mode", signingModeValue);
};

// Drag event handlers for reordering submitters
const onDragStart = (event: DragEvent, submitter: any) => {
  if (!props.editable) {
    event.preventDefault();
    return;
  }
  draggedSubmitter.value = submitter.id;
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", submitter.id);
  }
};

const onDragEnd = () => {
  draggedSubmitter.value = null;
};

const onDragOver = (event: DragEvent) => {
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = "move";
  }
};

const onDrop = (event: DragEvent) => {
  event.preventDefault();

  if (!props.editable || !draggedSubmitter.value) {
    return;
  }

  const targetElement = (event.target as HTMLElement).closest("[data-submitter-id]") as HTMLElement;
  if (!targetElement) {
    return;
  }

  const targetSubmitterId = targetElement.getAttribute("data-submitter-id");
  if (!targetSubmitterId || targetSubmitterId === draggedSubmitter.value) {
    return;
  }

  // Find indices in the ordered list
  const draggedIndex = orderedSubmitters.value.findIndex((s) => s.id === draggedSubmitter.value);
  const targetIndex = orderedSubmitters.value.findIndex((s) => s.id === targetSubmitterId);

  if (draggedIndex === -1 || targetIndex === -1) {
    return;
  }

  // Reorder the submitters array
  const newSubmitters = [...props.submitters];
  const [draggedItem] = newSubmitters.splice(
    newSubmitters.findIndex((s) => s.id === draggedSubmitter.value),
    1
  );

  // Insert at new position
  const targetPos = newSubmitters.findIndex((s) => s.id === targetSubmitterId);
  newSubmitters.splice(targetPos, 0, draggedItem);

  // Update order field
  newSubmitters.forEach((submitter, index) => {
    submitter.order = index;
  });

  // Emit the new order
  emit("update:submitter-order", newSubmitters);

  draggedSubmitter.value = null;
};

// Get color for submitter
const getSubmitterColor = (submitter: any, index: number) => {
  // Use submitter's colorIndex if available, otherwise use index
  const colorIndex = submitter.colorIndex !== undefined ? submitter.colorIndex : index;
  return submitterColors[colorIndex % submitterColors.length];
};
</script>
