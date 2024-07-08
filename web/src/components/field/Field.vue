<template>
  <div class="group pb-2">
    <div class="group relative rounded rounded-tr-none border border-[#e7e2df] py-1">
      <div class="group/contenteditable-container relative flex items-center justify-between">
        <div class="absolute bottom-0 left-0 right-0 top-0 cursor-pointer" @click="scrollToFirstArea()" />

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
            @focus="[onNameFocus(), scrollToFirstArea()]"
            @blur="onNameBlur"
          />
        </div>

        <div v-if="isNameFocus" class="relative flex items-center">
          <template v-if="field.type != 'phone'">
            <input
              :id="`required-checkbox-${field.id}`"
              v-model="field.required"
              type="checkbox"
              class="checkbox checkbox-xs no-animation rounded"
              @mousedown.prevent
            />
            <label
              :for="`required-checkbox-${field.id}`"
              class="label text-xs"
              @click.prevent="field.required = !field.required"
              @mousedown.prevent
              >Required</label
            >
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

          <span
            v-else
            class="dropdown dropdown-end"
            @mouseenter="renderDropdown = true"
            @touchstart="renderDropdown = true"
          >
            <label tabindex="0" title="Settings" class="cursor-pointer text-transparent group-hover:text-[#291334]">
              <SvgIcon name="settings" width="18" height="18" />
            </label>

            <ul
              v-if="renderDropdown"
              tabindex="0"
              class="dropdown-content menu menu-xs rounded-box z-10 mt-1.5 w-52 bg-[#faf7f5] p-2 shadow"
              draggable="true"
              @dragstart.prevent.stop
              @click="closeDropdown()"
            >
              <div v-if="field.type === 'text' && !defaultField" class="relative px-1 py-1.5" @click.stop>
                <input
                  v-model="field.default_value"
                  type="text"
                  placeholder="Default value"
                  dir="auto"
                  class="input input-bordered input-xs h-7 w-full max-w-xs !outline-0"
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
                  class="select select-bordered select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                  @change="save"
                >
                  <option v-for="format in dateFormats" :key="format" :value="format">
                    {{ formatDate(new Date(), format) }}
                  </option>
                </select>
                <label class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px"> Format </label>
              </div>
              <li v-if="field.type != 'phone'" @click.stop>
                <label class="cursor-pointer py-1.5">
                  <input v-model="field.required" type="checkbox" class="toggle toggle-xs" @update:model-value="save" />
                  <span class="label-text">Required</span>
                </label>
              </li>
              <li v-if="field.type === 'text' && !defaultField" @click.stop>
                <label class="cursor-pointer py-1.5">
                  <input v-model="field.readonly" type="checkbox" class="toggle toggle-xs" @update:model-value="save" />
                  <span class="label-text">Read only</span>
                </label>
              </li>
              <hr class="mt-0.5 pb-0.5" />
              <li v-for="(area, index) in field.areas || []" :key="index">
                <a href="#" class="px-2 py-1 text-sm" @click.prevent="emit('scroll-to', area)">
                  <SvgIcon name="shape" width="18" height="18" />
                  Page {{ area.page + 1 }}
                </a>
              </li>
              <li v-if="!field.areas?.length || !['radio', 'multiple'].includes(field.type)">
                <a href="#" class="px-2 py-1 text-sm" @click.prevent="emit('set-draw', { field })">
                  <SvgIcon name="section" width="18" height="18" />
                  Draw new area
                </a>
              </li>
              <li
                v-if="
                  field.areas?.length === 1 && ['date', 'signature', 'initials', 'text', 'cells'].includes(field.type)
                "
              >
                <a href="#" class="px-2 py-1 text-sm" @click.prevent="copyToAllPages(field)">
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
              class="input input-primary input-xs -mr-6 w-full bg-transparent !pr-7 text-sm"
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
            class="input input-primary input-xs w-full bg-transparent text-sm"
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
  </div>
</template>

<script setup lang="ts">
import { computed, inject, nextTick, ref, watch } from "vue";
import Contenteditable from "@/components/field/Contenteditable.vue";
import FieldType from "@/components/field/Type.vue";
import { fieldNames as fieldNamesConst } from "@/components/field/constants.ts";
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
  }
});

const emit = defineEmits(["set-draw", "remove", "scroll-to"]);

const template: any = inject("template");
const save: any = inject("save");
const selectedAreaRef: any = inject("selectedAreaRef");

const nameRef: any = ref(null);
const optionsRef: any = ref(null);
const isNameFocus = ref(false);
const renderDropdown = ref(false);

const fieldNames: any = computed(() => fieldNamesConst);
const areas = computed(() => props.field.areas || []);
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

const defaultName = computed(() => {
  if (props.field.type === "payment" && props.field.preferences?.price) {
    const { price, currency } = props.field.preferences || {};
    const formattedPrice = new Intl.NumberFormat([], {
      style: "currency",
      currency
    }).format(price);
    return `${fieldNames[props.field.type]} ${formattedPrice}`;
  } else {
    const typeIndex = template.value.fields.filter((f: any) => f.type === props.field.type).indexOf(props.field);
    const suffix: any = { multiple: "Select", radio: "Group" }[props.field.type] || "Field";
    return `${fieldNames[props.field.type]} ${suffix} ${typeIndex + 1}`;
  }
});

watch(
  () => props.field.type,
  (newType) => {
    props.field.preferences ||= {};

    if (newType === "date") {
      props.field.preferences.format ||= Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US")
        ? "MM/DD/YYYY"
        : "DD/MM/YYYY";
    }
  },
  { immediate: true }
);

// Methods should be converted into standalone functions or use `use` functions if they are composable.
function formatDate(date: any, format: any): any {
  const monthFormats = {
    M: "numeric",
    MM: "2-digit",
    MMM: "short",
    MMMM: "long"
  };

  const dayFormats = {
    D: "numeric",
    DD: "2-digit"
  };

  const yearFormats = {
    YYYY: "numeric",
    YY: "2-digit"
  };

  const parts = new Intl.DateTimeFormat([], {
    day: dayFormats[format.match(/D+/)],
    month: monthFormats[format.match(/M+/)],
    year: yearFormats[format.match(/Y+/)]
  }).formatToParts(date);

  return format
    .replace(/D+/, parts.find((p) => p.type === "day").value)
    .replace(/M+/, parts.find((p) => p.type === "month").value)
    .replace(/Y+/, parts.find((p) => p.type === "year").value);
}

function copyToAllPages(field: any): any {
  const areaString = JSON.stringify(field.areas[0]);
  template.value.documents.forEach((attachment: any) => {
    attachment.preview_images.forEach((page: any) => {
      if (
        !field.areas.find((area: any) => area.attachment_id === attachment.id && area.page === parseInt(page.filename))
      ) {
        field.areas.push({
          ...JSON.parse(areaString),
          attachment_id: attachment.id,
          page: parseInt(page.filename)
        });
      }
    });
  });

  nextTick(() => {
    emit("scroll-to", field.areas[field.areas.length - 1]);
  });
  save;
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

function closeDropdown(): void {
  renderDropdown.value = false;
  const activeElement = document.activeElement as HTMLElement;
  if (activeElement !== null) {
    activeElement.blur();
  }
}

function addOption(): void {
  props.field.options.push({ value: "", id: v4() });
  nextTick(() => {
    const inputs = optionsRef.value.querySelectorAll("input");
    inputs[inputs.length - 1]?.focus();
  });
  save;
}

function removeOption(option: any): void {
  props.field.options.splice(props.field.options.indexOf(option), 1);
  props.field.areas.splice(
    props.field.areas.findIndex((a: any) => a.option_id === option.id),
    1
  );
  save;
}

function maybeUpdateOptions(): void {
  delete props.field.default_value;
  if (!["radio", "multiple", "select"].includes(props.field.type)) {
    delete props.field.options;
  }
  if (["radio", "multiple", "select"].includes(props.field.type)) {
    props.field.options ||= [{ value: "", id: v4() }];
  }
  (props.field.areas || []).forEach((area: any) => {
    if (props.field.type === "cells") {
      area.cell_w = (area.w * 2) / Math.floor(area.w / area.h);
    } else {
      delete area.cell_w;
    }
  });
}

function onNameBlur(e: any): void {
  const text = nameRef.value.$refs.contenteditable.innerText.trim();
  if (text) {
    props.field.name = text;
  } else {
    props.field.name = "";
    nameRef.value.$refs.contenteditable.innerText = defaultName;
  }
  isNameFocus.value = false;
  save;
}

function removeArea(area: any): void {
  props.field.areas.splice(props.field.areas.indexOf(area), 1);
  save;
}
</script>
