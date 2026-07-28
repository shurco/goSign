<script lang="ts">
  import Button from "./Button.svelte";

  interface Props {
    currentPage: number;
    pageSize: number;
    total: number;
    maxVisible?: number;
    showInfo?: boolean;
    class?: string;
    onCurrentPageChange?: (page: number) => void;
  }

  let {
    currentPage = $bindable(),
    pageSize,
    total,
    maxVisible = 5,
    showInfo = true,
    class: className = "",
    onCurrentPageChange
  }: Props = $props();

  function setPage(page: number): void {
    currentPage = page;
    onCurrentPageChange?.(page);
  }

  const totalPages = $derived(Math.ceil(total / pageSize));
  const startIndex = $derived((currentPage - 1) * pageSize);
  const endIndex = $derived(startIndex + pageSize);

  const visiblePages = $derived.by(() => {
    const pages: number[] = [];
    let start = Math.max(1, currentPage - Math.floor(maxVisible / 2));
    let end = Math.min(totalPages, start + maxVisible - 1);

    if (end - start + 1 < maxVisible) {
      start = Math.max(1, end - maxVisible + 1);
    }

    for (let i = start; i <= end; i++) {
      pages.push(i);
    }

    return pages;
  });
</script>

<div class="flex items-center justify-between {className}">
  {#if showInfo}
    <div class="pagination-info">
      Showing {startIndex + 1} to {Math.min(endIndex, total)} of {total} entries
    </div>
  {/if}
  <div class="flex items-center gap-1">
    <Button variant="ghost" size="sm" disabled={currentPage === 1} onclick={() => setPage(currentPage - 1)}>«</Button>
    {#each visiblePages as page (page)}
      <Button variant={page === currentPage ? "primary" : "ghost"} size="sm" onclick={() => setPage(page)}>
        {page}
      </Button>
    {/each}
    <Button variant="ghost" size="sm" disabled={currentPage === totalPages} onclick={() => setPage(currentPage + 1)}>
      »
    </Button>
  </div>
</div>

<style>
  .pagination-info {
    font-size: var(--font-size-12);
    color: var(--base-txt-muted);
  }
</style>
