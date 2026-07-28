# Frontend Component Architecture

**Last Updated**: 2026-07-28

## Overview

The goSign frontend lives in `web/private/` and uses Svelte 5 (runes), SvelteKit 2 (static SPA, `ssr = false`), TypeScript, and Tailwind CSS v4. The component library follows three layers: UI primitives, common composites, and domain-specific components.

## Component Structure

```
web/private/src/lib/components/
├── ui/                    # 18 primitive UI components
├── common/                # 4 generic composite components
├── field/                 # Field-specific components
├── signing/               # Signing portal components
├── template/              # Document template components
├── organization/          # Organization modals
└── SvgIcon.svelte         # Inline SVG icon loader
```

## UI Layer (Primitives)

Located in `web/private/src/lib/components/ui/`.

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
- **ButtonGroup** - Grouped buttons
- **Badge** - Status indicators and labels

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

All UI components follow consistent design principles:
- **Variants**: `primary`, `success`, `warning`, `error`, `ghost`, `info`
- **Sizes**: `sm`, `md` (default), `lg`
- **TypeScript**: Full type safety with typed props
- **Accessibility**: ARIA labels and keyboard navigation

## Common Layer (Composites)

Located in `web/private/src/lib/components/common/`.

### FieldInput

Universal component for field types used in document signing. Fields support `readonly`, `validation` (pattern, min, max, message), and `preferences` (format, price, currency, date format, signature format, etc.).

**Supported Types**: `text`, `number`, `signature`, `initials`, `date`, `image`, `file`, `checkbox`, `radio`, `select`, `multiple`, `cells`, `stamp`, `payment`, `phone`

**Usage**:
```svelte
<FieldInput type="signature" bind:value={formData.signature} required />
```

### ResourceTable

Universal table with search, sorting, pagination, bulk selection, and custom cell rendering via snippets.

**Usage**:
```svelte
<ResourceTable data={submissions} {columns} searchable selectable onEdit={handleEdit}>
  {#snippet cellStatus({ value })}
    <Badge variant={getStatusVariant(value)}>{value}</Badge>
  {/snippet}
</ResourceTable>
```

### FieldProgressDots

Progress indicator for the signing flow: clickable dots per field with states (filled, current, pending). Used on the submitter signing page for quick field navigation.

### FormModal

Universal modal for forms: built-in state management, validation, close on outside click/ESC, loading state during submission.

## Composables

Located in `web/private/src/lib/composables/` (`.svelte.ts` modules using runes).

- **useCurrentUser** - Shared current user state for layout components (Sidebar, SettingsSidebar); restored from sessionStorage to avoid re-fetching between layouts. Exposes `userData`, `isAdmin`, `loadUserData`, `clearUser`.
- **useConditions** - Conditional field logic evaluation (show/hide/require/disable). Covered by unit tests.
- **useFormulas** - Formula parsing and calculation for computed fields. Covered by unit tests.
- **ui** - Shared UI actions (dropdowns, click-outside, etc.).

## Domain Layer

### Field Components (`components/field/`)

- **Field** - Individual field wrapper with drag-and-drop
- **Type** - Field type selector
- **Submitter** - Submitter assignment
- **List** - Field list with grouping
- **Contenteditable** - Inline text editing
- **ConditionBuilder** - Visual editor for conditional field logic
- **FormulaBuilder** - Visual editor for calculated fields
- **SigningModeSelector** - Signing mode (Parallel / Sequential) with i18n; optional `hideOrderList` to control order via parent
- **inputs/** - Type-specific inputs (Cells, Date, File, Select, Signature, Text)

### Signing Components (`components/signing/`)

Components for the public signing portal (`/s/:slug`):
- **FieldFormDrawer** - Bottom drawer with current field form, progress dots, and prev/next navigation

### Template Components (`components/template/`)

- **Document** - Complete document viewer
- **Page** - Single page renderer
- **Area** - Field placement area with drag-and-drop
- **Preview** - Document preview mode

### Organization Components (`components/organization/`)

- **CreateOrganizationModal / EditOrganizationModal** - Organization CRUD

## Design Principles

### KISS (Keep It Simple, Stupid)
- Simple API with reasonable defaults, minimal required props
- One component = one purpose

### DRY (Don't Repeat Yourself)
- Unified search/sort/pagination logic in ResourceTable
- Single FieldInput implementation for all field types
- Consistent variant/size system

### Composition Over Configuration
- Customization through snippets (Svelte 5), not prop explosions

### Type Safety
- Typed props for all components, discriminated unions for variants

## Usage Examples

### Dashboard Page
```svelte
<ResourceTable data={recentSubmissions} columns={dashboardColumns} />
```

### Settings Page
```svelte
<Card>
  <FormControl label={t("settings.emailProvider")}>
    <Select bind:value={config.provider} options={providers} />
  </FormControl>
</Card>
```

### Routing

Pages live in `web/private/src/lib/pages/` as plain components; `web/private/src/routes/` contains thin SvelteKit wrappers that import them. Route params come from `$app/state`.

## Testing

Unit tests live next to the code in `__tests__/` directories and run with Vitest (`bun run test`):
- `composables/__tests__/useConditions.svelte.test.ts`
- `composables/__tests__/useFormulas.svelte.test.ts`
- `i18n/__tests__/i18n.spec.ts`

Type checking: `bun run check` (svelte-check). Linting: `bun run lint` (Prettier + ESLint).
