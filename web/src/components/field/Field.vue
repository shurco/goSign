<template>
  <div class="group pb-2">
    <div
      class="group relative rounded rounded-tr-none border py-1"
      :class="[isSelected ? borderColors[submitterIndex] : 'border-[#e7e2df]', !field.required ? 'border-dashed' : '']"
    >
      <div class="group/contenteditable-container relative flex items-center justify-between" @focusout="onFocusOut">
        <div class="absolute top-0 right-0 bottom-0 left-0 cursor-pointer" @click="scrollToFirstArea()" />

        <div class="flex items-center space-x-1 p-1">
          <FieldType
            v-model="field.type"
            :editable="editable && !defaultField"
            :button-width="20"
            @update:model-value="[maybeUpdateOptions(), save]"
            @click="scrollToFirstArea()"
          />
          <Contenteditable
            ref="nameRef"
            :model-value="field.name || defaultName"
            :editable="editable && !defaultField"
            :icon-inline="true"
            :icon-width="18"
            :icon-stroke-width="1.6"
            :select-on-edit-click="true"
            @focus="[onNameFocus(), scrollToFirstArea()]"
            @blur="onNameBlur"
          />
        </div>

        <div v-if="isNameFocus" class="relative flex items-center gap-1.5 pr-2">
          <template v-if="field.type != 'phone'">
            <input
              :id="`required-checkbox-${field.id}`"
              v-model="field.required"
              type="checkbox"
              class="toggle toggle-xs"
              @mousedown.prevent
              @change="save"
            />
          </template>
        </div>

        <div v-else-if="editable" class="flex items-center space-x-1">
          <button
            v-if="field && !field.areas.length"
            title="Draw"
            class="relative cursor-pointer text-transparent group-hover:text-[#291334]"
            @click="emit('set-draw', { field })"
          >
            <SvgIcon name="section" width="18" height="18" />
          </button>

          <span v-else ref="dropdownRef" class="dropdown dropdown-end">
            <label
              tabindex="0"
              title="Settings"
              class="cursor-pointer text-transparent group-hover:text-[#291334]"
              @click="renderDropdown = !renderDropdown"
            >
              <SvgIcon name="settings" width="18" height="18" />
            </label>

            <ul
              v-if="renderDropdown"
              tabindex="0"
              class="dropdown-content menu menu-xs rounded-box z-10 mt-1.5 w-52 bg-[#faf7f5] p-2 shadow"
              draggable="true"
              @dragstart.prevent.stop
            >
              <div v-if="field.type === 'text' && !defaultField" class="relative px-1 py-1.5" @click.stop>
                <input
                  v-model="field.default_value"
                  type="text"
                  placeholder="Default value"
                  dir="auto"
                  class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                  @blur="save"
                />
                <label v-if="field.default_value" class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">
                  Default value
                </label>
              </div>
              <div v-if="field.type === 'date'" class="relative px-1 py-1.5" @click.stop>
                <select
                  v-model="field.preferences.format"
                  placeholder="Format"
                  class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                  @change="save"
                >
                  <option v-for="format in dateFormats" :key="format" :value="format">
                    {{ formatDate(new Date(), format) }}
                  </option>
                </select>
                <label class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px"> Format </label>
              </div>
              <li v-if="field.type === 'number' && !defaultField" class="px-2 py-1" @click.stop>
                <div class="space-y-1">
                  <label class="label text-xs py-0">
                    <span>Number format</span>
                  </label>
                  <select
                    v-model="field.preferences.format"
                    class="select select-xs w-full"
                    @change="[ensurePreferences(), save()]"
                  >
                    <option value="">None</option>
                    <option value="comma">1,000.00 (comma)</option>
                    <option value="dot">1.000,00 (dot)</option>
                    <option value="space">1 000,00 (space)</option>
                    <option value="usd">$1,000.00 (USD)</option>
                    <option value="eur">€1.000,00 (EUR)</option>
                    <option value="gbp">£1,000.00 (GBP)</option>
                  </select>
                  <label class="label text-xs py-0 mt-1"><span>Min</span></label>
                  <input
                    v-model.number="field.validation.min"
                    type="number"
                    class="input input-xs w-full"
                    @change="[ensureValidation(), save()]"
                  />
                  <label class="label text-xs py-0 mt-1"><span>Max</span></label>
                  <input
                    v-model.number="field.validation.max"
                    type="number"
                    class="input input-xs w-full"
                    @change="[ensureValidation(), save()]"
                  />
                  <label class="label text-xs py-0 mt-1"><span>Step</span></label>
                  <input
                    v-model="field.validation.step"
                    type="text"
                    placeholder="any"
                    class="input input-xs w-full"
                    @change="[ensureValidation(), save()]"
                  />
                </div>
              </li>
              <li v-if="field.type === 'signature' && !defaultField" class="px-2 py-1" @click.stop>
                <div class="space-y-1">
                  <label class="label text-xs py-0"><span>Signature format</span></label>
                  <select
                    v-model="signatureFormat"
                    class="select select-xs w-full"
                    @change="[ensurePreferences(), save()]"
                  >
                    <option value="any">Any</option>
                    <option value="drawn">Drawn</option>
                    <option value="typed">Typed</option>
                    <option value="drawn_or_typed">Drawn or typed</option>
                    <option value="drawn_or_upload">Drawn or upload</option>
                    <option value="upload">Upload</option>
                  </select>
                  <label class="flex cursor-pointer items-center gap-2 py-1">
                    <input
                      v-model="field.preferences.with_signature_id"
                      type="checkbox"
                      class="toggle toggle-xs"
                      @change="[ensurePreferences(), save()]"
                    />
                    <span class="label-text text-xs">With signature ID</span>
                  </label>
                </div>
              </li>
              <li v-if="field.type === 'payment' && editable && !defaultField" class="px-2 py-1" @click.stop>
                <div class="space-y-1">
                  <label class="label text-xs py-0"><span>Price</span></label>
                  <input
                    v-model.number="field.preferences.price"
                    type="number"
                    step="0.01"
                    min="0"
                    class="input input-xs w-full"
                    @change="[ensurePreferences(), save()]"
                  />
                  <label class="label text-xs py-0 mt-1"><span>Currency</span></label>
                  <select
                    v-model="field.preferences.currency"
                    class="select select-xs w-full"
                    @change="[ensurePreferences(), save()]"
                  >
                    <option value="USD">USD</option>
                    <option value="EUR">EUR</option>
                    <option value="GBP">GBP</option>
                    <option value="JPY">JPY</option>
                    <option value="RUB">RUB</option>
                  </select>
                </div>
              </li>
              <li v-if="field.type === 'stamp' && !defaultField" @click.stop>
                <div class="space-y-1">
                  <label class="flex cursor-pointer items-center gap-2 py-1">
                    <input
                      v-model="field.preferences.with_logo"
                      type="checkbox"
                      class="toggle toggle-xs"
                      @change="[ensurePreferences(), save()]"
                    />
                    <span class="label-text text-xs">With logo</span>
                  </label>
                  <label class="flex cursor-pointer items-center gap-2 py-1">
                    <input
                      v-model="field.preferences.with_signature_id"
                      type="checkbox"
                      class="toggle toggle-xs"
                      @change="[ensurePreferences(), save()]"
                    />
                    <span class="label-text text-xs">With stamp ID</span>
                  </label>
                </div>
              </li>
              <li v-if="['text', 'cells'].includes(field.type) && !defaultField" class="px-2 py-1" @click.stop>
                <div class="space-y-1">
                  <label class="label text-xs py-0"><span>Validation</span></label>
                  <select
                    v-model="validationType"
                    class="select select-xs w-full"
                    @change="applyValidationPreset"
                  >
                    <option value="">None</option>
                    <option value="length">Length</option>
                    <option value="email">Email</option>
                    <option value="ssn">SSN</option>
                    <option value="ein">EIN</option>
                    <option value="url">URL</option>
                    <option value="zip">ZIP</option>
                    <option value="numbers_only">Numbers only</option>
                    <option value="letters_only">Letters only</option>
                    <option value="custom">Custom pattern</option>
                  </select>
                  <template v-if="validationType === 'length'">
                    <label class="label text-xs py-0 mt-1"><span>Min length</span></label>
                    <input
                      v-model.number="field.validation.min"
                      type="number"
                      class="input input-xs w-full"
                      @change="[ensureValidation(), save()]"
                    />
                    <label class="label text-xs py-0 mt-1"><span>Max length</span></label>
                    <input
                      v-model.number="field.validation.max"
                      type="number"
                      class="input input-xs w-full"
                      @change="[ensureValidation(), save()]"
                    />
                  </template>
                  <template v-if="validationType === 'custom'">
                    <label class="label text-xs py-0 mt-1"><span>Pattern</span></label>
                    <input
                      v-model="field.validation.pattern"
                      type="text"
                      placeholder="^[0-9]{3}-[0-9]{2}-[0-9]{4}$"
                      class="input input-xs w-full font-mono text-xs"
                      @change="[ensureValidation(), save()]"
                    />
                    <label class="label text-xs py-0 mt-1"><span>Error message</span></label>
                    <input
                      v-model="field.validation.message"
                      type="text"
                      class="input input-xs w-full"
                      @change="[ensureValidation(), save()]"
                    />
                  </template>
                </div>
              </li>
              <li v-if="field.type != 'phone'" class="px-2" @click.stop>
                <label class="flex cursor-pointer items-center gap-2 py-1.5">
                  <input v-model="field.required" type="checkbox" class="toggle toggle-xs" @update:model-value="save" />
                  <span class="label-text">Required</span>
                </label>
              </li>
              <li v-if="(field.type === 'text' || field.type === 'stamp') && !defaultField" class="px-2" @click.stop>
                <label class="flex cursor-pointer items-center gap-2 py-1.5">
                  <input v-model="field.readonly" type="checkbox" class="toggle toggle-xs" @update:model-value="save" />
                  <span class="label-text">Read only</span>
                </label>
              </li>
              <hr v-if="field.type != 'phone'" class="mt-0.5 pb-0.5" />
              <li v-for="(area, index) in field.areas || []" :key="index">
                <a
                  href="#"
                  class="px-2 py-1 text-sm"
                  @click.prevent="
                    emit('scroll-to', area);
                    closeDropdown();
                  "
                >
                  <SvgIcon name="shape" width="18" height="18" />
                  Page {{ area.page + 1 }}
                </a>
              </li>
              <li v-if="!field.areas?.length || !['radio', 'multiple'].includes(field.type)">
                <a
                  href="#"
                  class="px-2 py-1 text-sm"
                  @click.prevent="
                    emit('set-draw', { field });
                    closeDropdown();
                  "
                >
                  <SvgIcon name="section" width="18" height="18" />
                  Draw new area
                </a>
              </li>
              <hr v-if="editable && !defaultField" class="mt-0.5 pb-0.5" />
              <li v-if="editable && !defaultField" @click.stop>
                <a
                  href="#"
                  class="px-2 py-1 text-sm"
                  @click.prevent="openConditionBuilder"
                >
                  <SvgIcon name="settings" width="18" height="18" />
                  Conditional Logic
                </a>
              </li>
              <li v-if="editable && !defaultField && ['number', 'text'].includes(field.type)" @click.stop>
                <a
                  href="#"
                  class="px-2 py-1 text-sm"
                  @click.prevent="openFormulaBuilder"
                >
                  <SvgIcon name="settings" width="18" height="18" />
                  Formula
                </a>
              </li>
              <li
                v-if="
                  field.areas?.length === 1 && ['date', 'signature', 'initials', 'text', 'cells'].includes(field.type)
                "
              >
                <a
                  href="#"
                  class="px-2 py-1 text-sm"
                  @click.prevent="
                    copyToAllPages(field);
                    closeDropdown();
                  "
                >
                  <SvgIcon name="copy" width="18" height="18" />
                  Copy to All Pages
                </a>
              </li>
            </ul>
          </span>
          <button
            class="relative pr-1 text-transparent group-hover:text-[#291334]"
            title="Remove"
            @click="emit('remove', field)"
          >
            <SvgIcon name="trash-x" width="18" height="18" />
          </button>
        </div>
      </div>

      <div
        v-if="field.options"
        ref="optionsRef"
        class="mx-2 space-y-1.5 border-t border-[#e7e2df] pt-2"
        draggable="true"
        @dragstart.prevent.stop
      >
        <div v-for="(option, index) in field.options" :key="option.id" class="flex items-center space-x-1.5">
          <span class="w-3.5 text-sm"> {{ index + 1 }}. </span>
          <div
            v-if="
              ['radio', 'multiple'].includes(field.type) &&
              (index > 0 || field.areas.find((a: any) => a.option_id) || !field.areas.length) &&
              !field.areas.find((a: any) => a.option_id === option.id)
            "
            class="flex w-full items-center"
          >
            <input
              v-model="option.value"
              class="input input-xs input-primary -mr-6 w-full bg-transparent !pr-7 text-sm"
              type="text"
              dir="auto"
              required
              :placeholder="`Option ${index + 1}`"
              @blur="save"
            />
            <button title="Draw" tabindex="-1" @click.prevent="emit('set-draw', { field, option })">
              <SvgIcon name="section" width="18" height="18" />
            </button>
          </div>
          <input
            v-else
            v-model="option.value"
            class="input input-xs input-primary w-full bg-transparent text-sm"
            :placeholder="`Option ${index + 1}`"
            type="text"
            required
            dir="auto"
            @focus="maybeFocusOnOptionArea(option)"
            @blur="save"
          />
          <button class="w-3.5 text-sm" tabindex="-1" @click="removeOption(option)">&times;</button>
        </div>
        <button v-if="field.options" class="w-full pb-1 text-center text-sm" @click="addOption">+ Add option</button>
      </div>
    </div>

    <!-- Condition Builder Modal -->
    <Modal v-model="showConditionBuilder" size="lg">
      <template #header>
        <h3 class="text-lg font-semibold">{{ t('fields.conditions.title') }}</h3>
      </template>
      <template #default>
        <ConditionBuilder
          :field="field"
          :available-fields="availableFieldsForConditions"
          @update:conditions="(conditions) => {
            field.condition_groups = conditions;
            save();
          }"
        />
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm hover:bg-gray-50"
            @click="closeConditionBuilder"
          >
            {{ t('common.close') }}
          </button>
        </div>
      </template>
    </Modal>

    <!-- Formula Builder Modal -->
    <Modal v-model="showFormulaBuilder" size="lg">
      <template #header>
        <h3 class="text-lg font-semibold">{{ t('fields.formula.title') }}</h3>
      </template>
      <template #default>
        <FormulaBuilder
          ref="formulaBuilderRef"
          :field="field"
          :available-fields="availableFieldsForFormula"
        />
      </template>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button
            class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm hover:bg-gray-50"
            @click="closeFormulaBuilder"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            class="rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-700"
            @click="applyFormulaAndClose"
          >
            {{ t('common.save') }}
          </button>
        </div>
      </template>
    </Modal>
  </div>
</template>

<script setup lang="ts">
// This component intentionally mutates props.field because field is part of a reactive template object
// managed by the parent component. Direct mutation is used for performance optimization.
// Rule vue/no-mutating-props is disabled for this file in eslint.config.mjs
import { computed, inject, nextTick, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import Contenteditable from "@/components/field/Contenteditable.vue";
import FieldType from "@/components/field/Type.vue";
import ConditionBuilder from "@/components/field/ConditionBuilder.vue";
import FormulaBuilder from "@/components/field/FormulaBuilder.vue";
import Modal from "@/components/ui/Modal.vue";
import { borderColors, fieldNames as fieldNamesConst, subNames } from "@/components/field/constants.ts";
import { useDropdown } from "@/composables/ui";
import { v4 } from "uuid";

const props = defineProps({
  field: {
    type: Object,
    required: true
  },
  defaultField: {
    type: Object,
    required: false,
    default: null
  },
  editable: {
    type: Boolean,
    required: false,
    default: true
  },
  isSelected: {
    type: Boolean,
    required: false,
    default: false
  }
});

const emit = defineEmits(["set-draw", "remove", "scroll-to"]);

const { t } = useI18n();
const template: any = inject("template");
const save: any = inject("save");
const selectedAreaRef: any = inject("selectedAreaRef");

const nameRef: any = ref(null);
const optionsRef: any = ref(null);
const dropdownRef = ref<HTMLElement | null>(null);
const formulaBuilderRef = ref<any>(null);
const isNameFocus = ref(false);
const showConditionBuilder = ref(false);
const showFormulaBuilder = ref(false);

const { isOpen: renderDropdown, close: closeDropdown } = useDropdown(dropdownRef);

function openConditionBuilder(): void {
  showConditionBuilder.value = true;
  nextTick(() => closeDropdown());
}

function openFormulaBuilder(): void {
  showFormulaBuilder.value = true;
  nextTick(() => closeDropdown());
}

function closeConditionBuilder(): void {
  showConditionBuilder.value = false;
}

function closeFormulaBuilder(): void {
  showFormulaBuilder.value = false;
}

function applyFormulaAndClose(): void {
  if (formulaBuilderRef.value?.getFormula) {
    props.field.formula = formulaBuilderRef.value.getFormula();
    if (props.field.formula && !props.field.calculationType) {
      props.field.calculationType = "number";
    }
    save();
    showFormulaBuilder.value = false;
  }
}

const validationType = ref("");
const validationPresets: Record<string, { pattern: string; message: string }> = {
  email: { pattern: "^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$", message: "Please enter a valid email address" },
  ssn: { pattern: "^[0-9]{3}-[0-9]{2}-[0-9]{4}$", message: "Please enter a valid SSN (XXX-XX-XXXX)" },
  ein: { pattern: "^[0-9]{2}-[0-9]{7}$", message: "Please enter a valid EIN (XX-XXXXXXX)" },
  url: { pattern: "^https?:\\/\\/.+", message: "Please enter a valid URL" },
  zip: { pattern: "^[0-9]{5}(-[0-9]{4})?$", message: "Please enter a valid ZIP code" },
  numbers_only: { pattern: "^[0-9]+$", message: "Please enter numbers only" },
  letters_only: { pattern: "^[a-zA-Z]+$", message: "Please enter letters only" }
};

function applyValidationPreset(): void {
  ensureValidation();
  if (validationType.value && validationType.value !== "length" && validationType.value !== "custom") {
    const preset = validationPresets[validationType.value];
    if (preset) {
      props.field.validation.pattern = preset.pattern;
      props.field.validation.message = preset.message;
    }
  }
  save();
}

const fieldNames: any = computed(() => fieldNamesConst);
const submitterIndex = computed(() => {
  return template.value.submitters.findIndex((s: any) => s.id === props.field.submitter_id);
});

// Normalize signature format for select: backend omitempty can omit format when empty, so show "any"
const signatureFormat = computed({
  get: () => {
    const v = props.field.preferences?.format;
    return typeof v === "string" && v !== "" ? v : "any";
  },
  set: (value: string) => {
    ensurePreferences();
    props.field.preferences!.format = value;
  }
});
const dateFormats = computed(() => [
  "MM/DD/YYYY",
  "DD/MM/YYYY",
  "YYYY-MM-DD",
  "DD-MM-YYYY",
  "DD.MM.YYYY",
  "MMM D, YYYY",
  "MMMM D, YYYY",
  "D MMM YYYY",
  "D MMMM YYYY"
]);

// Generate default field name for any field (used in conditions/formula dropdowns)
function getDefaultFieldName(f: any): string {
  if (!template?.value?.fields) return f.id || "Field";
  if (f.type === "payment" && f.preferences?.price) {
    const { price, currency } = f.preferences || {};
    const formattedPrice = new Intl.NumberFormat([], { style: "currency", currency }).format(price);
    return `${fieldNames.value[f.type]} ${formattedPrice}`;
  }
  const idx = template.value.submitters?.findIndex((s: any) => s.id === f.submitter_id) ?? 0;
  const partyName = subNames[idx]?.replace(" Party", "") || "First";
  const typeName = fieldNames.value[f.type] || "Field";
  const sameTypeAndPartyFields = template.value.fields.filter(
    (other: any) => other.type === f.type && other.submitter_id === f.submitter_id && other.id !== f.id
  );
  const fieldNumber = sameTypeAndPartyFields.length + 1;
  return `${partyName} ${typeName} ${fieldNumber}`;
}

// Generate default field name based on party, type, and number
const defaultName = computed(() => getDefaultFieldName(props.field));

// Available fields with display names for Condition/Formula modals
const availableFieldsForConditions = computed(() => {
  const list = (template?.value?.fields ?? []).filter(
    (f: any) => f.id !== props.field.id && f.submitter_id === props.field.submitter_id
  );
  return list.map((f: any) => ({ ...f, displayName: (f.name && String(f.name).trim()) || getDefaultFieldName(f) }));
});

const availableFieldsForFormula = computed(() => {
  const list = (template?.value?.fields ?? []).filter(
    (f: any) =>
      f.id !== props.field.id &&
      f.submitter_id === props.field.submitter_id &&
      ["number", "text"].includes(f.type)
  );
  return list.map((f: any) => ({ ...f, displayName: (f.name && String(f.name).trim()) || getDefaultFieldName(f) }));
});

// Check if current field name matches the default pattern
function isDefaultName(name: string): boolean {
  if (!name) {
    return true;
  }

  // Check if name matches pattern: {Party} {Type} {Number}
  const pattern = /^(First|Second|Third|Fourth|Fifth|Sixth|Seventh|Eighth|Ninth|Tenth)\s+\w+\s+\d+$/;
  return pattern.test(name);
}

// Store previous submitter_id to detect changes
const previousSubmitterId = ref(props.field.submitter_id);

// Watch for submitter changes and update name if it's default
watch(
  () => props.field.submitter_id,
  (newSubmitterId, oldSubmitterId) => {
    if (newSubmitterId !== oldSubmitterId && isDefaultName(props.field.name)) {
      // Update the field name to reflect new party
      props.field.name = "";
      // Name will be shown as defaultName in the template
      save();
    }
    previousSubmitterId.value = newSubmitterId;
  }
);

watch(
  () => props.field.type,
  (newType) => {
    props.field.preferences ||= {};

    if (newType === "date") {
      props.field.preferences.format ||= Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US")
        ? "MM/DD/YYYY"
        : "DD/MM/YYYY";
    }
    if (newType === "payment") {
      if (props.field.preferences.price == null) props.field.preferences.price = 0;
      if (!props.field.preferences.currency) props.field.preferences.currency = "USD";
    }
    if (newType === "number" && (!props.field.validation || typeof props.field.validation !== "object")) {
      props.field.validation = {};
    }
    if (["text", "cells"].includes(newType) && (!props.field.validation || typeof props.field.validation !== "object")) {
      props.field.validation = {};
    }
  },
  { immediate: true }
);

function ensurePreferences(): void {
  if (!props.field.preferences) {
    props.field.preferences = {};
  }
}

function ensureValidation(): void {
  if (!props.field.validation || typeof props.field.validation !== "object") {
    props.field.validation = {};
  }
}

// Methods should be converted into standalone functions or use `use` functions if they are composable.
function formatDate(date: Date, format: string): string {
  const monthFormats: Record<string, "numeric" | "2-digit" | "short" | "long"> = {
    M: "numeric",
    MM: "2-digit",
    MMM: "short",
    MMMM: "long"
  };

  const dayFormats: Record<string, "numeric" | "2-digit"> = {
    D: "numeric",
    DD: "2-digit"
  };

  const yearFormats: Record<string, "numeric" | "2-digit"> = {
    YYYY: "numeric",
    YY: "2-digit"
  };

  const dayMatch = format.match(/D+/);
  const monthMatch = format.match(/M+/);
  const yearMatch = format.match(/Y+/);

  const parts = new Intl.DateTimeFormat([], {
    day: dayMatch ? dayFormats[dayMatch[0]] || "numeric" : "numeric",
    month: monthMatch ? monthFormats[monthMatch[0]] || "numeric" : "numeric",
    year: yearMatch ? yearFormats[yearMatch[0]] || "numeric" : "numeric"
  }).formatToParts(date);

  const dayPart = parts.find((p) => p.type === "day");
  const monthPart = parts.find((p) => p.type === "month");
  const yearPart = parts.find((p) => p.type === "year");

  return format
    .replace(/D+/, dayPart?.value || "")
    .replace(/M+/, monthPart?.value || "")
    .replace(/Y+/, yearPart?.value || "");
}

function copyToAllPages(field: any): any {
  const areaString = JSON.stringify(field.areas[0]);
  template.value.documents.forEach((attachment: any) => {
    attachment.preview_images.forEach((page: any) => {
      if (
        !field.areas.find(
          (area: any) => area.attachment_id === attachment.id && area.page === parseInt(page.filename, 10)
        )
      ) {
        field.areas.push({
          ...JSON.parse(areaString),
          attachment_id: attachment.id,
          page: parseInt(page.filename, 10)
        });
      }
    });
  });

  nextTick(() => {
    emit("scroll-to", field.areas[field.areas.length - 1]);
  });
  save();
}

function onNameFocus(): void {
  isNameFocus.value = true;
  if (!props.field.name) {
    setTimeout(() => {
      nameRef.value.$refs.contenteditable.innerText = " ";
    }, 1);
  }
}

function maybeFocusOnOptionArea(option: any): void {
  const area = props.field.areas.find((a: any) => a.option_id === option.id);
  if (area) {
    selectedAreaRef.value = area;
  }
}

function scrollToFirstArea(): void {
  return props.field.areas?.[0] && emit("scroll-to", props.field.areas[0]);
}

function addOption(): void {
  props.field.options.push({ value: "", id: v4() });
  nextTick(() => {
    const inputs = optionsRef.value.querySelectorAll("input");
    inputs[inputs.length - 1]?.focus();
  });
  save();
}

function removeOption(option: any): void {
  props.field.options.splice(props.field.options.indexOf(option), 1);
  props.field.areas.splice(
    props.field.areas.findIndex((a: any) => a.option_id === option.id),
    1
  );
  save();
}

function getEffectiveCellW(area: { w: number; h: number; cell_w?: number }): number {
  if (area.cell_w != null && area.cell_w > 0) return area.cell_w;
  if (area.w <= 0) return 0;
  if (area.h > 0) {
    const denom = Math.floor(area.w / area.h);
    return denom > 0 ? (area.w * 2) / denom : area.w / 5;
  }
  return area.w / 5;
}

function getCellCountFromArea(area: { w: number; h: number; cell_w?: number }): number {
  const cellWidth = getEffectiveCellW(area);
  if (!cellWidth || cellWidth <= 0 || area.w <= 0) return 0;
  let currentWidth = 0;
  let count = 0;
  while (currentWidth + (cellWidth + cellWidth / 4) < area.w) {
    currentWidth += cellWidth;
    count++;
  }
  return Math.max(count, 1);
}

function maybeUpdateOptions(): void {
  if (props.field.type !== "cells") {
    delete props.field.default_value;
  }
  if (!["radio", "multiple", "select"].includes(props.field.type)) {
    delete props.field.options;
  }
  if (["radio", "multiple", "select"].includes(props.field.type)) {
    props.field.options ||= [{ value: "", id: v4() }];
  }
  (props.field.areas || []).forEach((area: any) => {
    if (props.field.type === "cells") {
      const denom = area.h > 0 ? Math.floor(area.w / area.h) : 0;
      if (denom > 0) {
        area.cell_w = (area.w * 2) / denom;
      } else if (area.w > 0 && (!area.cell_w || area.cell_w <= 0)) {
        area.cell_w = area.w / 5;
      }
      area.cell_count = getCellCountFromArea(area);
    } else {
      delete area.cell_w;
      delete area.cell_count;
    }
  });
}

function onNameBlur(): void {
  const text = nameRef.value.$refs.contenteditable.innerText.trim();
  if (text) {
    props.field.name = text;
  } else {
    props.field.name = "";
    nameRef.value.$refs.contenteditable.innerText = defaultName;
  }
  isNameFocus.value = false;
  save();
}

function onFocusOut(event: FocusEvent): void {
  // Check if focus moved outside the container
  const currentTarget = event.currentTarget as HTMLElement;
  const relatedTarget = event.relatedTarget as HTMLElement;

  if (!currentTarget.contains(relatedTarget)) {
    isNameFocus.value = false;
  }
}
</script>
