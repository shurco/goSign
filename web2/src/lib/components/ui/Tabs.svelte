<script lang="ts">
  import { setContext } from "svelte";
  import type { Snippet } from "svelte";

  interface Props {
    value?: string;
    children?: Snippet;
  }

  let { value = $bindable(""), children }: Props = $props();

  function setActiveTab(tabValue: string): void {
    value = tabValue;
  }

  setContext("activeTab", {
    get value() {
      return value;
    },
    set value(v: string) {
      value = v;
    }
  });
  setContext("setActiveTab", setActiveTab);
</script>

<div class="tabs-wrapper">
  <div class="flex gap-2 rounded-lg bg-gray-100 p-1">
    {@render children?.()}
  </div>
</div>
