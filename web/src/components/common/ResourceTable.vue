<template>
  <div class="resource-table-wrapper">
    <!-- Filters and Search -->
    <div v-if="showFilters" class="mb-4 flex flex-wrap gap-4">
      <Input
        v-if="searchable"
        v-model="searchQuery"
        type="text"
        :placeholder="searchPlaceholder"
        class="min-w-64 flex-1"
        @input="handleSearch"
      />
      <slot name="filters" :filters="filters" :update-filter="updateFilter" />
    </div>

    <!-- Table -->
    <div class="overflow-x-auto rounded-lg border border-[var(--color-base-300)] bg-white">
      <Table :zebra="true">
        <template #header>
          <tr>
            <th v-if="selectable" class="w-12 px-4 py-3">
              <Checkbox :checked="allSelected" size="sm" @change="toggleSelectAll" />
            </th>
            <th
              v-for="column in columns"
              :key="column.key"
              :class="[
                column.headerClass,
                column.sortable && 'cursor-pointer hover:bg-gray-100',
                'px-4 py-3 text-left text-sm font-medium text-gray-700'
              ]"
              @click="column.sortable && handleSort(column.key)"
            >
              <div class="flex items-center gap-2">
                {{ column.label }}
                <span v-if="column.sortable && sortBy === column.key" class="text-xs">
                  {{ sortOrder === "asc" ? "↑" : "↓" }}
                </span>
              </div>
            </th>
            <th v-if="hasActions" class="w-32 px-4 py-3 text-right text-sm font-medium text-gray-700">Actions</th>
          </tr>
        </template>

        <tr v-if="isLoading">
          <td :colspan="totalColumns" class="py-8 text-center">
            <div class="flex flex-col items-center justify-center gap-2">
              <LoadingSpinner size="md" />
              <p class="text-gray-600">Loading...</p>
            </div>
          </td>
        </tr>

        <tr v-else-if="paginatedData.length === 0">
          <td :colspan="totalColumns" class="py-8 text-center text-gray-500">
            {{ emptyMessage }}
          </td>
        </tr>

        <tr v-for="item in paginatedData" :key="getItemId(item)" class="hover:bg-gray-50">
          <td v-if="selectable" class="px-4 py-3">
            <Checkbox :checked="isSelected(item)" size="sm" @change="toggleSelect(item)" />
          </td>
          <td
            v-for="column in columns"
            :key="column.key"
            :class="[column.cellClass, 'px-4 py-3 text-sm text-gray-900']"
          >
            <slot :name="`cell-${column.key}`" :item="item" :value="getNestedValue(item, column.key)">
              {{ formatCellValue(item, column) }}
            </slot>
          </td>
          <td v-if="hasActions" class="px-4 py-3 text-right">
            <slot name="actions" :item="item">
              <div class="flex justify-end gap-2">
                <Button v-if="showEdit" variant="ghost" size="sm" @click="emit('edit', item)">Edit</Button>
                <Button
                  v-if="showDelete"
                  variant="ghost"
                  size="sm"
                  class="text-[var(--color-error)]"
                  @click="emit('delete', item)"
                >
                  Delete
                </Button>
              </div>
            </slot>
          </td>
        </tr>
      </Table>
    </div>

    <!-- Pagination -->
    <Pagination
      v-if="showPagination && totalPages > 1"
      v-model:current-page="currentPage"
      :page-size="pageSize"
      :total="filteredData.length"
      class="mt-4"
      @update:current-page="emit('page-change', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import Input from "@/components/ui/Input.vue";
import Checkbox from "@/components/ui/Checkbox.vue";
import Button from "@/components/ui/Button.vue";
import Table from "@/components/ui/Table.vue";
import LoadingSpinner from "@/components/ui/LoadingSpinner.vue";
import Pagination from "@/components/ui/Pagination.vue";

interface Column {
  key: string;
  label: string;
  sortable?: boolean;
  formatter?: (value: unknown) => string;
  headerClass?: string;
  cellClass?: string;
}

interface Props {
  data: unknown[];
  columns: Column[];
  searchable?: boolean;
  searchPlaceholder?: string;
  searchKeys?: string[];
  selectable?: boolean;
  showFilters?: boolean;
  showPagination?: boolean;
  pageSize?: number;
  isLoading?: boolean;
  emptyMessage?: string;
  hasActions?: boolean;
  showEdit?: boolean;
  showDelete?: boolean;
  idKey?: string;
}

interface Emits {
  (e: "select", selectedItems: unknown[]): void;
  (e: "edit" | "delete", item: unknown): void;
  (e: "page-change", page: number): void;
  (e: "search", query: string): void;
}

const props = withDefaults(defineProps<Props>(), {
  searchable: true,
  searchPlaceholder: "Search...",
  searchKeys: () => [],
  selectable: false,
  showFilters: true,
  showPagination: true,
  pageSize: 10,
  isLoading: false,
  emptyMessage: "No data available",
  hasActions: true,
  showEdit: true,
  showDelete: true,
  idKey: "id"
});

const emit = defineEmits<Emits>();

const searchQuery = ref("");
const currentPage = ref(1);
const sortBy = ref<string>("");
const sortOrder = ref<"asc" | "desc">("asc");
const selectedItems = ref<Set<unknown>>(new Set());
const filters = ref<Record<string, unknown>>({});

const totalColumns = computed(() => {
  let count = props.columns.length;
  if (props.selectable) {
    count++;
  }
  if (props.hasActions) {
    count++;
  }
  return count;
});

const filteredData = computed(() => {
  let result = [...props.data];

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    result = result.filter((item) => {
      const searchInKeys = props.searchKeys.length > 0 ? props.searchKeys : props.columns.map((c) => c.key);

      return searchInKeys.some((key) => {
        const value = getNestedValue(item, key);
        return String(value).toLowerCase().includes(query);
      });
    });
  }

  if (sortBy.value) {
    result.sort((a, b) => {
      const aVal = getNestedValue(a, sortBy.value);
      const bVal = getNestedValue(b, sortBy.value);

      if (aVal === bVal) {
        return 0;
      }
      const comparison = aVal > bVal ? 1 : -1;
      return sortOrder.value === "asc" ? comparison : -comparison;
    });
  }

  return result;
});

const totalPages = computed(() => Math.ceil(filteredData.value.length / props.pageSize));

const startIndex = computed(() => (currentPage.value - 1) * props.pageSize);
const endIndex = computed(() => startIndex.value + props.pageSize);

const paginatedData = computed(() => {
  return filteredData.value.slice(startIndex.value, endIndex.value);
});

const allSelected = computed(() => {
  if (paginatedData.value.length === 0) {
    return false;
  }
  return paginatedData.value.every((item) => selectedItems.value.has(getItemId(item)));
});

watch(
  selectedItems,
  () => {
    emit("select", Array.from(selectedItems.value));
  },
  { deep: true }
);

function getItemId(item: unknown): string {
  return getNestedValue(item, props.idKey);
}

function getNestedValue(obj: unknown, path: string): string {
  if (typeof obj !== "object" || obj === null) {
    return "";
  }

  return (
    path.split(".").reduce((current: any, key: string) => {
      return current?.[key];
    }, obj) ?? ""
  );
}

function formatCellValue(item: unknown, column: Column): string {
  const value = getNestedValue(item, column.key);
  if (column.formatter) {
    return column.formatter(value);
  }
  return String(value);
}

function handleSearch(): void {
  currentPage.value = 1;
  emit("search", searchQuery.value);
}

function handleSort(key: string): void {
  if (sortBy.value === key) {
    sortOrder.value = sortOrder.value === "asc" ? "desc" : "asc";
  } else {
    sortBy.value = key;
    sortOrder.value = "asc";
  }
  currentPage.value = 1;
}

function toggleSelect(item: unknown): void {
  const id = getItemId(item);
  if (selectedItems.value.has(id)) {
    selectedItems.value.delete(id);
  } else {
    selectedItems.value.add(id);
  }
}

function toggleSelectAll(): void {
  if (allSelected.value) {
    paginatedData.value.forEach((item) => {
      selectedItems.value.delete(getItemId(item));
    });
  } else {
    paginatedData.value.forEach((item) => {
      selectedItems.value.add(getItemId(item));
    });
  }
}

function isSelected(item: unknown): boolean {
  return selectedItems.value.has(getItemId(item));
}

function updateFilter(key: string, value: unknown): void {
  filters.value[key] = value;
  currentPage.value = 1;
}

defineExpose({
  clearSelection: () => selectedItems.value.clear(),
  selectAll: () => {
    filteredData.value.forEach((item) => selectedItems.value.add(getItemId(item)));
  },
  getSelectedItems: () => Array.from(selectedItems.value),
  resetPage: () => {
    currentPage.value = 1;
  }
});
</script>

<style scoped>
.resource-table-wrapper {
  @apply w-full;
}
</style>
