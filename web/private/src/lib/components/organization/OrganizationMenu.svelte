<script lang="ts">
  import { onMount } from "svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  interface Props {
    organization: {
      owner_id: string;
    };
    onClose?: () => void;
    onEdit?: () => void;
    onManageMembers?: () => void;
    onDelete?: () => void;
  }

  let { organization, onClose, onEdit, onManageMembers, onDelete }: Props = $props();

  let menuRef = $state<HTMLElement | null>(null);
  let position = $state({ top: 0, left: 0 });

  // TODO: Get current user ID from store/auth
  let currentUserId = $state("user-id");

  const canDelete = $derived(organization.owner_id === currentUserId);

  export function updatePosition(event: MouseEvent): void {
    const rect = (event.target as HTMLElement).getBoundingClientRect();
    position = {
      top: rect.bottom + window.scrollY,
      left: rect.left + window.scrollX
    };
  }

  const handleClickOutside = (event: MouseEvent) => {
    if (menuRef && !menuRef.contains(event.target as Node)) {
      onClose?.();
    }
  };

  onMount(() => {
    document.addEventListener("click", handleClickOutside);
    return () => {
      document.removeEventListener("click", handleClickOutside);
    };
  });
</script>

<!-- Backdrop -->
<div class="fixed inset-0 z-40" onclick={() => onClose?.()} role="presentation"></div>

<!-- Menu -->
<div
  bind:this={menuRef}
  class="ring-opacity-5 absolute z-50 mt-2 w-48 rounded-md border border-gray-200 bg-white ring-1 ring-black transition-colors hover:border-gray-300"
  style:top="{position.top}px"
  style:left="{position.left}px"
>
  <div class="py-1" role="menu" aria-orientation="vertical">
    <!-- Edit Organization -->
    <button
      class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
      role="menuitem"
      onclick={() => onEdit?.()}
    >
      <SvgIcon name="pencil" class="mr-3 h-4 w-4" />
      Edit Organization
    </button>

    <!-- Manage Members -->
    <button
      class="flex w-full items-center px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900"
      role="menuitem"
      onclick={() => onManageMembers?.()}
    >
      <SvgIcon name="users" class="mr-3 h-4 w-4" />
      Manage Members
    </button>

    <!-- Divider -->
    <div class="my-1 border-t border-gray-100"></div>

    <!-- Delete Organization -->
    {#if canDelete}
      <button
        class="flex w-full items-center px-4 py-2 text-left text-sm text-red-700 hover:bg-red-50 hover:text-red-900"
        role="menuitem"
        onclick={() => onDelete?.()}
      >
        <SvgIcon name="trash-x" class="mr-3 h-4 w-4" />
        Delete Organization
      </button>
    {/if}
  </div>
</div>
