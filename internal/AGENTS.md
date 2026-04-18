# internal/ — AGENTS.md

Application code that must not be imported by external packages. See the repo-level [`../AGENTS.md`](../AGENTS.md) for project-wide conventions.

## Subpackages

| Package                   | Responsibility                                                                                                                    |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `app` (`app.go`)          | Bootstrap / dependency injection. Wires config, DB pool, Redis, storage, queries, services, handlers, and routes, and runs Fiber. |
| `assets/`                 | Embedded fonts and images (certificate rendering). `EnsureOnDisk` extracts them for runtime use.                                  |
| `config/`                 | `GOSIGN_`-prefixed environment loader; returns `*config.Config`. `JWT_SECRET` is required; CORS origins follow dev-mode defaults. |
| `handlers/api/`           | REST `v1` JSON handlers (auth-guarded). One file per resource.                                                                    |
| `handlers/public/`        | Publicly reachable endpoints: auth (`/auth/*`), OAuth, signing (`/sign/`), health, `verify`, submitter signing links.             |
| `middleware/`             | Fiber middleware: CORS, helmet, recover, request logger, JWT/API-key auth, org permission guard, rate limiters.                   |
| `models/`                 | Plain data types (DTOs) shared by queries, handlers, and services.                                                                |
| `queries/`                | `pgx/v5` repositories. `Init(pool)` must be called once; `DB` is the global facade.                                               |
| `routes/`                 | Wires handler instances onto Fiber routers (`ApiRoutes`, `SiteRoutes`, `NotFoundRoute`).                                          |
| `services/`               | Business-logic services: `apikey`, `completed_document`, `reminder`, plus `email/`, `field/`, `formula/`, `submission/`.          |
| `trust/`                  | Background updater for Adobe trust lists (runs every 12 h on startup).                                                            |
| `worker/`                 | Background worker + `tasks/reminders.go` for scheduled reminders.                                                                 |
| `testutil/`               | Test helpers: `pgtestdb` fixtures, JWT minting, authenticated requests.                                                           |

## Handler conventions

- Each handler is a struct (`type XHandler struct { … }`) constructed via `NewXHandler(deps…)` in `app.go` and registered in `routes/`.
- Always return via `webutil.Response(c, status, message, data)` for a uniform envelope.
- Validate request payloads with `go-playground/validator/v10`; return `fiber.StatusBadRequest` with a human-readable message on validation failures.
- For non-2xx paths that are caused by internal errors, log with `logging.Log.Err(err).Msg("...")` and return `Internal server error` (never leak internal details to the client).
- Protected endpoints sit behind `middleware.Protected()` (JWT or API key) and `APIRateLimiter()` / `StrictRateLimiter()` as appropriate.

## Queries (`internal/queries/`)

- Repositories receive a `*pgxpool.Pool`. Never hold on to a single `*pgx.Conn`.
- Prefer `pool.QueryRow` for one-row lookups with `Scan`; for collections return typed slices from `models/`.
- Wrap errors with context: `fmt.Errorf("queries: get user by email: %w", err)`.
- Every DB change must ship with (a) a new file in `migrations/`, (b) updated typed queries, and (c) tests where practical using `pgtestdb`.

## Services

- Services own orchestration logic (transactions, outbound calls, multi-repo updates). Keep them free of Fiber types.
- Services are constructed in `app.go`. When a dependency is genuinely optional, make the zero value safe (e.g. nil provider = "disabled").

## Adding a new endpoint

1. Define request/response types in `models/` (or a `_types.go` next to the handler).
2. Add the query helper(s) to `internal/queries/` and unit-test the SQL with `pgtestdb`.
3. Implement the handler in `handlers/api/` or `handlers/public/`.
4. Wire it in `app.go` (constructor) and `routes/` (route registration).
5. Add happy-path + error-path tests in `*_test.go` next to the handler.
6. Update `docs/SWAGGER.md` / `docs/API_AUTHENTICATION.md` if the contract changed.

## Tests

- Unit tests: `go test -short ./internal/...`.
- Integration: drop `-short`; requires Docker (`pgtestdb`) and `GOSIGN_*` env vars from `.env.example`.
