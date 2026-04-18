# web/ — AGENTS.md

Vue 3 + Vite + Bun frontend for goSign. See the repo-level [`../AGENTS.md`](../AGENTS.md) for conventions that apply everywhere.

## Stack

- **Vue** 3.5 (Composition API, `<script setup lang="ts">`)
- **TypeScript** 5.9 in strict mode (`tsconfig.json`)
- **Vite** 8 build tool
- **Tailwind CSS** v4 (+ `prettier-plugin-tailwindcss`)
- **Pinia** 3 stores
- **vue-router** 5 with lazy-loaded layouts
- **vue-i18n** 11 (7 UI locales, 14 signing portal locales, RTL)
- **Vitest** 4 + `@vue/test-utils` + `jsdom` for unit tests
- **expr-eval** for client-side formula evaluation (see `composables/useFormulas.ts`)
- Package manager: **Bun** (`bun.lockb` is authoritative)

## Directory layout (`src/`)

| Path              | Purpose                                                                                                                  |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `App.vue`, `main.ts` | App bootstrap + router/i18n/Pinia install                                                                             |
| `router.ts`       | Route definitions, auth guards, lazy layout loading                                                                      |
| `layouts/`        | `Blank`, `Main`, `Profile`, `Sidebar`, `SettingsSidebar`                                                                 |
| `pages/`          | Route views (including nested `settings/`)                                                                               |
| `components/`     | Reusable widgets grouped by domain: `common/`, `field/`, `organization/`, `pdf/`, `signing/`, `template/`, `themes/`, `ui/` |
| `composables/`    | Reactive helpers: `useConditions`, `useFormulas`, `useTheme`, `useCurrentUser`, `ui`                                     |
| `i18n/`           | `index.ts` + locale files                                                                                                |
| `services/api.ts` | Typed HTTP client backed by the browser `fetch` API                                                                      |
| `stores/`         | Pinia stores (add new ones here)                                                                                         |
| `models/`         | Shared TypeScript interfaces mirroring server DTOs                                                                       |
| `utils/`          | Pure helpers                                                                                                             |
| `assets/`         | Static assets                                                                                                            |
| `test/`           | Vitest setup (`setup.ts`), shared fixtures                                                                               |

## Commands (run inside `web/`)

- `bun install` — install deps
- `bun run dev` — Vite dev server
- `bun run build` — `typecheck` + `vite build` (production)
- `bun run typecheck` — `vue-tsc --noEmit`
- `bun run lint` / `bun run lint:fix` — ESLint (flat config)
- `bun run format` — Prettier on `src/**/*.{ts,tsx,vue,css,scss}`
- `bun x vitest run` — full test suite
- `bun x vitest --ui` — interactive UI

Top-level `Makefile` wraps these: `make web-dev`, `make web-build`, `make web-test`, `make web-typecheck`, `make web-lint`.

## Coding rules (Vue / TS)

1. **No `any`** unless wrapping an untyped third-party API and explaining why.
2. Components use `<script setup lang="ts">`, `defineProps<…>()`, `defineEmits<…>()`.
3. Keep components below ~300 lines; extract into children or composables when they grow.
4. State for >1 component ⇒ Pinia store, not prop-drilling.
5. Use path alias `@/` for imports inside `src/` (configured in `tsconfig.json` + `vite.config.ts`).
6. ESLint rule `no-relative-import-paths` is on; fix, don't disable.
7. **Reactivity gotcha**: when a `computed` must react to properties of a `ref<Record>`, touch each key inside the computed (see `composables/useFormulas.ts` for the canonical pattern).
8. Input validation lives next to the component (`composables/useConditions.ts`, local utils), HTTP errors are returned by `services/api.ts` as typed `{ ok: false, error }` objects.

## Adding a feature

1. Describe the interaction in the matching page under `pages/` (or a new route in `router.ts`).
2. Extract data-fetching into `services/api.ts` (preserve strict types).
3. Put shared state in `stores/`.
4. Extract logic into a composable when it's reused or when it warrants isolated tests.
5. Add tests under `src/**/__tests__/*.spec.ts`.
6. Update the route guard / menu entries if the feature needs new permissions.

## Tests

- All tests run in jsdom; use `await nextTick()` after reactive mutations when asserting derived values.
- Prefer `@vue/test-utils` `mount` for component-level tests, plain `describe`/`it` for composables.
- Snapshot tests are discouraged unless the snapshot is a stable structural contract.
