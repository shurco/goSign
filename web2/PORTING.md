# Vue → Svelte 5 porting conventions (goSign web2)

Temporary working document for the Vue→Svelte migration. Source of truth for every ported file. Delete after migration completes.

## Context

- **Source**: `web/` — Vue 3.5 SPA (`<script setup lang="ts">`, vue-router, vue-i18n, composables).
- **Target**: `web2/` — Svelte 5 (runes) + SvelteKit 2 static SPA (`ssr=false`), TypeScript strict, Tailwind v4 (identical classes).
- Port **1:1**: same markup, same Tailwind classes, same behavior, same API calls. Do not redesign, do not add features, do not "improve" logic. Translate idioms only.
- `@/` alias → `web2/src/lib/` (configured in vite.config.ts). Keep import specifiers identical to Vue where possible, only changing extensions/suffixes.

## File mapping

| Vue (web/src/…)                          | Svelte (web2/src/lib/…)                                                             |
| ---------------------------------------- | ----------------------------------------------------------------------------------- |
| `components/**/X.vue`                    | `components/**/X.svelte`                                                            |
| `pages/X.vue`                            | `pages/X.svelte`                                                                    |
| `composables/ui.ts`                      | `composables/ui.svelte.ts` (actions, see below)                                     |
| `composables/useX.ts`                    | `composables/useX.svelte.ts`                                                        |
| `i18n/index.ts`                          | `i18n/index.svelte.ts`                                                              |
| `models/*`, `utils/*`, `services/api.ts` | already copied verbatim — import as `@/models/...`, `@/utils/...`, `@/services/api` |

Routing already exists under `web2/src/routes/` (thin wrappers importing `@/pages/*`). Pages are plain components; route params come from `$app/state`.

## Core APIs (already implemented — use these, do not reinvent)

```ts
// i18n (replaces useI18n / $t / $d)
import { t, te, d, n, getLocale, setLocale, SUPPORTED_LOCALES, SIGNING_LOCALES } from "@/i18n/index.svelte";
t("common.save"); t("x.y", { name });        // template: {t("common.save")}
d(dateValue, "short" | "long");              // replaces $d / d()
getLocale(); setLocale("ru");                // replaces locale.value

// HTTP (unchanged API)
import { apiGet, apiPost, apiPut, apiPatch, apiDelete } from "@/services/api";

// Auth (unchanged except setAuthRouter → setAuthNavigate, already wired in root layout)
import { logout, getAuthHeaders, fetchWithAuth, clearAdminCache } from "@/utils/auth";

// Current user / theme
import { useCurrentUser } from "@/composables/useCurrentUser.svelte";
const currentUser = useCurrentUser();        // currentUser.userData, currentUser.isAdmin (reactive getters), loadUserData(), clearUser()
import { useTheme } from "@/composables/useTheme.svelte";

// Conditions / formulas — Refs became getter functions:
import { useConditions } from "@/composables/useConditions.svelte";
const { fieldStates, ... } = useConditions(() => fields, () => formData);
import { useFormulas } from "@/composables/useFormulas.svelte";   // call during component init only

// UI helpers (Vue composables became Svelte actions)
import { clickOutside, escapeKey, focusTrap, portal, createDropdown } from "@/composables/ui.svelte";
// <div use:clickOutside={() => close()}>  <div use:focusTrap={isOpen}>  <div use:portal>

// Icons — SvgIcon is NOT global anymore, import it explicitly:
import SvgIcon from "@/components/SvgIcon.svelte";
// <SvgIcon name="x" class="h-5 w-5" />
```

Heroicons are replaced by sprite icons: `CheckCircleIcon→"check-circle"`, `XCircleIcon→"x-circle"`, `UserPlusIcon→"user-plus"`, `PencilIcon→"pencil"`, `TrashIcon→"trash-x"`, `UsersIcon→"users"` (via `SvgIcon`).

## Syntax conversion table

| Vue                                              | Svelte 5                                                                                                                           |
| ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| `defineProps<Props>()` + `withDefaults`          | `let { x, y = 5, ...rest }: Props = $props();`                                                                                     |
| `modelValue` prop + `update:modelValue`          | prop `value = $bindable()` (usage: `bind:value={...}`)                                                                             |
| named `v-model:foo`                              | prop `foo = $bindable()` (usage: `bind:foo`)                                                                                       |
| `emit("save", x)` / `@save`                      | callback prop `onSave?: (x: T) => void`; call `onSave?.(x)` (camelCase with `on` prefix; `select-submitter` → `onSelectSubmitter`) |
| default `<slot />`                               | `children?: Snippet` prop + `{@render children?.()}`                                                                               |
| named `<slot name="header" />`                   | `header?: Snippet` prop + `{@render header?.()}`                                                                                   |
| scoped slot `<slot :item="item" />`              | `row?: Snippet<[Item]>` + `{@render row?.(item)}`                                                                                  |
| `$slots.footer` check                            | `{#if footer}`                                                                                                                     |
| `ref(x)` / `reactive(o)`                         | `let x = $state(x)` / `const o = $state({...})`                                                                                    |
| `computed(() => …)`                              | `const v = $derived(…)` or `$derived.by(() => {…})`                                                                                |
| `watch(src, cb)` / `watchEffect`                 | `$effect(() => {…})`; wrap writes to signals you also read in `untrack(() => …)`                                                   |
| `onMounted` / `onBeforeUnmount`                  | `onMount(() => {… return cleanup})` or `$effect`                                                                                   |
| `nextTick()`                                     | `await tick()` (from "svelte")                                                                                                     |
| template ref `ref="el"`                          | `let el = $state<HTMLElement                                                                                                       | null>(null)`+`bind:this={el}` |
| `defineExpose({ fn })`                           | `export function fn() {…}` in the component script; parent: `bind:this={comp}` then `comp.fn()`                                    |
| `provide("key", v)` / `inject("key")`            | `setContext("key", v)` / `getContext("key")` — keep the same string keys; pass reactive values as getters/objects, not snapshots   |
| `v-if` / `v-else-if` / `v-else`                  | `{#if}` / `{:else if}` / `{:else}`                                                                                                 |
| `v-for="(it, i) in xs" :key="it.id"`             | `{#each xs as it, i (it.id)}`                                                                                                      |
| `v-show="c"`                                     | `style:display={c ? null : "none"}`                                                                                                |
| `:class="{ a: c }"` / arrays                     | string ternaries inside `class="… {c ? 'a' : ''}"` (match existing exemplars)                                                      |
| `v-model` on `<input>`/`<select>`/`<textarea>`   | `bind:value`; checkbox → `bind:checked`; radio set → `bind:group`                                                                  |
| `v-model.number`                                 | manual `oninput` with `Number(...)` parse                                                                                          |
| `@click="f"` / `@click.stop` / `@submit.prevent` | `onclick={f}` / call `e.stopPropagation()` / `e.preventDefault()` in handler                                                       |
| `v-html="x"`                                     | `{@html x}`                                                                                                                        |
| `<Teleport to="body">`                           | `use:portal` action                                                                                                                |
| `<Transition name="modal">` (fade 0.2s)          | `transition:fade={{ duration: 200 }}`                                                                                              |
| `<Transition name="drawer-expand">`              | `transition:slide` with matching duration                                                                                          |
| `contenteditable` two-way                        | `bind:textContent` / `bind:innerHTML` (element needs `contenteditable`)                                                            |
| `useRoute().params.id`                           | `page.params.id` (`import { page } from "$app/state"`)                                                                             |
| `useRoute().query.x`                             | `page.url.searchParams.get("x")`                                                                                                   |
| `useRoute().path`                                | `page.url.pathname`                                                                                                                |
| `useRouter().push("/x")` / `.replace`            | `goto("/x")` / `goto("/x", { replaceState: true })` (`$app/navigation`)                                                            |
| `<RouterLink to="/x">`                           | `<a href="/x">`                                                                                                                    |
| `$t("k")` / `$d(v, "short")`                     | `{t("k")}` / `{d(v, "short")}`                                                                                                     |
| `useI18n().locale.value = "ru"`                  | `setLocale("ru")`                                                                                                                  |

Component instance typing for `bind:this`: `let comp: ReturnType<typeof MyComponent> | undefined = $state();`

## Context contract (provide/inject)

Vue injects Refs; to keep consumer code 1:1, contexts that carried a Ref keep a **Ref-like shape** with a `.value` accessor backed by `$state`:

```ts
// provider (e.g. pages/Edit.svelte)
let template = $state<Template>(...);
setContext("template", {
  get value() { return template; },
  set value(v: Template) { template = v; }
});
// consumer (e.g. components/template/Area.svelte)
const template = getContext<{ value: Template }>("template");
template.value.name // reads stay identical to Vue
```

Known context keys (keep these exact strings and shapes):

- `"template"` — Ref-like `{ value: Template }` (provided by Edit/View pages)
- `"save"` — plain function
- `"baseFetch"` — plain function
- `"selectedAreaRef"` — Ref-like (provided by Edit/View pages)
- `"webhookModalTrigger"`, `"apiKeyModalTrigger"` — Ref-like (provided by pages/Settings.svelte)
- `"activeTab"` — Ref-like, `"setActiveTab"` — function (provided by ui/Tabs.svelte)

## Rules

1. `<script lang="ts">` (runes are enforced project-wide). No `any` unless the Vue source had it.
2. Keep every Tailwind class, DOM structure, and text exactly as in the Vue source. Keep code comments.
3. Preserve prop names/defaults (except `modelValue` → `value`). Preserve emitted event semantics via callback props.
4. Props that are bound with `v-model` anywhere in the codebase must be `$bindable()`.
5. `bind:value` cannot be combined with a dynamic `type` on `<input>` — use `{value}` + `oninput` (see `ui/Input.svelte`).
6. Look at existing exemplars before writing: `ui/Button.svelte` (rest props, snippets), `ui/Input.svelte` (bindable), `ui/Modal.svelte` (portal, transitions, snippets, actions), `layouts/Sidebar.svelte` (page state, i18n, each loops).
7. Do NOT run bun/vite/vitest/svelte-check or git; do not create extra files/docs; write only the files assigned to you. Validation happens centrally.
8. `graphify-out/graph.json` does not exist in this repo — skip graphify entirely; read source files directly.
9. Never modify anything under `web/` (read-only source).
