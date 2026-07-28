<script lang="ts">
  import { getContext } from "svelte";
  import Contenteditable from "@/components/field/Contenteditable.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { subColors, subNames } from "@/components/field/constants";
  import { clickOutside, createDropdown } from "@/composables/ui.svelte";
  import { v4 } from "uuid";

  interface Props {
    submitters: Record<string, unknown>[];
    editable?: boolean;
    compact?: boolean;
    mobileView?: boolean;
    value: string;
    menuClasses?: string;
    class?: string;
    onRemove?: (submitter: Record<string, unknown>) => void;
    onNewSubmitter?: (submitter: Record<string, unknown>) => void;
    onNameChange?: (submitter: Record<string, unknown>) => void;
    onValueChange?: (id: string) => void;
    onclick?: (event: MouseEvent) => void;
  }

  let {
    submitters,
    editable = true,
    compact = false,
    mobileView = false,
    value = $bindable(),
    menuClasses = "dropdown-content menu p-2 shadow bg-[#faf7f5] rounded-box w-full z-10",
    class: className = "",
    onRemove,
    onNewSubmitter,
    onNameChange,
    onValueChange,
    onclick
  }: Props = $props();

  const save = getContext<() => void>("save") ?? (() => {});

  let dropdownRef = $state<HTMLElement | null>(null);
  let mobileDropdownRef = $state<HTMLElement | null>(null);
  const dropdown = createDropdown();

  const selectedSubmitter = $derived(submitters.find((e) => e.id === value) as Record<string, unknown> | undefined);

  const submitterColors = $derived.by(() => {
    const colors: Record<string, string> = {};
    submitters.forEach((submitter, index) => {
      // Pure fallback (no mutation inside $derived): same pattern as SigningModeSelector
      const colorIndex = typeof submitter.colorIndex === "number" ? submitter.colorIndex : index;
      colors[submitter.id as string] = subColors[colorIndex % subColors.length];
    });
    return colors;
  });

  function getSubmitterColor(submitter: Record<string, unknown>): string {
    return submitterColors[submitter.id as string];
  }

  function selectSubmitter(submitter: Record<string, unknown>): void {
    value = submitter.id as string;
    onValueChange?.(value);
  }

  function remove(submitter: Record<string, unknown>): void {
    if (window.confirm("Are you sure?")) {
      onRemove?.(submitter);
    }
  }

  function move(submitter: Record<string, unknown>, direction: number): void {
    const currentIndex = submitters.indexOf(submitter);

    const newIndex = currentIndex + direction;
    if (newIndex < 0 || newIndex >= submitters.length) {
      return;
    }

    const wasSelected = submitter.id === value;

    submitters.splice(currentIndex, 1);
    submitters.splice(newIndex, 0, submitter);

    if (wasSelected) {
      selectSubmitter(submitter);
    }

    save();
  }

  function addSubmitter(): void {
    const newSubmitter = {
      name: subNames[submitters.length],
      id: v4(),
      colorIndex: submitters.length
    };

    submitters.push(newSubmitter);
    value = newSubmitter.id;
    onValueChange?.(value);
    onNewSubmitter?.(newSubmitter);
  }

  function handleNameChange(): void {
    if (selectedSubmitter) {
      onNameChange?.(selectedSubmitter);
    }
  }
</script>

{#if mobileView}
  <div
    bind:this={mobileDropdownRef}
    class={className}
    role="presentation"
    {onclick}
    use:clickOutside={() => dropdown.close()}
  >
    <div class="flex items-end space-x-2">
      <div
        class="group/contenteditable-container flex w-full items-end justify-between rounded-md border border-[#e7e2df] bg-[#faf7f5] p-2"
      >
        <div class="flex items-center space-x-2">
          {#if selectedSubmitter}
            <span class="h-3 w-3 flex-shrink-0 rounded-full {getSubmitterColor(selectedSubmitter)}"></span>
            <div role="presentation" onclick={(e) => e.stopPropagation()}>
              <Contenteditable
                bind:value={selectedSubmitter.name as string}
                class="cursor-text"
                iconInline={true}
                {editable}
                selectOnEditClick={true}
                iconWidth={18}
                onValueChange={handleNameChange}
                onBlur={save}
              />
            </div>
          {/if}
        </div>
      </div>
      <div class="dropdown dropdown-end dropdown-top">
        <button
          type="button"
          class="flex w-full cursor-pointer justify-center rounded-md border border-[#e7e2df] bg-[#faf7f5] p-2"
          onclick={(e) => {
            e.stopPropagation();
            dropdown.toggle();
          }}
        >
          <SvgIcon name="chevron-up" width="24" height="24" />
        </button>
        {#if editable && dropdown.isOpen}
          <ul class="mb-2 min-w-max rounded-md {menuClasses}">
            {#each submitters as submitter (submitter.id)}
              <li class="relative">
                <button
                  type="button"
                  class="menu-item flex items-center px-2 {submitter === selectedSubmitter ? 'active' : ''}"
                  class:pr-10={submitters.length > 1 && editable}
                  onclick={() => {
                    selectSubmitter(submitter);
                    dropdown.close();
                  }}
                >
                  <span class="flex items-center py-1">
                    <span class="mr-3 ml-1 h-3 w-3 rounded-full {getSubmitterColor(submitter)}"></span>
                    <span>
                      {submitter.name}
                    </span>
                  </span>
                </button>
                {#if submitters.length > 1 && editable}
                  <button
                    type="button"
                    title="Remove"
                    class="absolute top-1/2 right-1 -translate-y-1/2 px-2"
                    onclick={(e) => {
                      e.stopPropagation();
                      remove(submitter);
                    }}
                  >
                    <SvgIcon name="trash-x" width="18" height="18" />
                  </button>
                {/if}
              </li>
            {/each}
            {#if submitters.length < 10 && editable}
              <li>
                <button
                  type="button"
                  class="menu-item flex px-2"
                  onclick={() => {
                    addSubmitter();
                    dropdown.close();
                  }}
                >
                  <SvgIcon name="user-plus" width="20" height="20" stroke-width="1.6" />
                  <span class="py-1"> Add {subNames[submitters.length]} </span>
                </button>
              </li>
            {/if}
          </ul>
        {/if}
      </div>
    </div>
  </div>
{:else}
  <div
    bind:this={dropdownRef}
    class="dropdown {className}"
    role="presentation"
    {onclick}
    use:clickOutside={() => dropdown.close()}
  >
    {#if compact}
      <button
        type="button"
        title={selectedSubmitter?.name as string}
        class="flex h-full w-full cursor-pointer items-center justify-center text-[#faf7f5]"
        onclick={(e) => {
          e.stopPropagation();
          dropdown.toggle();
        }}
      >
        {#if selectedSubmitter}
          <span class="mx-1 h-3 w-3 rounded-full {getSubmitterColor(selectedSubmitter)}"></span>
        {/if}
      </button>
    {:else}
      <div
        class="group/contenteditable-container hover:border-content group flex w-full justify-between rounded-md border border-[#e7e2df] p-2"
      >
        <div class="flex items-center space-x-2">
          {#if selectedSubmitter}
            <span class="h-3 w-3 rounded-full {getSubmitterColor(selectedSubmitter)}"></span>
            <div role="presentation" onclick={(e) => e.stopPropagation()}>
              <Contenteditable
                bind:value={selectedSubmitter.name as string}
                class="cursor-text"
                iconInline={true}
                {editable}
                selectOnEditClick={true}
                iconWidth={18}
                onValueChange={handleNameChange}
                onBlur={save}
              />
            </div>
          {/if}
        </div>
        <button
          type="button"
          title="Select submitter"
          class="flex h-6 w-6 cursor-pointer items-center justify-center rounded border-dashed border-[#291334]/20 transition-all duration-75 group-hover:border"
          onclick={(e) => {
            e.stopPropagation();
            dropdown.toggle();
          }}
        >
          <SvgIcon name="plus" width="18" height="18" />
        </button>
      </div>
    {/if}
    {#if (editable || !compact) && dropdown.isOpen}
      <ul class={menuClasses}>
        {#each submitters as submitter (submitter.id)}
          <li class="group relative" class:active={submitter === selectedSubmitter}>
            <button
              type="button"
              class="menu-item flex items-center px-2 {submitter === selectedSubmitter ? 'active' : ''}"
              class:pr-16={!compact && submitters.length > 1 && editable}
              onclick={() => {
                selectSubmitter(submitter);
                dropdown.close();
              }}
            >
              <span class="flex items-center py-1">
                <span class="mr-3 ml-1 h-3 w-3 rounded-full {getSubmitterColor(submitter)}"></span>
                <span>
                  {submitter.name}
                </span>
              </span>
            </button>
            {#if !compact && submitters.length > 1 && editable}
              <div class="absolute top-1/2 right-1 flex -translate-y-1/2 items-center gap-1">
                <div class="invisible flex flex-col gap-1 group-hover:visible group-[.active]:visible">
                  <button
                    type="button"
                    title="Up"
                    class="flex h-4 w-6 items-center justify-center rounded border border-base-200 bg-white text-[--color-base-content] transition-colors group-[.active]:text-base-content hover:border-base-content hover:bg-base-content hover:text-base-100"
                    onclick={(e) => {
                      e.stopPropagation();
                      move(submitter, -1);
                    }}
                  >
                    <SvgIcon name="chevron-up" width="12" height="12" />
                  </button>
                  <button
                    type="button"
                    title="Down"
                    class="flex h-4 w-6 items-center justify-center rounded border border-base-200 bg-white text-[--color-base-content] transition-colors group-[.active]:text-base-content hover:border-base-content hover:bg-base-content hover:text-base-100"
                    onclick={(e) => {
                      e.stopPropagation();
                      move(submitter, 1);
                    }}
                  >
                    <SvgIcon name="chevron-down" width="12" height="12" />
                  </button>
                </div>
                <button
                  type="button"
                  title="Remove"
                  class="invisible px-2 group-hover:visible group-[.active]:visible"
                  onclick={(e) => {
                    e.stopPropagation();
                    remove(submitter);
                  }}
                >
                  <SvgIcon name="trash-x" width="18" height="18" />
                </button>
              </div>
            {/if}
          </li>
        {/each}
        {#if submitters.length < 10 && editable}
          <li>
            <button
              type="button"
              class="menu-item flex px-2"
              onclick={() => {
                addSubmitter();
                dropdown.close();
              }}
            >
              <SvgIcon name="user-plus" class="mr-2" width="20" height="20" stroke-width="1.6" />
              <span class="py-1"> Add {subNames[submitters.length]} </span>
            </button>
          </li>
        {/if}
      </ul>
    {/if}
  </div>
{/if}
