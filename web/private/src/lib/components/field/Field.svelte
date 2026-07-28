<script lang="ts">
  // This component intentionally mutates props.field because field is part of a reactive template object
  // managed by the parent component. Direct mutation is used for performance optimization.
  // Rule vue/no-mutating-props is disabled for this file in eslint.config.mjs
  import { getContext, tick } from "svelte";
  import Contenteditable from "@/components/field/Contenteditable.svelte";
  import Type from "@/components/field/Type.svelte";
  import ConditionBuilder from "@/components/field/ConditionBuilder.svelte";
  import FormulaBuilder from "@/components/field/FormulaBuilder.svelte";
  import Modal from "@/components/ui/Modal.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { borderColors, fieldNames as fieldNamesConst, subNames } from "@/components/field/constants";
  import { clickOutside, createDropdown } from "@/composables/ui.svelte";
  import { t } from "@/i18n/index.svelte";
  import { v4 } from "uuid";

  interface Props {
    field: Record<string, unknown>;
    defaultField?: Record<string, unknown> | null;
    editable?: boolean;
    isSelected?: boolean;
    onSetDraw?: (payload: Record<string, unknown>) => void;
    onRemove?: (field: Record<string, unknown>) => void;
    onScrollTo?: (area: unknown) => void;
  }

  let {
    field,
    defaultField = null,
    editable = true,
    isSelected = false,
    onSetDraw,
    onRemove,
    onScrollTo
  }: Props = $props();

  const template = getContext<{ value: Record<string, unknown> }>("template");
  const save = getContext<() => void>("save") ?? (() => {});
  const selectedAreaRef = getContext<{ value: unknown }>("selectedAreaRef");

  let nameRef = $state<{ getContenteditable(): HTMLSpanElement | null } | undefined>();
  let optionsRef = $state<HTMLElement | null>(null);
  let dropdownRef = $state<HTMLElement | null>(null);
  let formulaBuilderRef = $state<{ getFormula(): string } | undefined>();
  let isNameFocus = $state(false);
  let showConditionBuilder = $state(false);
  let showFormulaBuilder = $state(false);
  let showDescriptionModal = $state(false);
  let validationType = $state("");

  const dropdown = createDropdown();

  const validationPresets: Record<string, { pattern: string; message: string }> = {
    email: { pattern: "^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$", message: "Please enter a valid email address" },
    ssn: { pattern: "^[0-9]{3}-[0-9]{2}-[0-9]{4}$", message: "Please enter a valid SSN (XXX-XX-XXXX)" },
    ein: { pattern: "^[0-9]{2}-[0-9]{7}$", message: "Please enter a valid EIN (XX-XXXXXXX)" },
    url: { pattern: "^https?:\\/\\/.+", message: "Please enter a valid URL" },
    zip: { pattern: "^[0-9]{5}(-[0-9]{4})?$", message: "Please enter a valid ZIP code" },
    numbers_only: { pattern: "^[0-9]+$", message: "Please enter numbers only" },
    letters_only: { pattern: "^[a-zA-Z]+$", message: "Please enter letters only" }
  };

  const fieldNames = fieldNamesConst;

  const submitterIndex = $derived(
    (template.value.submitters as Record<string, unknown>[]).findIndex((s) => s.id === field.submitter_id)
  );

  const signatureFormatValue = $derived(
    typeof (field.preferences as Record<string, unknown> | undefined)?.format === "string" &&
      (field.preferences as Record<string, unknown>).format !== ""
      ? ((field.preferences as Record<string, unknown>).format as string)
      : "any"
  );

  function setSignatureFormat(value: string): void {
    ensurePreferences();
    (field.preferences as Record<string, unknown>).format = value;
  }

  const dateFormats = [
    "MM/DD/YYYY",
    "DD/MM/YYYY",
    "YYYY-MM-DD",
    "DD-MM-YYYY",
    "DD.MM.YYYY",
    "MMM D, YYYY",
    "MMMM D, YYYY",
    "MMMM YYYY",
    "D MMM YYYY",
    "D MMMM YYYY"
  ];

  /** DocuSeal-style "set signing date": readonly date auto-filled at signing time. */
  const isSetSigningDate = $derived(field.default_value === "{{date}}");

  function toggleSetSigningDate(checked: boolean): void {
    if (checked) {
      field.default_value = "{{date}}";
      field.readonly = true;
    } else {
      delete field.default_value;
      field.readonly = false;
    }
    save();
  }

  const isCheckboxChecked = $derived(field.default_value === true || field.default_value === "true");

  function toggleCheckboxChecked(checked: boolean): void {
    if (checked) {
      field.default_value = true;
    } else {
      delete field.default_value;
    }
    save();
  }

  const fontSizeValue = $derived.by(() => {
    const size = (field.preferences as Record<string, unknown> | undefined)?.font_size;
    return typeof size === "number" && size > 0 ? size : "";
  });

  const showFontSettings = $derived(["text", "number", "date", "select", "cells"].includes(field.type as string));

  function getDefaultFieldName(f: Record<string, unknown>): string {
    if (!template?.value?.fields) {
      return (f.id as string) || "Field";
    }
    if (f.type === "payment" && (f.preferences as Record<string, unknown> | undefined)?.price) {
      const prefs = (f.preferences as Record<string, unknown>) || {};
      const { price, currency } = prefs;
      const formattedPrice = new Intl.NumberFormat([], {
        style: "currency",
        currency: currency as string
      }).format(price as number);
      return `${fieldNames[f.type as string]} ${formattedPrice}`;
    }
    const idx =
      (template.value.submitters as Record<string, unknown>[] | undefined)?.findIndex((s) => s.id === f.submitter_id) ??
      0;
    const partyName = subNames[idx]?.replace(" Party", "") || "First";
    const typeName = fieldNames[f.type as string] || "Field";
    const sameTypeAndPartyFields = (template.value.fields as Record<string, unknown>[]).filter(
      (other) => other.type === f.type && other.submitter_id === f.submitter_id && other.id !== f.id
    );
    const fieldNumber = sameTypeAndPartyFields.length + 1;
    return `${partyName} ${typeName} ${fieldNumber}`;
  }

  const defaultName = $derived(getDefaultFieldName(field));

  const availableFieldsForConditions = $derived(
    ((template?.value?.fields as Record<string, unknown>[]) ?? [])
      .filter((f) => f.id !== field.id && f.submitter_id === field.submitter_id)
      .map((f) => ({
        ...f,
        displayName: (f.name && String(f.name).trim()) || getDefaultFieldName(f)
      }))
  );

  const availableFieldsForFormula = $derived(
    ((template?.value?.fields as Record<string, unknown>[]) ?? [])
      .filter(
        (f) =>
          f.id !== field.id && f.submitter_id === field.submitter_id && ["number", "text"].includes(f.type as string)
      )
      .map((f) => ({
        ...f,
        displayName: (f.name && String(f.name).trim()) || getDefaultFieldName(f)
      }))
  );

  // Intentional initial snapshot: tracks previous value, updated inside $effect below
  // svelte-ignore state_referenced_locally
  let previousSubmitterId = $state(field.submitter_id);

  $effect(() => {
    const newSubmitterId = field.submitter_id;
    const oldSubmitterId = previousSubmitterId;
    if (newSubmitterId !== oldSubmitterId && isDefaultName(field.name as string)) {
      field.name = "";
      save();
    }
    previousSubmitterId = newSubmitterId;
  });

  $effect(() => {
    const newType = field.type as string;
    field.preferences ||= {};

    if (newType === "date") {
      const prefs = field.preferences as Record<string, unknown>;
      prefs.format ||= Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US") ? "MM/DD/YYYY" : "DD/MM/YYYY";
    }
    if (newType === "payment") {
      const prefs = field.preferences as Record<string, unknown>;
      if (prefs.price == null) {
        prefs.price = 0;
      }
      if (!prefs.currency) {
        prefs.currency = "USD";
      }
    }
    if (newType === "number" && (!field.validation || typeof field.validation !== "object")) {
      field.validation = {};
    }
    if (["text", "cells"].includes(newType) && (!field.validation || typeof field.validation !== "object")) {
      field.validation = {};
    }
  });

  async function openConditionBuilder(): Promise<void> {
    showConditionBuilder = true;
    await tick();
    dropdown.close();
  }

  async function openFormulaBuilder(): Promise<void> {
    showFormulaBuilder = true;
    await tick();
    dropdown.close();
  }

  async function openDescriptionModal(): Promise<void> {
    showDescriptionModal = true;
    await tick();
    dropdown.close();
  }

  function closeConditionBuilder(): void {
    showConditionBuilder = false;
  }

  function closeFormulaBuilder(): void {
    showFormulaBuilder = false;
  }

  function closeDescriptionModal(): void {
    showDescriptionModal = false;
  }

  function saveDescriptionAndClose(): void {
    save();
    showDescriptionModal = false;
  }

  function applyFormulaAndClose(): void {
    if (formulaBuilderRef?.getFormula) {
      field.formula = formulaBuilderRef.getFormula();
      if (field.formula && !field.calculationType) {
        field.calculationType = "number";
      }
      save();
      showFormulaBuilder = false;
    }
  }

  function applyValidationPreset(): void {
    ensureValidation();
    if (validationType && validationType !== "length" && validationType !== "custom") {
      const preset = validationPresets[validationType];
      if (preset) {
        (field.validation as Record<string, unknown>).pattern = preset.pattern;
        (field.validation as Record<string, unknown>).message = preset.message;
      }
    }
    save();
  }

  function isDefaultName(name: string): boolean {
    if (!name) {
      return true;
    }
    const pattern = /^(First|Second|Third|Fourth|Fifth|Sixth|Seventh|Eighth|Ninth|Tenth)\s+\w+\s+\d+$/;
    return pattern.test(name);
  }

  function ensurePreferences(): void {
    if (!field.preferences) {
      field.preferences = {};
    }
  }

  function removeArea(area: Record<string, unknown>): void {
    const areasArr = field.areas as Record<string, unknown>[];
    areasArr.splice(areasArr.indexOf(area), 1);
    save();
  }

  function setFontSize(raw: string): void {
    ensurePreferences();
    const prefs = field.preferences as Record<string, unknown>;
    const size = parseInt(raw, 10);
    if (Number.isFinite(size) && size > 0) {
      prefs.font_size = size;
    } else {
      delete prefs.font_size;
    }
    save();
  }

  function setFontPref(key: "font_type" | "align" | "color", value: string): void {
    ensurePreferences();
    const prefs = field.preferences as Record<string, unknown>;
    if (value) {
      prefs[key] = value;
    } else {
      delete prefs[key];
    }
    save();
  }

  function ensureValidation(): void {
    if (!field.validation || typeof field.validation !== "object") {
      field.validation = {};
    }
  }

  function formatDate(date: Date, format: string): string {
    const monthFormats: Record<string, "numeric" | "2-digit" | "short" | "long"> = {
      M: "numeric",
      MM: "2-digit",
      MMM: "short",
      MMMM: "long"
    };

    const dayFormats: Record<string, "numeric" | "2-digit"> = {
      D: "numeric",
      DD: "2-digit"
    };

    const yearFormats: Record<string, "numeric" | "2-digit"> = {
      YYYY: "numeric",
      YY: "2-digit"
    };

    const dayMatch = format.match(/D+/);
    const monthMatch = format.match(/M+/);
    const yearMatch = format.match(/Y+/);

    const parts = new Intl.DateTimeFormat([], {
      day: dayMatch ? dayFormats[dayMatch[0]] || "numeric" : "numeric",
      month: monthMatch ? monthFormats[monthMatch[0]] || "numeric" : "numeric",
      year: yearMatch ? yearFormats[yearMatch[0]] || "numeric" : "numeric"
    }).formatToParts(date);

    const dayPart = parts.find((p) => p.type === "day");
    const monthPart = parts.find((p) => p.type === "month");
    const yearPart = parts.find((p) => p.type === "year");

    return format
      .replace(/D+/, dayPart?.value || "")
      .replace(/M+/, monthPart?.value || "")
      .replace(/Y+/, yearPart?.value || "");
  }

  async function copyToAllPages(f: Record<string, unknown>): Promise<void> {
    const areas = f.areas as Record<string, unknown>[];
    const areaString = JSON.stringify(areas[0]);
    (template.value.documents as Record<string, unknown>[]).forEach((attachment) => {
      (attachment.preview_images as Record<string, unknown>[]).forEach((page) => {
        if (
          !areas.find(
            (area) => area.attachment_id === attachment.id && area.page === parseInt(page.filename as string, 10)
          )
        ) {
          areas.push({
            ...JSON.parse(areaString),
            attachment_id: attachment.id,
            page: parseInt(page.filename as string, 10)
          });
        }
      });
    });

    await tick();
    onScrollTo?.(areas[areas.length - 1]);
    save();
  }

  function onNameFocus(): void {
    isNameFocus = true;
    if (!field.name) {
      setTimeout(() => {
        const el = nameRef?.getContenteditable();
        if (el) {
          el.innerText = " ";
        }
      }, 1);
    }
  }

  function maybeFocusOnOptionArea(option: Record<string, unknown>): void {
    const areas = field.areas as Record<string, unknown>[];
    const area = areas.find((a) => a.option_id === option.id);
    if (area) {
      selectedAreaRef.value = area;
    }
  }

  function scrollToFirstArea(): void {
    const areas = field.areas as Record<string, unknown>[] | undefined;
    if (areas?.[0]) {
      onScrollTo?.(areas[0]);
    }
  }

  async function addOption(): Promise<void> {
    (field.options as Record<string, unknown>[]).push({ value: "", id: v4() });
    await tick();
    const inputs = optionsRef?.querySelectorAll("input");
    inputs?.[inputs.length - 1]?.focus();
    save();
  }

  function removeOption(option: Record<string, unknown>): void {
    const options = field.options as Record<string, unknown>[];
    options.splice(options.indexOf(option), 1);
    const areas = field.areas as Record<string, unknown>[];
    areas.splice(
      areas.findIndex((a) => a.option_id === option.id),
      1
    );
    save();
  }

  /** DocuSeal-style paste: multi-line text becomes multiple options. */
  function handleOptionPaste(e: ClipboardEvent, option: Record<string, unknown>): void {
    const text = e.clipboardData?.getData("text") || "";
    const lines = text
      .split("\n")
      .map((line) => line.trim())
      .filter(Boolean);
    if (lines.length < 2) {
      return;
    }
    e.preventDefault();
    const opts = field.options as Record<string, unknown>[];
    const index = opts.indexOf(option);
    option.value = lines[0];
    opts.splice(index + 1, 0, ...lines.slice(1).map((value) => ({ value, id: v4() })));
    save();
  }

  function handleOptionEnter(e: KeyboardEvent, index: number): void {
    if (e.key !== "Enter") {
      return;
    }
    e.preventDefault();
    const opts = field.options as Record<string, unknown>[];
    if (index === opts.length - 1) {
      addOption();
    }
  }

  function getEffectiveCellW(area: { w: number; h: number; cell_w?: number }): number {
    if (area.cell_w != null && area.cell_w > 0) {
      return area.cell_w;
    }
    if (area.w <= 0) {
      return 0;
    }
    if (area.h > 0) {
      const denom = Math.floor(area.w / area.h);
      return denom > 0 ? (area.w * 2) / denom : area.w / 5;
    }
    return area.w / 5;
  }

  function getCellCountFromArea(area: { w: number; h: number; cell_w?: number }): number {
    const cellWidth = getEffectiveCellW(area);
    if (!cellWidth || cellWidth <= 0 || area.w <= 0) {
      return 0;
    }
    let currentWidth = 0;
    let count = 0;
    while (currentWidth + (cellWidth + cellWidth / 4) < area.w) {
      currentWidth += cellWidth;
      count++;
    }
    return Math.max(count, 1);
  }

  function maybeUpdateOptions(): void {
    if (field.type !== "cells") {
      delete field.default_value;
    }
    if (!["radio", "multiple", "select"].includes(field.type as string)) {
      delete field.options;
    }
    if (["radio", "multiple", "select"].includes(field.type as string)) {
      field.options ||= [{ value: "", id: v4() }];
    }
    ((field.areas as Record<string, unknown>[]) || []).forEach((area) => {
      if (field.type === "cells") {
        const denom = (area.h as number) > 0 ? Math.floor((area.w as number) / (area.h as number)) : 0;
        if (denom > 0) {
          area.cell_w = ((area.w as number) * 2) / denom;
        } else if ((area.w as number) > 0 && (!area.cell_w || (area.cell_w as number) <= 0)) {
          area.cell_w = (area.w as number) / 5;
        }
        area.cell_count = getCellCountFromArea(area as { w: number; h: number; cell_w?: number });
      } else {
        delete area.cell_w;
        delete area.cell_count;
      }
    });
  }

  function onNameBlur(): void {
    const text = nameRef?.getContenteditable()?.innerText.trim();
    if (text) {
      field.name = text;
    } else {
      field.name = "";
      const el = nameRef?.getContenteditable();
      if (el) {
        el.innerText = defaultName;
      }
    }
    isNameFocus = false;
    save();
  }

  function onFocusOut(event: FocusEvent): void {
    const currentTarget = event.currentTarget as HTMLElement;
    const relatedTarget = event.relatedTarget as HTMLElement | null;

    if (!currentTarget.contains(relatedTarget)) {
      isNameFocus = false;
    }
  }

  const areas = $derived((field.areas as Record<string, unknown>[]) || []);
  const options = $derived((field.options as Record<string, unknown>[]) || undefined);
</script>

<div class="group pb-2">
  <div
    class="group relative rounded rounded-tr-none border py-1 {isSelected
      ? borderColors[submitterIndex]
      : 'border-[#e7e2df]'} {field.required ? '' : 'border-dashed'}"
  >
    <div class="group/contenteditable-container relative flex items-center justify-between" onfocusout={onFocusOut}>
      <div
        class="absolute top-0 right-0 bottom-0 left-0 cursor-pointer"
        onclick={scrollToFirstArea}
        role="presentation"
      ></div>

      <div class="flex items-center space-x-1 p-1">
        <Type
          bind:value={field.type as string}
          editable={editable && !defaultField}
          buttonWidth={20}
          onValueChange={() => {
            maybeUpdateOptions();
            save();
          }}
          onclick={scrollToFirstArea}
        />
        <Contenteditable
          bind:this={nameRef}
          value={(field.name as string) || defaultName}
          editable={editable && !defaultField}
          iconInline={true}
          iconWidth={18}
          iconStrokeWidth={1.6}
          selectOnEditClick={true}
          onFocus={() => {
            onNameFocus();
            scrollToFirstArea();
          }}
          onBlur={onNameBlur}
        />
      </div>

      {#if isNameFocus}
        <div class="relative flex items-center gap-1.5 pr-2">
          {#if field.type != "phone"}
            <input
              id="required-checkbox-{field.id}"
              bind:checked={field.required as boolean}
              type="checkbox"
              class="toggle toggle-xs"
              onmousedown={(e) => e.preventDefault()}
              onchange={save}
            />
          {/if}
        </div>
      {:else if editable}
        <div class="flex items-center space-x-1">
          {#if field && !areas.length}
            <button
              title="Draw"
              class="relative cursor-pointer text-transparent group-hover:text-[#291334]"
              onclick={() => onSetDraw?.({ field })}
            >
              <SvgIcon name="section" width="18" height="18" />
            </button>
          {:else}
            <span bind:this={dropdownRef} class="dropdown dropdown-end" use:clickOutside={() => dropdown.close()}>
              <button
                type="button"
                title="Settings"
                class="cursor-pointer text-transparent group-hover:text-[#291334]"
                onclick={() => dropdown.toggle()}
              >
                <SvgIcon name="settings" width="18" height="18" />
              </button>

              {#if dropdown.isOpen}
                <ul
                  class="dropdown-content menu menu-xs z-10 mt-1.5 w-52 rounded-box bg-[#faf7f5] p-2 shadow"
                  draggable={true}
                  ondragstart={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                  }}
                >
                  {#if field.type === "text" && !defaultField}
                    <div class="relative px-1 py-1.5" onclick={(e) => e.stopPropagation()} role="presentation">
                      <input
                        bind:value={field.default_value as string}
                        type="text"
                        placeholder="Default value"
                        dir="auto"
                        class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                        onblur={save}
                      />
                      {#if field.default_value}
                        <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px"> Default value </span>
                      {/if}
                    </div>
                  {/if}
                  {#if field.type === "date"}
                    <div class="relative px-1 py-1.5" onclick={(e) => e.stopPropagation()} role="presentation">
                      <select
                        bind:value={(field.preferences as Record<string, unknown>).format as string}
                        class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                        onchange={save}
                      >
                        {#each dateFormats as format (format)}
                          <option value={format}>{formatDate(new Date(), format)}</option>
                        {/each}
                      </select>
                      <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px"> Format </span>
                    </div>
                    {#if !defaultField}
                      <li class="px-2" role="presentation" onclick={(e) => e.stopPropagation()}>
                        <label class="flex cursor-pointer items-center gap-2 py-1.5">
                          <input
                            checked={isSetSigningDate}
                            type="checkbox"
                            class="toggle toggle-xs"
                            onchange={(e) => toggleSetSigningDate(e.currentTarget.checked)}
                          />
                          <span class="label-text">Set signing date</span>
                        </label>
                      </li>
                    {/if}
                  {/if}
                  {#if field.type === "checkbox" && !defaultField}
                    <li class="px-2" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <label class="flex cursor-pointer items-center gap-2 py-1.5">
                        <input
                          checked={isCheckboxChecked}
                          type="checkbox"
                          class="toggle toggle-xs"
                          onchange={(e) => toggleCheckboxChecked(e.currentTarget.checked)}
                        />
                        <span class="label-text">Checked</span>
                      </label>
                    </li>
                  {/if}
                  {#if ["select", "radio"].includes(field.type as string) && !defaultField && (field.options as Record<string, unknown>[] | undefined)?.some((o) => o.value)}
                    <div class="relative px-1 py-1.5" onclick={(e) => e.stopPropagation()} role="presentation">
                      <select
                        bind:value={field.default_value as string}
                        class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                        onchange={() => {
                          if (!field.default_value) {
                            delete field.default_value;
                          }
                          save();
                        }}
                      >
                        <option value="">None</option>
                        {#each (field.options as Record<string, unknown>[]).filter((o) => o.value) as option (option.id)}
                          <option value={option.value as string}>{option.value}</option>
                        {/each}
                      </select>
                      <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px"> Default value </span>
                    </div>
                  {/if}
                  {#if field.type === "number" && !defaultField}
                    <li class="px-1 py-1" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <div class="space-y-1">
                        <div class="relative py-1.5">
                          <input
                            bind:value={field.default_value as string}
                            type="text"
                            inputmode="decimal"
                            placeholder="Default value"
                            dir="auto"
                            class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                            onblur={save}
                          />
                          {#if field.default_value}
                            <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">
                              Default value
                            </span>
                          {/if}
                        </div>
                        <div class="relative py-1.5">
                          <select
                            bind:value={(field.preferences as Record<string, unknown>).format as string}
                            class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                            onchange={() => {
                              ensurePreferences();
                              save();
                            }}
                          >
                            <option value="">None</option>
                            <option value="comma">1,000.00 (comma)</option>
                            <option value="dot">1.000,00 (dot)</option>
                            <option value="space">1 000,00 (space)</option>
                            <option value="usd">$1,000.00 (USD)</option>
                            <option value="eur">€1.000,00 (EUR)</option>
                            <option value="gbp">£1,000.00 (GBP)</option>
                            <option value="percent">1000%</option>
                          </select>
                          <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Number format</span>
                        </div>
                        <div class="flex items-center gap-1 py-1.5">
                          <div class="relative max-w-20 min-w-0 flex-1">
                            <input
                              bind:value={(field.validation as Record<string, unknown>).min as number}
                              type="number"
                              class="input-bordered input input-xs h-7 w-full !outline-0"
                              onchange={() => {
                                ensureValidation();
                                save();
                              }}
                            />
                            <span class="absolute -top-2 left-2.5 h-4 px-1" style="font-size: 8px">Min</span>
                          </div>
                          <span class="shrink-0 text-xs text-base-content/60"> – </span>
                          <div class="relative max-w-20 min-w-0 flex-1">
                            <input
                              bind:value={(field.validation as Record<string, unknown>).max as number}
                              type="number"
                              class="input-bordered input input-xs h-7 w-full !outline-0"
                              onchange={() => {
                                ensureValidation();
                                save();
                              }}
                            />
                            <span class="absolute -top-2 left-2.5 h-4 px-1" style="font-size: 8px">Max</span>
                          </div>
                        </div>
                        <div class="relative py-1.5">
                          <input
                            bind:value={(field.validation as Record<string, unknown>).step as string}
                            type="text"
                            placeholder="any"
                            class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                            onchange={() => {
                              ensureValidation();
                              save();
                            }}
                          />
                          <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Step</span>
                        </div>
                      </div>
                    </li>
                  {/if}
                  {#if field.type === "signature" && !defaultField}
                    <li class="px-1 py-1" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <div class="space-y-1">
                        <div class="relative py-1.5">
                          <select
                            value={signatureFormatValue}
                            class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                            onchange={(e) => {
                              setSignatureFormat(e.currentTarget.value);
                              ensurePreferences();
                              save();
                            }}
                          >
                            <option value="any">Any</option>
                            <option value="drawn">Drawn</option>
                            <option value="typed">Typed</option>
                            <option value="drawn_or_typed">Drawn or typed</option>
                            <option value="drawn_or_upload">Drawn or upload</option>
                            <option value="upload">Upload</option>
                          </select>
                          <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Signature format</span>
                        </div>
                        <label class="flex cursor-pointer items-center gap-2 py-1">
                          <input
                            bind:checked={(field.preferences as Record<string, unknown>).with_signature_id as boolean}
                            type="checkbox"
                            class="toggle toggle-xs"
                            onchange={() => {
                              ensurePreferences();
                              save();
                            }}
                          />
                          <span class="label-text text-xs">With signature ID</span>
                        </label>
                      </div>
                    </li>
                  {/if}
                  {#if field.type === "payment" && editable && !defaultField}
                    <li class="px-1 py-1" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <div class="space-y-1">
                        <div class="relative py-1.5">
                          <input
                            bind:value={(field.preferences as Record<string, unknown>).price as number}
                            type="number"
                            step="0.01"
                            min="0"
                            class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                            onchange={() => {
                              ensurePreferences();
                              save();
                            }}
                          />
                          <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Price</span>
                        </div>
                        <div class="relative py-1.5">
                          <select
                            bind:value={(field.preferences as Record<string, unknown>).currency as string}
                            class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                            onchange={() => {
                              ensurePreferences();
                              save();
                            }}
                          >
                            <option value="USD">USD</option>
                            <option value="EUR">EUR</option>
                            <option value="GBP">GBP</option>
                            <option value="JPY">JPY</option>
                            <option value="RUB">RUB</option>
                          </select>
                          <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Currency</span>
                        </div>
                      </div>
                    </li>
                  {/if}
                  {#if field.type === "stamp" && !defaultField}
                    <li role="presentation" onclick={(e) => e.stopPropagation()}>
                      <div class="space-y-1">
                        <label class="flex cursor-pointer items-center gap-2 py-1">
                          <input
                            bind:checked={(field.preferences as Record<string, unknown>).with_logo as boolean}
                            type="checkbox"
                            class="toggle toggle-xs"
                            onchange={() => {
                              ensurePreferences();
                              save();
                            }}
                          />
                          <span class="label-text text-xs">With logo</span>
                        </label>
                        <label class="flex cursor-pointer items-center gap-2 py-1">
                          <input
                            bind:checked={(field.preferences as Record<string, unknown>).with_signature_id as boolean}
                            type="checkbox"
                            class="toggle toggle-xs"
                            onchange={() => {
                              ensurePreferences();
                              save();
                            }}
                          />
                          <span class="label-text text-xs">With stamp ID</span>
                        </label>
                      </div>
                    </li>
                  {/if}
                  {#if ["text", "cells"].includes(field.type as string) && !defaultField}
                    <li class="px-1 py-1" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <div class="space-y-1">
                        <div class="relative py-1.5">
                          <select
                            bind:value={validationType}
                            class="select-bordered select select-xs !h-7 w-full max-w-xs font-normal !outline-0"
                            onchange={applyValidationPreset}
                          >
                            <option value="">None</option>
                            <option value="length">Length</option>
                            <option value="email">Email</option>
                            <option value="ssn">SSN</option>
                            <option value="ein">EIN</option>
                            <option value="url">URL</option>
                            <option value="zip">ZIP</option>
                            <option value="numbers_only">Numbers only</option>
                            <option value="letters_only">Letters only</option>
                            <option value="custom">Custom pattern</option>
                          </select>
                          <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Validation</span>
                        </div>
                        {#if validationType === "length"}
                          <div class="relative py-1.5">
                            <input
                              bind:value={(field.validation as Record<string, unknown>).min as number}
                              type="number"
                              class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                              onchange={() => {
                                ensureValidation();
                                save();
                              }}
                            />
                            <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Min length</span>
                          </div>
                          <div class="relative py-1.5">
                            <input
                              bind:value={(field.validation as Record<string, unknown>).max as number}
                              type="number"
                              class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                              onchange={() => {
                                ensureValidation();
                                save();
                              }}
                            />
                            <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Max length</span>
                          </div>
                        {/if}
                        {#if validationType === "custom"}
                          <div class="relative py-1.5">
                            <input
                              bind:value={(field.validation as Record<string, unknown>).pattern as string}
                              type="text"
                              placeholder="^[0-9]{3}-[0-9]{2}-[0-9]{4}$"
                              class="input-bordered input input-xs h-7 w-full max-w-xs font-mono text-xs !outline-0"
                              onchange={() => {
                                ensureValidation();
                                save();
                              }}
                            />
                            <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Pattern</span>
                          </div>
                          <div class="relative py-1.5">
                            <input
                              bind:value={(field.validation as Record<string, unknown>).message as string}
                              type="text"
                              class="input-bordered input input-xs h-7 w-full max-w-xs !outline-0"
                              onchange={() => {
                                ensureValidation();
                                save();
                              }}
                            />
                            <span class="absolute -top-1 left-2.5 h-4 px-1" style="font-size: 8px">Error message</span>
                          </div>
                        {/if}
                      </div>
                    </li>
                  {/if}
                  {#if field.type != "phone"}
                    <li class="px-2" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <label class="flex cursor-pointer items-center gap-2 py-1.5">
                        <input
                          bind:checked={field.required as boolean}
                          type="checkbox"
                          class="toggle toggle-xs"
                          onchange={save}
                        />
                        <span class="label-text">Required</span>
                      </label>
                    </li>
                  {/if}
                  {#if ["text", "number", "stamp"].includes(field.type as string) && !defaultField}
                    <li class="px-2" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <label class="flex cursor-pointer items-center gap-2 py-1.5">
                        <input
                          bind:checked={field.readonly as boolean}
                          type="checkbox"
                          class="toggle toggle-xs"
                          onchange={save}
                        />
                        <span class="label-text">Read only</span>
                      </label>
                    </li>
                  {/if}
                  {#if showFontSettings && !defaultField}
                    <li class="px-1 py-1" role="presentation" onclick={(e) => e.stopPropagation()}>
                      <div class="space-y-1">
                        <div class="flex items-center gap-1 py-1.5">
                          <div class="relative max-w-16 min-w-0 flex-1">
                            <input
                              value={fontSizeValue}
                              type="number"
                              min="4"
                              max="72"
                              placeholder="11"
                              class="input-bordered input input-xs h-7 w-full !outline-0"
                              onchange={(e) => setFontSize(e.currentTarget.value)}
                            />
                            <span class="absolute -top-2 left-2.5 h-4 px-1" style="font-size: 8px">Font size</span>
                          </div>
                          <div class="relative min-w-0 flex-1">
                            <select
                              value={((field.preferences as Record<string, unknown> | undefined)
                                ?.font_type as string) || ""}
                              class="select-bordered select select-xs !h-7 w-full font-normal !outline-0"
                              onchange={(e) => setFontPref("font_type", e.currentTarget.value)}
                            >
                              <option value="">Regular</option>
                              <option value="bold">Bold</option>
                              <option value="italic">Italic</option>
                              <option value="bold_italic">Bold italic</option>
                            </select>
                            <span class="absolute -top-2 left-2.5 h-4 px-1" style="font-size: 8px">Style</span>
                          </div>
                        </div>
                        <div class="flex items-center gap-1 py-1.5">
                          <div class="relative min-w-0 flex-1">
                            <select
                              value={((field.preferences as Record<string, unknown> | undefined)?.align as string) ||
                                ""}
                              class="select-bordered select select-xs !h-7 w-full font-normal !outline-0"
                              onchange={(e) => setFontPref("align", e.currentTarget.value)}
                            >
                              <option value="">Left</option>
                              <option value="center">Center</option>
                              <option value="right">Right</option>
                            </select>
                            <span class="absolute -top-2 left-2.5 h-4 px-1" style="font-size: 8px">Align</span>
                          </div>
                          <div class="relative max-w-16 min-w-0 flex-1">
                            <input
                              value={((field.preferences as Record<string, unknown> | undefined)?.color as string) ||
                                "#000000"}
                              type="color"
                              class="input-bordered input input-xs h-7 w-full cursor-pointer p-0.5 !outline-0"
                              onchange={(e) => setFontPref("color", e.currentTarget.value)}
                            />
                            <span class="absolute -top-2 left-2.5 h-4 px-1" style="font-size: 8px">Color</span>
                          </div>
                        </div>
                      </div>
                    </li>
                  {/if}
                  {#if field.type != "phone"}
                    <hr class="mt-0.5 pb-0.5" />
                  {/if}
                  {#each areas as area, index (index)}
                    <li>
                      <div class="flex items-center justify-between gap-1 p-0">
                        <button
                          type="button"
                          class="menu-item flex-1 px-2 py-1 text-sm"
                          onclick={() => {
                            onScrollTo?.(area);
                            dropdown.close();
                          }}
                        >
                          <SvgIcon name="shape" width="18" height="18" />
                          Page {(area.page as number) + 1}
                        </button>
                        {#if editable && !defaultField}
                          <button
                            title="Remove area"
                            class="px-1.5 py-1"
                            onclick={(e) => {
                              e.preventDefault();
                              e.stopPropagation();
                              removeArea(area);
                            }}
                          >
                            <SvgIcon name="x" width="14" height="14" />
                          </button>
                        {/if}
                      </div>
                    </li>
                  {/each}
                  {#if !areas.length || !["radio", "multiple"].includes(field.type as string)}
                    <li>
                      <button
                        type="button"
                        class="menu-item px-2 py-1 text-sm"
                        onclick={() => {
                          onSetDraw?.({ field });
                          dropdown.close();
                        }}
                      >
                        <SvgIcon name="section" width="18" height="18" />
                        Draw new area
                      </button>
                    </li>
                  {/if}
                  {#if editable && !defaultField}
                    <hr class="mt-0.5 pb-0.5" />
                  {/if}
                  {#if editable && !defaultField}
                    <li role="presentation" onclick={(e) => e.stopPropagation()}>
                      <button type="button" class="menu-item px-2 py-1 text-sm" onclick={openConditionBuilder}>
                        <SvgIcon name="settings" width="18" height="18" />
                        Conditional Logic
                      </button>
                    </li>
                  {/if}
                  {#if editable && !defaultField && ["number", "text"].includes(field.type as string)}
                    <li role="presentation" onclick={(e) => e.stopPropagation()}>
                      <button type="button" class="menu-item px-2 py-1 text-sm" onclick={openFormulaBuilder}>
                        <SvgIcon name="settings" width="18" height="18" />
                        Formula
                      </button>
                    </li>
                  {/if}
                  {#if editable && !defaultField}
                    <li role="presentation" onclick={(e) => e.stopPropagation()}>
                      <button type="button" class="menu-item px-2 py-1 text-sm" onclick={openDescriptionModal}>
                        <SvgIcon name="settings" width="18" height="18" />
                        Description
                      </button>
                    </li>
                  {/if}
                  {#if areas.length >= 1 && !["radio", "multiple"].includes(field.type as string)}
                    <li>
                      <button
                        type="button"
                        class="menu-item px-2 py-1 text-sm"
                        onclick={() => {
                          copyToAllPages(field);
                          dropdown.close();
                        }}
                      >
                        <SvgIcon name="copy" width="18" height="18" />
                        Copy to All Pages
                      </button>
                    </li>
                  {/if}
                </ul>
              {/if}
            </span>
          {/if}
          <button
            class="relative pr-1 text-transparent group-hover:text-[#291334]"
            title="Remove"
            onclick={() => onRemove?.(field)}
          >
            <SvgIcon name="trash-x" width="18" height="18" />
          </button>
        </div>
      {/if}
    </div>

    {#if options}
      <div
        bind:this={optionsRef}
        class="mx-2 space-y-1.5 border-t border-[#e7e2df] pt-2"
        role="presentation"
        draggable={true}
        ondragstart={(e) => {
          e.preventDefault();
          e.stopPropagation();
        }}
      >
        {#each options as option, index (option.id)}
          <div class="flex items-center space-x-1.5">
            <span class="w-3.5 text-sm"> {index + 1}. </span>
            {#if ["radio", "multiple"].includes(field.type as string) && (index > 0 || areas.find((a) => a.option_id) || !areas.length) && !areas.find((a) => a.option_id === option.id)}
              <div class="flex w-full items-center">
                <input
                  bind:value={option.value as string}
                  class="input input-xs input-primary -mr-6 w-full bg-transparent !pr-7 text-sm"
                  type="text"
                  dir="auto"
                  required
                  placeholder="Option {index + 1}"
                  onblur={save}
                  onpaste={(e) => handleOptionPaste(e, option)}
                  onkeydown={(e) => handleOptionEnter(e, index)}
                />
                <button
                  title="Draw"
                  tabindex="-1"
                  onclick={(e) => {
                    e.preventDefault();
                    onSetDraw?.({ field, option });
                  }}
                >
                  <SvgIcon name="section" width="18" height="18" />
                </button>
              </div>
            {:else}
              <input
                bind:value={option.value as string}
                class="input input-xs input-primary w-full bg-transparent text-sm"
                placeholder="Option {index + 1}"
                type="text"
                required
                dir="auto"
                onfocus={() => maybeFocusOnOptionArea(option)}
                onblur={save}
                onpaste={(e) => handleOptionPaste(e, option)}
                onkeydown={(e) => handleOptionEnter(e, index)}
              />
            {/if}
            <button class="w-3.5 text-sm" tabindex="-1" onclick={() => removeOption(option)}>&times;</button>
          </div>
        {/each}
        <button class="w-full pb-1 text-center text-sm" onclick={addOption}>+ Add option</button>
      </div>
    {/if}
  </div>

  <Modal bind:open={showConditionBuilder} size="lg">
    {#snippet header()}
      <h3 class="text-lg font-semibold">{t("fields.conditions.title")}</h3>
    {/snippet}
    <ConditionBuilder
      field={field as never}
      availableFields={availableFieldsForConditions as never}
      onUpdateConditions={(conditions) => {
        field.condition_groups = conditions;
        save();
      }}
    />
    {#snippet footer()}
      <div class="flex justify-end gap-2">
        <button
          class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm hover:bg-gray-50"
          onclick={closeConditionBuilder}
        >
          {t("common.close")}
        </button>
      </div>
    {/snippet}
  </Modal>

  <Modal bind:open={showFormulaBuilder} size="lg">
    {#snippet header()}
      <h3 class="text-lg font-semibold">{t("fields.formula.title")}</h3>
    {/snippet}
    <FormulaBuilder
      bind:this={formulaBuilderRef}
      field={field as never}
      availableFields={availableFieldsForFormula as never}
    />
    {#snippet footer()}
      <div class="flex justify-end gap-2">
        <button
          class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm hover:bg-gray-50"
          onclick={closeFormulaBuilder}
        >
          {t("common.cancel")}
        </button>
        <button
          class="rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-700"
          onclick={applyFormulaAndClose}
        >
          {t("common.save")}
        </button>
      </div>
    {/snippet}
  </Modal>

  <Modal bind:open={showDescriptionModal} size="md">
    {#snippet header()}
      <h3 class="text-lg font-semibold">Description</h3>
    {/snippet}
    <div class="space-y-4">
      <div>
        <label class="label py-0 text-xs" for="field-description-{field.id}"><span>Description</span></label>
        <textarea
          id="field-description-{field.id}"
          bind:value={field.description as string}
          class="textarea textarea-bordered textarea-sm mt-1 min-h-24 w-full resize-y font-normal !outline-0"
          placeholder="Field description or instructions"
          rows="4"
          onchange={save}></textarea>
      </div>
      <div>
        <label class="label py-0 text-xs" for="field-title-{field.id}"
          ><span>Display title</span> <span class="text-base-content/60">(optional)</span></label
        >
        <input
          id="field-title-{field.id}"
          bind:value={field.title as string}
          type="text"
          class="input-bordered input input-sm mt-1 h-9 w-full font-normal !outline-0"
          placeholder="Optional display title"
          onchange={save}
        />
      </div>
    </div>
    {#snippet footer()}
      <div class="flex justify-end gap-2">
        <button
          class="rounded-md border border-gray-300 bg-white px-3 py-1.5 text-sm hover:bg-gray-50"
          onclick={closeDescriptionModal}
        >
          {t("common.close")}
        </button>
        <button
          class="rounded-md bg-indigo-600 px-3 py-1.5 text-sm text-white hover:bg-indigo-700"
          onclick={saveDescriptionAndClose}
        >
          {t("common.save")}
        </button>
      </div>
    {/snippet}
  </Modal>
</div>
