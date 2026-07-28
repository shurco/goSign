<script lang="ts">
  import { onMount, tick } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import Button from "@/components/ui/Button.svelte";
  import Input from "@/components/ui/Input.svelte";
  import Select from "@/components/ui/Select.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";
  import FormModal from "@/components/common/FormModal.svelte";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import { apiDelete, apiGet, apiPost, apiPut } from "@/services/api";
  import type { Template, TemplateFolder } from "@/models";
  import { fileToBase64Payload } from "@/utils/file";

  // Data
  let templates = $state<Template[]>([]);
  let folders = $state<TemplateFolder[]>([]);
  let loading = $state(false);
  let selectedFolderId = $state<string | null>(null);

  // Modals
  let showCreateFolderModal = $state(false);
  let showCreateTemplateModal = $state(false);
  let showRenameFolderModal = $state(false);
  let showMoveTemplateModal = $state(false);
  let showEditTemplateModal = $state(false);
  let folderToRename = $state<TemplateFolder | null>(null);
  let templateToMove = $state<Template | null>(null);
  let templateToEdit = $state<Template | null>(null);
  let renameFolderModalRef: ReturnType<typeof FormModal> | undefined = $state();
  let editTemplateModalRef: ReturnType<typeof FormModal> | undefined = $state();
  let createTemplateModalRef: ReturnType<typeof FormModal> | undefined = $state();
  let templateFileInput = $state<HTMLInputElement | null>(null);
  let selectedFile = $state<File | null>(null);

  // Filters and search
  let searchQuery = $state("");
  let selectedCategory = $state("");
  let sortBy = $state("name");

  // Type for unified list item
  interface LibraryItem {
    id: string;
    type: "parent" | "folder" | "template";
    folderName?: string;
    data?: TemplateFolder | Template;
  }

  // Table columns
  const columns = $derived([
    { key: "name", label: t("templates.templateName"), sortable: true },
    { key: "category", label: t("templates.category"), sortable: true },
    { key: "signers", label: t("templates.signers"), sortable: true },
    {
      key: "created_at",
      label: t("submissions.created"),
      sortable: true,
      formatter: (value: unknown): string => (value ? formatDate(value as string) : "")
    }
  ]);

  // Mock data removed - always use real API data

  // Unified list of folders and templates
  const libraryItems = $derived.by((): LibraryItem[] => {
    const items: LibraryItem[] = [];

    // Add parent navigation item (..) when inside a folder
    if (selectedFolderId !== null) {
      const currentFolder = folders.find((f) => f.id === selectedFolderId);
      items.push({
        id: "parent-navigation",
        type: "parent",
        folderName: currentFolder?.name
      });
    }

    // Add folders first - show all folders at root level (no parent) when no folder selected
    if (selectedFolderId === null) {
      folders
        .filter((folder) => !folder.parent_id)
        .forEach((folder) => {
          // Search filter for folders
          if (!searchQuery || folder.name.toLowerCase().includes(searchQuery.toLowerCase())) {
            items.push({ id: `folder-${folder.id}`, type: "folder", data: folder });
          }
        });
    }

    // Add templates
    let filtered = Array.isArray(templates) ? [...templates] : [];

    // Folder filter
    if (selectedFolderId !== null) {
      // Show templates in selected folder
      filtered = filtered.filter((template) => {
        const templateFolderId = template.folder_id || null;
        return String(templateFolderId) === String(selectedFolderId);
      });
    } else {
      // Show templates in root (no folder)
      filtered = filtered.filter((template) => {
        const templateFolderId = template.folder_id;
        return !templateFolderId || templateFolderId === null || templateFolderId === "";
      });
    }

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter((template) => template.name.toLowerCase().includes(query));
    }

    // Category filter
    if (selectedCategory) {
      filtered = filtered.filter((template) => template.category === selectedCategory);
    }

    // Add templates to items
    filtered.forEach((template) => {
      items.push({ id: `template-${template.id}`, type: "template", data: template });
    });

    // Sort - parent navigation always first, then folders, then templates
    items.sort((a, b) => {
      // Parent navigation always comes first
      if (a.type === "parent") {
        return -1;
      }
      if (b.type === "parent") {
        return 1;
      }

      // Folders always come before templates
      if (a.type === "folder" && b.type === "template") {
        return -1;
      }
      if (a.type === "template" && b.type === "folder") {
        return 1;
      }

      // Both are folders - sort by name
      if (a.type === "folder" && b.type === "folder") {
        return (a.data as TemplateFolder).name.localeCompare((b.data as TemplateFolder).name);
      }

      // Both are templates - sort by selected criteria
      if (a.type === "template" && b.type === "template") {
        const at = a.data as Template;
        const bt = b.data as Template;
        switch (sortBy) {
          case "name":
            return at.name.localeCompare(bt.name);
          case "created_at":
            return new Date(bt.created_at).getTime() - new Date(at.created_at).getTime();
          case "usage":
            return (
              ((bt as any).submitter_count ?? bt.submitters?.length ?? 0) -
              ((at as any).submitter_count ?? at.submitters?.length ?? 0)
            );
          default:
            return 0;
        }
      }

      return 0;
    });

    return items;
  });

  // Watch for folder selection changes to update route
  $effect(() => {
    const newFolderId = selectedFolderId;
    if (newFolderId && page.url.pathname !== `/templates/${newFolderId}/folder`) {
      goto(`/templates/${newFolderId}/folder`);
    } else if (!newFolderId && page.url.pathname !== "/templates") {
      goto("/templates");
    }
  });

  // Methods
  let loadTemplatesPromise: Promise<void> | null = null;

  const loadFolders = async () => {
    try {
      const response = await apiGet("/api/v1/templates/folders");

      if (response && response.data) {
        // API returns: { success: true, message: "folders", data: [...TemplateFolder] }
        const result = response.data;
        if (Array.isArray(result)) {
          folders = result as TemplateFolder[];
        } else if (result.folders && Array.isArray(result.folders)) {
          folders = result.folders as TemplateFolder[];
        } else {
          folders = [];
        }
      } else {
        console.warn("Failed to load folders from API:", response);
        folders = [];
      }

      if (!Array.isArray(folders)) {
        folders = [];
      }
    } catch (error) {
      console.error("Failed to load folders:", error);
      folders = [];
    }
  };

  const loadTemplates = async () => {
    // Prevent multiple simultaneous loads
    if (loading || loadTemplatesPromise) {
      return loadTemplatesPromise || Promise.resolve();
    }

    loading = true;
    loadTemplatesPromise = (async () => {
      try {
        // Load folders and templates in parallel
        await Promise.all([loadFolders(), loadTemplatesData()]);
      } finally {
        loading = false;
        loadTemplatesPromise = null;
      }
    })();

    return loadTemplatesPromise;
  };

  const loadTemplatesData = async () => {
    try {
      // Build query parameters
      // eslint-disable-next-line svelte/prefer-svelte-reactivity -- transient local value, never stored in reactive state
      const params = new URLSearchParams();
      if (searchQuery) {
        params.append("query", searchQuery);
      }
      if (selectedCategory) {
        params.append("category", selectedCategory);
      }
      params.append("sort_by", sortBy === "name" ? "name" : sortBy === "created_at" ? "created_at" : "updated_at");
      params.append("sort_order", "desc");
      params.append("limit", "100"); // Load all templates (max 100)
      params.append("offset", "0");

      const queryString = params.toString();
      const endpoint = `/api/v1/templates/search${queryString ? `?${queryString}` : ""}`;

      const response = await apiGet(endpoint);

      if (response && response.data) {
        // API returns: { success: true, message: "templates", data: { templates: [...], total: number, ... } }
        const result = response.data;
        if (result.templates && Array.isArray(result.templates)) {
          templates = result.templates as Template[];
        } else if (Array.isArray(result)) {
          // Fallback if API returns array directly
          templates = result as Template[];
        } else {
          templates = [];
        }
      } else {
        // API returned unsuccessful response
        console.warn("Failed to load templates from API:", response);
        templates = [];
      }

      // Ensure templates is always an array
      if (!Array.isArray(templates)) {
        templates = [];
      }
    } catch (error) {
      console.error("Failed to load templates:", error);
      // Show empty list instead of mock data
      templates = [];
    }
  };

  const openTemplateView = (template: Template) => {
    // Open read-only view in a new tab/window
    const href = `/templates/${template.id}/view`;
    window.open(href, "_blank", "noopener,noreferrer");
  };

  const openTemplateEditor = (template: Template) => {
    // Open template editor in the current tab
    goto(`/templates/${template.id}/edit`);
  };

  const editTemplate = (template: Template) => {
    // Open edit modal for name and category
    templateToEdit = template;
    showEditTemplateModal = true;
  };

  const handleEditTemplate = async (formData: Record<string, unknown>) => {
    if (!templateToEdit) {
      return;
    }

    const name = formData.name as string;
    if (!name || name.trim() === "") {
      return;
    }

    try {
      const category = formData.category as string;
      const updateData: any = {
        name: name.trim()
      };

      // Only include category if it's not empty
      if (category && category.trim() !== "") {
        updateData.category = category.trim();
      } else {
        updateData.category = null;
      }

      const response = await apiPut(`/api/v1/templates/${templateToEdit.id}`, updateData);
      if (response && response.data) {
        showEditTemplateModal = false;
        templateToEdit = null;
        await loadTemplates();
      } else {
        alert(t("templates.updateError") || "Failed to update template");
      }
    } catch (error) {
      console.error("Failed to update template:", error);
      alert(t("templates.updateError") || "Failed to update template");
    }
  };

  const handleFileSelect = (event: Event) => {
    const input = event.target as HTMLInputElement;
    if (input && input.files && input.files.length > 0) {
      const file = input.files[0];
      // Validate file type
      if (file.type === "application/pdf" || file.name.endsWith(".pdf")) {
        selectedFile = file;
      } else {
        alert(t("templates.invalidFileType"));
        if (input) {
          input.value = "";
        }
      }
    } else {
      selectedFile = null;
    }
  };

  const handleDrop = (event: DragEvent) => {
    event.preventDefault();
    if (event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files.length > 0) {
      const file = event.dataTransfer.files[0];
      // Validate file type
      if (file.type === "application/pdf" || file.name.endsWith(".pdf")) {
        selectedFile = file;
        // Also update the input element
        if (templateFileInput) {
          const dataTransfer = new DataTransfer();
          dataTransfer.items.add(file);
          templateFileInput.files = dataTransfer.files;
        }
      } else {
        alert(t("templates.invalidFileType"));
      }
    }
  };

  const removeSelectedFile = () => {
    selectedFile = null;
    if (templateFileInput) {
      templateFileInput.value = "";
    }
  };

  const handleCancelCreateTemplate = () => {
    showCreateTemplateModal = false;
    removeSelectedFile();
  };

  const handleCreateTemplate = async (formData: any): Promise<void> => {
    const templateName = (formData.name as string)?.trim() || t("templates.newTemplate");

    try {
      if (selectedFile) {
        // Create template from file
        const file = selectedFile;

        // Convert file to base64 (payload only)
        const base64String = await fileToBase64Payload(file);

        // Determine file type
        let fileType = "pdf";
        if (file.name.endsWith(".pdf")) {
          fileType = "pdf";
        } else if (file.name.endsWith(".html") || file.name.endsWith(".htm")) {
          fileType = "html";
        } else if (file.name.endsWith(".docx")) {
          fileType = "docx";
        }

        const response = await apiPost("/api/v1/templates/from-file", {
          name: templateName,
          type: fileType,
          file_base64: base64String,
          description: ""
        });

        if (response && response.data) {
          let newTemplate = response.data;
          if (newTemplate && typeof newTemplate === "object" && "template" in newTemplate) {
            newTemplate = newTemplate.template;
          }

          const templateId =
            newTemplate?.id ||
            (newTemplate && typeof newTemplate === "object" && "id" in newTemplate ? newTemplate.id : null);

          if (templateId) {
            // Clean up
            removeSelectedFile();

            // Reset modal submitting state before closing
            if (createTemplateModalRef) {
              createTemplateModalRef.resetSubmitting();
            }

            // Close modal
            showCreateTemplateModal = false;

            // Reload templates
            await loadTemplates();

            // Navigate to edit page for the new template
            goto(`/templates/${templateId}/edit`);
          } else {
            console.error("Failed to get template ID from response:", response);
            const errorMsg = t("templates.createTemplateError");
            alert(errorMsg);
            if (createTemplateModalRef) {
              createTemplateModalRef.resetSubmitting();
            }
            throw new Error("Template ID not found in response");
          }
        } else {
          console.error("Failed to create template: unexpected response", response);
          const errorMsg = t("templates.createTemplateError");
          alert(errorMsg);
          if (createTemplateModalRef) {
            createTemplateModalRef.resetSubmitting();
          }
          throw new Error("Unexpected response format");
        }
      } else {
        // Create empty template
        const category = formData.category as string;
        const response = await apiPost("/api/v1/templates/empty", {
          name: templateName,
          category: category && category.trim() !== "" ? category.trim() : null
        });

        if (response && response.data) {
          let newTemplate = response.data;
          if (newTemplate && typeof newTemplate === "object" && "template" in newTemplate) {
            newTemplate = newTemplate.template;
          }

          const templateId =
            newTemplate?.id ||
            (newTemplate && typeof newTemplate === "object" && "id" in newTemplate ? newTemplate.id : null);

          if (templateId) {
            // Reset modal submitting state before closing
            if (createTemplateModalRef) {
              createTemplateModalRef.resetSubmitting();
            }

            // Close modal
            showCreateTemplateModal = false;

            // Reload templates
            await loadTemplates();

            // Navigate to edit page for the new template
            goto(`/templates/${templateId}/edit`);
          } else {
            console.error("Failed to get template ID from response:", response);
            const errorMsg = t("templates.createTemplateError");
            alert(errorMsg);
            if (createTemplateModalRef) {
              createTemplateModalRef.resetSubmitting();
            }
            throw new Error("Template ID not found in response");
          }
        } else {
          console.error("Failed to create template: unexpected response", response);
          const errorMsg = t("templates.createTemplateError");
          alert(errorMsg);
          if (createTemplateModalRef) {
            createTemplateModalRef.resetSubmitting();
          }
          throw new Error("Unexpected response format");
        }
      }
    } catch (error: any) {
      console.error("Failed to create template:", error);
      const errorMessage = error?.message || t("templates.createTemplateError");
      alert(errorMessage);
      // Reset submitting state on error
      if (createTemplateModalRef) {
        createTemplateModalRef.resetSubmitting();
      }
      // Don't re-throw - let user retry
    }
  };

  const viewFolder = (folder: TemplateFolder) => {
    goto(`/templates/${folder.id}/folder`);
  };

  const goBackToParent = () => {
    selectedFolderId = null;
    goto("/templates");
  };

  const toggleFavorite = async (template: Template) => {
    try {
      if (template.is_favorite) {
        // Remove from favorites
        const response = await apiDelete(`/api/v1/templates/favorites/${template.id}`);
        if (response && (response.data || response.message)) {
          template.is_favorite = false;
          // Update in templates array
          const index = templates.findIndex((t) => t.id === template.id);
          if (index !== -1) {
            templates[index].is_favorite = false;
          }
        }
      } else {
        // Add to favorites
        const response = await apiPost("/api/v1/templates/favorites", {
          template_id: template.id
        });
        if (response && (response.data || response.message)) {
          template.is_favorite = true;
          // Update in templates array
          const index = templates.findIndex((t) => t.id === template.id);
          if (index !== -1) {
            templates[index].is_favorite = true;
          }
        }
      }
    } catch (error: any) {
      console.error("Failed to toggle favorite:", error);
      const errorMessage = error?.message || "Failed to update favorite status. Please try again.";
      alert(errorMessage);
    }
  };

  const formatDate = (date: string | Date) => {
    return new Date(date).toLocaleDateString();
  };

  const translateCategory = (category: string): string => {
    if (!category) {
      return "";
    }
    // Map category values to translation keys
    const categoryMap: Record<string, string> = {
      business: "templates.business",
      legal: "templates.legal",
      personal: "templates.personal",
      education: "templates.education"
    };

    const translationKey = categoryMap[category.toLowerCase()];
    if (translationKey) {
      return t(translationKey);
    }
    // Fallback to original category if no translation found
    return category;
  };

  const handleCreateFolder = async (formData: Record<string, unknown>) => {
    const name = (formData.name as string)?.trim();
    if (!name || name === "") {
      alert(t("templates.folderNameRequired"));
      return;
    }

    try {
      const response = await apiPost("/api/v1/templates/folders", { name });
      // Backend returns {success: bool, message: string, data: any}
      // For 201 Created, success might be false (because code != 200), but data will be present
      if (response && (response.data || response.message)) {
        showCreateFolderModal = false;
        await loadFolders();
      } else {
        console.error("Failed to create folder: unexpected response", response);
        alert(t("templates.createFolderUnexpectedError"));
      }
    } catch (error: any) {
      console.error("Failed to create folder:", error);
      const errorMessage = error?.message || t("templates.createFolderError");
      alert(errorMessage);
    }
  };

  const renameFolder = (folder: TemplateFolder) => {
    folderToRename = folder;
    showRenameFolderModal = true;
  };

  const handleRenameFolder = async (formData: Record<string, unknown>) => {
    if (!folderToRename) {
      return;
    }

    const name = formData.name as string;
    if (!name || name.trim() === "") {
      return;
    }

    try {
      const response = await apiPut(`/api/v1/templates/folders/${folderToRename.id}`, {
        name: name.trim()
      });
      if (response && response.data) {
        showRenameFolderModal = false;
        folderToRename = null;
        await loadFolders();
      }
    } catch (error) {
      console.error("Failed to rename folder:", error);
    }
  };

  const deleteFolder = async (folder: TemplateFolder) => {
    if (!confirm(t("templates.deleteFolderConfirm", { name: folder.name }))) {
      return;
    }

    try {
      const response = await apiDelete(`/api/v1/templates/folders/${folder.id}`);
      if (response && response.data) {
        await loadFolders();
        // If deleted folder was selected, reset selection
        if (selectedFolderId === folder.id) {
          selectedFolderId = null;
        }
      }
    } catch (error) {
      console.error("Failed to delete folder:", error);
    }
  };

  const showMoveModal = (template: Template) => {
    templateToMove = template;
    showMoveTemplateModal = true;
  };

  const handleMoveTemplate = async (formData: Record<string, unknown>) => {
    if (!templateToMove) {
      return;
    }

    // Get folder_id - null, empty string, or actual folder ID
    let folderId = formData.folder_id as string | null | undefined;
    // Convert null/undefined/empty string to empty string for backend
    if (!folderId || folderId === "null" || folderId === "") {
      folderId = "";
    }

    try {
      // Send empty string for root (null), or folder ID
      const response = await apiPut(`/api/v1/templates/${templateToMove.id}/move`, {
        folder_id: folderId
      });
      // Check success or data presence
      if (response && (response.data || response.message)) {
        showMoveTemplateModal = false;
        const movedToFolderId = folderId || null;
        templateToMove = null;

        // Reload templates and folders
        await Promise.all([loadTemplates(), loadFolders()]);

        // If moved to a folder, navigate to that folder to show the moved template
        if (movedToFolderId) {
          goto(`/templates/${movedToFolderId}/folder`);
        } else {
          // If moved to root, navigate to root view
          goto("/templates");
        }
      } else {
        console.error("Failed to move template: unexpected response", response);
        alert(t("templates.moveTemplateUnexpectedError"));
      }
    } catch (error: any) {
      console.error("Failed to move template:", error);
      const errorMessage = error?.message || t("templates.moveTemplateError");
      alert(errorMessage);
    }
  };

  // Watch for rename modal opening to initialize form
  $effect(() => {
    const isOpen = showRenameFolderModal;
    const folder = folderToRename;
    const modalRef = renameFolderModalRef;
    if (isOpen && folder && modalRef) {
      void tick().then(() => {
        if (modalRef.setFormData) {
          modalRef.setFormData({ name: folder.name });
        }
      });
    }
  });

  $effect(() => {
    const isOpen = showEditTemplateModal;
    const template = templateToEdit;
    const modalRef = editTemplateModalRef;
    if (isOpen && template && modalRef) {
      void tick().then(() => {
        if (modalRef.setFormData) {
          modalRef.setFormData({
            name: template.name,
            category: template.category || ""
          });
        }
      });
    }
  });

  // Lifecycle
  let hasLoadedOnce = false;

  onMount(() => {
    // Check if we're on a folder route
    if (page.url.pathname.match(/^\/templates\/[^/]+\/folder$/) && page.params.id) {
      selectedFolderId = page.params.id as string;
    }

    loadTemplates().then(() => {
      hasLoadedOnce = true;
    });
  });

  // Watch for route changes to update selectedFolderId
  $effect(() => {
    const folderId = page.params.id;
    const path = page.url.pathname;
    if (path.match(/^\/templates\/[^/]+\/folder$/) && folderId) {
      selectedFolderId = folderId as string;
    } else if (path === "/templates") {
      selectedFolderId = null;
    }
  });

  // Reload templates when component is activated (reused by router)
  // Only reload if we haven't loaded yet or data might be stale
  $effect(() => {
    const path = page.url.pathname;
    if (
      (path === "/templates" || path.match(/^\/templates\/[^/]+\/folder$/)) &&
      (!hasLoadedOnce || templates.length === 0)
    ) {
      void tick().then(() => {
        loadTemplates();
      });
    }
  });

  // Reload templates when navigating to this page via router
  // Only if component is reused and data is empty or stale
  let previousPath = $state<string | null>(null);

  $effect(() => {
    const newPath = page.url.pathname;
    const oldPath = previousPath;
    previousPath = newPath;

    if (oldPath === null) {
      return;
    }

    // Only reload if navigating TO this page (not from it) and we need fresh data
    if ((newPath === "/templates" || newPath.match(/^\/templates\/[^/]+\/folder$/)) && oldPath !== newPath) {
      // Only reload if data is empty or component was reused
      if (templates.length === 0 || !hasLoadedOnce) {
        void tick().then(() => {
          setTimeout(() => {
            if (newPath === page.url.pathname && !loading) {
              loadTemplates().then(() => {
                hasLoadedOnce = true;
              });
            }
          }, 50);
        });
      }
    }
  });
</script>

<div class="template-library">
  <!-- Header -->
  <div class="mb-6 flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold">{t("templates.title")}</h1>
      <p class="mt-1 text-sm text-gray-600">{t("templates.description")}</p>
    </div>
    <div class="flex items-center gap-3">
      <Button variant="ghost" onclick={() => (showCreateFolderModal = true)}>
        <SvgIcon name="folder" class="mr-2 h-5 w-5" />
        {t("templates.newFolder")}
      </Button>
      <div class="relative">
        <Button variant="primary" onclick={() => (showCreateTemplateModal = true)}>
          <SvgIcon name="plus" class="mr-2 h-5 w-5" />
          {t("templates.newTemplate")}
        </Button>
      </div>
    </div>
  </div>

  <!-- Search and Filters -->
  <div class="mb-4 flex flex-col gap-4 lg:flex-row">
    <!-- Search Input -->
    <div class="flex-1">
      <Input bind:value={searchQuery} placeholder={t("templates.searchTemplates")} class="w-full" />
    </div>

    <!-- Filters -->
    <div class="flex gap-3">
      <Select bind:value={selectedCategory} placeholder={t("templates.allCategories")} class="w-40">
        <option value="">{t("templates.allCategories")}</option>
        <option value="business">{t("templates.business")}</option>
        <option value="legal">{t("templates.legal")}</option>
        <option value="personal">{t("templates.personal")}</option>
        <option value="education">{t("templates.education")}</option>
      </Select>

      <Select bind:value={sortBy} class="w-40">
        <option value="name">{t("templates.sortByName")}</option>
        <option value="created_at">{t("templates.sortByDate")}</option>
        <option value="usage">{t("templates.mostUsed")}</option>
      </Select>
    </div>
  </div>

  <!-- Templates and Folders Table -->
  <ResourceTable
    data={libraryItems}
    {columns}
    isLoading={loading}
    searchable={false}
    emptyMessage={t("templates.noItemsFound")}
    showEdit={false}
    showDelete={false}
    idKey="id"
    cellSnippets={{
      name: cellName,
      category: cellCategory,
      signers: cellSigners,
      created_at: cellCreatedAt
    }}
    actions={rowActions}
  />

  <!-- Create Folder Modal -->
  <FormModal
    bind:open={showCreateFolderModal}
    title={t("templates.createFolder")}
    submitText={t("templates.createFolderButton")}
    onSubmitEvent={handleCreateFolder}
  >
    {#snippet children(formData)}
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.folderName")}</label>
          <Input
            value={(formData.name as string) || ""}
            placeholder={t("templates.enterFolderName")}
            class="w-full"
            oninput={(e) => {
              formData.name = String(e.currentTarget.value);
            }}
          />
        </div>
      </div>
    {/snippet}
  </FormModal>

  <!-- Rename Folder Modal -->
  <FormModal
    bind:this={renameFolderModalRef}
    bind:open={showRenameFolderModal}
    title={t("templates.renameFolder")}
    submitText={t("templates.renameFolderButton")}
    onSubmitEvent={handleRenameFolder}
  >
    {#snippet children(formData)}
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.folderName")}</label>
          <Input
            value={(formData.name as string) || ""}
            placeholder={t("templates.enterFolderName")}
            class="w-full"
            oninput={(e) => {
              formData.name = String(e.currentTarget.value);
            }}
          />
        </div>
      </div>
    {/snippet}
  </FormModal>

  <!-- Create Template Modal -->
  <FormModal
    bind:this={createTemplateModalRef}
    bind:open={showCreateTemplateModal}
    title={t("templates.createTemplate")}
    submitText={t("templates.create")}
    onSubmit={handleCreateTemplate}
    onCancel={handleCancelCreateTemplate}
  >
    {#snippet children(formData)}
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.templateName")}</label>
          <Input bind:value={formData.name} placeholder={t("templates.enterTemplateName")} class="w-full" />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.category")}</label>
          <Select bind:value={formData.category} class="w-full">
            <option value="">{t("templates.allCategories")}</option>
            <option value="business">{t("templates.business")}</option>
            <option value="legal">{t("templates.legal")}</option>
            <option value="personal">{t("templates.personal")}</option>
            <option value="education">{t("templates.education")}</option>
          </Select>
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700"
            >{t("templates.uploadFromFile")} ({t("templates.optional")})</label
          >
          <label
            for="templateFileInput"
            class="relative block h-32 w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50"
            class:border-blue-400={!!selectedFile}
            class:bg-gray-50={!!selectedFile}
            ondragover={(e) => e.preventDefault()}
            ondrop={handleDrop}
          >
            <div class="absolute top-0 right-0 bottom-0 left-0 flex items-center justify-center">
              <div class="flex flex-col items-center">
                {#if !selectedFile}
                  <span class="flex flex-col items-center">
                    <SvgIcon name="cloud-upload" class="h-8 w-8 text-gray-400" />
                    <div class="mt-2 text-sm font-medium text-gray-700">{t("templates.clickToUpload")}</div>
                    <div class="text-xs text-gray-500">{t("templates.dragAndDrop")}</div>
                  </span>
                {:else}
                  <span class="flex flex-col items-center">
                    <SvgIcon name="document" class="h-8 w-8 text-blue-500" />
                    <div class="mt-2 text-sm font-medium text-gray-700">{selectedFile.name}</div>
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
                  </span>
                {/if}
              </div>
            </div>
            <input
              id="templateFileInput"
              bind:this={templateFileInput}
              type="file"
              accept=".pdf"
              class="hidden"
              onchange={handleFileSelect}
            />
          </label>
          <p class="mt-1 text-xs text-gray-500">{t("templates.uploadFromFileHint")}</p>
        </div>
      </div>
    {/snippet}
  </FormModal>

  <!-- Edit Template Modal -->
  <FormModal
    bind:this={editTemplateModalRef}
    bind:open={showEditTemplateModal}
    title={t("templates.editTemplate")}
    submitText={t("common.save")}
    onSubmitEvent={handleEditTemplate}
  >
    {#snippet children(formData)}
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.templateName")}</label>
          <Input
            value={(formData.name as string) || ""}
            placeholder={t("templates.enterTemplateName")}
            class="w-full"
            oninput={(e) => {
              formData.name = String(e.currentTarget.value);
            }}
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.category")}</label>
          <Select bind:value={formData.category} class="w-full">
            <option value="">{t("templates.allCategories")}</option>
            <option value="business">{t("templates.business")}</option>
            <option value="legal">{t("templates.legal")}</option>
            <option value="personal">{t("templates.personal")}</option>
            <option value="education">{t("templates.education")}</option>
          </Select>
        </div>
      </div>
    {/snippet}
  </FormModal>

  <!-- Move Template Modal -->
  <FormModal
    bind:open={showMoveTemplateModal}
    title={t("templates.moveTemplate")}
    submitText={t("templates.moveTemplateButton")}
    onSubmitEvent={handleMoveTemplate}
  >
    {#snippet children(formData)}
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("templates.selectFolder")}</label>
          <Select bind:value={formData.folder_id} class="w-full">
            <option value="">{t("templates.rootFolder")}</option>
            {#each folders as folder (folder.id)}
              <option value={folder.id}>{folder.name}</option>
            {/each}
          </Select>
        </div>
      </div>
    {/snippet}
  </FormModal>
</div>

{#snippet cellName(item: unknown, _value: string)}
  {@const libItem = item as LibraryItem}
  <div class="flex items-center gap-2">
    {#if libItem.type === "folder"}
      <SvgIcon name="folder" class="h-4 w-4 flex-shrink-0 text-blue-500" />
    {:else if libItem.type === "template"}
      <SvgIcon name="document" class="h-4 w-4 flex-shrink-0 text-gray-400" />
    {/if}
    {#if libItem.type === "parent"}
      <button class="cursor-pointer text-left font-medium text-gray-700 hover:text-blue-600" onclick={goBackToParent}>
        /..{libItem.folderName ? ` (${libItem.folderName})` : ""}
      </button>
    {:else}
      <button
        class="cursor-pointer text-left font-medium text-gray-900 hover:text-blue-600"
        onclick={() =>
          libItem.type === "folder"
            ? viewFolder(libItem.data as TemplateFolder)
            : openTemplateView(libItem.data as Template)}
      >
        {libItem.type === "folder" ? (libItem.data as TemplateFolder).name : (libItem.data as Template).name}
      </button>
    {/if}
  </div>
{/snippet}

{#snippet cellCategory(item: unknown, _value: string)}
  {@const libItem = item as LibraryItem}
  {#if libItem.type === "template" && (libItem.data as Template).category}
    <span class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800">
      {translateCategory((libItem.data as Template).category!)}
    </span>
  {:else if libItem.type !== "parent"}
    <span class="text-sm text-gray-400">—</span>
  {/if}
{/snippet}

{#snippet cellSigners(item: unknown, _value: string)}
  {@const libItem = item as LibraryItem}
  {#if libItem.type === "template"}
    <span class="text-sm text-gray-600">
      {(libItem.data as Template).submitter_count ?? (libItem.data as Template).submitters?.length ?? 0}
    </span>
  {:else if libItem.type !== "parent"}
    <span class="text-sm text-gray-400">—</span>
  {/if}
{/snippet}

{#snippet cellCreatedAt(item: unknown, _value: string)}
  {@const libItem = item as LibraryItem}
  {#if libItem.type === "template"}
    <span class="text-sm text-gray-500">
      {formatDate((libItem.data as Template).created_at)}
    </span>
  {:else if libItem.type !== "parent"}
    <span class="text-sm text-gray-400">—</span>
  {/if}
{/snippet}

{#snippet rowActions(item: unknown)}
  {@const libItem = item as LibraryItem}
  <div class="flex items-center justify-end gap-2">
    <!-- No actions for parent navigation -->
    {#if libItem.type === "parent"}

    {:else if libItem.type === "folder"}
      <!-- Folder actions -->
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
        title={t("templates.renameFolder")}
        onclick={(e) => {
          e.stopPropagation();
          renameFolder(libItem.data as TemplateFolder);
        }}
      >
        <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
      </button>
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
        title={t("templates.deleteFolder")}
        onclick={(e) => {
          e.stopPropagation();
          deleteFolder(libItem.data as TemplateFolder);
        }}
      >
        <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
      </button>
    {:else}
      <!-- Template actions -->
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-yellow-600"
        title={(libItem.data as Template).is_favorite ? t("templates.removeFavorite") : t("templates.addFavorite")}
        onclick={(e) => {
          e.stopPropagation();
          toggleFavorite(libItem.data as Template);
        }}
      >
        <SvgIcon
          name={(libItem.data as Template).is_favorite ? "star-solid" : "star"}
          class="h-5 w-5 stroke-[2] {(libItem.data as Template).is_favorite ? 'text-yellow-500' : ''}"
        />
      </button>
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
        title={t("templates.moveToFolder")}
        onclick={(e) => {
          e.stopPropagation();
          showMoveModal(libItem.data as Template);
        }}
      >
        <SvgIcon name="folder" class="h-5 w-5 stroke-[2]" />
      </button>
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-blue-600"
        title="Open editor"
        onclick={(e) => {
          e.stopPropagation();
          openTemplateEditor(libItem.data as Template);
        }}
      >
        <SvgIcon name="pencil" class="h-5 w-5 stroke-[2]" />
      </button>
      <button
        class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
        title={t("templates.editTemplate")}
        onclick={(e) => {
          e.stopPropagation();
          editTemplate(libItem.data as Template);
        }}
      >
        <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
      </button>
    {/if}
  </div>
{/snippet}

<style>
  .line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
