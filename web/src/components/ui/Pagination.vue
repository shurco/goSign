<template>
  <div class="flex items-center justify-between">
    <div v-if="showInfo" class="text-sm text-gray-600">
      Showing {{ startIndex + 1 }} to {{ Math.min(endIndex, total) }} of {{ total }} entries
    </div>
    <div class="flex items-center gap-1">
      <Button
        variant="ghost"
        size="sm"
        :disabled="currentPage === 1"
        @click="emit('update:currentPage', currentPage - 1)"
      >
        «
      </Button>
      <Button
        v-for="page in visiblePages"
        :key="page"
        :variant="page === currentPage ? 'primary' : 'ghost'"
        size="sm"
        @click="emit('update:currentPage', page)"
      >
        {{ page }}
      </Button>
      <Button
        variant="ghost"
        size="sm"
        :disabled="currentPage === totalPages"
        @click="emit('update:currentPage', currentPage + 1)"
      >
        »
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import Button from "./Button.vue";

interface Props {
  currentPage: number;
  pageSize: number;
  total: number;
  maxVisible?: number;
  showInfo?: boolean;
}

interface Emits {
  (e: "update:currentPage", page: number): void;
}

const props = withDefaults(defineProps<Props>(), {
  maxVisible: 5,
  showInfo: true
});

const emit = defineEmits<Emits>();

const totalPages = computed(() => Math.ceil(props.total / props.pageSize));
const startIndex = computed(() => (props.currentPage - 1) * props.pageSize);
const endIndex = computed(() => startIndex.value + props.pageSize);

const visiblePages = computed(() => {
  const pages: number[] = [];
  let start = Math.max(1, props.currentPage - Math.floor(props.maxVisible / 2));
  let end = Math.min(totalPages.value, start + props.maxVisible - 1);

  if (end - start + 1 < props.maxVisible) {
    start = Math.max(1, end - props.maxVisible + 1);
  }

  for (let i = start; i <= end; i++) {
    pages.push(i);
  }

  return pages;
});
</script>
