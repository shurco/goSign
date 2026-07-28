# AGENTS.md — web2 (frontend)

> Svelte 5 rewrite of the goSign frontend (ported 1:1 from the Vue 3 app in `web/`). Read the root [`AGENTS.md`](../AGENTS.md) first.

## Stack

- **Svelte 5** (runes only — `$state`, `$derived`, `$effect`, `$props`, `$bindable`) + **SvelteKit 2** in pure SPA mode (`ssr = false`, `@sveltejs/adapter-static` with `fallback: index.html`, output dir `dist/`).
- TypeScript strict (pinned to **6.x** — typescript-eslint does not support TS 7 yet), Tailwind CSS v4 (`@tailwindcss/vite`), Vitest 4, ESLint + Prettier.
- Package manager: **Bun**. SvelteKit/Vite/adapter config lives entirely in `vite.config.ts` (no `svelte.config.js`).

## Commands (run inside `web2/`)

- `bun run dev` — dev server with API proxy to `localhost:8088`
- `bun run build` — production build into `dist/`
- `bun run check` — `svelte-check` (type + template validation)
- `bun test` / `bun run test:unit` — Vitest (server project: node; client project: jsdom for `*.svelte.test.ts`)
- `bun run lint` / `bun run format` — Prettier + ESLint

## Layout

| Path                      | Purpose                                                                                                                                                                                         |
| ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `src/routes/`             | SvelteKit filesystem routes. Layout groups map to the old Vue layouts: `(blank)`, `(main)`, `(sidebar)`, `(settings)`. Route files are thin wrappers importing page components from `@/pages/`. |
| `src/lib/`                | Application code, aliased as `@/` (kept identical to the old `web/src` import style)                                                                                                            |
| `src/lib/components/`     | `ui/` primitives, `common/`, `field/` (+`inputs/`), `organization/`, `signing/`, `template/`, `themes/`, `SvgIcon.svelte`                                                                       |
| `src/lib/pages/`          | Page components (1:1 with old `web/src/pages/*.vue`)                                                                                                                                            |
| `src/lib/layouts/`        | `Blank`, `Main`, `Sidebar`, `SettingsSidebar`                                                                                                                                                   |
| `src/lib/composables/`    | Runes-based state modules (`*.svelte.ts`): `ui.svelte.ts` (actions: clickOutside, escapeKey, focusTrap, portal), `useCurrentUser`, `useTheme`, `useConditions`, `useFormulas`                   |
| `src/lib/i18n/`           | Custom lightweight i18n (`index.svelte.ts`): `t()`, `te()`, `d()`, `n()`, `setLocale()`; JSON locales in `locales/` (7 UI languages)                                                            |
| `src/lib/services/api.ts` | Typed fetch client (`apiGet/apiPost/...`), auto token refresh via `utils/auth.ts`                                                                                                               |
| `src/lib/utils/`          | `auth.ts` (JWT refresh flow), `guards.ts` (route guards used by `+layout.ts`/`+page.ts`), `time.ts`, `status.ts`, `file.ts`                                                                     |
| `src/lib/models/`         | Shared TS types (copied verbatim from the Vue app)                                                                                                                                              |
| `src/lib/assets/`         | `app.css` (Tailwind entry + theme), `svg/` sprite icons (via `vite-plugin-svg-icons`, `virtual:svg-icons-register`)                                                                             |
| `static/`                 | `favicon.ico`, `gosign-embed.js`                                                                                                                                                                |

## Rules

1. Runes mode is enforced project-wide (see `compilerOptions.runes` in `vite.config.ts`). No legacy Svelte syntax, no stores unless justified.
2. Components: props via typed `$props()`; two-way bindings via `$bindable()`; events are `onXxx` callback props; slots are snippets.
3. Auth guards live in `+layout.ts`/`+page.ts` load functions (`requireAuth`, `requireAdmin`, `redirectIfAuthenticated` from `@/utils/guards`). The app is client-only, so loads always run in the browser.
4. Cross-component state passed through Svelte context keeps the Vue-era Ref-like `{ value }` accessor shape (keys: `"template"`, `"save"`, `"baseFetch"`, `"selectedAreaRef"`, `"webhookModalTrigger"`, `"apiKeyModalTrigger"`).
5. i18n: use `t("key", { params })` from `@/i18n/index.svelte`; locale files must stay key-compatible across all 7 languages (covered by `src/lib/i18n/__tests__`).
6. Tests: colocated under `__tests__/`; files using runes must be named `*.svelte.test.ts` (they run in the jsdom Vitest project).
7. No `any` unless unavoidable; `bun run check` must stay green.
