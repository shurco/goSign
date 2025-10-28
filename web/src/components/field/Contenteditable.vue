<template>
  <div class="group/contenteditable relative overflow-visible" :class="{ 'flex items-center': !iconInline }">
    <span
      ref="contenteditableRef"
      dir="auto"
      :contenteditable="editable"
      style="min-width: 2px"
      :class="iconInline ? 'inline' : 'block'"
      class="peer outline-none focus:block"
      @keydown.enter.prevent="blurContenteditable"
      @focus="$emit('focus', $event)"
      @blur="onBlur"
    >
      {{ value }}
    </span>
    <span v-if="withRequired" title="Required" class="text-red-500 peer-focus:hidden" @click="focusContenteditable()"
      >*</span
    >
    <SvgIcon
      name="pencil"
      class="flex-none cursor-pointer align-middle opacity-0 group-hover/contenteditable:opacity-100 group-hover/contenteditable-container:opacity-100 peer-focus:hidden"
      :style="iconInline ? {} : { right: -(1.1 * iconWidth) + 'px' }"
      title="Edit"
      :class="{
        invisible: !editable,
        'ml-1': !withRequired,
        absolute: !iconInline,
        'inline align-bottom': iconInline
      }"
      :width="iconWidth"
      :height="iconWidth"
      :stroke-width="iconStrokeWidth"
      @click="[focusContenteditable(), selectOnEditClick && selectContent()]"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";

const props = defineProps({
  modelValue: {
    type: String,
    required: false,
    default: ""
  },
  iconInline: {
    type: Boolean,
    required: false,
    default: false
  },
  iconWidth: {
    type: Number,
    required: false,
    default: 30
  },
  withRequired: {
    type: Boolean,
    required: false,
    default: false
  },
  selectOnEditClick: {
    type: Boolean,
    required: false,
    default: false
  },
  editable: {
    type: Boolean,
    required: false,
    default: true
  },
  iconStrokeWidth: {
    type: Number,
    required: false,
    default: 2
  }
});

const emit = defineEmits(["update:model-value", "focus", "blur"]);

const value = ref(props.modelValue);
const contenteditableRef: any = ref();

watch(
  () => props.modelValue,
  (newValue) => {
    value.value = newValue;
  },
  { immediate: true }
);

function selectContent(): void {
  setTimeout(() => {
    const el = contenteditableRef.value;
    if (!el) return;

    const range = document.createRange();
    range.selectNodeContents(el);
    const sel = window.getSelection();
    sel?.removeAllRanges();
    sel?.addRange(range);
  }, 10);
}

function onBlur(e: any): void {
  setTimeout(() => {
    if (contenteditableRef.value) {
      value.value = contenteditableRef.value.innerText.trim() || props.modelValue;
      emit("update:model-value", value.value);
    }
    emit("blur", e);
  }, 1);
}

function focusContenteditable(): void {
  contenteditableRef.value?.focus();
}

function blurContenteditable(): void {
  contenteditableRef.value?.blur();
}
</script>
