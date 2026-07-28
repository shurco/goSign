<script lang="ts">
  import FileDropZone from "@/components/ui/FileDropZone.svelte";

  interface Props {
    value?: string;
    type?: "file" | "image";
    disabled?: boolean;
    error?: string;
    onBlur?: () => void;
  }

  let { value = $bindable(""), type = "file", disabled = false, error = "", onBlur }: Props = $props();

  let selectedFileName = $state("");

  const accept = $derived(type === "image" ? "image/*" : undefined);

  $effect(() => {
    if (!value || value === "") {
      selectedFileName = "";
    }
  });

  function handleFileChange(file: File): void {
    selectedFileName = file.name;

    if (type === "image") {
      const reader = new FileReader();
      reader.onload = (e) => {
        const result = e.target?.result as string;
        if (result) {
          value = result;
        }
        onBlur?.();
      };
      reader.readAsDataURL(file);
    } else {
      value = file.name;
      onBlur?.();
    }
  }

  function handleClear(): void {
    selectedFileName = "";
    value = "";
    onBlur?.();
  }
</script>

<div class="field-input-wrapper">
  <FileDropZone
    {accept}
    {disabled}
    selectedLabel={selectedFileName}
    onChange={handleFileChange}
    onClear={handleClear}
  />
  {#if type === "image" && value}
    <div class="mt-2 rounded-md border border-[var(--color-base-300)] bg-[--color-base-100] p-2">
      <img src={value} alt="" class="max-h-32 w-full object-contain" />
    </div>
  {/if}
  {#if error}
    <div class="mt-2 text-sm text-[var(--color-error)]">{error}</div>
  {/if}
</div>
