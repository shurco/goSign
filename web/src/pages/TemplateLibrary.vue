<template>
  <div class="template-library">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ $t('templates.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ $t('templates.description') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="ghost" @click="showCreateFolderModal = true">
          <SvgIcon name="folder" class="mr-2 h-5 w-5" />
          {{ $t('templates.newFolder') }}
        </Button>
        <div class="relative">
          <Button variant="primary" @click="showCreateTemplateModal = true">
            <SvgIcon name="plus" class="mr-2 h-5 w-5" />
            {{ $t('templates.newTemplate') }}
        </Button>
      </div>
    </div>
    </div>

    <!-- Search and Filters -->
    <div class="mb-4 flex flex-col gap-4 lg:flex-row">
        <!-- Search Input -->
        <div class="flex-1">
        <Input v-model="searchQuery" :placeholder="$t('templates.searchTemplates')" class="w-full">
            <template #prefix>
              <SvgIcon name="search" class="h-4 w-4 text-gray-400" />
            </template>
          </Input>
        </div>

        <!-- Filters -->
        <div class="flex gap-3">
        <Select v-model="selectedCategory" :placeholder="$t('templates.allCategories')" class="w-40">
          <option value="">{{ $t('templates.allCategories') }}</option>
          <option value="business">{{ $t('templates.business') }}</option>
          <option value="legal">{{ $t('templates.legal') }}</option>
          <option value="personal">{{ $t('templates.personal') }}</option>
          <option value="education">{{ $t('templates.education') }}</option>
          </Select>

          <Select v-model="sortBy" class="w-40">
          <option value="name">{{ $t('templates.sortByName') }}</option>
          <option value="created_at">{{ $t('templates.sortByDate') }}</option>
          <option value="usage">{{ $t('templates.mostUsed') }}</option>
          </Select>
        </div>
      </div>

    <!-- Templates and Folders Table -->
    <ResourceTable
      :data="libraryItems"
      :columns="columns"
      :is-loading="loading"
      :searchable="false"
      :empty-message="$t('templates.noItemsFound')"
      :show-edit="false"
      :show-delete="false"
      id-key="id"
    >
      <template #cell-name="{ item }">
        <div class="flex items-center gap-2">
          <SvgIcon
            v-if="(item as LibraryItem).type === 'folder'"
            name="folder"
            class="h-4 w-4 text-blue-500 flex-shrink-0"
          />
          <SvgIcon
            v-else-if="(item as LibraryItem).type === 'template'"
            name="document"
            class="h-4 w-4 text-gray-400 flex-shrink-0"
          />
          <button
            v-if="(item as LibraryItem).type === 'parent'"
            class="cursor-pointer text-left font-medium text-gray-700 hover:text-blue-600"
            @click="goBackToParent"
          >
            /..{{ (item as LibraryItem).folderName ? ` (${(item as LibraryItem).folderName})` : '' }}
          </button>
          <button
            v-else
            class="cursor-pointer text-left font-medium text-gray-900 hover:text-blue-600"
            @click="(item as LibraryItem).type === 'folder' ? viewFolder((item as LibraryItem).data as TemplateFolder) : openTemplateView((item as LibraryItem).data as Template)"
          >
            {{ (item as LibraryItem).type === 'folder' ? ((item as LibraryItem).data as TemplateFolder).name : ((item as LibraryItem).data as Template).name }}
          </button>
      </div>
      </template>

      <template #cell-category="{ item }">
        <span
          v-if="(item as LibraryItem).type === 'template' && ((item as LibraryItem).data as Template).category"
          class="inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800"
        >
          {{ translateCategory(((item as LibraryItem).data as Template).category) }}
        </span>
        <span v-else-if="(item as LibraryItem).type !== 'parent'" class="text-sm text-gray-400">—</span>
      </template>

      <template #cell-signers="{ item }">
        <span v-if="(item as LibraryItem).type === 'template'" class="text-sm text-gray-600">
          {{
            ((item as LibraryItem).data as Template).submitter_count ??
            ((item as LibraryItem).data as Template).submitters?.length ??
            0
          }}
        </span>
        <span v-else-if="(item as LibraryItem).type !== 'parent'" class="text-sm text-gray-400">—</span>
      </template>

      <template #cell-created_at="{ item }">
        <span v-if="(item as LibraryItem).type === 'template'" class="text-sm text-gray-500">
          {{ formatDate(((item as LibraryItem).data as Template).created_at) }}
        </span>
        <span v-else-if="(item as LibraryItem).type !== 'parent'" class="text-sm text-gray-400">—</span>
      </template>

      <template #actions="{ item }">
        <div class="flex items-center justify-end gap-2">
          <!-- No actions for parent navigation -->
          <template v-if="(item as LibraryItem).type === 'parent'">
          </template>
          <!-- Folder actions -->
          <template v-else-if="(item as LibraryItem).type === 'folder'">
                      <button
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
              @click.stop="renameFolder((item as LibraryItem).data as TemplateFolder)"
              :title="$t('templates.renameFolder')"
                      >
              <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
                      </button>
                      <button
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
              @click.stop="deleteFolder((item as LibraryItem).data as TemplateFolder)"
              :title="$t('templates.deleteFolder')"
                      >
              <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
                      </button>
        </template>
          <!-- Template actions -->
        <template v-else>
            <button
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-yellow-600"
              @click.stop="toggleFavorite((item as LibraryItem).data as Template)"
              :title="((item as LibraryItem).data as Template).is_favorite ? $t('templates.removeFavorite') : $t('templates.addFavorite')"
            >
                <SvgIcon
                :name="((item as LibraryItem).data as Template).is_favorite ? 'star-solid' : 'star'"
                class="h-5 w-5 stroke-[2]"
                :class="((item as LibraryItem).data as Template).is_favorite ? 'text-yellow-500' : ''"
              />
            </button>
            <button
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
              @click.stop="showMoveModal((item as LibraryItem).data as Template)"
              :title="$t('templates.moveToFolder')"
            >
              <SvgIcon name="folder" class="h-5 w-5 stroke-[2]" />
            </button>
            <button
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-blue-600"
              @click.stop="openTemplateEditor((item as LibraryItem).data as Template)"
              title="Open editor"
            >
              <SvgIcon name="pencil" class="h-5 w-5 stroke-[2]" />
            </button>
            <button
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-gray-600"
              @click.stop="editTemplate((item as LibraryItem).data as Template)"
              :title="$t('templates.editTemplate')"
            >
              <SvgIcon name="settings" class="h-5 w-5 stroke-[2]" />
            </button>
        </template>
      </div>
      </template>
    </ResourceTable>

    <!-- Create Folder Modal -->
    <FormModal
      v-model="showCreateFolderModal"
      :title="$t('templates.createFolder')"
      :submit-text="$t('templates.createFolderButton')"
      @submit="handleCreateFolder"
    >
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.folderName') }}</label>
            <Input
              :model-value="(formData.name as string) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.name = String(val);
                }
              "
              :placeholder="$t('templates.enterFolderName')"
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
      :title="$t('templates.renameFolder')"
      :submit-text="$t('templates.renameFolderButton')"
      @submit="handleRenameFolder"
    >
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.folderName') }}</label>
            <Input
              :model-value="(formData.name as string) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.name = String(val);
                }
              "
              :placeholder="$t('templates.enterFolderName')"
              class="w-full"
            />
          </div>
        </div>
      </template>
    </FormModal>

    <!-- Create Template Modal -->
    <FormModal
      ref="createTemplateModalRef"
      v-model="showCreateTemplateModal"
      :title="$t('templates.createTemplate')"
      :submit-text="$t('templates.create')"
      :on-submit="handleCreateTemplate"
      @cancel="handleCancelCreateTemplate"
    >
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.templateName') }}</label>
            <Input
              v-model="formData.name"
              :placeholder="$t('templates.enterTemplateName')"
              class="w-full"
            />
          </div>
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.category') }}</label>
            <Select
              v-model="formData.category"
              class="w-full"
            >
              <option value="">{{ $t('templates.allCategories') }}</option>
              <option value="business">{{ $t('templates.business') }}</option>
              <option value="legal">{{ $t('templates.legal') }}</option>
              <option value="personal">{{ $t('templates.personal') }}</option>
              <option value="education">{{ $t('templates.education') }}</option>
            </Select>
          </div>
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.uploadFromFile') }} ({{ $t('templates.optional') }})</label>
            <label
              for="templateFileInput"
              class="relative block h-32 w-full cursor-pointer rounded-xl border-2 border-dashed border-gray-300 hover:bg-gray-50"
              :class="{ 'bg-gray-50 border-blue-400': selectedFile }"
              @dragover.prevent
              @drop.prevent="handleDrop"
            >
              <div class="absolute top-0 right-0 bottom-0 left-0 flex items-center justify-center">
                <div class="flex flex-col items-center">
                  <span v-if="!selectedFile" class="flex flex-col items-center">
                    <SvgIcon name="cloud-upload" class="h-8 w-8 text-gray-400" />
                    <div class="mt-2 text-sm font-medium text-gray-700">{{ $t('templates.clickToUpload') }}</div>
                    <div class="text-xs text-gray-500">{{ $t('templates.dragAndDrop') }}</div>
                  </span>
                  <span v-else class="flex flex-col items-center">
                    <SvgIcon name="document" class="h-8 w-8 text-blue-500" />
                    <div class="mt-2 text-sm font-medium text-gray-700">{{ selectedFile.name }}</div>
                    <button
                      type="button"
                      class="mt-1 text-xs text-red-600 hover:text-red-800"
                      @click.stop="removeSelectedFile"
                    >
                      {{ $t('templates.removeFile') }}
                    </button>
                  </span>
                </div>
              </div>
              <input
                id="templateFileInput"
                ref="templateFileInput"
                type="file"
                accept=".pdf"
                class="hidden"
                @change="handleFileSelect"
              />
            </label>
            <p class="mt-1 text-xs text-gray-500">{{ $t('templates.uploadFromFileHint') }}</p>
          </div>
        </div>
      </template>
    </FormModal>

    <!-- Edit Template Modal -->
    <FormModal
      ref="editTemplateModalRef"
      v-model="showEditTemplateModal"
      :title="$t('templates.editTemplate')"
      :submit-text="$t('common.save')"
      @submit="handleEditTemplate"
    >
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.templateName') }}</label>
            <Input
              :model-value="(formData.name as string) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.name = String(val);
                }
              "
              :placeholder="$t('templates.enterTemplateName')"
              class="w-full"
            />
          </div>
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.category') }}</label>
            <Select
              :model-value="(formData.category as string) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.category = String(val);
                }
              "
              class="w-full"
            >
              <option value="">{{ $t('templates.allCategories') }}</option>
              <option value="business">{{ $t('templates.business') }}</option>
              <option value="legal">{{ $t('templates.legal') }}</option>
              <option value="personal">{{ $t('templates.personal') }}</option>
              <option value="education">{{ $t('templates.education') }}</option>
            </Select>
          </div>
        </div>
      </template>
    </FormModal>

    <!-- Move Template Modal -->
    <FormModal v-model="showMoveTemplateModal" :title="$t('templates.moveTemplate')" :submit-text="$t('templates.moveTemplateButton')" @submit="handleMoveTemplate">
      <template #default="{ formData }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('templates.selectFolder') }}</label>
            <Select
              :model-value="(formData.folder_id as string | number) || ''"
              @update:model-value="
                (val: string | number) => {
                  formData.folder_id = String(val);
                }
              "
              class="w-full"
            >
              <option value="">{{ $t('templates.rootFolder') }}</option>
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
import { useI18n } from "vue-i18n";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import Checkbox from "@/components/ui/Checkbox.vue";
import Badge from "@/components/ui/Badge.vue";
import Pagination from "@/components/ui/Pagination.vue";
import LoadingSpinner from "@/components/ui/LoadingSpinner.vue";
import FileInput from "@/components/ui/FileInput.vue";
import SvgIcon from "@/components/SvgIcon.vue";
import FormModal from "@/components/common/FormModal.vue";
import ResourceTable from "@/components/common/ResourceTable.vue";
import { apiGet, apiPost, apiPut, apiDelete } from "@/services/api";
import type { Template, TemplateFolder } from "@/models";
import { fileToBase64Payload } from "@/utils/file";

const router = useRouter();
const route = useRoute();
const { t } = useI18n();

// Data
const templates = ref<Template[]>([]);
const folders = ref<TemplateFolder[]>([]);
const loading = ref(false);
const selectedTemplates = ref<string[]>([]);
const selectedFolderId = ref<string | null>(null);

// Modals
const showCreateFolderModal = ref(false);
const showCreateTemplateModal = ref(false);
const showRenameFolderModal = ref(false);
const showMoveTemplateModal = ref(false);
const showEditTemplateModal = ref(false);
const folderMenuOpen = ref<string | null>(null);
const folderToRename = ref<TemplateFolder | null>(null);
const templateToMove = ref<Template | null>(null);
const templateToEdit = ref<Template | null>(null);
const renameFolderModalRef = ref<any>(null);
const editTemplateModalRef = ref<any>(null);
const createTemplateModalRef = ref<any>(null);
const templateFileInput = ref<HTMLInputElement | null>(null);
const selectedFile = ref<File | null>(null);


// Filters and search
const searchQuery = ref("");
const selectedCategory = ref("");
const sortBy = ref("name");
const currentPage = ref(1);
const pageSize = ref(12);

// Type for unified list item
type LibraryItem = { 
  id: string;
  type: "parent";
  folderName?: string;
} | { 
  id: string;
  type: "folder"; 
  data: TemplateFolder 
} | { 
  id: string;
  type: "template"; 
  data: Template 
};

// Table columns
const columns = computed(() => [
  { key: "name", label: t('templates.templateName'), sortable: true },
  { key: "category", label: t('templates.category'), sortable: true },
  { key: "signers", label: t('templates.signers'), sortable: true },
  {
    key: "created_at",
    label: t('submissions.created'),
    sortable: true,
    formatter: (value: unknown): string => value ? formatDate(value as string) : ""
  }
]);

// Mock data removed - always use real API data


// Unified list of folders and templates
const libraryItems = computed((): LibraryItem[] => {
  const items: LibraryItem[] = [];

  // Add parent navigation item (..) when inside a folder
  if (selectedFolderId.value !== null) {
    const currentFolder = folders.value.find((f) => f.id === selectedFolderId.value);
    items.push({
      id: "parent-navigation",
      type: "parent",
      folderName: currentFolder?.name
    });
  }

  // Add folders first - show all folders at root level (no parent) when no folder selected
  if (selectedFolderId.value === null) {
    folders.value
      .filter((folder) => !folder.parent_id)
      .forEach((folder) => {
        // Search filter for folders
        if (!searchQuery.value || folder.name.toLowerCase().includes(searchQuery.value.toLowerCase())) {
          items.push({ id: `folder-${folder.id}`, type: "folder", data: folder });
        }
      });
  }

  // Add templates
  let filtered = Array.isArray(templates.value) ? [...templates.value] : [];

  // Folder filter
  if (selectedFolderId.value !== null) {
    // Show templates in selected folder
    filtered = filtered.filter((template) => {
      const templateFolderId = template.folder_id || null;
      return String(templateFolderId) === String(selectedFolderId.value);
    });
  } else {
    // Show templates in root (no folder)
    filtered = filtered.filter((template) => {
      const templateFolderId = template.folder_id;
      return !templateFolderId || templateFolderId === null || templateFolderId === "";
    });
  }

  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase();
    filtered = filtered.filter(
      (template) => template.name.toLowerCase().includes(query)
    );
  }

  // Category filter
  if (selectedCategory.value) {
    filtered = filtered.filter((template) => template.category === selectedCategory.value);
  }

  // Add templates to items
  filtered.forEach((template) => {
    items.push({ id: `template-${template.id}`, type: "template", data: template });
  });

  // Sort - parent navigation always first, then folders, then templates
  items.sort((a, b) => {
    // Parent navigation always comes first
    if (a.type === "parent") return -1;
    if (b.type === "parent") return 1;
    
    // Folders always come before templates
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
          return (
            (b.data.submitter_count ?? b.data.submitters?.length ?? 0) -
            (a.data.submitter_count ?? a.data.submitters?.length ?? 0)
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
watch(selectedFolderId, (newFolderId) => {
  if (newFolderId && route.path !== `/templates/${newFolderId}/folder`) {
    router.push(`/templates/${newFolderId}/folder`);
  } else if (!newFolderId && route.path !== "/templates") {
    router.push("/templates");
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

const openTemplateView = (template: Template) => {
  // Open read-only view in a new tab/window
  const href = router.resolve({ name: "template-view", params: { id: template.id } }).href;
  window.open(href, "_blank", "noopener,noreferrer");
};

const openTemplateEditor = (template: Template) => {
  // Open template editor in the current tab
  router.push(`/templates/${template.id}/edit`);
};

const editTemplate = (template: Template) => {
  // Open edit modal for name and category
  templateToEdit.value = template;
  showEditTemplateModal.value = true;
};

const handleEditTemplate = async (formData: Record<string, unknown>) => {
  if (!templateToEdit.value) return;

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

    const response = await apiPut(`/api/v1/templates/${templateToEdit.value.id}`, updateData);
    if (response && response.data) {
      showEditTemplateModal.value = false;
      templateToEdit.value = null;
      await loadTemplates();
    } else {
      alert(t('templates.updateError') || 'Failed to update template');
    }
  } catch (error) {
    console.error("Failed to update template:", error);
    alert(t('templates.updateError') || 'Failed to update template');
  }
};

const createNewTemplate = async () => {
  try {
    // Create a new empty template
    const response = await apiPost("/api/v1/templates", {
      name: t('templates.newTemplate'),
      description: "",
      schema: [],
      fields: [],
      submitters: []
    });

    if (response && response.data) {
      // API returns: { success: true, message: "template", data: Template }
      // or: { data: Template } (if wrapped)
      let newTemplate = response.data;
      
      // Handle case where data might be wrapped
      if (newTemplate && typeof newTemplate === 'object' && 'template' in newTemplate) {
        newTemplate = newTemplate.template;
      }
      
      const templateId = newTemplate?.id || (newTemplate && typeof newTemplate === 'object' && 'id' in newTemplate ? newTemplate.id : null);
      
      if (templateId) {
        // Navigate to edit page for the new template
        router.push(`/templates/${templateId}/edit`);
  } else {
        console.error("Failed to get template ID from response:", response);
        alert(t('templates.createTemplateError'));
      }
    } else {
      console.error("Failed to create template: unexpected response", response);
      alert(t('templates.createTemplateError'));
    }
  } catch (error: any) {
    console.error("Failed to create template:", error);
    const errorMessage = error?.message || t('templates.createTemplateError');
    alert(errorMessage);
  }
};

const handleFileSelect = (event: Event) => {
  const input = event.target as HTMLInputElement;
  if (input && input.files && input.files.length > 0) {
    const file = input.files[0];
    // Validate file type
    if (file.type === 'application/pdf' || file.name.endsWith('.pdf')) {
      selectedFile.value = file;
    } else {
      alert(t('templates.invalidFileType'));
      if (input) {
        input.value = '';
      }
    }
  } else {
    selectedFile.value = null;
  }
};

const handleDrop = (event: DragEvent) => {
  event.preventDefault();
  if (event.dataTransfer && event.dataTransfer.files && event.dataTransfer.files.length > 0) {
    const file = event.dataTransfer.files[0];
    // Validate file type
    if (file.type === 'application/pdf' || file.name.endsWith('.pdf')) {
      selectedFile.value = file;
      // Also update the input element
      if (templateFileInput.value) {
        const dataTransfer = new DataTransfer();
        dataTransfer.items.add(file);
        templateFileInput.value.files = dataTransfer.files;
      }
    } else {
      alert(t('templates.invalidFileType'));
    }
  }
};

const removeSelectedFile = () => {
  selectedFile.value = null;
  if (templateFileInput.value) {
    templateFileInput.value.value = '';
  }
};

const handleCancelCreateTemplate = () => {
  showCreateTemplateModal.value = false;
  removeSelectedFile();
};

const handleCreateTemplate = async (formData: any): Promise<void> => {
  const templateName = (formData.name as string)?.trim() || t('templates.newTemplate');
  
  try {
    if (selectedFile.value) {
      // Create template from file
      const file = selectedFile.value;
      
      // Convert file to base64 (payload only)
      const base64String = await fileToBase64Payload(file);
      
      // Determine file type
      let fileType = 'pdf';
      if (file.name.endsWith('.pdf')) {
        fileType = 'pdf';
      } else if (file.name.endsWith('.html') || file.name.endsWith('.htm')) {
        fileType = 'html';
      } else if (file.name.endsWith('.docx')) {
        fileType = 'docx';
      }
      
      
      const response = await apiPost("/api/v1/templates/from-file", {
        name: templateName,
        type: fileType,
        file_base64: base64String,
        description: ""
      });
      
      console.log('Template creation response:', response);
      
      if (response && response.data) {
        let newTemplate = response.data;
        if (newTemplate && typeof newTemplate === 'object' && 'template' in newTemplate) {
          newTemplate = newTemplate.template;
        }
        
        const templateId = newTemplate?.id || (newTemplate && typeof newTemplate === 'object' && 'id' in newTemplate ? newTemplate.id : null);
        
        if (templateId) {
          // Clean up
          removeSelectedFile();
          
          // Reset modal submitting state before closing
          if (createTemplateModalRef.value) {
            createTemplateModalRef.value.resetSubmitting();
          }
          
          // Close modal
          showCreateTemplateModal.value = false;
          
          // Reload templates
          await loadTemplates();
          
          // Navigate to edit page for the new template
          router.push(`/templates/${templateId}/edit`);
        } else {
          console.error("Failed to get template ID from response:", response);
          const errorMsg = t('templates.createTemplateError');
          alert(errorMsg);
          if (createTemplateModalRef.value) {
            createTemplateModalRef.value.resetSubmitting();
          }
          throw new Error('Template ID not found in response');
        }
      } else {
        console.error("Failed to create template: unexpected response", response);
        const errorMsg = t('templates.createTemplateError');
        alert(errorMsg);
        if (createTemplateModalRef.value) {
          createTemplateModalRef.value.resetSubmitting();
        }
        throw new Error('Unexpected response format');
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
        if (newTemplate && typeof newTemplate === 'object' && 'template' in newTemplate) {
          newTemplate = newTemplate.template;
        }
        
        const templateId = newTemplate?.id || (newTemplate && typeof newTemplate === 'object' && 'id' in newTemplate ? newTemplate.id : null);
        
        if (templateId) {
          // Reset modal submitting state before closing
          if (createTemplateModalRef.value) {
            createTemplateModalRef.value.resetSubmitting();
          }
          
          // Close modal
          showCreateTemplateModal.value = false;
          
          // Reload templates
          await loadTemplates();
          
          // Navigate to edit page for the new template
          router.push(`/templates/${templateId}/edit`);
        } else {
          console.error("Failed to get template ID from response:", response);
          const errorMsg = t('templates.createTemplateError');
          alert(errorMsg);
          if (createTemplateModalRef.value) {
            createTemplateModalRef.value.resetSubmitting();
          }
          throw new Error('Template ID not found in response');
        }
      } else {
        console.error("Failed to create template: unexpected response", response);
        const errorMsg = t('templates.createTemplateError');
        alert(errorMsg);
        if (createTemplateModalRef.value) {
          createTemplateModalRef.value.resetSubmitting();
        }
        throw new Error('Unexpected response format');
      }
    }
  } catch (error: any) {
    console.error("Failed to create template:", error);
    const errorMessage = error?.message || t('templates.createTemplateError');
    alert(errorMessage);
    // Reset submitting state on error
    if (createTemplateModalRef.value) {
      createTemplateModalRef.value.resetSubmitting();
    }
    // Don't re-throw - let user retry
  }
};


const viewFolder = (folder: TemplateFolder) => {
  router.push(`/templates/${folder.id}/folder`);
};

const goBackToParent = () => {
  selectedFolderId.value = null;
  router.push("/templates");
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
  router.push("/templates");
};

const formatDate = (date: string | Date) => {
  return new Date(date).toLocaleDateString();
};

const translateCategory = (category: string): string => {
  if (!category) return '';
  // Map category values to translation keys
  const categoryMap: Record<string, string> = {
    'business': 'templates.business',
    'legal': 'templates.legal',
    'personal': 'templates.personal',
    'education': 'templates.education'
  };
  
  const translationKey = categoryMap[category.toLowerCase()];
  if (translationKey) {
    return t(translationKey);
  }
  // Fallback to original category if no translation found
  return category;
};

// Folder operations
const openFolderMenu = (folder: TemplateFolder, event: MouseEvent) => {
  event.stopPropagation();
  folderMenuOpen.value = folderMenuOpen.value === folder.id ? null : folder.id;
};

const handleCreateFolder = async (formData: Record<string, unknown>) => {
  const name = (formData.name as string)?.trim();
  if (!name || name === "") {
    alert(t('templates.folderNameRequired'));
    return;
  }

  try {
    const response = await apiPost("/api/v1/templates/folders", { name });
    // Backend returns {success: bool, message: string, data: any}
    // For 201 Created, success might be false (because code != 200), but data will be present
    if (response && (response.data || response.message)) {
      showCreateFolderModal.value = false;
      await loadFolders();
    } else {
      console.error("Failed to create folder: unexpected response", response);
      alert(t('templates.createFolderUnexpectedError'));
    }
  } catch (error: any) {
    console.error("Failed to create folder:", error);
    const errorMessage = error?.message || t('templates.createFolderError');
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

  if (!confirm(t('templates.deleteFolderConfirm', { name: folder.name }))) {
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
      alert(t('templates.moveTemplateUnexpectedError'));
    }
  } catch (error: any) {
    console.error("Failed to move template:", error);
    const errorMessage = error?.message || t('templates.moveTemplateError');
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

watch(showEditTemplateModal, async (isOpen) => {
  if (isOpen && templateToEdit.value && editTemplateModalRef.value) {
    await nextTick();
    if (editTemplateModalRef.value.setFormData) {
      editTemplateModalRef.value.setFormData({ 
        name: templateToEdit.value.name,
        category: templateToEdit.value.category || ''
      });
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
