# UI Components

Simple, reusable UI components following KISS and DRY principles.

## Philosophy

- **KISS**: Simple components with single responsibility
- **DRY**: Reusable patterns, change in one place
- **Composition**: Complex components from simple ones
- **Type Safety**: Full TypeScript support
- **Consistency**: Unified variant and size system

## Base Components

### Button
Simple button with variants, sizes, and loading state.
```vue
<Button variant="primary" size="md" :loading="false">
  Click me
</Button>
```

### Input
Text input with type support and error state.
```vue
<Input v-model="value" type="text" :error="false" />
```

### Checkbox
Checkbox for selection with consistent styling.
```vue
<Checkbox v-model="checked" size="md" />
```

### Switch
Toggle switch for binary states (on/off).
```vue
<Switch v-model="enabled" size="md" />
```

### Radio
Radio button with consistent styling.
```vue
<Radio v-model="selected" value="option1" />
```

### Select
Dropdown with options.
```vue
<Select v-model="value">
  <option value="1">Option 1</option>
</Select>
```

## Composite Components

### Modal
Modal dialog with slots for header, body, and footer.
```vue
<Modal v-model="isOpen" size="md">
  <template #header>Title</template>
  <template #default>Content</template>
  <template #footer>Actions</template>
</Modal>
```

### Card
Content container with header and footer slots.
```vue
<Card>
  <template #header>Header</template>
  Content
  <template #footer>Footer</template>
</Card>
```

## Variant System

All components use consistent variants:
- `primary` - Main actions
- `success` - Success states
- `warning` - Warning states
- `error` - Error states
- `ghost` - Transparent/minimal style
- `info` - Informational

## Size System

Standard sizes across components:
- `sm` - Small
- `md` - Medium (default)
- `lg` - Large

