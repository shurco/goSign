<template>
  <svg aria-hidden="true" :class="svgClass">
    <use :xlink:href="symbolId" />
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  prefix?: string;
  name: string;
  svgClass?: string;
}

const props = withDefaults(defineProps<Props>(), {
  prefix: "icon",
  svgClass: ""
});

const symbolId = computed(() => `#${props.prefix}-${props.name}`);
</script>
