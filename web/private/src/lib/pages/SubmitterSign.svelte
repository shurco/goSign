<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import { page } from "$app/state";
  import FieldFormDrawer from "@/components/signing/FieldFormDrawer.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import Button from "@/components/ui/Button.svelte";
  import Modal from "@/components/ui/Modal.svelte";
  // Public signer-facing endpoints do not require authentication.
  import { useConditions } from "@/composables/useConditions.svelte";
  import { useFormulas } from "@/composables/useFormulas.svelte";
  import type { Field } from "@/models/template";
  import { t, getLocale, setLocale, SUPPORTED_LOCALES, type Locale } from "@/i18n/index.svelte";
  import { fieldIcons, fieldNames, subNames } from "@/components/field/constants";
  import { formatDateByPattern } from "@/utils/time";

  const SIGNING_DRAFT_STORAGE_KEY_PREFIX = "signing-draft-";

  function getDraftStorageKey(s: string): string {
    return SIGNING_DRAFT_STORAGE_KEY_PREFIX + s;
  }

  function loadDraftFromStorage(s: string): Record<string, unknown> | null {
    try {
      const raw = localStorage.getItem(getDraftStorageKey(s));
      if (!raw) {
        return null;
      }
      const parsed = JSON.parse(raw);
      return parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : null;
    } catch {
      return null;
    }
  }

  function clearDraftStorage(s: string): void {
    try {
      localStorage.removeItem(getDraftStorageKey(s));
    } catch {
      // ignore
    }
  }

  // Field type is imported from @/models/template

  interface Submitter {
    id: string;
    name: string;
    email: string;
    slug: string;
    status: string;
    completed_at?: string;
    declined_at?: string;
  }

  interface Template {
    id: string;
    name: string;
    description?: string;
    fields: Field[];
    schema?: Array<{
      attachment_id: string;
      [key: string]: any;
    }>;
    documents: {
      id: string;
      url?: string;
      preview_images: {
        id: string;
        url?: string; // populated client-side as `${document.url}/${document.id}`
        filename: string;
        metadata: {
          width: number;
          height: number;
        };
      }[];
      metadata?: {
        pdf?: {
          number_of_pages?: number;
        };
      };
    }[];
  }

  const slug = $derived(page.params.slug as string);

  let isLoading = $state(true);
  let isSubmitting = $state(false);
  let error = $state("");
  const SIGNING_LOCALE_STORAGE_KEY = "signing_locale";
  let previousLocale = $state<string | null>(null);
  let signingLocale = $state(getLocale() as string);
  let showLanguageSelector = $state(true);
  let declineModalOpen = $state(false);
  let declineReason = $state("");

  let template = $state<Template | null>(null);
  let submitter = $state<Submitter | null>(null);
  let submissionStatus = $state<string>("");
  let completedDocumentUrl = $state<string>("");
  let formData = $state<Record<string, any>>({});
  let fieldErrors = $state<Record<string, string>>({});
  let currentFieldIndex = $state(0);
  /** Single expanded form block (only this field shows input). Null = all collapsed. */
  let expandedFieldId = $state<string | null>(null);
  /** Signature IDs for fields with "with_signature_id" (generated when value is set). */
  let signatureIds = $state<Record<string, string>>({});
  /** Timeout for auto-removing field highlight so only one block is highlighted at a time. */
  let highlightTimeout: ReturnType<typeof setTimeout> | null = null;
  let isUpdatingSubmitter = $state(false);
  let submitterInfo = $state({ name: "", email: "" });
  let submitterInfoErrors = $state<Record<string, string>>({});

  /** Drawer reads/writes formData for the currently expanded field (function binding, no stale bridge state). */
  function getDrawerValue(): string | boolean | string[] | undefined {
    return expandedFieldId ? formData[expandedFieldId] : undefined;
  }

  function setDrawerValue(v: string | boolean | string[] | undefined): void {
    if (expandedFieldId) {
      formData[expandedFieldId] = v;
    }
  }

  const myFields = $derived.by(() => {
    if (!template || !submitter) {
      return [];
    }
    return template.fields.filter((f) => f.submitter_id === submitter?.id);
  });

  // Use conditions composable
  const { fieldStates } = useConditions(
    () => template?.fields || [],
    () => formData
  );

  // Use formulas composable
  const { calculatedValues } = useFormulas(
    () => template?.fields || [],
    () => formData
  );

  // Filter visible fields based on conditions.
  // Readonly fields are excluded from fillable steps (DocuSeal behavior): their
  // values come from defaults resolved in initializeFormData and are submitted as-is.
  const visibleFields = $derived.by(() => {
    return myFields.filter((field) => {
      if ((field as any).readonly) {
        return false;
      }
      const state = fieldStates[field.id];
      return state ? state.visible : true;
    });
  });

  const filledFieldIds = $derived.by(() => {
    void formData;
    void fieldStates;
    return visibleFields.filter((f) => isFieldFilled(f)).map((f) => f.id);
  });

  const activeField = $derived.by(() => {
    if (!expandedFieldId) {
      return null;
    }
    return visibleFields.find((f) => f.id === expandedFieldId) ?? null;
  });

  const completedFieldsCount = $derived.by(() => {
    return visibleFields.filter((field) => {
      const value = formData[field.id];
      const required = fieldStates[field.id]?.required || field.required;
      if (!required) {
        return true;
      }

      // Special handling for image/file types
      if (field.type === "image" || field.type === "file") {
        // For image, value is base64 string (starts with "data:"); for file, filename string
        return typeof value === "string" && value.trim() !== "";
      }
      // Signature, initials, stamp: value must be a non-empty data URL image
      if (field.type === "signature" || field.type === "initials" || field.type === "stamp") {
        return typeof value === "string" && value.trim() !== "" && value.startsWith("data:");
      }

      // Special handling for cells type - check if all cells are filled
      if (field.type === "cells") {
        const cellCount = getCellCount(field);
        return typeof value === "string" && value.length === cellCount;
      }

      if (typeof value === "string") {
        return value.trim() !== "";
      }
      if (Array.isArray(value)) {
        return value.length > 0;
      }
      if (typeof value === "boolean") {
        return value === true;
      }
      return value !== undefined && value !== null;
    }).length;
  });

  const isFormValid = $derived(completedFieldsCount === visibleFields.length && Object.keys(fieldErrors).length === 0);

  // Sort documents by schema order (same as editor)
  const sortedDocuments = $derived.by(() => {
    if (!template || !template.documents) {
      return [];
    }
    if (template.schema && template.schema.length > 0) {
      return template.schema
        .map((item: any) => template?.documents.find((d: any) => d.id === item.attachment_id))
        .filter((d): d is NonNullable<typeof d> => Boolean(d));
    }
    // Fallback to original order if no schema
    return template.documents;
  });

  const needsEmailOrName = $derived.by(() => {
    if (!submitter) {
      return false;
    }
    return !submitter.email || !submitter.name;
  });

  const isSubmitterInfoValid = $derived.by(() => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return (
      submitterInfo.name.trim() !== "" &&
      submitterInfo.email.trim() !== "" &&
      emailRegex.test(submitterInfo.email) &&
      Object.keys(submitterInfoErrors).length === 0
    );
  });

  const prevUnfilledIndex = $derived.by(() => {
    void formData;
    void fieldStates;
    return getPrevUnfilledIndex();
  });

  const nextUnfilledIndex = $derived.by(() => {
    void formData;
    void fieldStates;
    return getNextUnfilledIndex();
  });

  // Persist form draft to localStorage (debounced) so reload restores filled fields
  let draftSaveTimeout: ReturnType<typeof setTimeout> | null = null;
  $effect(() => {
    for (const key of Object.keys(formData)) {
      void formData[key];
    }
    if (!slug || !submitter) {
      return;
    }
    const status = submitter.status;
    if (status === "completed" || status === "declined") {
      return;
    }
    if (draftSaveTimeout) {
      clearTimeout(draftSaveTimeout);
    }
    draftSaveTimeout = setTimeout(() => {
      draftSaveTimeout = null;
      const fieldIds = new Set(myFields.map((f) => f.id));
      const draft: Record<string, unknown> = {};
      fieldIds.forEach((id) => {
        if (formData[id] !== undefined) {
          draft[id] = formData[id];
        }
      });
      try {
        localStorage.setItem(getDraftStorageKey(slug), JSON.stringify(draft));
      } catch {
        /* ignore */
      }
    }, 500);
  });

  // Generate signature ID when user fills a signature/initials field with "with_signature_id"
  $effect(() => {
    for (const key of Object.keys(formData)) {
      void formData[key];
    }
    myFields.forEach((field) => {
      if (!hasWithSignatureId(field)) {
        return;
      }
      const v = formData[field.id];
      if (v != null && String(v).trim() !== "" && !signatureIds[field.id]) {
        signatureIds[field.id] = generateSignatureId(field);
      }
    });
  });

  // Sort preview images within each document (same as editor)
  function getSortedPreviewImages(doc: any): any[] {
    if (!doc || !doc.preview_images || doc.preview_images.length === 0) {
      return [];
    }

    const numberOfPages = doc.metadata?.pdf?.number_of_pages || doc.preview_images.length;
    const previewImagesIndex = doc.preview_images.reduce(
      (acc: any, e: any) => {
        acc[parseInt(e.filename, 10)] = e;
        return acc;
      },
      {} as Record<number, any>
    );

    const lazyloadMetadata = doc.preview_images[doc.preview_images.length - 1].metadata;
    return [...Array(numberOfPages).keys()].map((i) => {
      return (
        previewImagesIndex[i] || {
          metadata: lazyloadMetadata,
          id: Math.random().toString(),
          url: doc.url ? `${doc.url}/${doc.id}` : undefined,
          filename: doc.preview_images[i]?.filename || `${i}`
        }
      );
    });
  }

  function detectBrowserSigningLocale(): Locale {
    const browser = (navigator.language || "en").split("-")[0];
    if (browser in SUPPORTED_LOCALES) {
      return browser as Locale;
    }
    return "en";
  }

  function initialSigningLocale(): Locale {
    const stored = localStorage.getItem(SIGNING_LOCALE_STORAGE_KEY);
    if (stored && stored in SUPPORTED_LOCALES) {
      return stored as Locale;
    }
    return detectBrowserSigningLocale();
  }

  function applySigningLocale(next: Locale): void {
    signingLocale = next;
    setLocale(next);
    document.documentElement.setAttribute("lang", next);
    localStorage.setItem(SIGNING_LOCALE_STORAGE_KEY, next);
  }

  function onSigningLocaleChange(event: Event): void {
    const value = (event.target as HTMLSelectElement | null)?.value as Locale | undefined;
    if (!value || !(value in SUPPORTED_LOCALES)) {
      return;
    }
    applySigningLocale(value);
  }

  onMount(async () => {
    // For public signing we intentionally default to browser language (not app/user locale).
    // We also keep this preference isolated from the admin UI locale.
    previousLocale = getLocale() as string;
    applySigningLocale(initialSigningLocale());
    await loadSubmission();
    // Auto-open drawer for first unfilled field when signing form is shown
    await tick();
    if (
      submitter &&
      submitter.status !== "completed" &&
      submitter.status !== "declined" &&
      !needsEmailOrName &&
      visibleFields.length > 0
    ) {
      const firstUnfilled = visibleFields.find((f) => !isFieldFilled(f));
      if (firstUnfilled) {
        expandedFieldId = firstUnfilled.id;
        currentFieldIndex = visibleFields.findIndex((f) => f.id === firstUnfilled.id);
        scrollToFieldOnDocument(firstUnfilled.id);
      }
    }
  });

  onDestroy(() => {
    // Restore original app locale when leaving the signing page.
    if (previousLocale) {
      setLocale(previousLocale as Locale);
      document.documentElement.setAttribute("lang", previousLocale);
    }
    if (draftSaveTimeout) {
      clearTimeout(draftSaveTimeout);
    }
    if (highlightTimeout != null) {
      clearTimeout(highlightTimeout);
    }
  });

  function todayISO(): string {
    const now = new Date();
    const month = String(now.getMonth() + 1).padStart(2, "0");
    const day = String(now.getDate()).padStart(2, "0");
    return `${now.getFullYear()}-${month}-${day}`;
  }

  /** Resolve template default (supports DocuSeal {{date}} placeholder). */
  function resolveDefaultString(def: unknown): string {
    const s = String(def ?? "").trim();
    return s === "{{date}}" ? todayISO() : s;
  }

  function initializeFormData(): void {
    // Reset field values for the current submitter fields, applying template
    // default values the same way DocuSeal does.
    const next: Record<string, any> = {};

    myFields.forEach((field) => {
      const def = (field as any).default_value;
      if (field.type === "checkbox") {
        next[field.id] = def === true || def === "true";
        return;
      }
      if (field.type === "multiple" || (field as any).type === "multi_select") {
        if (Array.isArray(def)) {
          next[field.id] = [...def].map(String);
        } else {
          next[field.id] = def != null && String(def).trim() !== "" ? [String(def)] : [];
        }
        return;
      }
      // Signature-like and attachment fields never take text defaults
      if (["signature", "initials", "stamp", "image", "file", "payment"].includes(field.type)) {
        next[field.id] = "";
        return;
      }
      // text, number, date, select, radio, cells, phone
      next[field.id] = def != null ? resolveDefaultString(def) : "";
    });

    formData = next;
    fieldErrors = {};
  }

  function getFieldLabel(field: Field): string {
    // Prefer i18n label in this order:
    // 1) Field-level translations (field.translations[locale])
    // 2) Field.label
    // 3) Field.name
    // 4) Generated default name based on field type and submitter
    // 5) Field.id (last resort)
    const loc = (signingLocale || getLocale() || "en").toString();
    const anyField: any = field as any;

    const fieldTranslations = anyField.translations as Record<string, string> | undefined;
    const translated = fieldTranslations?.[loc];
    if (translated && translated.trim() !== "") {
      return translated;
    }

    if (anyField.label && String(anyField.label).trim() !== "") {
      return String(anyField.label);
    }
    if (field.name && field.name.trim() !== "") {
      return field.name;
    }

    // Generate default name if field.name is empty
    const defaultName = generateDefaultFieldName(field);
    if (defaultName) {
      return defaultName;
    }

    return field.id;
  }

  /** Value to show on document overlay when field is filled (or empty string). */
  function getFieldDisplayValue(field: Field): string {
    const value = formData[field.id];
    if (value == null || value === "") {
      return "";
    }
    if (field.type === "signature" || field.type === "initials" || field.type === "stamp" || field.type === "image") {
      return typeof value === "string" && value.startsWith("data:") ? value : "";
    }
    if (field.type === "checkbox") {
      return value ? "✓" : "";
    }
    if (field.type === "file") {
      return typeof value === "string" ? value : "";
    }
    if (field.type === "date") {
      const format = (field as { preferences?: { format?: string } }).preferences?.format || "DD/MM/YYYY";
      return formatDateByPattern(String(value), format);
    }
    if (Array.isArray(value)) {
      return value.join(", ");
    }
    if (field.type === "number" && value !== "") {
      const format = (field as any).preferences?.format;
      const num = Number(value);
      if (Number.isFinite(num) && format) {
        return formatNumberForDisplay(num, format, (field as any).preferences?.currency);
      }
    }
    return String(value).slice(0, 200);
  }

  function isFieldDisplayImage(field: Field): boolean {
    const value = formData[field.id];
    return typeof value === "string" && value.startsWith("data:") && /^data:image\//i.test(value);
  }

  function generateDefaultFieldName(field: Field): string {
    if (!template || !submitter) {
      return "";
    }

    // Get submitter index for party name
    const submitters = (template as { submitters?: { id: string }[] })?.submitters ?? [];
    const submitterIndex = submitters.findIndex((s: any) => s.id === submitter?.id);
    const partyName = subNames[submitterIndex]?.replace(" Party", "") || "First";

    // Get type name from constants
    const typeName = fieldNames[field.type] || "Field";

    // Count how many fields of this type and party already exist
    const sameTypeAndPartyFields = template.fields.filter(
      (f: any) => f.type === field.type && f.submitter_id === submitter?.id && f.id !== field.id
    );

    const fieldNumber = sameTypeAndPartyFields.length + 1;

    return `${partyName} ${typeName} ${fieldNumber}`;
  }

  function isFieldFilled(field: Field): boolean {
    const value = formData[field.id];
    const required = fieldStates[field.id]?.required ?? field.required;
    if (!required) {
      return true;
    }
    if (field.type === "image" || field.type === "file") {
      return typeof value === "string" && value.trim() !== "";
    }
    if (field.type === "signature" || field.type === "initials" || field.type === "stamp") {
      return typeof value === "string" && value.trim() !== "" && value.startsWith("data:");
    }
    if (field.type === "cells") {
      const cellCount = getCellCount(field);
      return typeof value === "string" && value.length === cellCount;
    }
    if (typeof value === "string") {
      return value.trim() !== "";
    }
    if (Array.isArray(value)) {
      return value.length > 0;
    }
    if (typeof value === "boolean") {
      return value === true;
    }
    return value !== undefined && value !== null;
  }

  /** Signature/initials/stamp format from field preferences for signing UI (drawn, typed, upload, etc.). Stamp: upload only. */
  function getSignatureFormat(field: Field): string {
    if (field.type === "stamp") {
      return "upload";
    }
    if (field.type !== "signature" && field.type !== "initials") {
      return "";
    }
    const prefs = field.preferences as { format?: string } | undefined;
    const format = prefs?.format;
    return typeof format === "string" ? format : "";
  }

  function hasWithSignatureId(field: Field): boolean {
    if (field.type !== "signature" && field.type !== "initials" && field.type !== "stamp") {
      return false;
    }
    const prefs = field.preferences as { with_signature_id?: boolean } | undefined;
    return !!prefs?.with_signature_id;
  }

  /** Initials derived from submitter name, e.g. "John Smith" -> "JS" (DocuSeal auto-initials). */
  const initialsDefault = $derived.by(() => {
    const name = submitter?.name || "";
    return name
      .split(/\s+/)
      .filter(Boolean)
      .map((part) => part[0].toUpperCase())
      .slice(0, 3)
      .join("");
  });

  function generateSignatureId(field: Field): string {
    const prefix = field.type === "stamp" ? "STMP-" : "SIG-";
    const bytes = new Uint8Array(4);
    crypto.getRandomValues(bytes);
    return (
      prefix +
      Array.from(bytes)
        .map((b) => b.toString(16).padStart(2, "0").toUpperCase())
        .join("")
    );
  }

  function getCellCount(field: Field): number {
    if (field.type !== "cells" || !field.areas || field.areas.length === 0) {
      return 6;
    }

    const area = field.areas[0] as any;
    // Use persisted cell_count so editor and signing show the same number of cells
    if (area.cell_count != null && area.cell_count > 0) {
      return area.cell_count;
    }

    const cellWidth = area.cell_w;
    const areaWidth = area.w;
    if (!cellWidth || !areaWidth) {
      return 6;
    }

    let currentWidth = 0;
    let count = 0;
    while (currentWidth + (cellWidth + cellWidth / 4) < areaWidth) {
      currentWidth += cellWidth;
      count++;
    }
    return Math.max(count, 1);
  }

  function normalizeTemplateForSigning(tpl: Template | null): void {
    if (!tpl) {
      return;
    }

    // Ensure preview images have a usable base URL (same convention as builder `Document.vue`).
    for (const doc of tpl.documents || []) {
      const base = (doc.url || "/drive/pages").replace(/\/$/, "");
      const docBase = `${base}/${doc.id}`;
      for (const img of doc.preview_images || []) {
        // In API, preview_images do not include url. We add it here.
        (img as any).url ||= docBase;
      }
    }

    // Ensure field areas have a height.
    // Backend historically used `z` for height. Some parts of the UI expect `h`.
    for (const f of tpl.fields || []) {
      for (const a of (f.areas as any[]) || []) {
        if (!a) {
          continue;
        }
        if (a.h === undefined && a.z !== undefined) {
          a.h = a.z;
        }
      }
    }
  }

  function pageNumberFromPreview(preview: any, fallbackIndex: number): number {
    const n = Number.parseInt(String(preview?.filename ?? ""), 10);
    return Number.isFinite(n) ? n : fallbackIndex;
  }

  async function loadSubmission(): Promise<void> {
    try {
      isLoading = true;
      const response = await fetch(`/public/sign/${slug}`);

      if (!response.ok) {
        throw new Error(t("signing.submissionNotFound"));
      }

      const data = await response.json();
      // API returns: { success, message, data: { template, submitter } }
      const payload = data.data || data;
      template = payload.template;
      submitter = payload.submitter;
      submissionStatus = String(payload.submission_status || "");
      completedDocumentUrl = String(payload.completed_document_url || "");
      normalizeTemplateForSigning(template);

      // Mark as opened
      if (submitter?.status === "pending") {
        await fetch(`/public/sign/${slug}/open`, {
          method: "POST"
        });
      }

      // Initialize form data
      initializeFormData();

      // Restore saved draft from localStorage (only when not completed/declined)
      const status = submitter?.status;
      if (status !== "completed" && status !== "declined") {
        const draft = loadDraftFromStorage(slug);
        if (draft) {
          const allowedIds = new Set(myFields.map((f) => f.id));
          Object.keys(draft).forEach((id) => {
            if (allowedIds.has(id) && draft[id] !== undefined) {
              formData[id] = draft[id];
            }
          });
        }
      } else {
        clearDraftStorage(slug);
      }

      // Pre-fill submitter info if available
      if (submitter) {
        submitterInfo.name = submitter.name || "";
        submitterInfo.email = submitter.email || "";
      }
    } catch (err) {
      error = (err as Error).message;
    } finally {
      isLoading = false;
    }
  }

  function getFieldsForPage(docId: string, pageNumber: number): Field[] {
    if (!template || !submitter) {
      return [];
    }
    return myFields.filter((field) =>
      field.areas?.some((area: any) => area?.attachment_id === docId && area?.page === pageNumber)
    );
  }

  // Vue bound this as a :style object; Svelte style attributes take a CSS string
  function getFieldStyle(field: Field, docId: string, pageNumber: number): string {
    const area: any = field.areas?.find((a: any) => a?.attachment_id === docId && a?.page === pageNumber);
    if (!area) {
      return "";
    }

    const x = Number(area.x) || 0;
    const y = Number(area.y) || 0;
    const w = Number(area.w) || 0;
    const h = Number(area.h ?? area.z) || 0;

    return `left: ${x * 100}%; top: ${y * 100}%; width: ${w * 100}%; height: ${h * 100}%`;
  }

  function onImageLoad(e: Event): void {
    const target = e.target as HTMLImageElement;
    target.setAttribute("width", target.naturalWidth.toString());
    target.setAttribute("height", target.naturalHeight.toString());
  }

  function clearAllFieldHighlights(): void {
    if (highlightTimeout != null) {
      clearTimeout(highlightTimeout);
      highlightTimeout = null;
    }
    document.querySelectorAll(".doc-field-overlay").forEach((el) => {
      el.classList.remove("ring-2", "ring-primary", "rounded");
    });
  }

  function scrollToField(fieldId: string): void {
    // Readonly fields are not fillable: don't open the drawer for them
    const target = myFields.find((f) => f.id === fieldId);
    if (target && (target as any).readonly) {
      return;
    }
    expandedFieldId = fieldId;
    const idx = visibleFields.findIndex((f) => f.id === fieldId);
    if (idx >= 0) {
      currentFieldIndex = idx;
    }
    clearAllFieldHighlights();
    const docEl = document.querySelector(`[data-field-id="${fieldId}"].doc-field-overlay`);

    if (docEl) {
      docEl.scrollIntoView({ behavior: "smooth", block: "center" });
      docEl.classList.add("ring-2", "ring-primary", "rounded");
    }
    highlightTimeout = setTimeout(clearAllFieldHighlights, 2000);
  }

  function getPrevUnfilledIndex(): number {
    const fields = visibleFields;
    for (let i = currentFieldIndex - 1; i >= 0; i--) {
      if (!isFieldFilled(fields[i])) {
        return i;
      }
    }
    return -1;
  }

  function getNextUnfilledIndex(): number {
    const fields = visibleFields;
    for (let i = currentFieldIndex + 1; i < fields.length; i++) {
      if (!isFieldFilled(fields[i])) {
        return i;
      }
    }
    return -1;
  }

  function goToPrevField(): void {
    const idx = getPrevUnfilledIndex();
    if (idx < 0) {
      return;
    }
    currentFieldIndex = idx;
    const field = visibleFields[idx];
    if (field) {
      scrollToField(field.id);
    }
  }

  function goToNextField(): void {
    const idx = getNextUnfilledIndex();
    if (idx < 0) {
      return;
    }
    currentFieldIndex = idx;
    const field = visibleFields[idx];
    if (field) {
      scrollToField(field.id);
    }
  }

  function scrollToFieldOnDocument(fieldId: string): void {
    expandedFieldId = fieldId;
    clearAllFieldHighlights();
    const element = document.querySelector(`[data-field-id="${fieldId}"].doc-field-overlay`);
    if (element) {
      element.scrollIntoView({ behavior: "smooth", block: "center" });
      element.classList.add("ring-2", "ring-primary", "rounded");
      highlightTimeout = setTimeout(clearAllFieldHighlights, 2000);
    }
  }

  function onSigningAreaClick(e: MouseEvent): void {
    if (!expandedFieldId) {
      return;
    }
    const target = e.target as HTMLElement;
    if (target.closest(".field-form-drawer")) {
      return;
    }
    if (target.closest(".doc-field-overlay")) {
      return;
    }
    closeDrawer();
  }

  function closeDrawer(): void {
    expandedFieldId = null;
  }

  function onDrawerNavigate(direction: "prev" | "next"): void {
    if (direction === "prev") {
      goToPrevField();
    } else {
      goToNextField();
    }
  }

  function onDrawerFieldSelect(fieldId: string): void {
    scrollToField(fieldId);
  }

  function validateField(field: Field): void {
    // Readonly fields carry template defaults and are not user-editable
    if ((field as any).readonly) {
      return;
    }

    const value = formData[field.id];

    const required = fieldStates[field.id]?.required || field.required;
    if (required) {
      if (!value || (typeof value === "string" && value.trim() === "")) {
        fieldErrors[field.id] = t("signing.required");
        return;
      }
      if (field.type === "signature" || field.type === "initials" || field.type === "stamp") {
        if (typeof value !== "string" || !value.startsWith("data:")) {
          fieldErrors[field.id] = t("signing.required");
          return;
        }
      }
      if (Array.isArray(value) && value.length === 0) {
        fieldErrors[field.id] = t("signing.selectAtLeastOne");
        return;
      }
      if (field.type === "cells") {
        const cellCount = getCellCount(field);
        if (typeof value !== "string" || value.length !== cellCount) {
          fieldErrors[field.id] = t("signing.fillAllCells");
          return;
        }
      }
    }

    if (field.type === "number" && value != null && String(value).trim() !== "") {
      const num = Number(value);
      if (Number.isNaN(num)) {
        fieldErrors[field.id] = t("signing.invalidNumber");
        return;
      }
      const validation = (field as any).validation;
      if (validation?.min != null && num < Number(validation.min)) {
        fieldErrors[field.id] = t("signing.numberMin", { min: validation.min });
        return;
      }
      if (validation?.max != null && num > Number(validation.max)) {
        fieldErrors[field.id] = t("signing.numberMax", { max: validation.max });
        return;
      }
    }

    // DocuSeal-style regex validation for text-like fields (validation.pattern + message)
    if (["text", "phone", "cells"].includes(field.type) && typeof value === "string" && value !== "") {
      const validation = (field as any).validation;
      if (validation?.pattern) {
        let matches = true;
        try {
          matches = new RegExp(validation.pattern).test(value);
        } catch {
          // invalid regex in template: skip validation
        }
        if (!matches) {
          fieldErrors[field.id] = validation.message || t("signing.invalidFormat");
          return;
        }
      }
    }

    const { [field.id]: _removed, ...rest } = fieldErrors;
    fieldErrors = rest;
  }

  async function handleSubmit(): Promise<void> {
    if (!submitter || isSubmitting) {
      return;
    }

    // Validate all fields
    myFields.forEach((field) => validateField(field));

    if (!isFormValid) {
      return;
    }

    isSubmitting = true;

    try {
      const payload: Record<string, unknown> = { ...formData };
      myFields.forEach((field) => {
        if (hasWithSignatureId(field)) {
          const v = formData[field.id];
          if (v != null && String(v).trim() !== "") {
            if (!signatureIds[field.id]) {
              signatureIds[field.id] = generateSignatureId(field);
            }
            payload[`${field.id}_signature_id`] = signatureIds[field.id];
          }
        }
      });
      const response = await fetch(`/public/sign/${slug}/complete`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          fields: payload
        })
      });

      if (!response.ok) {
        throw new Error(t("signing.submitFailed"));
      }

      clearDraftStorage(slug);
      // Reload to show completed state
      await loadSubmission();
    } catch (err) {
      error = (err as Error).message;
    } finally {
      isSubmitting = false;
    }
  }

  function openDeclineModal(): void {
    declineReason = "";
    declineModalOpen = true;
  }

  async function handleDeclineSubmit(): Promise<void> {
    if (!submitter || isSubmitting) {
      return;
    }

    isSubmitting = true;

    try {
      const response = await fetch(`/public/sign/${slug}/decline`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reason: declineReason.trim() || undefined })
      });

      if (!response.ok) {
        throw new Error(t("signing.declineFailed"));
      }

      clearDraftStorage(slug);
      declineModalOpen = false;
      declineReason = "";
      await loadSubmission();
    } catch (err) {
      error = (err as Error).message;
    } finally {
      isSubmitting = false;
    }
  }

  function handleReset(event?: Event): void {
    event?.preventDefault?.();
    if (isSubmitting) {
      return;
    }

    const confirmed = confirm(t("signing.resetConfirm"));
    if (!confirmed) {
      return;
    }

    initializeFormData();
  }

  function formatDate(dateString?: string | null): string {
    if (!dateString) {
      return "—";
    }

    const d = new Date(dateString);
    if (Number.isNaN(d.getTime())) {
      return "—";
    }

    const loc = (signingLocale || getLocale() || "en").toString();
    return d.toLocaleString(loc, {
      year: "numeric",
      month: "long",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit"
    });
  }

  /** Format number for overlay display (comma, dot, space, usd, eur, gbp). */
  function formatNumberForDisplay(num: number, format: string, currency?: string): string {
    if (format === "usd" || format === "eur" || format === "gbp") {
      const cur = currency || (format === "eur" ? "EUR" : format === "gbp" ? "GBP" : "USD");
      return new Intl.NumberFormat([], { style: "currency", currency: cur }).format(num);
    }
    if (format === "comma") {
      return new Intl.NumberFormat("en-US", { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(num);
    }
    if (format === "dot") {
      return new Intl.NumberFormat("de-DE", { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(num);
    }
    if (format === "space") {
      return new Intl.NumberFormat("fr-FR", { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(num);
    }
    if (format === "percent") {
      return `${num}%`;
    }
    return String(num);
  }

  function validateSubmitterInfo(): void {
    submitterInfoErrors = {};

    if (!submitterInfo.name || submitterInfo.name.trim() === "") {
      submitterInfoErrors.name = t("signing.required");
    }

    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!submitterInfo.email || submitterInfo.email.trim() === "") {
      submitterInfoErrors.email = t("signing.required");
    } else if (!emailRegex.test(submitterInfo.email)) {
      submitterInfoErrors.email = t("signing.invalidEmail");
    }
  }

  async function handleUpdateSubmitter(event?: Event): Promise<void> {
    event?.preventDefault?.();

    if (isUpdatingSubmitter) {
      return;
    }

    validateSubmitterInfo();

    if (!isSubmitterInfoValid) {
      return;
    }

    isUpdatingSubmitter = true;
    error = "";

    try {
      const response = await fetch(`/public/sign/${slug}/update`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          name: submitterInfo.name.trim(),
          email: submitterInfo.email.trim()
        })
      });

      // Check content type before parsing
      const contentType = response.headers.get("content-type");
      let data: any = {};

      if (contentType && contentType.includes("application/json")) {
        try {
          data = await response.json();
        } catch {
          // If JSON parsing fails, read as text
          const text = await response.text();
          error = text || t("signing.updateFailed");
          return;
        }
      } else {
        // If not JSON, read as text
        const text = await response.text();
        error = text || t("signing.updateFailed");
        return;
      }

      if (!response.ok) {
        // Try to extract validation errors
        const errorMsg = data.message || data.error || t("signing.updateFailed");

        // Check if it's an email validation error
        if (errorMsg.toLowerCase().includes("email") && errorMsg.toLowerCase().includes("valid")) {
          submitterInfoErrors.email = t("signing.invalidEmail");
        } else if (errorMsg.toLowerCase().includes("name") || errorMsg.toLowerCase().includes("required")) {
          submitterInfoErrors.name = t("signing.required");
        } else {
          error = errorMsg;
        }
        return;
      }

      // Reload to show signing form
      await loadSubmission();
    } catch (err) {
      error = (err as Error).message || t("signing.updateFailed");
    } finally {
      isUpdatingSubmitter = false;
    }
  }
</script>

<div class="submitter-sign-page min-h-screen bg-[--color-base-200]">
  <!-- Loading State -->
  {#if isLoading}
    <div class="flex h-screen items-center justify-center">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {:else if error}
    <!-- Error State -->
    <div class="container mx-auto px-4 py-8">
      <div class="alert alert-error">
        <SvgIcon name="error-circle" class="h-6 w-6 shrink-0" />
        <span>{error}</span>
      </div>
    </div>
  {:else if submitter?.status === "completed"}
    <!-- Completed State -->
    <div class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5 text-center">
          <div class="mb-4 text-6xl text-success">✓</div>
          <h2 class="card-title justify-center text-2xl">{t("signing.completedTitle")}</h2>
          <p>{t("signing.completedThanks")}</p>
          <p class="text-sm text-[--color-base-content]/60">
            {t("signing.completedOn")}: {formatDate(submitter.completed_at)}
          </p>

          <div class="mt-5 flex flex-col items-center gap-2">
            {#if submissionStatus === "completed" && completedDocumentUrl}
              <a class="btn btn-primary btn-sm" href={completedDocumentUrl} target="_blank" rel="noopener">
                {t("common.download")}
              </a>
            {:else}
              <p class="text-sm text-[--color-base-content]/60">
                {t("signing.waitingForOthers")}
              </p>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {:else if submitter?.status === "declined"}
    <!-- Declined State -->
    <div class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5 text-center">
          <div class="mb-4 text-6xl text-error">✕</div>
          <h2 class="card-title justify-center text-2xl">{t("signing.declinedTitle")}</h2>
          <p>{t("signing.declinedText")}</p>
          <p class="text-sm text-[--color-base-content]/60">
            {t("signing.declinedOn")}: {formatDate(submitter.declined_at)}
          </p>
        </div>
      </div>
    </div>
  {:else if needsEmailOrName}
    <!-- Email/Name Form (if missing) -->
    <div class="container mx-auto px-4 py-8">
      <div class="mx-auto max-w-2xl rounded-lg border border-[var(--color-base-300)] bg-white">
        <div class="px-6 py-5">
          <h2 class="card-title mb-4 text-2xl">{t("signing.enterYourInfo")}</h2>
          <p class="mb-6 text-[--color-base-content]/60">{t("signing.enterYourInfoDescription")}</p>

          <form
            novalidate
            onsubmit={(e) => {
              e.preventDefault();
              handleUpdateSubmitter(e);
            }}
          >
            <div class="space-y-4">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-semibold">
                    {t("auth.firstName")}
                    <span class="text-error">*</span>
                  </span>
                </label>
                <input
                  bind:value={submitterInfo.name}
                  type="text"
                  class="input input-bordered {submitterInfoErrors.name ? 'input-error' : ''}"
                  placeholder={t("auth.firstName")}
                  onblur={validateSubmitterInfo}
                  oninput={() => {
                    submitterInfoErrors.name = "";
                  }}
                />
                {#if submitterInfoErrors.name}
                  <label class="label">
                    <span class="label-text-alt text-error">{submitterInfoErrors.name}</span>
                  </label>
                {/if}
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text font-semibold">
                    {t("auth.email")}
                    <span class="text-error">*</span>
                  </span>
                </label>
                <input
                  bind:value={submitterInfo.email}
                  type="text"
                  class="input input-bordered {submitterInfoErrors.email ? 'input-error' : ''}"
                  placeholder={t("auth.email")}
                  onblur={validateSubmitterInfo}
                  oninput={() => {
                    submitterInfoErrors.email = "";
                  }}
                />
                {#if submitterInfoErrors.email}
                  <label class="label">
                    <span class="label-text-alt text-error">{submitterInfoErrors.email}</span>
                  </label>
                {/if}
              </div>

              <div class="card-actions mt-6">
                <Button
                  type="submit"
                  variant="primary"
                  loading={isUpdatingSubmitter}
                  disabled={!isSubmitterInfoValid || isUpdatingSubmitter}
                >
                  {t("common.continue")}
                </Button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
  {:else}
    <!-- Signing Form -->
    <div class="container mx-auto px-4">
      <!-- Header -->
      <div class="mb-6 bg-white">
        <div class="px-6 py-5">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h1 class="card-title text-2xl">{template?.name}</h1>
              {#if template?.description}
                <p class="text-[--color-base-content]/60">{template.description}</p>
              {/if}
              <div class="mt-2 flex gap-2">
                {#if submitter?.name}
                  <div class="badge badge-outline">{submitter?.name}</div>
                {/if}
                {#if submitter?.email}
                  <div class="badge badge-outline">{submitter?.email}</div>
                {/if}
              </div>
            </div>

            <!-- Language + Decline -->
            <div class="flex flex-wrap items-end gap-3">
              {#if showLanguageSelector}
                <div class="w-full sm:w-48">
                  <label class="mb-1 block text-xs font-medium text-gray-600">{t("settings.language")}</label>
                  <select
                    class="select select-bordered select-sm w-full"
                    value={signingLocale}
                    onchange={onSigningLocaleChange}
                  >
                    {#each Object.entries(SUPPORTED_LOCALES) as [code, name] (code)}
                      <option value={code}>{name}</option>
                    {/each}
                  </select>
                </div>
              {/if}
              <Button
                type="button"
                variant="ghost"
                size="sm"
                class="border-red-300 text-red-700 hover:bg-red-50"
                disabled={isSubmitting}
                onclick={openDeclineModal}
              >
                {t("signing.decline")}
              </Button>
            </div>
          </div>
        </div>
      </div>

      <!-- Full-width document preview (padding for fixed bottom panel) -->
      <div class="relative" onclick={onSigningAreaClick}>
        <div class="overflow-hidden">
          {#each sortedDocuments as doc (doc.id)}
            {#each getSortedPreviewImages(doc) as page, pageIndex (page.id)}
              {@const pageNum = pageNumberFromPreview(page, pageIndex)}
              <div class="relative mb-4">
                <div class="relative">
                  <img
                    src={`${page.url}/${page.filename}`}
                    alt={`Page ${pageIndex + 1}`}
                    width={page.metadata?.width}
                    height={page.metadata?.height}
                    class="mb-4 rounded border border-[#e7e2df]"
                    loading="lazy"
                    onload={onImageLoad}
                  />
                  <!-- Field Overlays: label outside, bordered block unchanged -->
                  {#each getFieldsForPage(doc.id, pageNum) as field (`${field.id}-${doc.id}-${pageNum}`)}
                    <div
                      id={`doc-field-${field.id}-${doc.id}-${pageNum}`}
                      class="doc-field-overlay absolute cursor-pointer rounded transition"
                      data-field-id={field.id}
                      style={getFieldStyle(field, doc.id, pageNum)}
                      onclick={() => scrollToField(field.id)}
                    >
                      <span
                        class="absolute left-0 w-full truncate text-xs text-[--color-base-content]/80"
                        style="bottom: 100%; margin-bottom: 2px"
                      >
                        {getFieldLabel(field)}
                      </span>
                      <div
                        class="absolute inset-0 flex items-center justify-center overflow-hidden rounded border-2 border-primary bg-primary/10 hover:bg-primary/20"
                      >
                        {#if getFieldDisplayValue(field)}
                          {#if isFieldDisplayImage(field)}
                            <img
                              src={getFieldDisplayValue(field)}
                              class="max-h-full max-w-full object-contain"
                              alt=""
                            />
                          {:else}
                            <span class="truncate px-1 text-xs">
                              {getFieldDisplayValue(field)}
                            </span>
                          {/if}
                        {:else if fieldIcons[field.type]}
                          <SvgIcon
                            name={fieldIcons[field.type]}
                            width="20"
                            height="20"
                            class="flex-shrink-0 opacity-50"
                            stroke-width="1.6"
                          />
                        {/if}
                      </div>
                    </div>
                  {/each}
                </div>
              </div>
            {/each}
          {/each}
        </div>

        <!-- Single floating panel: drawer (when field open) + action bar (always) -->
        <FieldFormDrawer
          isOpen={expandedFieldId !== null}
          field={activeField}
          bind:value={getDrawerValue, setDrawerValue}
          allFields={visibleFields}
          {filledFieldIds}
          {fieldStates}
          {fieldErrors}
          {calculatedValues}
          {signatureIds}
          {getFieldLabel}
          {getCellCount}
          {getSignatureFormat}
          {hasWithSignatureId}
          {isFieldFilled}
          {initialsDefault}
          canGoPrev={prevUnfilledIndex >= 0}
          canGoNext={nextUnfilledIndex >= 0}
          {isFormValid}
          {isSubmitting}
          {prevUnfilledIndex}
          {nextUnfilledIndex}
          onClose={closeDrawer}
          onNavigate={onDrawerNavigate}
          onFieldSelect={onDrawerFieldSelect}
          onBlur={validateField}
          onReset={handleReset}
          onSubmit={handleSubmit}
        />
      </div>

      <!-- Decline modal -->
      <Modal
        bind:open={declineModalOpen}
        title={t("signing.decline")}
        size="md"
        onClose={() => {
          declineModalOpen = false;
        }}
      >
        <div class="space-y-3">
          <label class="block text-sm font-medium text-[--color-base-content]">
            {t("signing.declineReasonLabel")}
          </label>
          <textarea
            bind:value={declineReason}
            class="textarea textarea-bordered min-h-[100px] w-full resize-y"
            placeholder={t("signing.declineReasonPlaceholder")}
            rows={4}></textarea>
        </div>
        {#snippet footer()}
          <div class="flex justify-end gap-2">
            <Button
              type="button"
              variant="ghost"
              disabled={isSubmitting}
              onclick={() => {
                declineModalOpen = false;
              }}
            >
              {t("common.cancel")}
            </Button>
            <Button
              type="button"
              variant="error"
              loading={isSubmitting}
              disabled={isSubmitting}
              onclick={handleDeclineSubmit}
            >
              {t("signing.decline")}
            </Button>
          </div>
        {/snippet}
      </Modal>
    </div>
  {/if}
</div>

<style>
  .submitter-sign-page {
    @apply min-h-screen;
  }
</style>
