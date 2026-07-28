<script lang="ts">
  import type { Snippet } from "svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  interface Props {
    /** Heading text (20/bold). Ignored when the title snippet is provided */
    heading?: string;
    /** Subtitle under the title (13/muted) */
    subtitle?: string;
    /** Back button handler; omitted when not rendered */
    onback?: () => void;
    /** Custom title content (e.g. editable input) */
    title?: Snippet;
    /** Actions on the right */
    actions?: Snippet;
  }

  let { heading = "", subtitle = "", onback, title, actions }: Props = $props();
</script>

<div class="page-header">
  <div class="page-header-main">
    {#if onback}
      <button type="button" class="page-back" onclick={onback} title="Back" aria-label="Back">
        <SvgIcon name="arrow-left" class="page-back-icon" />
      </button>
    {/if}
    {#if title}
      {@render title()}
    {:else}
      <div class="page-header-text">
        <h1>{heading}</h1>
        {#if subtitle}
          <p class="page-subtitle">{subtitle}</p>
        {/if}
      </div>
    {/if}
  </div>
  {#if actions}
    <div class="page-actions">
      {@render actions()}
    </div>
  {/if}
</div>

<style>
  /* Base styles for .page-header / .page-back / .page-actions live in app.css */
  .page-header-main {
    display: flex;
    align-items: center;
    gap: var(--space-8);
    flex: 1;
    min-width: 0;
  }
  .page-header-text {
    min-width: 0;
  }
  .page-back :global(.page-back-icon) {
    width: 18px;
    height: 18px;
  }
</style>
