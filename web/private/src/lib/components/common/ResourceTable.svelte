<script lang="ts">
  import type { Snippet } from "svelte";
  import { SvelteSet } from "svelte/reactivity";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import Checkbox from "@/components/ui/Checkbox.svelte";
  import Button from "@/components/ui/Button.svelte";
  import Table from "@/components/ui/Table.svelte";
  import LoadingSpinner from "@/components/ui/LoadingSpinner.svelte";
  import Pagination from "@/components/ui/Pagination.svelte";

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
    cellSnippets?: Record<string, Snippet<[item: unknown, value: string]>>;
    filters?: Snippet<[filters: Record<string, unknown>, updateFilter: (key: string, value: unknown) => void]>;
    actions?: Snippet<[item: unknown]>;
    onSelect?: (selectedItems: unknown[]) => void;
    onEdit?: (item: unknown) => void;
    onDelete?: (item: unknown) => void;
    onPageChange?: (page: number) => void;
    onSearch?: (query: string) => void;
  }

  let {
    data,
    columns,
    searchable = true,
    searchPlaceholder = "Search...",
    searchKeys = [],
    selectable = false,
    showFilters = true,
    showPagination = true,
    pageSize = 10,
    isLoading = false,
    emptyMessage = "No data available",
    hasActions = true,
    showEdit = true,
    showDelete = true,
    idKey = "id",
    cellSnippets = {},
    filters,
    actions,
    onSelect,
    onEdit,
    onDelete,
    onPageChange,
    onSearch
  }: Props = $props();

  let searchQuery = $state("");
  let currentPage = $state(1);
  let sortBy = $state("");
  let sortOrder = $state<"asc" | "desc">("asc");
  const selectedItems = new SvelteSet<unknown>();
  let filtersState = $state<Record<string, unknown>>({});

  const totalColumns = $derived.by(() => {
    let count = columns.length;
    if (selectable) {
      count++;
    }
    if (hasActions) {
      count++;
    }
    return count;
  });

  const filteredData = $derived.by(() => {
    let result = [...data];

    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      result = result.filter((item) => {
        const searchInKeys = searchKeys.length > 0 ? searchKeys : columns.map((c) => c.key);

        return searchInKeys.some((key) => {
          const value = getNestedValue(item, key);
          return String(value).toLowerCase().includes(query);
        });
      });
    }

    if (sortBy) {
      result.sort((a, b) => {
        const aVal = getNestedValue(a, sortBy);
        const bVal = getNestedValue(b, sortBy);

        if (aVal === bVal) {
          return 0;
        }
        const comparison = aVal > bVal ? 1 : -1;
        return sortOrder === "asc" ? comparison : -comparison;
      });
    }

    return result;
  });

  const totalPages = $derived(Math.ceil(filteredData.length / pageSize));

  const startIndex = $derived((currentPage - 1) * pageSize);
  const endIndex = $derived(startIndex + pageSize);

  const paginatedData = $derived(filteredData.slice(startIndex, endIndex));

  const allSelected = $derived(
    paginatedData.length > 0 && paginatedData.every((item) => selectedItems.has(getItemId(item)))
  );

  $effect(() => {
    onSelect?.(Array.from(selectedItems));
  });

  function getItemId(item: unknown): string {
    return getNestedValue(item, idKey);
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
    currentPage = 1;
    onSearch?.(searchQuery);
  }

  function handleSort(key: string): void {
    if (sortBy === key) {
      sortOrder = sortOrder === "asc" ? "desc" : "asc";
    } else {
      sortBy = key;
      sortOrder = "asc";
    }
    currentPage = 1;
  }

  function toggleSelect(item: unknown): void {
    const id = getItemId(item);
    if (selectedItems.has(id)) {
      selectedItems.delete(id);
    } else {
      selectedItems.add(id);
    }
  }

  function toggleSelectAll(): void {
    if (allSelected) {
      paginatedData.forEach((item) => {
        selectedItems.delete(getItemId(item));
      });
    } else {
      paginatedData.forEach((item) => {
        selectedItems.add(getItemId(item));
      });
    }
  }

  function isSelected(item: unknown): boolean {
    return selectedItems.has(getItemId(item));
  }

  function updateFilter(key: string, value: unknown): void {
    filtersState[key] = value;
    currentPage = 1;
  }

  export function clearSelection(): void {
    selectedItems.clear();
  }

  export function selectAll(): void {
    filteredData.forEach((item) => selectedItems.add(getItemId(item)));
  }

  export function getSelectedItems(): unknown[] {
    return Array.from(selectedItems);
  }

  export function resetPage(): void {
    currentPage = 1;
  }
</script>

<div class="resource-table-wrapper w-full">
  <!-- Filters and Search (twing-m toolbar) -->
  {#if showFilters}
    <div class="toolbar">
      {#if searchable}
        <div class="search-box">
          <span class="search-icon-wrap"><SvgIcon name="search" width="14" height="14" /></span>
          <input
            type="text"
            class="search-input"
            placeholder={searchPlaceholder}
            bind:value={searchQuery}
            oninput={handleSearch}
          />
        </div>
      {/if}
      {#if filters}
        {@render filters(filtersState, updateFilter)}
      {/if}
    </div>
  {/if}

  <!-- Table (twing-m table-card + data-grid) -->
  <div class="table-card overflow-x-auto">
    <Table>
      {#snippet header()}
        <tr>
          {#if selectable}
            <th class="w-12">
              <Checkbox checked={allSelected} size="sm" onchange={toggleSelectAll} />
            </th>
          {/if}
          {#each columns as column (column.key)}
            <th
              class="{column.headerClass ?? ''} {column.sortable ? 'th-sortable' : ''}"
              onclick={() => column.sortable && handleSort(column.key)}
            >
              <div class="flex items-center gap-2">
                {column.label}
                {#if column.sortable && sortBy === column.key}
                  <span>
                    {sortOrder === "asc" ? "↑" : "↓"}
                  </span>
                {/if}
              </div>
            </th>
          {/each}
          {#if hasActions}
            <th class="w-32 text-right">Actions</th>
          {/if}
        </tr>
      {/snippet}

      {#if isLoading}
        <tr>
          <td colspan={totalColumns} class="cell-state">
            <div class="flex flex-col items-center justify-center gap-2 py-8">
              <LoadingSpinner size="md" />
              <p class="state-text">Loading...</p>
            </div>
          </td>
        </tr>
      {:else if paginatedData.length === 0}
        <tr>
          <td colspan={totalColumns} class="cell-state">
            <div class="state-text py-8 text-center">{emptyMessage}</div>
          </td>
        </tr>
      {:else}
        {#each paginatedData as item (getItemId(item))}
          <tr>
            {#if selectable}
              <td>
                <Checkbox checked={isSelected(item)} size="sm" onchange={() => toggleSelect(item)} />
              </td>
            {/if}
            {#each columns as column (column.key)}
              <td class={column.cellClass ?? ""}>
                {#if cellSnippets[column.key]}
                  {@render cellSnippets[column.key](item, getNestedValue(item, column.key))}
                {:else}
                  {formatCellValue(item, column)}
                {/if}
              </td>
            {/each}
            {#if hasActions}
              <td class="text-right">
                {#if actions}
                  {@render actions(item)}
                {:else}
                  <div class="flex justify-end gap-2">
                    {#if showEdit}
                      <Button variant="ghost" size="sm" onclick={() => onEdit?.(item)}>Edit</Button>
                    {/if}
                    {#if showDelete}
                      <Button
                        variant="ghost"
                        size="sm"
                        class="text-[var(--color-error)]"
                        onclick={() => onDelete?.(item)}
                      >
                        Delete
                      </Button>
                    {/if}
                  </div>
                {/if}
              </td>
            {/if}
          </tr>
        {/each}
      {/if}
    </Table>
  </div>

  <!-- Pagination -->
  {#if showPagination && totalPages > 1}
    <Pagination
      bind:currentPage
      {pageSize}
      total={filteredData.length}
      class="mt-4"
      onCurrentPageChange={(page) => onPageChange?.(page)}
    />
  {/if}
</div>

<style>
  .th-sortable {
    cursor: pointer;
  }
  .th-sortable:hover {
    color: var(--base-txt-secondary);
  }
  .state-text {
    font-size: var(--font-size-13);
    color: var(--base-txt-muted);
  }
</style>
