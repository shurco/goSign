<script lang="ts">
  import type { SVGAttributes } from "svelte/elements";

  interface Props extends SVGAttributes<SVGSVGElement> {
    prefix?: string;
    name: string;
    /** Tooltip attribute passed through to the svg element (Vue attr fallthrough parity) */
    title?: string;
  }

  let { prefix = "icon", name, ...rest }: Props = $props();

  const symbolId = $derived(`#${prefix}-${name}`);
</script>

<svg aria-hidden="true" {...rest}>
  <use href={symbolId} xlink:href={symbolId} />
</svg>
