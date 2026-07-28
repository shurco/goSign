<script lang="ts">
  /**
   * Section secondary nav (WS2 secondaryMenu): vertical panel to the
   * right of the dark icon rail — for sections with subsections
   * (Settings). Section content goes in .section-page.
   */
  import type { Snippet } from "svelte";
  import { page } from "$app/state";

  export type SectionNavItem = {
    href: string;
    label: string;
  };

  let { title, items, extra }: { title: string; items: SectionNavItem[]; extra?: Snippet } = $props();

  // Active item is the longest matching path prefix:
  // /settings/email/templates highlights the nested item, not the root.
  const activeHref = $derived.by(() => {
    const pathname = page.url.pathname;
    let best = "";
    for (const it of items) {
      if ((pathname === it.href || pathname.startsWith(it.href + "/")) && it.href.length > best.length) {
        best = it.href;
      }
    }
    return best;
  });
</script>

<aside class="section-nav">
  <h2 class="section-nav-title">{title}</h2>
  {#if extra}
    <div class="section-nav-extra">{@render extra()}</div>
  {/if}
  {#each items as item (item.href)}
    <a href={item.href} class="section-nav-item" class:active={activeHref === item.href}>
      {item.label}
    </a>
  {/each}
</aside>

<style>
  /* 200px panel in twing-m style: white, pill items, active = blue pill */
  .section-nav {
    position: fixed;
    left: var(--layout-rail-width);
    top: 0;
    bottom: 0;
    width: var(--layout-sidebar-width);
    background: var(--base-surf-panel);
    border-right: 1px solid var(--base-line-tertiary);
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: var(--space-16) 10px;
    overflow-y: auto;
    z-index: 50;
  }
  /* Typography and vertical rhythm match .page-header h1: 20/bold,
     34px row (header button height) so titles share one baseline */
  .section-nav-title {
    display: flex;
    align-items: center;
    min-height: 34px;
    font-size: 20px;
    font-weight: var(--font-weight-bold);
    color: var(--base-txt-primary);
    margin: 0 0 var(--space-12);
    padding: 0 10px;
  }
  /* Full width of the panel content zone — matches item pills */
  .section-nav-extra {
    padding: 0 0 var(--space-12);
  }
  .section-nav-item {
    display: flex;
    align-items: center;
    gap: var(--space-8);
    padding: var(--space-8) 10px;
    border-radius: var(--radius-8);
    font-size: var(--font-size-13);
    color: var(--base-txt-tertiary);
    text-decoration: none;
    flex-shrink: 0;
    transition: var(--transition-colors);
  }
  .section-nav-item:hover {
    background: var(--base-surf-page);
    color: var(--base-txt-primary);
  }
  .section-nav-item.active {
    background: var(--base-hlt-notr-easy);
    color: var(--base-hlt-invert);
    font-weight: var(--font-weight-semibold);
  }
</style>
