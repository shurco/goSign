<template>
  <div class="template-library">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-3xl font-bold">Templates</h1>
      <div class="flex items-center gap-3">
        <Button variant="ghost" size="sm" @click="showCreateFolderModal = true">
          <SvgIcon name="file" class="mr-2 h-4 w-4" />
          New Folder
        </Button>
        <Button variant="primary" size="sm">
          <SvgIcon name="plus" class="mr-2 h-4 w-4" />
          New Template
        </Button>
      </div>
    </div>

    <!-- Breadcrumb / Back Button -->
    <div v-if="selectedFolderId" class="mb-4 flex items-center gap-2">
      <Button variant="ghost" size="sm" @click="router.push('/templates')" class="flex items-center gap-1">
        <SvgIcon name="arrow-left" class="h-4 w-4" />
        Back
      </Button>
      <span class="text-sm text-gray-600">
        {{ folders.find((f) => f.id === selectedFolderId)?.name || "Folder" }}
      </span>
    </div>

    <!-- Search and Filters -->
    <div class="mb-6">
      <div class="flex flex-col gap-4 lg:flex-row">
        <!-- Search Input -->
        <div class="flex-1">
          <Input v-model="searchQuery" placeholder="Search templates..." class="w-full">
            <template #prefix>
              <SvgIcon name="search" class="h-4 w-4 text-gray-400" />
            </template>
          </Input>
        </div>

        <!-- Filters -->
        <div class="flex gap-3">
          <Select v-model="selectedCategory" placeholder="All Categories" class="w-40">
            <option value="">All Categories</option>
            <option value="business">Business</option>
            <option value="legal">Legal</option>
            <option value="personal">Personal</option>
            <option value="education">Education</option>
          </Select>

          <Select v-model="sortBy" class="w-40">
            <option value="name">Sort by Name</option>
            <option value="created_at">Sort by Date</option>
            <option value="usage">Most Used</option>
          </Select>
        </div>
      </div>

      <!-- Active Filters -->
      <div v-if="activeFilters.length > 0" class="mt-4 flex flex-wrap gap-2">
        <span class="text-sm text-gray-600">Filters:</span>
        <Badge v-for="filter in activeFilters" :key="filter.key" variant="ghost" class="flex items-center gap-1">
          {{ filter.label }}
          <button class="ml-1" @click="removeFilter(filter.key)">
            <SvgIcon name="x" class="h-3 w-3" />
          </button>
        </Badge>
        <Button variant="ghost" size="sm" @click="clearFilters"> Clear all </Button>
      </div>
    </div>

    <!-- Bulk Actions Bar -->
    <div
      v-if="selectedTemplates.length > 0"
      class="mb-6 flex items-center justify-between rounded-lg border border-blue-200 bg-blue-50 p-4"
    >
      <div class="flex items-center gap-3">
        <Checkbox
          :checked="selectedTemplates.length === filteredTemplates.length"
          :indeterminate="selectedTemplates.length > 0 && selectedTemplates.length < filteredTemplates.length"
          @change="toggleSelectAll"
        />
        <span class="text-sm font-medium text-blue-900">
          {{ selectedTemplates.length }} template{{ selectedTemplates.length > 1 ? "s" : "" }} selected
        </span>
      </div>

      <div class="flex items-center gap-2">
        <Button variant="ghost" size="sm" class="text-red-600 hover:text-red-700">
          <SvgIcon name="trash-x" class="mr-2 h-4 w-4" />
          Delete Selected
        </Button>
      </div>
    </div>

    <!-- Templates Grid -->
    <div v-if="loading" class="flex justify-center py-12">
      <LoadingSpinner class="h-8 w-8" />
    </div>

    <div v-else-if="libraryItems.length === 0" class="py-12 text-center">
      <SvgIcon name="document" class="mx-auto mb-4 h-16 w-16 text-gray-300" />
      <h3 class="mb-2 text-lg font-medium text-gray-900">No items found</h3>
      <p class="mb-6 text-gray-600">
        {{ searchQuery ? "Try adjusting your search terms" : "Get started by creating your first template" }}
      </p>
      <Button variant="primary">
        <SvgIcon name="plus" class="mr-2 h-4 w-4" />
        Create Template
      </Button>
    </div>

    <div v-else class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
      <div
        v-for="item in paginatedItems"
        :key="`${item.type}-${item.data.id}`"
        class="cursor-pointer rounded-lg border border-gray-200 bg-white transition-colors hover:border-gray-300 hover:shadow-sm"
        :class="
          item.type === 'folder' && selectedFolderId === (item.data as TemplateFolder).id
            ? 'border-blue-300 bg-blue-50'
            : ''
        "
        @click="item.type === 'folder' ? viewFolder(item.data) : viewTemplate(item.data)"
      >
        <!-- Folder Card -->
        <template v-if="item.type === 'folder'">
          <div class="group relative">
            <div
              class="flex aspect-square items-center justify-center rounded-t-lg bg-gradient-to-br from-blue-50 to-indigo-50 p-3"
            >
              <SvgIcon name="file" class="h-7 w-7 text-blue-500" />
            </div>
            <div class="p-2.5">
              <div class="flex items-start justify-between gap-1">
                <h3 class="truncate text-xs font-medium text-gray-900" :title="item.data.name">
                  {{ item.data.name }}
                </h3>
                <div class="folder-menu-container flex-shrink-0 opacity-0 transition-opacity group-hover:opacity-100">
                  <div class="relative">
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-5 w-5 p-0"
                      @click.stop="openFolderMenu(item.data, $event)"
                    >
                      <svg class="h-3.5 w-3.5" fill="currentColor" viewBox="0 0 20 20">
                        <path
                          d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z"
                        />
                      </svg>
                    </Button>
                    <div
                      v-if="folderMenuOpen === (item.data as TemplateFolder).id"
                      class="absolute right-0 z-50 mt-1 w-40 rounded-md border border-gray-200 bg-white shadow-lg"
                      @click.stop
                    >
                      <button
                        class="w-full px-3 py-2 text-left text-xs text-gray-700 hover:bg-gray-50"
                        @click="renameFolder(item.data as TemplateFolder)"
                      >
                        Rename
                      </button>
                      <button
                        class="w-full px-3 py-2 text-left text-xs text-red-600 hover:bg-red-50"
                        @click="deleteFolder(item.data as TemplateFolder)"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                </div>
              </div>
              <p class="mt-0.5 text-xs text-gray-500">Folder</p>
            </div>
          </div>
        </template>

        <!-- Template Card -->
        <template v-else>
          <div class="flex aspect-square items-center justify-center rounded-t-lg bg-gray-50 p-3">
            <SvgIcon name="document" class="h-7 w-7 text-gray-400" />
          </div>
          <div class="p-3">
            <div class="mb-1 flex items-start justify-between gap-1">
              <h3 class="truncate text-sm font-medium text-gray-900" :title="item.data.name">
                {{ item.data.name }}
              </h3>
              <Checkbox
                :checked="isSelected(item.data.id)"
                class="ml-1 flex-shrink-0"
                @change="toggleTemplateSelection(item.data.id)"
                @click.stop
              />
            </div>

            <p v-if="item.data.description" class="mb-2 line-clamp-1 text-xs text-gray-600">
              {{ item.data.description }}
            </p>

            <div class="mb-2 flex items-center justify-between text-xs text-gray-500">
              <span>{{ item.data.submitters?.length || 0 }} signers</span>
            </div>

            <div class="flex items-center gap-1">
              <Button variant="ghost" size="sm" class="h-6 flex-1 p-0 text-xs" @click.stop="toggleFavorite(item.data)">
                <SvgIcon
                  :name="item.data.is_favorite ? 'star-solid' : 'star'"
                  class="h-3.5 w-3.5"
                  :class="item.data.is_favorite ? 'text-yellow-500' : 'text-gray-400'"
                />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                class="h-6 p-0 text-xs"
                @click.stop="showMoveModal(item.data)"
                title="Move to folder"
              >
                <SvgIcon name="folder" class="h-3.5 w-3.5 text-gray-400" />
              </Button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="mt-6 flex justify-center">
      <Pagination
        :current-page="currentPage"
        :page-size="pageSize"
        :total="libraryItems.length"
        @update:current-page="currentPage = $event"
      />
    </div>

    <!-- Create Folder Modal -->
    <FormModal
      v-model="showCreateFolderModal"
      title="Create New Folder"
      submit-text="Create"
      @submit="handleCreateFolder"
    >
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">Folder Name</label>
            <Input
              :model-value="(formData.name as string) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.name = String(val);
                }
              "
              placeholder="Enter folder name"
              class="w-full"
            />
          </div>
        </div>
      </template>
    </FormModal>

    <!-- Rename Folder Modal -->
    <FormModal
      ref="renameFolderModalRef"
      v-model="showRenameFolderModal"
      title="Rename Folder"
      submit-text="Save"
      @submit="handleRenameFolder"
    >
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">Folder Name</label>
            <Input
              :model-value="(formData.name as string) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.name = String(val);
                }
              "
              placeholder="Enter folder name"
              class="w-full"
            />
          </div>
        </div>
      </template>
    </FormModal>

    <!-- Move Template Modal -->
    <FormModal v-model="showMoveTemplateModal" title="Move Template" submit-text="Move" @submit="handleMoveTemplate">
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">Select Folder</label>
            <Select
              :model-value="(formData.folder_id as string | number) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.folder_id = String(val);
                }
              "
              class="w-full"
            >
              <option value="">Root (No folder)</option>
              <option v-for="folder in folders" :key="folder.id" :value="folder.id">
                {{ folder.name }}
              </option>
            </Select>
          </div>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import Checkbox from "@/components/ui/Checkbox.vue";
import Badge from "@/components/ui/Badge.vue";
import Pagination from "@/components/ui/Pagination.vue";
import LoadingSpinner from "@/components/ui/LoadingSpinner.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import FormModal from "@/components/common/FormModal.vue";
import { apiGet, apiPost, apiPut, apiDelete } from "@/services/api";
import type { Template, TemplateFolder } from "@/models";

const router = useRouter();
const route = useRoute();

// Data
const templates = ref<Template[]>([]);
const folders = ref<TemplateFolder[]>([]);
const loading = ref(false);
const selectedTemplates = ref<string[]>([]);
const selectedFolderId = ref<string | null>(null);

// Modals
const showCreateFolderModal = ref(false);
const showRenameFolderModal = ref(false);
const showMoveTemplateModal = ref(false);
const folderMenuOpen = ref<string | null>(null);
const folderToRename = ref<TemplateFolder | null>(null);
const templateToMove = ref<Template | null>(null);
const renameFolderModalRef = ref<any>(null);

// Type for unified list item
type LibraryItem = { type: "folder"; data: TemplateFolder } | { type: "template"; data: Template };

// Filters and search
const searchQuery = ref("");
const selectedCategory = ref("");
const sortBy = ref("name");
const currentPage = ref(1);
const pageSize = ref(12);

// Mock data removed - always use real API data

// Computed - unified list of folders and templates
const libraryItems = computed((): LibraryItem[] => {
  const items: LibraryItem[] = [];

  // Add folders first - show all folders at root level (no parent)
  if (selectedFolderId.value === null) {
    folders.value
      .filter((folder) => !folder.parent_id)
      .forEach((folder) => {
        // Search filter for folders
        if (!searchQuery.value || folder.name.toLowerCase().includes(searchQuery.value.toLowerCase())) {
          items.push({ type: "folder", data: folder });
        }
      });
  }

  // Add templates
  let filtered = Array.isArray(templates.value) ? [...templates.value] : [];

  // Folder filter
  if (selectedFolderId.value !== null) {
    // Show templates in selected folder - compare as strings
    filtered = filtered.filter((template) => {
      const templateFolderId = template.folder_id || null;
      return String(templateFolderId) === String(selectedFolderId.value);
    });
  } else {
    // Show templates in root (no folder) - templates with null or empty folder_id
    filtered = filtered.filter((template) => {
      const templateFolderId = template.folder_id;
      return !templateFolderId || templateFolderId === null || templateFolderId === "";
    });
  }

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (template) => template.name.toLowerCase().includes(query) || template.description?.toLowerCase().includes(query)
    );
  }

  // Category filter
  if (selectedCategory.value) {
    filtered = filtered.filter((template) => template.category === selectedCategory.value);
  }

  // Add templates to items
  filtered.forEach((template) => {
    items.push({ type: "template", data: template });
  });

  // Sort - folders first, then by sort criteria
  items.sort((a, b) => {
    // Folders always come first
    if (a.type === "folder" && b.type === "template") return -1;
    if (a.type === "template" && b.type === "folder") return 1;

    // Both are folders - sort by name
    if (a.type === "folder" && b.type === "folder") {
      return a.data.name.localeCompare(b.data.name);
    }

    // Both are templates - sort by selected criteria
    if (a.type === "template" && b.type === "template") {
      switch (sortBy.value) {
        case "name":
          return a.data.name.localeCompare(b.data.name);
        case "created_at":
          return new Date(b.data.created_at).getTime() - new Date(a.data.created_at).getTime();
        case "usage":
          return (b.data.submitters?.length || 0) - (a.data.submitters?.length || 0);
        default:
          return 0;
      }
    }

    return 0;
  });

  return items;
});

const filteredTemplates = computed(() => {
  // For backward compatibility, extract only templates from libraryItems
  return libraryItems.value.filter((item) => item.type === "template").map((item) => item.data as Template);
});

const paginatedItems = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value;
  const end = start + pageSize.value;
  return libraryItems.value.slice(start, end);
});

const totalPages = computed(() => {
  return Math.ceil(libraryItems.value.length / pageSize.value);
});

const activeFilters = computed(() => {
  const filters = [];
  if (searchQuery.value) {
    filters.push({ key: "search", label: `Search: "${searchQuery.value}"` });
  }
  if (selectedCategory.value) {
    filters.push({ key: "category", label: `Category: ${selectedCategory.value}` });
  }
  return filters;
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
        folders.value = result as TemplateFolder[];
      } else if (result.folders && Array.isArray(result.folders)) {
        folders.value = result.folders as TemplateFolder[];
      } else {
        folders.value = [];
      }
    } else {
      console.warn("Failed to load folders from API:", response);
      folders.value = [];
    }

    if (!Array.isArray(folders.value)) {
      folders.value = [];
    }
  } catch (error) {
    console.error("Failed to load folders:", error);
    folders.value = [];
  }
};

const loadTemplates = async () => {
  // Prevent multiple simultaneous loads
  if (loading.value || loadTemplatesPromise) {
    return loadTemplatesPromise || Promise.resolve();
  }

  loading.value = true;
  loadTemplatesPromise = (async () => {
    try {
      // Load folders and templates in parallel
      await Promise.all([loadFolders(), loadTemplatesData()]);
    } finally {
      loading.value = false;
      loadTemplatesPromise = null;
    }
  })();

  return loadTemplatesPromise;
};

const loadTemplatesData = async () => {
  try {
    // Build query parameters
    const params = new URLSearchParams();
    if (searchQuery.value) {
      params.append("query", searchQuery.value);
    }
    if (selectedCategory.value) {
      params.append("category", selectedCategory.value);
    }
    params.append(
      "sort_by",
      sortBy.value === "name" ? "name" : sortBy.value === "created_at" ? "created_at" : "updated_at"
    );
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
        templates.value = result.templates as Template[];
      } else if (Array.isArray(result)) {
        // Fallback if API returns array directly
        templates.value = result as Template[];
      } else {
        templates.value = [];
      }
    } else {
      // API returned unsuccessful response
      console.warn("Failed to load templates from API:", response);
      templates.value = [];
    }

    // Ensure templates is always an array
    if (!Array.isArray(templates.value)) {
      templates.value = [];
    }
  } catch (error) {
    console.error("Failed to load templates:", error);
    // Show empty list instead of mock data
    templates.value = [];
  }
};

const viewTemplate = (template: Template) => {
  router.push(`/templates/${template.id}/edit`);
};

const viewFolder = (folder: TemplateFolder) => {
  router.push(`/templates/${folder.id}/folder`);
};

const toggleTemplateSelection = (templateId: string) => {
  const index = selectedTemplates.value.indexOf(templateId);
  if (index > -1) {
    selectedTemplates.value.splice(index, 1);
  } else {
    selectedTemplates.value.push(templateId);
  }
};

const toggleSelectAll = () => {
  if (selectedTemplates.value.length === filteredTemplates.value.length) {
    selectedTemplates.value = [];
  } else {
    selectedTemplates.value = filteredTemplates.value.map((t) => t.id);
  }
};

const isSelected = (templateId: string) => {
  return selectedTemplates.value.includes(templateId);
};

const toggleFavorite = async (template: Template) => {
  try {
    if (template.is_favorite) {
      // Remove from favorites
      const response = await apiDelete(`/api/v1/templates/favorites/${template.id}`);
      if (response && (response.data || response.message)) {
        template.is_favorite = false;
        // Update in templates array
        const index = templates.value.findIndex((t) => t.id === template.id);
        if (index !== -1) {
          templates.value[index].is_favorite = false;
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
        const index = templates.value.findIndex((t) => t.id === template.id);
        if (index !== -1) {
          templates.value[index].is_favorite = true;
        }
      }
    }
  } catch (error: any) {
    console.error("Failed to toggle favorite:", error);
    const errorMessage = error?.message || "Failed to update favorite status. Please try again.";
    alert(errorMessage);
  }
};

const removeFilter = (filterKey: string) => {
  if (filterKey === "search") {
    searchQuery.value = "";
  } else if (filterKey === "category") {
    selectedCategory.value = "";
  }
};

const clearFilters = () => {
  searchQuery.value = "";
  selectedCategory.value = "";
  selectedFolderId.value = null;
  currentPage.value = 1;
};

const formatDate = (date: string | Date) => {
  return new Date(date).toLocaleDateString();
};

// Folder operations
const openFolderMenu = (folder: TemplateFolder, event: MouseEvent) => {
  event.stopPropagation();
  folderMenuOpen.value = folderMenuOpen.value === folder.id ? null : folder.id;
};

const handleCreateFolder = async (formData: Record<string, unknown>) => {
  const name = (formData.name as string)?.trim();
  if (!name || name === "") {
    alert("Please enter a folder name");
    return;
  }

  try {
    const response = await apiPost("/api/v1/templates/folders", { name });
    console.log("Create folder response:", response);
    // Backend returns {success: bool, message: string, data: any}
    // For 201 Created, success might be false (because code != 200), but data will be present
    if (response && (response.data || response.message)) {
      showCreateFolderModal.value = false;
      await loadFolders();
    } else {
      console.error("Failed to create folder: unexpected response", response);
      alert("Failed to create folder: unexpected response");
    }
  } catch (error: any) {
    console.error("Failed to create folder:", error);
    const errorMessage = error?.message || "Failed to create folder. Please try again.";
    alert(errorMessage);
  }
};

const renameFolder = (folder: TemplateFolder) => {
  folderMenuOpen.value = null;
  folderToRename.value = folder;
  showRenameFolderModal.value = true;
};

const handleRenameFolder = async (formData: Record<string, unknown>) => {
  if (!folderToRename.value) return;

  const name = formData.name as string;
  if (!name || name.trim() === "") {
    return;
  }

  try {
    const response = await apiPut(`/api/v1/templates/folders/${folderToRename.value.id}`, {
      name: name.trim()
    });
    if (response && response.data) {
      showRenameFolderModal.value = false;
      folderToRename.value = null;
      await loadFolders();
    }
  } catch (error) {
    console.error("Failed to rename folder:", error);
  }
};

const deleteFolder = async (folder: TemplateFolder) => {
  folderMenuOpen.value = null;

  if (!confirm(`Are you sure you want to delete folder "${folder.name}"?`)) {
    return;
  }

  try {
    const response = await apiDelete(`/api/v1/templates/folders/${folder.id}`);
    if (response && response.data) {
      await loadFolders();
      // If deleted folder was selected, reset selection
      if (selectedFolderId.value === folder.id) {
        selectedFolderId.value = null;
      }
    }
  } catch (error) {
    console.error("Failed to delete folder:", error);
  }
};

const showMoveModal = (template: Template) => {
  templateToMove.value = template;
  showMoveTemplateModal.value = true;
};

const handleMoveTemplate = async (formData: Record<string, unknown>) => {
  if (!templateToMove.value) return;

  // Get folder_id - null, empty string, or actual folder ID
  let folderId = formData.folder_id as string | null | undefined;
  // Convert null/undefined/empty string to empty string for backend
  if (!folderId || folderId === "null" || folderId === "") {
    folderId = "";
  }

  try {
    // Send empty string for root (null), or folder ID
    const response = await apiPut(`/api/v1/templates/${templateToMove.value.id}/move`, {
      folder_id: folderId
    });
    console.log("Move template response:", response);
    // Check success or data presence
    if (response && (response.data || response.message)) {
      showMoveTemplateModal.value = false;
      const movedToFolderId = folderId || null;
      templateToMove.value = null;

      // Reload templates and folders
      await Promise.all([loadTemplates(), loadFolders()]);

      // If moved to a folder, navigate to that folder to show the moved template
      if (movedToFolderId) {
        router.push(`/templates/${movedToFolderId}/folder`);
      } else {
        // If moved to root, navigate to root view
        router.push("/templates");
      }

      // Reset to first page
      currentPage.value = 1;
    } else {
      console.error("Failed to move template: unexpected response", response);
      alert("Failed to move template: unexpected response");
    }
  } catch (error: any) {
    console.error("Failed to move template:", error);
    const errorMessage = error?.message || "Failed to move template. Please try again.";
    alert(errorMessage);
  }
};

// Watch for rename modal opening to initialize form
watch(showRenameFolderModal, async (isOpen) => {
  if (isOpen && folderToRename.value && renameFolderModalRef.value) {
    await nextTick();
    if (renameFolderModalRef.value.setFormData) {
      renameFolderModalRef.value.setFormData({ name: folderToRename.value.name });
    }
  }
});

// Close folder menu on outside click
const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement;
  if (!target.closest(".folder-menu-container")) {
    folderMenuOpen.value = null;
  }
};

// Lifecycle
let hasLoadedOnce = false;

onMounted(() => {
  // Check if we're on a folder route
  if (route.name === "template-folder" && route.params.id) {
    selectedFolderId.value = route.params.id as string;
  }

  loadTemplates().then(() => {
    hasLoadedOnce = true;
  });
  document.addEventListener("click", handleClickOutside);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleClickOutside);
});

// Watch for route changes to update selectedFolderId
watch(
  () => route.params.id,
  (folderId) => {
    if (route.name === "template-folder" && folderId) {
      selectedFolderId.value = folderId as string;
    } else if (route.path === "/templates") {
      selectedFolderId.value = null;
    }
  },
  { immediate: true }
);

// Reload templates when component is activated (reused by router)
// Only reload if we haven't loaded yet or data might be stale
onActivated(() => {
  if (
    (route.path === "/templates" || route.name === "template-folder") &&
    (!hasLoadedOnce || templates.value.length === 0)
  ) {
    nextTick(() => {
      loadTemplates();
    });
  }
});

// Reload templates when navigating to this page via router
// Only if component is reused and data is empty or stale
watch(
  () => route.path,
  async (newPath, oldPath) => {
    // Only reload if navigating TO this page (not from it) and we need fresh data
    if ((newPath === "/templates" || newPath.match(/^\/templates\/[^\/]+\/folder$/)) && oldPath !== newPath) {
      // Only reload if data is empty or component was reused
      if (templates.value.length === 0 || !hasLoadedOnce) {
        await nextTick();
        setTimeout(() => {
          if (newPath === route.path && !loading.value) {
            loadTemplates().then(() => {
              hasLoadedOnce = true;
            });
          }
        }, 50);
      }
    }
  },
  { immediate: false }
);
</script>

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
