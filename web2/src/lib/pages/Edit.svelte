<script lang="ts">
  import { onMount, onDestroy, tick, setContext } from "svelte";
  import { page } from "$app/state";
  import { beforeNavigate } from "$app/navigation";
  import { t } from "@/i18n/index.svelte";
  import Document from "@/components/template/Document.svelte";
  import DocumentPreview from "@/components/template/Preview.svelte";
  import Fields from "@/components/field/List.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import LoadingSpinner from "@/components/ui/LoadingSpinner.svelte";
  import FormModal from "@/components/common/FormModal.svelte";
  import type { Template } from "@/models";
  import { apiGet, apiPost } from "@/services/api";
  import { fetchWithAuth } from "@/utils/auth";
  import { fileToBase64Payload } from "@/utils/file";
  import { v4 } from "uuid";

  let template = $state<any>(null);
  let undoStack = $state<any[]>([]);
  let redoStack = $state<any[]>([]);
  let lastRedoData = $state<any>();

  let previewsRef = $state<HTMLElement | null>(null);

  let drawField = $state<any>();
  let drawOption = $state<any>();

  let selectedSubmitter = $state<any>();
  let selectedAreaRef = $state<any>();
  let editable = $state(true); // or whatever the initial value should be
  let dragField = $state<any>();

  let onSave = $state<((tpl: any) => void) | undefined>();

  let defaultSubmitters = $state<any[]>([]);
  let defaultFields = $state<string[]>(["text", "signature"]);
  let onlyDefinedFields = $state(false);

  const fetchOptions = { headers: {} };

  let templateFileInput = $state<HTMLInputElement | null>(null);
  let selectedFile = $state<File | null>(null);
  let uploading = $state(false);
  let uploadError = $state<string | null>(null);
  let loadingTemplate = $state(true);

  let showAddPagesModal = $state(false);
  let addPagesModalRef = $state<ReturnType<typeof FormModal> | undefined>();
  let addPagesFileInput = $state<HTMLInputElement | null>(null);
  let addPagesSelectedFile = $state<File | null>(null);
  let addPagesError = $state<string | null>(null);

  setContext("template", {
    get value() {
      return template;
    },
    set value(v: any) {
      template = v;
    }
  });
  setContext("selectedAreaRef", {
    get value() {
      return selectedAreaRef;
    },
    set value(v: any) {
      selectedAreaRef = v;
    }
  });
  setContext("save", save);
  setContext("baseFetch", baseFetch);

  // Manual save state (document + form edits)
  let isSaving = $state(false);
  let saveQueued = $state(false);
  let isDirty = $state(false);
  let lastSaveError = $state<string | null>(null);
  let lastSavedAt = $state<number | null>(null);

  let documentsEl = $state<HTMLElement | null>(null);

  const selectedField = $derived.by(() => {
    if (!template || !template.fields) {
      return null;
    }
    return template.fields.find((f: any) => f.areas?.includes(selectedAreaRef));
  });

  const sortedDocuments = $derived.by(() => {
    if (!template || !template.schema || !template.documents) {
      return [];
    }
    return template.schema.map((item: any) => {
      return template.documents.find((doc: any) => doc.id === item.attachment_id);
    });
  });

  const fieldAreasIndex = $derived.by(() => {
    if (!template || !template.fields) {
      return {};
    }
    const areas: any = {};
    template.fields.forEach((f: any) => {
      (f.areas || []).forEach((a: any) => {
        areas[a.attachment_id] ||= {};
        const acc = (areas[a.attachment_id][a.page] ||= []);
        acc.push({ area: a, field: f });
      });
    });
    return areas;
  });

  // Writable derived: reset when the document count changes; bind:this fills the slots in between
  let documentRefEls: Array<ReturnType<typeof Document> | undefined> = $derived(new Array(sortedDocuments.length));

  function applyLoadedTemplate(tpl: any): void {
    template = tpl;

    // Ensure at least one submitter exists (editor requires one)
    if (!template.submitters || template.submitters.length === 0) {
      template.submitters = [
        {
          id: v4(),
          name: "Signer 1",
          colorIndex: 0
        }
      ];
    }

    if (template.submitters && template.submitters.length > 0) {
      selectedSubmitter = template.submitters[0];
    }

    // Reset save state when template is loaded
    isDirty = false;
    lastSavedAt = Date.now();
    lastSaveError = null;
  }

  async function uploadAndApplyPdf(file: File, append = false): Promise<void> {
    const templateId = page.params.id as string;
    if (!templateId) {
      return;
    }

    uploading = true;
    uploadError = null;

    try {
      const base64 = await fileToBase64Payload(file);
      if (!base64) {
        throw new Error("Failed to read file");
      }

      const res = await apiPost(`/api/v1/templates/${templateId}/from-file`, {
        type: "pdf",
        file_base64: base64,
        append
      });

      if (!res || !res.data) {
        throw new Error("Unexpected response");
      }

      applyLoadedTemplate(res.data);
      // Reset save state after applying PDF
      isDirty = false;
      lastSavedAt = Date.now();
      lastSaveError = null;
      undoStack = [JSON.stringify(template)];
      redoStack = [];
    } catch (err: any) {
      console.error("Failed to upload/apply PDF:", err);
      throw err;
    } finally {
      uploading = false;
    }
  }

  onMount(async () => {
    loadingTemplate = true;
    try {
      // Get template ID from route params
      const templateId = page.params.id as string;

      if (!templateId) {
        console.error("Template ID is required");
        return;
      }

      // Load specific template by ID
      const res = await apiGet<Template>(`/api/v1/templates/${templateId}`);
      // API v1 returns: { message: "template", data: Template }
      if (res && res.data) {
        applyLoadedTemplate(res.data);
      } else {
        console.error("Template not found");
      }
    } catch (error) {
      console.error("Failed to load template:", error);
      // If auth failed, redirect will happen automatically
      if (page.url.pathname.match(/\/templates\/[^/]+\/edit$/)) {
        // Only show error if still on edit page (not redirected)
        console.error("Could not load template data");
      }
    }

    try {
      if (template) {
        undoStack = [JSON.stringify(template)];
      }
      redoStack = [];

      await tick();
      document.addEventListener("keyup", onKeyUp);
      window.addEventListener("keydown", onKeyDown);
    } finally {
      loadingTemplate = false;
    }

    window.addEventListener("beforeunload", onBeforeUnload);
    document.addEventListener("visibilitychange", onVisibilityChange);
  });

  onDestroy(() => {
    document.removeEventListener("keyup", onKeyUp);
    window.removeEventListener("keydown", onKeyDown);
    window.removeEventListener("beforeunload", onBeforeUnload);
    document.removeEventListener("visibilitychange", onVisibilityChange);
  });

  beforeNavigate(async () => {
    try {
      await flushSave({ keepalive: false });
    } catch {
      // error is shown in UI; don't block navigation
    }
  });

  function buildUpdatePayload(): Record<string, any> {
    return {
      name: template?.name,
      schema: template?.schema ?? [],
      submitters: template?.submitters ?? [],
      fields: template?.fields ?? []
    };
  }

  async function putTemplateUpdate(payload: any, { keepalive }: { keepalive: boolean }): Promise<void> {
    const id = template?.id;
    if (!id) {
      return;
    }

    const res = await fetchWithAuth(`/api/v1/templates/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
      credentials: "include",
      keepalive
    });

    if (!res.ok) {
      const data = await res.json().catch(() => null);
      const msg = data?.message || data?.error || `HTTP ${res.status}`;
      throw new Error(msg);
    }
  }

  async function flushSave({ keepalive } = { keepalive: false }): Promise<void> {
    if (!template) {
      return;
    }

    if (isSaving) {
      saveQueued = true;
      return;
    }

    // Nothing to do
    if (!isDirty && !saveQueued) {
      return;
    }

    isSaving = true;
    lastSaveError = null;
    try {
      await putTemplateUpdate(buildUpdatePayload(), { keepalive });
      lastSavedAt = Date.now();
      isDirty = false;
      if (onSave) {
        onSave(template);
      }
    } catch (err: any) {
      lastSaveError = err?.message || "Failed to save";
      isDirty = true;
      throw err;
    } finally {
      isSaving = false;
      if (saveQueued) {
        saveQueued = false;
        await flushSave({ keepalive: false });
      }
    }
  }

  function onBeforeUnload(): void {
    // Fire-and-forget; keepalive lets it continue after navigation/reload.
    if (isDirty || saveQueued) {
      void flushSave({ keepalive: true });
    }
  }

  function onVisibilityChange(): void {
    if (document.visibilityState === "hidden") {
      if (isDirty || saveQueued) {
        void flushSave({ keepalive: true });
      }
    }
  }

  function undo(): void {
    if (undoStack.length > 1) {
      undoStack.pop();
      const stringData = undoStack[undoStack.length - 1];
      const currentStringData = JSON.stringify(template);

      if (stringData && stringData !== currentStringData) {
        redoStack.push(currentStringData);
        Object.assign(template, JSON.parse(stringData));
        save();
      }
    }
  }

  function redo(): void {
    const stringData = redoStack.pop();
    lastRedoData = stringData;
    const currentStringData: any = JSON.stringify(template);

    if (stringData && stringData !== currentStringData) {
      if (undoStack[undoStack.length - 1] !== currentStringData) {
        undoStack.push(currentStringData);
      }
      Object.assign(template, JSON.parse(stringData));
      save();
    }
  }

  function scrollIntoDocument(item: any): void {
    const docIndex = sortedDocuments.findIndex((doc: any) => doc?.id === item.attachment_id);
    const scrollEl = docIndex !== -1 ? (documentsEl?.children[docIndex] as HTMLElement | undefined) : undefined;
    if (scrollEl) {
      scrollEl.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  }

  function getPageMask(documentRef: ReturnType<typeof Document> | undefined, pageIndex: number): HTMLDivElement | null {
    if (!documentRef) {
      return null;
    }
    const pageRef = documentRef.pageRefs?.[pageIndex];
    const withMask = pageRef as unknown as { mask?: HTMLDivElement | null };
    if (withMask?.mask) {
      return withMask.mask;
    }
    const docId = documentRef.document?.id;
    if (!docId || !documentsEl) {
      return null;
    }
    const docIndex = sortedDocuments.findIndex((d: any) => d?.id === docId);
    if (docIndex === -1) {
      return null;
    }
    const docRoot = documentsEl.children[docIndex] as HTMLElement | undefined;
    const masks = docRoot?.querySelectorAll("#mask");
    return (masks?.[pageIndex] as HTMLDivElement) ?? null;
  }

  function clearDrawField(): void {
    if (drawField && !drawOption && drawField.areas.length === 0) {
      const fieldIndex = template.fields.indexOf(drawField);

      if (fieldIndex !== -1) {
        template.fields.splice(fieldIndex, 1);
      }
    }
    drawField = null;
    drawOption = null;
  }

  function onKeyUp(e: KeyboardEvent): void {
    if (e.code === "Escape") {
      clearDrawField();
      selectedAreaRef = null;
    }
    if (
      editable &&
      ["Backspace", "Delete"].includes(e.key) &&
      selectedAreaRef &&
      document.activeElement === document.body
    ) {
      removeArea(selectedAreaRef);
      selectedAreaRef = null;
    }
  }

  function onKeyDown(event: KeyboardEvent): void {
    if ((event.metaKey && event.shiftKey && event.key === "z") || (event.ctrlKey && event.key === "Z")) {
      event.stopImmediatePropagation();
      event.preventDefault();
      redo();
    } else if ((event.ctrlKey || event.metaKey) && event.key === "z") {
      event.stopImmediatePropagation();
      event.preventDefault();
      undo();
    }
  }

  function removeArea(area: any): void {
    const field = template.fields.find((f: any) => f.areas?.includes(area));
    field.areas.splice(field.areas.indexOf(area), 1);

    if (!field.areas.length) {
      template.fields.splice(template.fields.indexOf(field), 1);
    }
    save();
  }

  function handleSelectSubmitter(submitterId: string): void {
    const submitter = template.submitters.find((s: any) => s.id === submitterId);
    if (submitter) {
      selectedSubmitter = submitter;
    }
  }

  function pushUndo(): void {
    const stringData: any = JSON.stringify(template);

    if (undoStack[undoStack.length - 1] !== stringData) {
      undoStack.push(stringData);
      if (lastRedoData !== stringData) {
        redoStack = [];
      }
    }
  }

  function onDraw(area: any): void {
    const documentRefForArea = documentRefEls.find((e) => e?.document?.id === area.attachment_id);
    const maxPage = documentRefForArea?.pageRefs?.length ?? 1;
    const clampedPage = Math.max(0, Math.min(typeof area.page === "number" ? area.page : 0, maxPage - 1));
    area.page = clampedPage;

    if (drawField) {
      if (drawOption) {
        const areaWithoutOption = drawField.areas?.find((a: any) => !a.option_id);

        if (areaWithoutOption && !drawField.areas.find((a: any) => a.option_id === drawField.options[0].id)) {
          areaWithoutOption.option_id = drawField.options[0].id;
        }

        area.option_id = drawOption.id;
      }

      if (area.w === 0 || area.h === 0) {
        const previousArea = drawField.areas?.[drawField.areas.length - 1];

        if (selectedField?.type === drawField.type) {
          area.w = selectedAreaRef.w;
          area.h = selectedAreaRef.h;
        } else if (previousArea) {
          area.w = previousArea.w;
          area.h = previousArea.h;
        } else {
          const documentRef = documentRefEls.find((e) => e && e.document && e.document.id === area.attachment_id);
          if (!documentRef) {
            return;
          }
          const maxPageInner = documentRef.pageRefs.length;
          area.page = Math.max(0, Math.min(area.page, maxPageInner - 1));
          const pageMask = getPageMask(documentRef, area.page);

          if (!pageMask) {
            return;
          }

          if (drawField.type === "checkbox" || drawOption) {
            area.w = pageMask.clientWidth / 30 / pageMask.clientWidth;
            area.h =
              (pageMask.clientWidth / 30 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight);
          } else if (drawField.type === "image") {
            area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
            area.h = (pageMask.clientWidth / 5 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight);
          } else if (drawField.type === "signature" || drawField.type === "stamp") {
            area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
            area.h =
              ((pageMask.clientWidth / 5 / pageMask.clientWidth) * (pageMask.clientWidth / pageMask.clientHeight)) / 2;
          } else if (drawField.type === "initials") {
            area.w = pageMask.clientWidth / 10 / pageMask.clientWidth;
            area.h = pageMask.clientWidth / 35 / pageMask.clientWidth;
          } else {
            area.w = pageMask.clientWidth / 5 / pageMask.clientWidth;
            area.h = pageMask.clientWidth / 35 / pageMask.clientWidth;
          }
        }

        area.x -= area.w / 2;
        area.y -= area.h / 2;
      }

      drawField.areas ||= [];

      const areaToAdd = {
        ...area,
        page: Math.max(0, Math.min(typeof area.page === "number" ? area.page : 0, maxPage - 1)),
        attachment_id: area.attachment_id || ""
      };

      const insertBeforeAreaIndex = drawField.areas.findIndex((a: any) => {
        return a.attachment_id === areaToAdd.attachment_id && a.page > areaToAdd.page;
      });

      if (insertBeforeAreaIndex !== -1) {
        drawField.areas.splice(insertBeforeAreaIndex, 0, areaToAdd);
      } else {
        drawField.areas.push(areaToAdd);
      }

      if (template.fields.indexOf(drawField) === -1) {
        template.fields.push(drawField);
      }

      drawField = null;
      drawOption = null;
      selectedAreaRef = areaToAdd;
      save();
    } else {
      const documentRef = documentRefEls.find((e) => e?.document?.id === area.attachment_id);
      if (!documentRef) {
        return;
      }
      const maxPageInner = documentRef.pageRefs.length;
      area.page = Math.max(0, Math.min(area.page, maxPageInner - 1));
      const pageMask = getPageMask(documentRef, area.page);

      if (!pageMask) {
        return;
      }

      let type = pageMask.clientWidth * area.w < 35 ? "checkbox" : "text";
      if (type === "checkbox") {
        const previousField = [...template.fields].reverse().find((f: any) => f.type === type);
        const previousArea = previousField?.areas?.[previousField.areas.length - 1];

        if (previousArea || area.w) {
          const areaW = previousArea?.w || 30 / pageMask.clientWidth;
          const areaH = previousArea?.h || 30 / pageMask.clientHeight;

          if (pageMask.clientWidth * area.w < 5) {
            area.x = area.x - areaW / 2;
            area.y = area.y - areaH / 2;
          }

          area.w = areaW;
          area.h = areaH;
        }
      }

      if (area.w) {
        // Ensure page number and attachment_id are preserved
        const areaToAdd = {
          ...area,
          page: typeof area.page === "number" ? area.page : 0,
          attachment_id: area.attachment_id || ""
        };

        const field = {
          name: "",
          id: v4(),
          required: type !== "checkbox",
          type,
          submitter_id: selectedSubmitter.id,
          areas: [areaToAdd]
        };

        template.fields.push(field);
        selectedAreaRef = areaToAdd;
        save();
      }
    }
  }

  function onDropfield(area: any): void {
    const field = {
      name: "",
      id: v4(),
      submitter_id: selectedSubmitter.id,
      required: true,
      ...dragField
    };

    if (["select", "multiple", "radio"].includes(field.type)) {
      field.options = [{ value: "", id: v4() }];
    }

    if (field.type === "stamp") {
      field.readonly = false;
    }

    if (field.type === "date") {
      field.preferences = {
        format: Intl.DateTimeFormat().resolvedOptions().locale.endsWith("-US") ? "MM/DD/YYYY" : "DD/MM/YYYY"
      };
    }

    const fieldArea: any = {
      x: (area.x - 6) / area.maskW,
      y: area.y / area.maskH,
      page: area.page,
      attachment_id: area.attachment_id
    };

    const previousField = [...template.fields].reverse().find((f: any) => f.type === field.type);
    let baseArea;
    if (selectedField?.type === field.type) {
      baseArea = selectedAreaRef;
    } else if (previousField?.areas?.length) {
      baseArea = previousField.areas[previousField.areas.length - 1];
    } else {
      if (["checkbox"].includes(field.type)) {
        baseArea = {
          w: area.maskW / 30 / area.maskW,
          h: (area.maskW / 30 / area.maskW) * (area.maskW / area.maskH)
        };
      } else if (field.type === "image") {
        baseArea = {
          w: area.maskW / 5 / area.maskW,
          h: (area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)
        };
      } else if (field.type === "signature" || field.type === "stamp") {
        baseArea = {
          w: area.maskW / 5 / area.maskW,
          h: ((area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)) / 2
        };
      } else if (field.type === "initials") {
        baseArea = {
          w: area.maskW / 10 / area.maskW,
          h: area.maskW / 35 / area.maskW
        };
      } else {
        baseArea = {
          w: area.maskW / 5 / area.maskW,
          h: area.maskW / 35 / area.maskW
        };
      }
    }

    fieldArea.w = baseArea.w;
    fieldArea.h = baseArea.h;

    // Fix: if baseArea.h is undefined, calculate default height based on field type
    if (!fieldArea.h) {
      if (["checkbox"].includes(field.type)) {
        fieldArea.h = (area.maskW / 30 / area.maskW) * (area.maskW / area.maskH);
      } else if (field.type === "image") {
        fieldArea.h = (area.maskW / 5 / area.maskW) * (area.maskW / area.maskH);
      } else if (field.type === "signature" || field.type === "stamp") {
        fieldArea.h = ((area.maskW / 5 / area.maskW) * (area.maskW / area.maskH)) / 2;
      } else if (field.type === "initials") {
        fieldArea.h = area.maskW / 35 / area.maskW;
      } else {
        // Default for text and other fields
        fieldArea.h = area.maskW / 35 / area.maskW;
      }
    }

    fieldArea.y = fieldArea.y - fieldArea.h / 2;

    if (field.type === "cells") {
      fieldArea.cell_w = baseArea.cell_w || baseArea.w / 5;
    }

    field.areas = [fieldArea];
    selectedAreaRef = fieldArea;
    template.fields.push(field);
    save();
  }

  function onDocumentRemove(item: any): void {
    if (!template) {
      return;
    }
    if (window.confirm("Are you sure?")) {
      if (template.schema) {
        template.schema.splice(template.schema.indexOf(item), 1);
      }
    }

    if (template.fields) {
      template.fields.forEach((field: any) => {
        [...(field.areas || [])].forEach((area) => {
          if (area.attachment_id === item.attachment_id) {
            field.areas.splice(field.areas.indexOf(area), 1);
          }
        });
      });
    }
    save();
  }

  function moveDocument(item: any, direction: any): void {
    if (!template || !template.schema) {
      return;
    }
    const currentIndex = template.schema.indexOf(item);
    if (currentIndex === -1) {
      return;
    }
    template.schema.splice(currentIndex, 1);

    if (currentIndex + direction > template.schema.length) {
      template.schema.unshift(item);
    } else if (currentIndex + direction < 0) {
      template.schema.push(item);
    } else {
      template.schema.splice(currentIndex + direction, 0, item);
    }
    save();
  }

  function scrollToArea(area: any): void {
    const documentRef = documentRefEls.find((a) => a && a.document && a.document.id === area.attachment_id);
    if (documentRef) {
      documentRef.scrollToArea(area);
      selectedAreaRef = area;
    }
  }

  function baseFetch(path: string, options: RequestInit = {}): Promise<Response> {
    // Normalize path: replace /api/ with /api/v1/ if needed
    let normalizedPath = path;
    if (path.startsWith("/api/") && !path.startsWith("/api/v1/")) {
      normalizedPath = path.replace("/api/", "/api/v1/");
    }
    // Use fetchWithAuth to ensure token is included in headers
    return fetchWithAuth(normalizedPath, {
      ...options,
      headers: { ...fetchOptions.headers, ...options.headers }
    });
  }

  const handleFileSelect = async (event: Event) => {
    const input = event.target as HTMLInputElement;
    const file = input?.files?.[0];
    if (!file) {
      selectedFile = null;
      return;
    }

    if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
      uploadError = t("templates.invalidFileType");
      selectedFile = null;
      if (input) {
        input.value = "";
      }
      return;
    }

    selectedFile = file;
    uploadError = null;
    try {
      await uploadAndApplyPdf(file, false);
      removeSelectedFile();
    } catch (err: any) {
      uploadError = err?.message || t("templates.uploadError");
    }
  };

  const handleDrop = async (event: DragEvent) => {
    if (uploading) {
      return;
    }
    event.preventDefault();
    const file = event.dataTransfer?.files?.[0];
    if (!file) {
      return;
    }

    if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
      uploadError = t("templates.invalidFileType");
      return;
    }

    selectedFile = file;
    if (templateFileInput) {
      const dt = new DataTransfer();
      dt.items.add(file);
      templateFileInput.files = dt.files;
    }
    uploadError = null;
    try {
      await uploadAndApplyPdf(file, false);
      removeSelectedFile();
    } catch (err: any) {
      uploadError = err?.message || t("templates.uploadError");
    }
  };

  const removeSelectedFile = () => {
    selectedFile = null;
    uploadError = null;
    if (templateFileInput) {
      templateFileInput.value = "";
    }
  };

  const resetAddPagesSelection = () => {
    addPagesSelectedFile = null;
    addPagesError = null;
    if (addPagesFileInput) {
      addPagesFileInput.value = "";
    }
  };

  const handleAddPagesFileSelect = (event: Event) => {
    const input = event.target as HTMLInputElement;
    const file = input?.files?.[0];
    if (!file) {
      addPagesSelectedFile = null;
      return;
    }
    if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
      addPagesError = t("templates.invalidFileType");
      addPagesSelectedFile = null;
      input.value = "";
      return;
    }
    addPagesError = null;
    addPagesSelectedFile = file;
  };

  const handleAddPagesDrop = (event: DragEvent) => {
    if (uploading) {
      return;
    }
    event.preventDefault();
    const file = event.dataTransfer?.files?.[0];
    if (!file) {
      return;
    }

    if (file.type !== "application/pdf" && !file.name.toLowerCase().endsWith(".pdf")) {
      addPagesError = t("templates.invalidFileType");
      return;
    }

    addPagesError = null;
    addPagesSelectedFile = file;
    if (addPagesFileInput) {
      const dt = new DataTransfer();
      dt.items.add(file);
      addPagesFileInput.files = dt.files;
    }
  };

  const handleAddPagesSubmit = async (): Promise<void> => {
    if (!addPagesSelectedFile) {
      addPagesError = t("templates.selectPdfFile");
      // Keep modal open; stop loading state on button
      if (addPagesModalRef?.resetSubmitting) {
        addPagesModalRef.resetSubmitting();
      }
      return;
    }

    addPagesError = null;
    try {
      await uploadAndApplyPdf(addPagesSelectedFile, true);
      showAddPagesModal = false;
      resetAddPagesSelection();
    } catch (err: any) {
      addPagesError = err?.message || t("templates.uploadError");
      if (addPagesModalRef?.resetSubmitting) {
        addPagesModalRef.resetSubmitting();
      }
    }
  };

  function updateTemplateBuilderDataset(): void {
    tick().then(() => {
      const templateBuilder = document.querySelector("template-builder") as HTMLElement | null;
      if (templateBuilder && template) {
        templateBuilder.dataset.template = JSON.stringify(template);
      }
    });
  }

  async function save({ force } = { force: false }): Promise<object> {
    if (!template) {
      return {};
    }

    // Mark as dirty when any change occurs (even without auto-save)
    isDirty = true;
    updateTemplateBuilderDataset();
    pushUndo();

    if (!force) {
      return Promise.resolve({});
    }

    try {
      await flushSave({ keepalive: false });
    } catch {
      // error is stored in lastSaveError
    }
    return {};
  }
</script>

<div class="edit-page flex h-full min-h-0 flex-col">
  <div class="mb-6 flex items-center justify-between">
    <h1 class="text-3xl font-bold">
      {#if template}
        {t("templates.title")}
        <span class="mx-2 text-gray-500">→</span>
        <span class="text-gray-900">{template.name}</span>
      {:else}
        {t("templates.title")}
        <span class="mx-2 text-gray-500">→</span>
        <span class="text-gray-900">{t("templates.editor")}</span>
      {/if}
    </h1>

    {#if !loadingTemplate && template}
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="inline-flex items-center rounded-lg border bg-white px-3 py-2 text-sm font-medium hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 {lastSaveError
            ? 'border-red-300 text-red-700'
            : !isDirty && lastSavedAt
              ? 'border-green-300 text-green-700'
              : 'border-gray-200 text-gray-700'}"
          disabled={uploading || isSaving || (!isDirty && !lastSaveError)}
          title={lastSaveError || ""}
          onclick={() => save({ force: true })}
        >
          {#if isSaving}
            <LoadingSpinner class="mr-2 h-4 w-4" />
          {:else if !isDirty && lastSavedAt && !lastSaveError}
            <SvgIcon name="check-circle" class="mr-2 h-4 w-4" />
          {:else if lastSaveError}
            <SvgIcon name="error-circle" class="mr-2 h-4 w-4" />
          {/if}
          <span>
            {isSaving
              ? t("common.loading")
              : lastSaveError
                ? t("common.save")
                : !isDirty && lastSavedAt
                  ? t("success.saved")
                  : t("common.save")}
          </span>
        </button>
      </div>
    {/if}
  </div>

  {#if loadingTemplate}
    <div class="flex flex-1 items-center justify-center">
      <div class="flex items-center gap-3 rounded-xl bg-white px-5 py-4 text-gray-700">
        <LoadingSpinner class="h-5 w-5" />
        <div class="text-sm font-medium">{t("templates.loadingTemplate")}</div>
      </div>
    </div>
  {:else if template && template.schema && template.schema.length > 0}
    <div class="flex min-h-0 flex-1 overflow-hidden">
      <!-- Left previews: fixed column; scrolls only when hovered and overflowing -->
      <div bind:this={previewsRef} class="hidden h-full w-28 flex-none overflow-x-hidden overflow-y-auto pr-3 lg:block">
        {#each template && template.schema ? template.schema : [] as item, index (index)}
          <DocumentPreview
            withArrows={(template && template.schema && template.schema.length) > 1}
            {item}
            document={sortedDocuments[index]}
            {editable}
            onScrollTo={scrollIntoDocument}
            onRemove={onDocumentRemove}
            onUp={(i) => moveDocument(i, -1)}
            onDown={(i) => moveDocument(i, 1)}
            onChange={() => save()}
          />
        {/each}
        <button
          type="button"
          class="mt-2 flex w-full cursor-pointer flex-col items-center justify-center rounded border-2 border-dashed border-gray-300 bg-gray-50/80 py-4 transition-colors hover:border-gray-400 hover:bg-gray-100 focus:ring-2 focus:ring-gray-400 focus:ring-offset-1 focus:outline-none disabled:cursor-not-allowed disabled:opacity-60 {template
            ?.schema?.length
            ? 'aspect-[210/297]'
            : ''} min-h-[5rem]"
          disabled={uploading}
          aria-label={t("templates.addPages")}
          onclick={() => (showAddPagesModal = true)}
        >
          <SvgIcon name="plus" class="h-6 w-6 text-gray-400" />
          <span class="mt-1.5 text-center text-xs font-medium text-gray-500">{t("templates.addPages")}</span>
        </button>
      </div>

      <!-- Center documents: the ONLY column that should scroll by default -->
      <div class="min-h-0 flex-1 overflow-hidden">
        <div bind:this={documentsEl} class="h-full overflow-x-hidden overflow-y-auto pr-3.5 pl-0.5">
          {#each sortedDocuments as document, index (document.id)}
            <Document
              bind:this={documentRefEls[index]}
              areasIndex={fieldAreasIndex[document.id]}
              {document}
              isDrag={!!dragField}
              defaultFields={[]}
              allowDraw={!onlyDefinedFields}
              {drawField}
              {editable}
              {onDraw}
              onDropField={onDropfield}
              onRemoveArea={removeArea}
              onSelectSubmitter={handleSelectSubmitter}
            />
          {/each}
        </div>
      </div>

      <!-- Right fields panel: fixed column; scrolls only when hovered and overflowing -->
      <div class="relative hidden h-full w-80 flex-none overflow-x-hidden overflow-y-auto pl-0.5 md:block">
        {#if drawField}
          <div class="sticky inset-0 z-20 h-full">
            <div class="space-y-4 rounded-lg bg-[--color-base-300] p-5 text-center">
              <p>Draw {drawField.name} field on the document</p>
              <div>
                <button class="base-button" onclick={() => clearDrawField()}>Cancel</button>
                {#if !drawOption && !drawField.areas.length && !["stamp", "signature", "initials"].includes(drawField.type)}
                  <a
                    href="#"
                    class="link mt-3 block text-sm"
                    onclick={(e) => {
                      e.preventDefault();
                      drawField = null;
                      drawOption = null;
                    }}
                  >
                    Or add field without drawing
                  </a>
                {/if}
              </div>
            </div>
          </div>
        {/if}

        <Fields
          fields={template.fields}
          submitters={template.submitters}
          {selectedSubmitter}
          {selectedField}
          {defaultSubmitters}
          {defaultFields}
          {onlyDefinedFields}
          {editable}
          onSetDraw={(e) => {
            drawField = e.field;
            drawOption = e.option;
          }}
          onSetDrag={(field) => (dragField = field)}
          onChangeSubmitter={(submitter) => (selectedSubmitter = submitter)}
          onDragEnd={() => (dragField = null)}
          onScrollToArea={scrollToArea}
        />
      </div>
    </div>
  {:else}
    <div class="flex h-full w-full items-center justify-center">
      <div class="w-full max-w-2xl rounded-2xl bg-white p-8">
        <div class="mb-6">
          <div class="text-xl font-semibold text-gray-900">{t("templates.uploadPdfToStartEditing")}</div>
          <div class="mt-1 text-sm text-gray-600">
            {t("templates.uploadPdfToStartEditingHint")}
          </div>
        </div>

        <label
          for="templateEditFileInput"
          class="relative block h-40 w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50 {selectedFile
            ? 'border-blue-400 bg-gray-50'
            : ''} {uploading ? 'cursor-not-allowed opacity-60' : ''}"
          ondragover={(e) => e.preventDefault()}
          ondrop={handleDrop}
        >
          <div class="absolute inset-0 flex items-center justify-center">
            <div class="flex flex-col items-center text-center">
              {#if !selectedFile}
                <span class="flex flex-col items-center">
                  <SvgIcon name="cloud-upload" class="h-10 w-10 text-gray-400" />
                  <div class="mt-2 text-sm font-medium text-gray-700">{t("templates.clickToUpload")}</div>
                  <div class="text-xs text-gray-500">{t("templates.dragAndDrop")}</div>
                </span>
              {:else}
                <span class="flex flex-col items-center">
                  <SvgIcon name="document" class="h-10 w-10 text-blue-500" />
                  <div class="mt-2 text-sm font-medium text-gray-900">{selectedFile.name}</div>
                  {#if uploading}
                    <div class="mt-1 text-xs text-gray-600">{t("templates.uploading")}</div>
                  {:else}
                    <button
                      type="button"
                      class="mt-1 text-xs text-red-600 hover:text-red-800"
                      onclick={(e) => {
                        e.stopPropagation();
                        removeSelectedFile();
                      }}
                    >
                      {t("templates.removeFile")}
                    </button>
                  {/if}
                </span>
              {/if}
            </div>
          </div>

          <input
            id="templateEditFileInput"
            bind:this={templateFileInput}
            type="file"
            accept=".pdf"
            class="hidden"
            disabled={uploading}
            onchange={handleFileSelect}
          />
        </label>

        {#if uploadError}
          <div class="mt-3 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            {uploadError}
          </div>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Add pages modal -->
  <FormModal
    bind:this={addPagesModalRef}
    bind:open={showAddPagesModal}
    title={t("templates.addPagesTitle")}
    submitText={t("templates.addPagesButton")}
    onSubmit={handleAddPagesSubmit}
    onCancel={resetAddPagesSelection}
  >
    <div class="space-y-3">
      <div class="text-sm text-gray-600">{t("templates.addPagesHint")}</div>

      <label
        for="addPagesFileInput"
        class="relative block h-32 w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50 {addPagesSelectedFile
          ? 'border-blue-400 bg-gray-50'
          : ''}"
        ondragover={(e) => e.preventDefault()}
        ondrop={handleAddPagesDrop}
      >
        <div class="absolute inset-0 flex items-center justify-center">
          <div class="flex flex-col items-center text-center">
            {#if !addPagesSelectedFile}
              <span class="flex flex-col items-center">
                <SvgIcon name="cloud-upload" class="h-8 w-8 text-gray-400" />
                <div class="mt-2 text-sm font-medium text-gray-700">{t("templates.clickToUpload")}</div>
                <div class="text-xs text-gray-500">{t("templates.dragAndDrop")}</div>
              </span>
            {:else}
              <span class="flex flex-col items-center">
                <SvgIcon name="document" class="h-8 w-8 text-blue-500" />
                <div class="mt-2 text-sm font-medium text-gray-900">{addPagesSelectedFile.name}</div>
                <button
                  type="button"
                  class="mt-1 text-xs text-red-600 hover:text-red-800"
                  onclick={(e) => {
                    e.stopPropagation();
                    resetAddPagesSelection();
                  }}
                >
                  {t("templates.removeFile")}
                </button>
              </span>
            {/if}
          </div>
        </div>

        <input
          id="addPagesFileInput"
          bind:this={addPagesFileInput}
          type="file"
          accept=".pdf"
          class="hidden"
          disabled={uploading}
          onchange={handleAddPagesFileSelect}
        />
      </label>

      {#if addPagesError}
        <div class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {addPagesError}
        </div>
      {/if}
    </div>
  </FormModal>
</div>
