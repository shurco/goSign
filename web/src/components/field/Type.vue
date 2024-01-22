<template>
  <span class="dropdown" @mouseenter="renderDropdown = true" @touchstart="renderDropdown = true">
    <slot>
      <label tabindex="0" :title="fieldNames[modelValue]" class="cursor-pointer">
        <SvgIcon :name="fieldIcons[modelValue]" :width="buttonWidth" :height="buttonWidth" :class="buttonClasses" stroke-width="1.6" />
      </label>
    </slot>
    <ul v-if="editable && renderDropdown" tabindex="0" class="dropdown-content menu menu-xs p-2 shadow rounded-box w-52 z-10 mb-3" :class="menuClasses" @click="closeDropdown">
      <template v-for="(icon, type) in fieldIcons" :key="type">
        <li>
          <a href="#" class="text-sm py-1 px-2" :class="{ active: type === modelValue }" @click.prevent="$emit('update:model-value', type)">
            <SvgIcon :name="icon" stroke-width="1.6" width="20" height="20" />
            {{ fieldNames[type] }}
          </a>
        </li>
      </template>
    </ul>
  </span>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { fieldNames, fieldIcons } from "@/components/field/constants.ts";

const props = defineProps({
  modelValue: {
    type: String,
    required: true,
  },
  menuClasses: {
    type: String,
    required: false,
    default: "mt-1.5 bg-[#faf7f5]",
  },
  buttonClasses: {
    type: String,
    required: false,
    default: "",
  },
  editable: {
    type: Boolean,
    required: false,
    default: true,
  },
  buttonWidth: {
    type: Number,
    required: false,
    default: 18,
  },
})

const emit = defineEmits(['update:model-value'])
const renderDropdown = ref(false)

function closeDropdown() {
  const activeElement = document.activeElement as HTMLElement;
  if (activeElement !== null) {
    activeElement.blur();
  }
}
</script>