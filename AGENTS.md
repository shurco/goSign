# AGENTS.md — goSign

> Purpose: give coding agents (and humans) the context they need to make correct, well-scoped changes to this repository. Keep this file short; link to specialised `AGENTS.md` files inside submodules when the context differs.

## Project at a glance

- **What it is**: an enterprise-grade document signing platform (PKCS#7/PAdES) with multi-signer workflows, webhooks, organisations, bulk imports, embedded signing, and a REST API.
- **Backend**: Go **1.26+**, [Fiber **v3**](https://docs.gofiber.io/), `pgx/v5`, goose migrations, MinIO storage, Redis, zerolog. Entry point: `cmd/goSign/main.go` → `internal.App()` in `internal/app.go`.
- **Frontend**: Vue 3.5 + TypeScript 5.9 + Vite 8 + Tailwind v4 + Pinia 3 + vue-router 5 + vue-i18n 11. Package manager: **Bun**. Source: `web/src/`.
- **Binary tools**: `cmd/cert/` (CA/cert generator), `cmd/pdf-cert/` (PDF certificate printer).

## Repository layout (top-level)

| Path            | Purpose                                                                               |
| --------------- | ------------------------------------------------------------------------------------- |
| `cmd/`          | Entry points: `goSign`, `cert`, `pdf-cert`                                            |
| `internal/`     | Application code (handlers, services, queries, middleware, worker). **Not importable by external packages.** |
| `pkg/`          | Reusable libraries: `pdf`, `security`, `storage`, `notification`, `webhook`, `geolocation`, `logging`, `utils`, `appdir` |
| `web/`          | Vue 3 frontend (Vite + Bun)                                                           |
| `migrations/`   | Numbered goose SQL migrations, embedded via `embed.go`                                |
| `docker/`       | `Dockerfile.backend`, `Dockerfile.frontend`, `nginx.conf`                             |
| `docs/`         | Topic-oriented Markdown (API auth, testing, embedded signing, i18n, white-label, …) |
| `scripts/`      | Bash helpers (migrations, key/cert tooling, cleanup)                                  |
| `fixtures/`     | Test PDFs, signed samples, uploads                                                    |

Deeper `AGENTS.md` files: [`internal/AGENTS.md`](internal/AGENTS.md), [`pkg/AGENTS.md`](pkg/AGENTS.md), [`web/AGENTS.md`](web/AGENTS.md).

## Commands you'll use most

Run from repo root unless noted. A top-level `Makefile` wraps these; run `make help` to list targets.

- `make build` — build `bin/goSign`
- `make run` — start the server (needs Postgres + Redis)
- `make test` — `go test -short -race -count=1 ./...` (skips network-dependent tests)
- `make test-all` — run everything, including tests that call external TSAs / real DBs
- `make lint` — `golangci-lint run ./...` (config in `.golangci.yml`)
- `make web-test` — run the Vitest suite
- `make check` — `vet` + Go tests + frontend typecheck + Vitest
- `make ci` — lint + `check`

Direct Bun commands (inside `web/`): `bun run dev`, `bun run build`, `bun run typecheck`, `bun run lint`, `bun x vitest run`.

## Configuration (env vars, `GOSIGN_` prefix)

Mandatory: `GOSIGN_JWT_SECRET`. Common optional: `GOSIGN_HTTP_ADDR` (default `0.0.0.0:8088`), `GOSIGN_DEV_MODE`, `GOSIGN_POSTGRES_URL`, `GOSIGN_REDIS_ADDRESS`, `GOSIGN_REDIS_PASSWORD`, `GOSIGN_CORS_ALLOWED_ORIGINS`. See [`.env.example`](.env.example) and [`internal/config/config.go`](internal/config/config.go).

## Coding rules (applies to every change)

1. **KISS / DRY / Single Responsibility.** Extract helpers when the same logic appears more than twice; never pass redundant fields that can be derived from the others.
2. **Go style** — run `gofmt`/`goimports`. Error messages start with a lowercase letter and end without punctuation. Wrap errors with `fmt.Errorf("...: %w", err)`. Use `errors.Is`/`errors.As` in tests.
3. **No dead code / legacy shims.** Remove unused exports and update every call site in the same change.
4. **Fiber handlers** return `webutil.Response(c, status, message, data)` for consistent envelopes. Non-success paths should log with `logging.Log.Err(err)` before returning 5xx.
5. **Frontend**: `<script setup lang="ts">`, composition API, **no `any`** unless justified; route definitions live in `web/src/router.ts`; state lives in `web/src/stores/` (Pinia); UI in `web/src/components/`.
6. **Tests are required** for new business logic. Network-dependent tests must guard with `if testing.Short() { t.Skip(...) }`.
7. **Database**: new schema changes go into a new migration file in `migrations/` (never edit applied migrations). Queries live in `internal/queries/`.
8. **Do not** commit secrets, `.env`, or anything under `bin/`.

## Working on a task — quick checklist

- [ ] Read the relevant submodule `AGENTS.md`.
- [ ] Check existing tests for the area (`rg -l _test.go internal/... pkg/...`).
- [ ] Implement the change following the rules above.
- [ ] Run `make check` (or the narrowest equivalent) and confirm green.
- [ ] Update docs under `docs/` and `AGENTS.md` files if behaviour or structure changed.
- [ ] Keep commits focused; write messages in imperative mood.

## References

- [`README.md`](README.md) — full feature tour and deployment guide.
- [`docs/TESTING.md`](docs/TESTING.md) — test strategy and fixtures.
- [`docs/API_AUTHENTICATION.md`](docs/API_AUTHENTICATION.md) — JWT & API key flows.
- [`docs/SWAGGER.md`](docs/SWAGGER.md) — how to access and update the OpenAPI spec.
