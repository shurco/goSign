<template>
  <span class="dropdown" @mouseenter="toggleDropdown(true)" @touchstart="toggleDropdown(true)">
    <slot name="label">
      <label tabindex="0" :title="fieldNames[modelValue]" class="cursor-pointer">
        <SvgIcon
          :name="fieldIcons[modelValue]"
          :width="props.buttonWidth"
          :height="props.buttonWidth"
          :class="props.buttonClasses"
          stroke-width="1.6"
        />
      </label>
    </slot>
    <ul
      v-if="props.editable && renderDropdown"
      tabindex="0"
      class="dropdown-content menu menu-xs rounded-box z-10 mb-3 w-52 p-2 shadow"
      :class="props.menuClasses"
      @click="closeDropdown()"
    >
      <li v-for="(icon, type) in fieldIcons" :key="type">
        <a
          href="#"
          class="flex flex-wrap px-2 py-1 text-sm"
          :class="{ active: type === modelValue }"
          @click.prevent="$emit('update:model-value', type)"
        >
          <SvgIcon :name="icon" stroke-width="1.6" width="20" height="20" />
          {{ fieldNames[type] }}
        </a>
      </li>
    </ul>
  </span>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { fieldIcons, fieldNames } from "@/components/field/constants.ts";

const props = defineProps({
  modelValue: {
    type: String,
    required: true
  },
  menuClasses: {
    type: String,
    default: "mt-1.5 bg-[#faf7f5]"
  },
  buttonClasses: {
    type: String,
    default: ""
  },
  editable: {
    type: Boolean,
    default: true
  },
  buttonWidth: {
    type: Number,
    default: 18
  }
});

const renderDropdown = ref(false);
const emit = defineEmits(["update:model-value"]);

function closeDropdown(): void {
  renderDropdown.value = false;
  const activeElement = document.activeElement as HTMLElement;
  if (activeElement) {
    activeElement.blur();
  }
}

function toggleDropdown(state: boolean): void {
  renderDropdown.value = state;
}
</script>
