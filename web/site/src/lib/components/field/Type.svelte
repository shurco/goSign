<script lang="ts">
  import type { Snippet } from "svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { fieldIcons, fieldNames } from "@/components/field/constants";
  import { clickOutside, createDropdown } from "@/composables/ui.svelte";

  interface Props {
    value: string;
    menuClasses?: string;
    buttonClasses?: string;
    editable?: boolean;
    buttonWidth?: number;
    label?: Snippet;
    onclick?: (event: MouseEvent) => void;
    onValueChange?: (type: string) => void;
  }

  let {
    value = $bindable(),
    menuClasses = "mt-1.5 bg-[#faf7f5]",
    buttonClasses = "",
    editable = true,
    buttonWidth = 18,
    label,
    onclick,
    onValueChange
  }: Props = $props();

  let dropdownRef = $state<HTMLElement | null>(null);
  const dropdown = createDropdown();

  function handleTypeSelect(type: string, event: MouseEvent): void {
    event.preventDefault();
    value = type;
    onValueChange?.(type);
    dropdown.close();
  }
</script>

<span bind:this={dropdownRef} class="dropdown" use:clickOutside={() => dropdown.close()}>
  {#if label}
    {@render label()}
  {:else}
    <label
      tabindex="0"
      title={fieldNames[value]}
      class="cursor-pointer"
      onclick={(e) => {
        e.stopPropagation();
        dropdown.toggle();
        onclick?.(e);
      }}
    >
      <SvgIcon
        name={fieldIcons[value]}
        width={buttonWidth}
        height={buttonWidth}
        class={buttonClasses}
        stroke-width="1.6"
      />
    </label>
  {/if}
  {#if editable && dropdown.isOpen}
    <ul
      tabindex="0"
      class="dropdown-content menu menu-xs z-10 mb-3 w-52 rounded-box p-2 shadow {menuClasses}"
      onclick={() => dropdown.close()}
    >
      {#each Object.entries(fieldIcons) as [type, icon] (type)}
        <li>
          <a
            href="#"
            class="flex flex-wrap px-2 py-1 text-sm {type === value ? 'active' : ''}"
            onclick={(e) => handleTypeSelect(type, e)}
          >
            <SvgIcon name={icon} stroke-width="1.6" width="20" height="20" />
            {fieldNames[type]}
          </a>
        </li>
      {/each}
    </ul>
  {/if}
</span>
