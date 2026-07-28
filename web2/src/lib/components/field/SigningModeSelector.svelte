<script lang="ts">
  import type { SigningMode } from "@/models";
  import ButtonGroup from "@/components/ui/ButtonGroup.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { t } from "@/i18n/index.svelte";

  interface SigningModeOption {
    value: SigningMode;
    label: string;
    icon: string;
    title: string;
    description: string;
  }

  interface Props {
    signingMode: SigningMode;
    submitters: Record<string, unknown>[];
    editable: boolean;
    hideOrderList?: boolean;
    onUpdateSigningMode?: (value: SigningMode) => void;
    onUpdateSubmitterOrder?: (orderedSubmitters: Record<string, unknown>[]) => void;
  }

  let {
    signingMode,
    submitters,
    editable,
    hideOrderList = false,
    onUpdateSigningMode,
    onUpdateSubmitterOrder
  }: Props = $props();

  let draggedSubmitter = $state<string | null>(null);
  let submittersList = $state<HTMLElement | null>(null);

  const submitterColors = [
    "bg-red-500",
    "bg-blue-500",
    "bg-green-500",
    "bg-yellow-500",
    "bg-purple-500",
    "bg-pink-500",
    "bg-indigo-500",
    "bg-orange-500"
  ];

  const orderedSubmitters = $derived(
    [...submitters].sort((a, b) => ((a.order as number) || 0) - ((b.order as number) || 0))
  );

  const signingModes = $derived<SigningModeOption[]>([
    {
      value: "parallel",
      label: t("signingMode.parallel"),
      icon: "arrows-right-left",
      title: t("signingMode.parallelTitle"),
      description: t("signingMode.parallelDescription")
    },
    {
      value: "sequential",
      label: t("signingMode.sequential"),
      icon: "arrow-right",
      title: t("signingMode.sequentialTitle"),
      description: t("signingMode.sequentialDescription")
    }
  ]);

  const currentMode = $derived(signingModes.find((mode) => mode.value === signingMode) || signingModes[0]);

  function updateSigningMode(mode: string | number): void {
    if (!editable) {
      return;
    }
    const signingModeValue = typeof mode === "string" ? (mode as SigningMode) : (String(mode) as SigningMode);
    onUpdateSigningMode?.(signingModeValue);
  }

  function onDragStart(event: DragEvent, submitter: Record<string, unknown>): void {
    if (!editable) {
      event.preventDefault();
      return;
    }
    draggedSubmitter = submitter.id as string;
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
      event.dataTransfer.setData("text/plain", submitter.id as string);
    }
  }

  function onDragEnd(): void {
    draggedSubmitter = null;
  }

  function onDragOver(event: DragEvent): void {
    event.preventDefault();
    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
  }

  function onDrop(event: DragEvent): void {
    event.preventDefault();

    if (!editable || !draggedSubmitter) {
      return;
    }

    const targetElement = (event.target as HTMLElement).closest("[data-submitter-id]") as HTMLElement;
    if (!targetElement) {
      return;
    }

    const targetSubmitterId = targetElement.getAttribute("data-submitter-id");
    if (!targetSubmitterId || targetSubmitterId === draggedSubmitter) {
      return;
    }

    const draggedIndex = orderedSubmitters.findIndex((s) => s.id === draggedSubmitter);
    const targetIndex = orderedSubmitters.findIndex((s) => s.id === targetSubmitterId);

    if (draggedIndex === -1 || targetIndex === -1) {
      return;
    }

    const newSubmitters = [...submitters];
    const [draggedItem] = newSubmitters.splice(
      newSubmitters.findIndex((s) => s.id === draggedSubmitter),
      1
    );

    const targetPos = newSubmitters.findIndex((s) => s.id === targetSubmitterId);
    newSubmitters.splice(targetPos, 0, draggedItem);

    newSubmitters.forEach((submitter, index) => {
      submitter.order = index;
    });

    onUpdateSubmitterOrder?.(newSubmitters);

    draggedSubmitter = null;
  }

  function getSubmitterColor(submitter: Record<string, unknown>, index: number): string {
    const colorIndex = submitter.colorIndex !== undefined ? (submitter.colorIndex as number) : index;
    return submitterColors[colorIndex % submitterColors.length];
  }
</script>

<div>
  <h3 class="mb-3 text-sm font-medium text-gray-700">{t("signingMode.title")}</h3>
  <div class="space-y-3">
    <ButtonGroup
      bind:value={() => signingMode, (v) => updateSigningMode(v)}
      options={signingModes}
      disabled={!editable}
    />

    <div class="rounded-md bg-blue-50 p-3 text-sm text-blue-800">
      <p class="mb-1 font-medium">{currentMode.title}</p>
      <p>{currentMode.description}</p>
    </div>

    {#if !hideOrderList && signingMode === "sequential" && submitters.length > 1}
      <div class="space-y-3">
        <div class="rounded-md bg-amber-50 p-3">
          <div class="flex items-start gap-2">
            <SvgIcon name="info" width="16" height="16" class="mt-0.5 text-amber-600" />
            <div class="text-sm text-amber-800">
              <p class="mb-1 font-medium">{t("signingMode.orderTitle")}</p>
              <p>{t("signingMode.orderHint")}</p>
            </div>
          </div>
        </div>

        <div class="space-y-2">
          <h4 class="text-sm font-medium text-gray-700">{t("signingMode.signingOrder")}</h4>
          <div bind:this={submittersList} class="space-y-2" ondragover={onDragOver} ondrop={onDrop} role="list">
            {#each orderedSubmitters as submitter, index (submitter.id)}
              <div
                data-submitter-id={submitter.id}
                class="flex cursor-move items-center gap-3 rounded-md border border-gray-200 bg-white p-3 transition-colors hover:bg-gray-50 {draggedSubmitter ===
                submitter.id
                  ? 'border-blue-300 bg-blue-50'
                  : ''}"
                draggable={true}
                ondragstart={(e) => onDragStart(e, submitter)}
                ondragend={onDragEnd}
                role="listitem"
              >
                <div class="flex items-center gap-2">
                  <div
                    class="flex h-6 w-6 items-center justify-center rounded-full bg-blue-100 text-xs font-medium text-blue-700"
                  >
                    {index + 1}
                  </div>
                  <div class="h-3 w-3 rounded-full {getSubmitterColor(submitter, index)}" />
                  <span class="text-sm font-medium text-gray-900">{submitter.name}</span>
                </div>
                <div class="ml-auto">
                  <SvgIcon name="drag" width="16" height="16" class="text-gray-400" />
                </div>
              </div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  </div>
</div>
