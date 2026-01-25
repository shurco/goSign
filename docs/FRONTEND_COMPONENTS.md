# Frontend Component Architecture

**Last Updated**: 2026-01-25

## Overview

goSign frontend follows a component-based architecture with Vue 3 Composition API, TypeScript, and Tailwind CSS v4. The component library consists of three layers: UI primitives, common composites, and domain-specific components.

## Component Structure

```
web/src/components/
├── ui/                    # 21 primitive UI components
├── common/                # 3 generic composite components
├── field/                 # Field-specific components
└── template/              # Document template components
```

## UI Layer (Primitives)

Located in `web/src/components/ui/`, these 21 components provide the foundation for all user interfaces.

### Form Components
- **Input** - Text input with type support, error state, and optional password visibility toggle
- **Checkbox** - Selection with consistent styling
- **Radio** - Radio button with unified design
- **Select** - Dropdown with options
- **Switch** - Toggle for binary states
- **FileInput** - File upload with drag-and-drop
- **FileDropZone** - Reusable file drop zone for uploads
- **FormControl** - Form field wrapper with label and validation

### Action Components
- **Button** - Primary action component with variants and loading state
- **Badge** - Status indicators and labels
- **Tab/Tabs** - Tabbed navigation

### Layout Components
- **Card** - Content container with header/footer
- **Modal** - Dialog with overlay
- **Table** - Data table with sorting
- **Pagination** - Navigation between pages

### Feedback Components
- **Alert** - Notifications and messages
- **LoadingSpinner** - Loading state indicator

### Data Display
- **Label** - Text labels with consistent styling
- **Stat/Stats** - Statistics display

All UI components follow consistent design principles:
- **Variants**: `primary`, `success`, `warning`, `error`, `ghost`, `info`
- **Sizes**: `sm`, `md` (default), `lg`
- **TypeScript**: Full type safety with interfaces
- **Accessibility**: ARIA labels and keyboard navigation

## Common Layer (Composites)

Located in `web/src/components/common/`, these 3 components compose UI primitives into powerful, reusable patterns.

### FieldInput

Universal component for field types used in document signing. Fields support `readonly`, `validation` (pattern, min, max, message), and `preferences` (format, price, currency, date format, signature format, etc.).

**Supported Types**:
- `text` - Text field
- `number` - Number field with optional format preferences
- `signature` - Canvas signature (format, with_signature_id)
- `initials` - Canvas initials
- `date` - Date picker with optional format pattern
- `image` - Image upload
- `file` - File upload
- `checkbox` - Checkbox
- `radio` - Radio buttons
- `select` - Dropdown
- `multiple` - Multi-select
- `cells` - Code/number cells (cell_count persisted)
- `stamp` - Stamp image
- `payment` - Payment with amount
- `phone` - Phone number

**Key Features**:
- Single component for all field types
- Structured validation (FieldValidation) and preferences (FieldPreferences)
- Built-in required/optional and readonly support
- Automatic value type conversion
- Formula fields with calculated values; date format by pattern

**Usage**:
```vue
<FieldInput
  type="signature"
  v-model="formData.signature"
  :required="true"
/>
```

### ResourceTable

Universal table with search, sorting, pagination, and actions.

**Key Features**:
- Built-in search across all columns
- Sortable columns with indicators
- Pagination with configurable page size
- Bulk selection with checkboxes
- Custom cell rendering via slots
- Loading states
- Empty state messages
- Export capabilities

**Usage**:
```vue
<ResourceTable
  :data="submissions"
  :columns="columns"
  searchable
  selectable
  @edit="handleEdit"
  @delete="handleDelete"
>
  <template #cell-status="{ value }">
    <Badge :variant="getStatusVariant(value)">{{ value }}</Badge>
  </template>
</ResourceTable>
```

### FormModal

Universal modal for forms with validation.

**Key Features**:
- Built-in form state management
- Validation framework integration
- Customizable size and layout
- Close on outside click/ESC
- Loading state during submission
- Success/error callbacks

**Usage**:
```vue
<FormModal
  v-model="isOpen"
  title="Create API Key"
  @submit="handleSubmit"
>
  <template #default="{ formData, errors }">
    <FieldInput
      v-model="formData.name"
      type="text"
      label="Name"
      :error="errors.name"
    />
  </template>
</FormModal>
```

## Domain Layer

### Field Components (`web/src/components/field/`)

Components specific to document field management:
- **Field** - Individual field wrapper with drag-and-drop
- **Type** - Field type selector
- **Submitter** - Submitter assignment
- **List** - Field list with grouping
- **Contenteditable** - Inline text editing
- **SigningModeSelector** - Signing mode (Parallel / Sequential) with i18n. Always visible (no collapse). Optional `hideOrderList` to control order via parent (e.g. draggable signer cards on Submissions page).

### Template Components (`web/src/components/template/`)

Components for document template editing and viewing:
- **Document** - Complete document viewer
- **Page** - Single page renderer
- **Area** - Field placement area with drag-and-drop
- **Preview** - Document preview mode

## Design Principles

### KISS (Keep It Simple, Stupid)
- Simple API with reasonable defaults
- Minimal required props
- Clear separation of responsibilities
- One component = one purpose

### DRY (Don't Repeat Yourself)
- Reusable components across all pages
- Unified search/sort/pagination logic
- Single implementation for all field types
- Consistent variant/size system

### Composition Over Configuration
- Customization through slots, not props
- Slot props for internal state access
- Flexible base API
- Easy to extend without breaking changes

### Type Safety
- Full TypeScript interfaces for all props
- Type-safe events and emits
- Discriminated unions for variants
- Compile-time error checking

## Component Usage Examples

### Dashboard Page
```vue
<Stats>
  <Stat label="Total Submissions" :value="stats.total" />
  <Stat label="Completed" :value="stats.completed" />
  <Stat label="Pending" :value="stats.pending" />
</Stats>

<ResourceTable
  :data="recentSubmissions"
  :columns="dashboardColumns"
/>
```

### Settings Page
```vue
<Tabs v-model="activeTab">
  <Tab label="Profile" value="profile" />
  <Tab label="Email" value="email" />
  <Tab label="Storage" value="storage" />
</Tabs>

<Card>
  <FormControl label="Email Provider">
    <Select v-model="config.provider">
      <option value="smtp">SMTP</option>
      <option value="sendgrid">SendGrid</option>
    </Select>
  </FormControl>
</Card>
```

### Submissions Page
Create-submission modal uses FormModal with FormControl (template select, signers), SigningModeSelector (`hideOrderList` + draggable signer cards in Sequential mode with drop indicator), and Save button. All copy is i18n (e.g. `signingMode.*`, `submissions.*`, `common.save`).

```vue
<ResourceTable
  :data="submissions"
  :columns="columns"
  searchable
  selectable
  @edit="openEditModal"
>
  <template #filters>
    <Select v-model="statusFilter">
      <option value="">All Statuses</option>
      <option value="pending">Pending</option>
      <option value="completed">Completed</option>
    </Select>
  </template>

  <template #cell-status="{ value }">
    <Badge :variant="getStatusVariant(value)">
      {{ value }}
    </Badge>
  </template>

  <template #actions="{ item }">
    <Button
      variant="ghost"
      size="sm"
      @click="sendReminder(item)"
    >
      Send Reminder
    </Button>
  </template>
</ResourceTable>
```

### Edit/Template Page
```vue
<template>
  <Document :pages="template.pages">
    <Page
      v-for="page in template.pages"
      :key="page.number"
      :page="page"
    >
      <Area
        :fields="getPageFields(page)"
        :submitters="template.submitters"
        @field-select="handleFieldSelect"
        @field-update="handleFieldUpdate"
      />
    </Page>
  </Document>

  <FormModal
    v-model="showFieldModal"
    title="Edit Field"
    @submit="updateField"
  >
    <FieldInput
      v-model="editingField.name"
      type="text"
      label="Field Name"
      required
    />
    <Select v-model="editingField.type" label="Field Type">
      <option value="text">Text</option>
      <option value="signature">Signature</option>
      <option value="date">Date</option>
    </Select>
  </FormModal>
</template>
```

### Signing Portal (`/s/:slug`)
```vue
<template>
  <Card>
    <template #header>
      <h2>Sign: {{ submission.name }}</h2>
    </template>

    <div class="space-y-4">
      <FieldInput
        v-for="field in requiredFields"
        :key="field.id"
        :type="field.type"
        v-model="submissionData[field.id]"
        :label="field.name"
        :required="field.required"
        :options="field.options"
        :error="errors[field.id]"
      />
    </div>

    <template #footer>
      <Button
        variant="primary"
        size="lg"
        :loading="submitting"
        @click="submitSignature"
      >
        Complete Signature
      </Button>
    </template>
  </Card>
</template>
```

## Performance Considerations

### Lazy Loading
All pages and heavy components are lazy-loaded via Vue Router:
```typescript
{
  path: '/submissions',
  component: () => import('@/pages/Submissions.vue')
}
```

### Component Optimization
- `v-memo` for expensive list items
- `v-once` for static content
- Computed properties for expensive calculations
- `shallowRef` for large data structures

### Virtual Scrolling
ResourceTable implements virtual scrolling for large datasets (1000+ rows).

### Bundle Size
- Base UI components: ~15KB gzipped
- Common components: ~8KB gzipped
- Domain components: ~12KB gzipped
- Total component library: ~35KB gzipped

## Testing Strategy

### Unit Tests
Each UI component has unit tests covering:
- Props validation
- Event emission
- Slot rendering
- Variant/size combinations

### Integration Tests
Common components have integration tests:
- User interactions (click, type, select)
- Form validation flows
- Table search/sort/pagination
- Modal open/close behavior

### E2E Tests
Critical user flows are covered by E2E tests:
- Document signing flow
- Template creation
- Submission management
- API key generation

## Migration from Legacy Components

The new component library replaces previous ad-hoc implementations:

### Before
```vue
<!-- Multiple implementations across pages -->
<div class="custom-table">...</div>
<div class="another-table">...</div>
<div class="yet-another-table">...</div>
```

### After
```vue
<!-- Single reusable component -->
<ResourceTable :data="data" :columns="columns" />
```

## Documentation

Each component directory includes a README with:
- Component API (props, events, slots)
- Usage examples
- Design principles
- Integration patterns

**Component Documentation**:
- [UI Components](../web/src/components/ui/README.md)
- [Common Components](../web/src/components/common/README.md)

## Future Improvements

### Planned
- [ ] Dark mode support for all components
- [ ] Animation library integration
- [ ] Storybook documentation
- [ ] Component playground
- [ ] Accessibility audit and improvements
- [ ] i18n support for all text
- [ ] Component performance profiling

### Under Consideration
- [ ] Headless UI component variants
- [ ] Custom theme system
- [ ] Component composition utilities
- [ ] Advanced table features (column reordering, resizing)
- [ ] Form builder component

---

**Status**: ✅ Complete  
**Version**: 1.0.0  
**Components**: 21 UI + 3 Common + 9 Domain = 33 total

