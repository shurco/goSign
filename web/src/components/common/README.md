# Universal Vue Components

This folder contains reusable components that are used throughout the application.

## FieldInput.vue

Universal component for displaying and inputting data for all field types.

### Supported Types

- **text** - text field
- **signature** - signature (canvas drawing)
- **initials** - initials (canvas drawing)
- **date** - date selection
- **image** - image upload
- **file** - file upload
- **checkbox** - checkbox
- **radio** - radio buttons
- **select** - dropdown list
- **multiple** - multiple selection
- **cells** - cells (for codes, numbers)
- **stamp** - stamp (image upload)
- **payment** - payment (with amount)
- **phone** - phone number

### Usage

```vue
<template>
  <FieldInput
    type="text"
    v-model="formData.name"
    placeholder="Enter your name"
    :required="true"
    @blur="validateField"
  />

  <FieldInput
    type="signature"
    v-model="formData.signature"
    :required="true"
  />

  <FieldInput
    type="select"
    v-model="formData.country"
    :options="countryOptions"
    placeholder="Select country"
  />

  <FieldInput
    type="radio"
    v-model="formData.gender"
    :options="[
      { id: '1', value: 'Male' },
      { id: '2', value: 'Female' }
    ]"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import FieldInput from '@/components/common/FieldInput.vue'

const formData = ref({
  name: '',
  signature: '',
  country: '',
  gender: ''
})

const countryOptions = [
  { id: '1', value: 'USA' },
  { id: '2', value: 'UK' },
  { id: '3', value: 'Canada' }
]
</script>
```

### Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `type` | `FieldType` | required | Field type |
| `modelValue` | `string \| boolean \| string[]` | `''` | Field value |
| `placeholder` | `string` | `''` | Placeholder |
| `label` | `string` | `''` | Field label |
| `required` | `boolean` | `false` | Required field |
| `readonly` | `boolean` | `false` | Read only |
| `disabled` | `boolean` | `false` | Disabled |
| `options` | `Option[]` | `[]` | Options for select, radio, multiple |
| `cellCount` | `number` | `6` | Number of cells for type="cells" |
| `paymentAmount` | `string` | `$0.00` | Amount for type="payment" |
| `error` | `string` | `''` | Error message |

### Events

- `update:modelValue` - value update
- `blur` - focus loss

## ResourceTable.vue

Universal table with search, sorting, pagination and actions.

### Usage

```vue
<template>
  <ResourceTable
    :data="submissions"
    :columns="columns"
    :isLoading="loading"
    searchable
    selectable
    @select="handleSelect"
    @edit="handleEdit"
    @delete="handleDelete"
  >
    <!-- Custom cell -->
    <template #cell-status="{ value }">
      <span class="badge" :class="getStatusClass(value)">
        {{ value }}
      </span>
    </template>

    <!-- Custom actions -->
    <template #actions="{ item }">
      <button class="btn btn-sm" @click="sendReminder(item)">
        Send Reminder
      </button>
    </template>
  </ResourceTable>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ResourceTable from '@/components/common/ResourceTable.vue'

const submissions = ref([
  { id: '1', title: 'Contract 1', status: 'pending', created_at: '2024-01-01' },
  { id: '2', title: 'Contract 2', status: 'completed', created_at: '2024-01-02' }
])

const columns = [
  { key: 'title', label: 'Title', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { 
    key: 'created_at', 
    label: 'Created', 
    sortable: true,
    formatter: (value) => new Date(value).toLocaleDateString()
  }
]

const handleSelect = (selectedItems) => {
  console.log('Selected:', selectedItems)
}

const handleEdit = (item) => {
  console.log('Edit:', item)
}

const handleDelete = (item) => {
  console.log('Delete:', item)
}
</script>
```

### Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `data` | `unknown[]` | required | Table data |
| `columns` | `Column[]` | required | Column configuration |
| `searchable` | `boolean` | `true` | Enable search |
| `searchPlaceholder` | `string` | `'Search...'` | Search placeholder |
| `searchKeys` | `string[]` | `[]` | Search keys (all columns by default) |
| `selectable` | `boolean` | `false` | Checkboxes for selection |
| `showFilters` | `boolean` | `true` | Show filters |
| `showPagination` | `boolean` | `true` | Show pagination |
| `pageSize` | `number` | `10` | Items per page |
| `isLoading` | `boolean` | `false` | Loading state |
| `emptyMessage` | `string` | `'No data available'` | Message for empty table |
| `hasActions` | `boolean` | `true` | Actions column |
| `showEdit` | `boolean` | `true` | Edit button |
| `showDelete` | `boolean` | `true` | Delete button |
| `idKey` | `string` | `'id'` | Identifier key |

### Column Interface

```typescript
interface Column {
  key: string                           // Key in data
  label: string                         // Column header
  sortable?: boolean                    // Sorting capability
  formatter?: (value: unknown) => string // Value formatting
  headerClass?: string                  // CSS class for header
  cellClass?: string                    // CSS class for cell
}
```

### Events

- `select(selectedItems)` - item selection
- `edit(item)` - item editing
- `delete(item)` - item deletion
- `page-change(page)` - page change
- `search(query)` - search

### Slots

- `filters` - custom filters
- `cell-{key}` - custom cell display
- `actions` - custom actions

### Exposed Methods

```typescript
clearSelection()           // Clear selection
selectAll()               // Select all
getSelectedItems()        // Get selected items
resetPage()               // Reset to first page
```

## FormModal.vue

Universal modal for forms with validation.

### Usage

```vue
<template>
  <button class="btn" @click="openModal">Create Submission</button>

  <FormModal
    v-model="isOpen"
    title="Create New Submission"
    submit-text="Create"
    size="lg"
    @submit="handleSubmit"
    @cancel="handleCancel"
  >
    <template #default="{ formData, errors }">
      <div class="space-y-4">
        <div class="form-control">
          <label class="label">
            <span class="label-text">Title</span>
          </label>
          <input
            v-model="formData.title"
            type="text"
            class="input input-bordered"
            :class="{ 'input-error': errors.title }"
          />
          <label v-if="errors.title" class="label">
            <span class="label-text-alt text-error">{{ errors.title }}</span>
          </label>
        </div>

        <div class="form-control">
          <label class="label">
            <span class="label-text">Template</span>
          </label>
          <select v-model="formData.template_id" class="select select-bordered">
            <option value="">Select template</option>
            <option v-for="tpl in templates" :key="tpl.id" :value="tpl.id">
              {{ tpl.name }}
            </option>
          </select>
        </div>
      </div>
    </template>
  </FormModal>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import FormModal from '@/components/common/FormModal.vue'

const isOpen = ref(false)
const templates = ref([
  { id: '1', name: 'Contract Template' },
  { id: '2', name: 'NDA Template' }
])

const openModal = () => {
  isOpen.value = true
}

const handleSubmit = async (formData) => {
  console.log('Submit:', formData)
  // API call...
  isOpen.value = false
}

const handleCancel = () => {
  console.log('Cancelled')
}
</script>
```

### Props

| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `modelValue` | `boolean` | required | Open state |
| `title` | `string` | required | Title |
| `size` | `'sm' \| 'md' \| 'lg' \| 'xl'` | `'md'` | Modal size |
| `submitText` | `string` | `'Submit'` | Button text Submit |
| `cancelText` | `string` | `'Cancel'` | Button text Cancel |
| `showCancel` | `boolean` | `true` | Show Cancel button |
| `closeButton` | `boolean` | `true` | Close button (X) |
| `closeOnOutsideClick` | `boolean` | `true` | Close on outside click |
| `closeOnEscape` | `boolean` | `true` | Close on Escape |
| `customClass` | `string` | `''` | Additional CSS classes |
| `validateOnMount` | `boolean` | `false` | Validate on mount |

### Events

- `update:modelValue(value)` - state change
- `submit(formData)` - form submission
- `cancel()` - cancellation
- `close()` - closing

### Slots

- `default` - form content (receives `formData` and `errors`)
- `footer` - custom footer (receives `submit`, `cancel`, `isSubmitting`)

### Exposed Methods

```typescript
open()                              // Open modal
close()                             // Close modal
setFormData(data)                   // Set form data
getFormData()                       // Get form data
setError(field, message)            // Set error
clearError(field)                   // Clear error
clearAllErrors()                    // Clear all errors
resetForm()                         // Reset form
validateForm()                      // Validate form
isValid()                           // Check validity
```

## Usage Principles

### DRY (Don't Repeat Yourself)
- All three components are reused across different parts of the application
- Unified search, sorting, pagination logic in `ResourceTable`
- Unified way of displaying all field types in `FieldInput`

### KISS (Keep It Simple, Stupid)
- Simple API with reasonable default values
- Minimum required props
- Clear separation of responsibilities

### Composition through slots
- Customization through slots instead of props
- Flexibility while maintaining simple base API
- Slot props for access to internal state

## Integration Examples

### Submissions Page
```vue
<ResourceTable
  :data="submissions"
  :columns="submissionColumns"
  @edit="openEditModal"
  @delete="confirmDelete"
>
  <template #cell-status="{ value }">
    <span class="badge" :class="getStatusClass(value)">{{ value }}</span>
  </template>
</ResourceTable>
```

### Settings Page
```vue
<FormModal v-model="showAPIKeyModal" title="Create API Key" @submit="createAPIKey">
  <FieldInput type="text" v-model="apiKeyForm.name" label="Name" required />
  <FieldInput type="date" v-model="apiKeyForm.expires_at" label="Expires At" />
</FormModal>
```

### Signing Portal
```vue
<FieldInput
  v-for="field in template.fields"
  :key="field.id"
  :type="field.type"
  v-model="submissionData[field.id]"
  :label="field.name"
  :required="field.required"
  :options="field.options"
/>
```

