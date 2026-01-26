<template>
  <div
    ref="drawerEl"
    tabindex="-1"
    class="field-form-drawer pointer-events-none fixed inset-x-0 bottom-0 z-50 flex flex-col outline-none ring-0 focus:outline-none focus:ring-0 focus-visible:outline-none focus-visible:ring-0"
    role="region"
    aria-label="Signing actions"
    @keydown="onKeydown"
  >
    <div class="field-form-drawer__panel relative flex flex-col">
      <div class="container mx-auto w-full max-w-4xl px-4 pb-safe">
        <div class="field-form-drawer__card pointer-events-auto overflow-hidden rounded-t-lg border border-b-0 border-[var(--color-base-300)] bg-white shadow-[0_-2px_10px_rgba(0,0,0,0.06)] dark:border-neutral-600 dark:bg-neutral-800">
          <!-- Expandable: dots + form (no handle, no extra nav row) -->
          <Transition name="drawer-expand">
            <div v-show="isOpen" class="flex flex-col max-h-[45vh] overflow-y-auto">
              <!-- Current field form -->
              <template v-if="field">
                <div class="border-t border-[var(--color-base-200)] dark:border-neutral-600 px-3 py-3 first:border-t-0">
                  <div class="mx-auto max-w-md">
                    <label class="mb-1.5 block text-xs font-medium text-[--color-base-content]/80">
                      {{ getFieldLabel(field) }}
                      <span v-if="fieldStates[field.id]?.required || field.required" class="text-error">*</span>
                    </label>
                    <FieldInput
                      :model-value="modelValue"
                      :type="field.type as any"
                      :required="fieldStates[field.id]?.required || field.required"
                      :readonly="field.readonly"
                      :disabled="fieldStates[field.id]?.disabled"
                      :options="field.options"
                      :placeholder="getFieldLabel(field)"
                      :error="fieldErrors[field.id]"
                      :formula="(field as any).formula ?? field.preferences?.formula"
                      :calculation-type="(field as any).calculationType as 'number' | 'currency' | undefined"
                      :calculated-value="calculatedValues[field.id]"
                      :cell-count="getCellCount(field)"
                      :price="field.preferences?.price"
                      :currency="field.preferences?.currency"
                      :date-format="field.type === 'date' ? (field.preferences as { format?: string })?.format : undefined"
                      :signature-format="getSignatureFormat(field)"
                      @update:model-value="(v) => onUpdate(field!.id, v)"
                      @blur="onBlur(field!)"
                    />
                    <p
                      v-if="hasWithSignatureId(field) && isFieldFilled(field) && signatureIds[field.id]"
                      class="mt-1.5 text-[11px] text-[--color-base-content]/60"
                    >
                      {{ field.type === 'stamp' ? t('signing.stampId') : t('signing.signatureId') }}: <span class="font-mono">{{ signatureIds[field.id] }}</span>
                    </p>
                  </div>
                </div>
              </template>
            </div>
          </Transition>

          <!-- Action bar: compact, always visible -->
          <div class="flex flex-shrink-0 flex-wrap items-center justify-center gap-2 border-t border-[var(--color-base-200)] dark:border-neutral-600 px-3 py-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              class="min-w-0"
              :disabled="prevUnfilledIndex < 0"
              @click="onNavigate('prev')"
            >
              {{ t('common.previous') }}
            </Button>
            <div class="flex flex-1 justify-center min-w-0 py-0.5">
              <FieldProgressDots
                :fields="allFields"
                :filled-field-ids="filledFieldIds"
                :current-field-id="field?.id ?? null"
                :get-field-label="getFieldLabel"
                @field-select="onFieldSelect"
              />
            </div>
            <template v-if="isFormValid">
              <Button type="button" variant="ghost" size="sm" class="min-w-0" :disabled="isSubmitting" @click="onReset">
                {{ t('common.reset') }}
              </Button>
              <Button
                type="button"
                variant="primary"
                size="sm"
                class="min-w-0"
                :loading="isSubmitting"
                :disabled="isSubmitting"
                @click="onSubmit"
              >
                {{ t('signing.sign') }}
              </Button>
            </template>
            <Button
              type="button"
              variant="outline"
              size="sm"
              class="min-w-0"
              :disabled="nextUnfilledIndex < 0"
              @click="onNavigate('next')"
            >
              {{ t('common.next') }}
            </Button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import FieldInput from "@/components/common/FieldInput.vue";
import FieldProgressDots from "@/components/common/FieldProgressDots.vue";
import Button from "@/components/ui/Button.vue";
import type { Field } from "@/models/template";

const { t } = useI18n();
const drawerEl = ref<HTMLElement | null>(null);

const props = defineProps<{
  isOpen: boolean;
  field: Field | null;
  modelValue: any;
  allFields: Field[];
  filledFieldIds: string[];
  fieldStates: Record<string, { visible?: boolean; required?: boolean; disabled?: boolean }>;
  fieldErrors: Record<string, string>;
  calculatedValues: Record<string, number | undefined>;
  signatureIds: Record<string, string>;
  getFieldLabel: (field: Field) => string;
  getCellCount: (field: Field) => number;
  getSignatureFormat: (field: Field) => string;
  hasWithSignatureId: (field: Field) => boolean;
  isFieldFilled: (field: Field) => boolean;
  canGoPrev: boolean;
  canGoNext: boolean;
  isFormValid: boolean;
  isSubmitting: boolean;
  prevUnfilledIndex: number;
  nextUnfilledIndex: number;
}>();

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      nextTick(() => {
        drawerEl.value?.focus({ preventScroll: true });
      });
    }
  }
);

const emit = defineEmits<{
  "update:modelValue": [value: any];
  close: [];
  navigate: [direction: "prev" | "next"];
  fieldSelect: [fieldId: string];
  blur: [field: Field];
  reset: [];
  submit: [];
}>();

function onUpdate(fieldId: string, value: any): void {
  emit("update:modelValue", value);
}

function onClose(): void {
  emit("close");
}

function onNavigate(direction: "prev" | "next"): void {
  emit("navigate", direction);
}

function onFieldSelect(fieldId: string): void {
  emit("fieldSelect", fieldId);
}

function onBlur(field: Field): void {
  emit("blur", field);
}

function onKeydown(e: KeyboardEvent): void {
  if (e.key === "Escape") {
    onClose();
  }
}

function onReset(): void {
  emit("reset");
}

function onSubmit(): void {
  emit("submit");
}
</script>

<style scoped>
.field-form-drawer:focus,
.field-form-drawer:focus-visible {
  outline: none;
  box-shadow: none;
}
.field-form-drawer__panel {
  background: transparent;
}
.field-form-drawer__card {
  min-height: 0;
}
.drawer-expand-enter-active,
.drawer-expand-leave-active {
  transition: opacity 0.2s ease, max-height 0.3s cubic-bezier(0.32, 0.72, 0, 1);
  overflow: hidden;
}
.drawer-expand-enter-from,
.drawer-expand-leave-to {
  opacity: 0;
  max-height: 0;
}
.drawer-expand-enter-to,
.drawer-expand-leave-from {
  opacity: 1;
  max-height: 320px;
}
.pb-safe {
  padding-bottom: env(safe-area-inset-bottom, 0);
}
</style>
