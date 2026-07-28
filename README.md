# goSign

✍️ **Sign documents without stress**

A modern, full-featured document signing platform with multi-signer workflows, email notifications, and a comprehensive REST API. Built with Go and SvelteKit, goSign provides enterprise-grade capabilities for secure digital document signing.

## ✨ Key Features

### 🔐 Core Signing

- 🔐 Digital signatures with X.509 certificates (PKCS7/CMS, PAdES)
- ✅ Document verification with full certificate chain validation
- 🎨 Visual signature placement and customizable appearance
- 📜 Certificate management: generate, manage, revoke (CRL)
- 🔄 Automatic trust certificate updates every 12 hours

### 📜 Document Workflow

- 👥 Multi-signer workflow: sequential or parallel signing with state machine
- 📧 Email notifications: invitations, reminders, status updates
- 📱 SMS notifications (optional)
- ⏰ Configurable reminder scheduling
- 📊 Real-time submission and signer status tracking

### 🔑 API & Integration

- 🔑 JWT tokens and API keys with rate limiting
- 📚 OpenAPI (swag) annotations for API documentation
- 🔗 Webhook support for real-time event notifications
- 🖼️ Embedded signing via JavaScript SDK (iframe)
- 📦 Bulk operations: CSV/XLSX import for mass submissions
- 🧾 Signing links (direct signing without email)

### 🏢 Enterprise Features

- 🏢 Organizations and teams: multi-tenant management
- 👥 Role-based access control: Owner, Admin, Member, Viewer
- 🔐 Organization context in JWT tokens
- 📧 Team invitations via email
- 🗂️ Organization-scoped templates
- 🌐 Multilingual (i18n): 7 UI languages, 14 signing portal languages, RTL support
- 🧩 Conditional fields: show/hide fields based on dynamic conditions
- 🧮 Formula engine: dynamic field calculations
- 🎨 White-label branding: custom logos, colors, fonts, themes
- ✉️ Customizable email templates per organization

### 🛡️ Security

- 🔑 JWT access tokens (15 min) + refresh tokens (7 days)
- 🧾 Two-factor authentication (TOTP with QR codes)
- 🌐 OAuth integration: Google and GitHub
- ✅ Email verification and password reset
- 🔒 bcrypt password hashing
- 🚦 Rate limiting: 100 req/min standard, 10 req/min for sensitive endpoints

## 🛠️ Tech Stack

### ⚙️ Backend

- **Language**: Go 1.26+
- **Framework**: Fiber v3
- **Database**: PostgreSQL 14+ with JSONB
- **Cache**: Redis 6+
- **Migrations**: goose
- **Authentication**: JWT + API keys
- **Email**: SMTP (go-mail)
- **Storage**: Local filesystem, S3 (MinIO-compatible)
- **PDF**: digitorus/pdf (signing/verification), signintech/gopdf (creation)
- **Formula engine**: expr-lang/expr
- **Logging**: zerolog
- **API docs**: OpenAPI annotations (swag)

### 🖥️ Frontend

- **Framework**: Svelte 5 + SvelteKit 2
- **State management**: Svelte runes
- **Routing**: SvelteKit filesystem routing
- **Styling**: Tailwind CSS v4
- **Build tool**: Vite
- **Package manager**: Bun
- **i18n**: custom lightweight i18n module

## 🏗️ Architecture

The application ships as independent services, each with its own binary and Dockerfile:

- **`gosign-server`** (`cmd/server`, `docker/Dockerfile.server`) — stateless HTTP API (Fiber). Scales horizontally. Serves everything under `/v1` plus `/drive` (files), `/embed` (iframe page) and `/health`.
- **`gosign-worker`** (`cmd/worker`, `docker/Dockerfile.worker`) — background maintenance: Adobe trust list updates and GeoLite2 database downloads. Runs as a single instance; shares the `base` volume with the API, which picks up refreshed GeoLite2 files automatically. Communicates with the API only through the shared database and volume — no RPC layer is needed.
- **`migrate`** (`docker/Dockerfile.migrate`) — self-contained migration image: goose + SQL migrations baked in. Runs once and exits.
- **`frontend`** (`docker/Dockerfile.frontend`) — SvelteKit SPA served by unprivileged nginx.

In the Docker Compose stack the services are split across two domains behind the `nginx` edge proxy:

- **`api.<domain>`** → API server (`/v1/...`, `/drive/...`, `/embed/{slug}`)
- **`app.<domain>`** → web app SPA

Domains are configured via `API_DOMAIN` / `APP_DOMAIN` in `.env` (defaults `api.localhost` / `app.localhost` work in browsers out of the box). The API allows CORS from the app origin, and `GOSIGN_APP_URL` tells the backend where the web app lives (signing links in emails, OAuth redirects, the `/embed` iframe). A future admin panel and marketing site can be added as separate services on their own domains without touching the API.

Database migrations are **not** embedded in the server: they are applied externally — locally via `./scripts/migration`, in Docker via the `migrate` service. Both server and worker verify at startup that the schema is initialized and exit with a clear error otherwise.

## 🗺️ Project Structure

```
goSign/
├── cmd/
│   ├── server/              # API server entrypoint
│   └── worker/              # Background worker entrypoint
├── exp/
│   ├── cert/                # Experiments: certificate utilities
│   └── pdf-cert/            # Experiments: PDF certificate utilities
├── internal/
│   ├── config/              # Configuration (env vars)
│   ├── server/              # HTTP API bootstrap
│   ├── worker/              # Background worker bootstrap and tasks
│   ├── handlers/
│   │   ├── api/             # REST API v1 handlers
│   │   └── public/          # Public and auth endpoints
│   ├── middleware/          # JWT, rate limiting, CORS
│   ├── models/              # Data models
│   ├── queries/             # Database repositories
│   ├── routes/              # Route registration
│   ├── services/            # Business logic
│   │   ├── submission/      # Multi-signer workflow state machine
│   │   ├── field/           # Field validation
│   │   └── formula/         # Formula evaluation
│   └── trust/               # Trust certificate management
├── pkg/
│   ├── pdf/
│   │   ├── sign/            # Digital signing
│   │   ├── verify/          # Signature verification
│   │   ├── fill/            # PDF form filling
│   │   └── revocation/      # CRL management
│   ├── notification/        # Email/SMS service
│   ├── webhook/             # Webhook dispatcher
│   ├── storage/             # Storage abstraction (local, S3)
│   ├── security/
│   │   ├── cert/            # Certificate operations
│   │   └── password/        # Hashing and validation
│   ├── appdir/              # Application data directories
│   ├── geolocation/         # GeoIP lookups
│   ├── logging/             # Logger setup
│   └── utils/               # Helper functions
├── web/
│   ├── private/             # Active frontend application (Svelte 5 + SvelteKit)
│   │   └── src/
│   │       ├── lib/
│   │       │   ├── components/ # Reusable UI and domain components
│   │       │   ├── composables/ # Runes-based shared logic
│   │       │   ├── i18n/      # Translations (7 UI languages)
│   │       │   ├── layouts/   # Application layouts
│   │       │   ├── models/    # TypeScript interfaces
│   │       │   └── pages/     # Page components
│   │       └── routes/        # SvelteKit filesystem routes
├── migrations/              # SQL migrations (goose, applied externally)
├── fixtures/                # Test/development data
├── docker/                  # Dockerfiles and nginx config
├── compose.yaml             # Docker Compose (all services)
└── scripts/                 # Utility scripts
```

## 🚀 Installation

### ✅ Prerequisites

- Go 1.26+
- PostgreSQL 14+
- Redis 6+
- Bun (or Node.js 18+ as alternative)
- `pdftoppm` from **poppler-utils** — required for PDF preview generation when creating templates from PDF files


| OS              | Package         | Install                          |
| --------------- | --------------- | -------------------------------- |
| Debian / Ubuntu | `poppler-utils` | `sudo apt install poppler-utils` |
| RHEL / Fedora   | `poppler-utils` | `sudo dnf install poppler-utils` |
| Alpine          | `poppler-utils` | `apk add poppler-utils`          |
| Arch            | `poppler`       | `pacman -S poppler`              |
| macOS           | `poppler`       | `brew install poppler`           |


### Backend Setup

1. Clone the repository:

```bash
git clone https://github.com/shurco/gosign.git
cd gosign
```

2. Install Go dependencies:

```bash
go mod download
```

3. Configure environment variables (see `.env.example`):

```bash
cp .env.example .env
# Set GOSIGN_JWT_SECRET (required); adjust GOSIGN_POSTGRES_URL, GOSIGN_REDIS_ADDRESS if needed
```

4. Run database migrations:

```bash
./scripts/migration up
```

5. (Optional) Load development fixtures with test data:

```bash
./scripts/migration dev up
```

Test users created by fixtures:

- **Admin**: `admin@gosign.local` / `admin123`
- **User 1**: `user1@gosign.local` / `user123`
- **User 2**: `user2@gosign.local` / `user234`

### Frontend Setup

```bash
cd web/private
bun install
bun run dev
```

### Common commands

```bash
go build -o bin/gosign-server ./cmd/server   # build the API server
go build -o bin/gosign-worker ./cmd/worker   # build the background worker
go run ./cmd/server                          # start the API server locally
go run ./cmd/worker                          # start the background worker
go test ./...                                # Go tests (DB tests auto-skip without test DB)
go vet ./...                                 # go vet
golangci-lint run ./...                      # linter

cd web/private && bun run dev                   # frontend dev server
cd web/private && bun run check                 # Svelte typecheck
cd web/private && bun run test                  # Vitest suite
cd web/private && bun run lint                  # ESLint
```

## 🧭 Usage

### ▶️ Starting the Application

```bash
go run ./cmd/server           # HTTP API on http://localhost:8088
go run ./cmd/worker           # background maintenance tasks
cd web/private && bun run dev # frontend dev server on http://localhost:5173
```

The API server exposes:


| Interface    | URL                             |
| ------------ | ------------------------------- |
| REST API     | `http://localhost:8088/v1/` |
| Health check | `http://localhost:8088/health`  |


### 🔗 API Endpoints

#### Authentication (`/v1/auth`)


| Method | Path                    | Description                         |
| ------ | ----------------------- | ----------------------------------- |
| POST   | `/v1/auth/signup`          | Register new user                   |
| POST   | `/v1/auth/signin`          | Login (returns JWT + refresh token) |
| POST   | `/v1/auth/refresh`         | Refresh access token                |
| POST   | `/v1/auth/signout`         | Logout                              |
| GET    | `/v1/auth/verify-email`    | Verify email address                |
| POST   | `/v1/auth/password/forgot` | Request password reset              |
| POST   | `/v1/auth/password/reset`  | Reset password                      |
| POST   | `/v1/auth/2fa/enable`      | Enable 2FA                          |
| POST   | `/v1/auth/2fa/verify`      | Verify 2FA code                     |
| POST   | `/v1/auth/2fa/disable`     | Disable 2FA                         |
| GET    | `/v1/auth/oauth/google`    | Google OAuth                        |
| GET    | `/v1/auth/oauth/github`    | GitHub OAuth                        |


#### Public


| Method | Path                  | Description                        |
| ------ | --------------------- | ---------------------------------- |
| POST   | `/v1/verify/pdf`         | Verify signed document             |
| GET    | `/s/:slug`            | Submitter signing portal (SPA)     |
| GET    | `/embed/:slug`        | Embeddable signing page (iframe)   |
| GET    | `/embed/:slug/config` | Embed configuration                |
| GET    | `/health`             | Health check                       |


#### API v1 (requires JWT or API key)

**📝 Submissions**


| Method | Path                       | Description               |
| ------ | -------------------------- | ------------------------- |
| GET    | `/v1/submissions`      | List submissions          |
| POST   | `/v1/submissions`      | Create submission         |
| GET    | `/v1/submissions/:id`  | Get submission            |
| PUT    | `/v1/submissions/:id`  | Update submission         |
| DELETE | `/v1/submissions/:id`  | Delete submission         |
| POST   | `/v1/submissions/send` | Send to signers           |
| POST   | `/v1/submissions/bulk` | Bulk import from CSV/XLSX |


**👤 Submitters**


| Method | Path                          | Description       |
| ------ | ----------------------------- | ----------------- |
| GET    | `/v1/submitters`          | List submitters   |
| GET    | `/v1/submitters/:id`      | Get submitter     |
| POST   | `/v1/submitters/resend`   | Resend invitation |
| POST   | `/v1/submitters/complete` | Complete signing  |
| POST   | `/v1/submitters/decline`  | Decline signing   |


**📄 Templates**


| Method | Path                                        | Description         |
| ------ | ------------------------------------------- | ------------------- |
| GET    | `/v1/templates`                         | List templates      |
| POST   | `/v1/templates`                         | Create template     |
| GET    | `/v1/templates/:id`                     | Get template        |
| PUT    | `/v1/templates/:id`                     | Update template     |
| DELETE | `/v1/templates/:id`                     | Delete template     |
| POST   | `/v1/templates/clone`                   | Clone template      |
| POST   | `/v1/templates/from-file`               | Create from PDF     |
| POST   | `/v1/templates/formulas/validate`       | Validate formula    |
| POST   | `/v1/templates/:id/conditions/validate` | Validate conditions |


**🔗 Signing Links** (direct signing without email)


| Method | Path                                            | Description                 |
| ------ | ----------------------------------------------- | --------------------------- |
| GET    | `/v1/signing-links`                         | List signing links          |
| POST   | `/v1/signing-links`                         | Create signing link         |
| GET    | `/v1/signing-links/:submission_id`          | Get signing link            |
| GET    | `/v1/signing-links/:submission_id/document` | Download completed document |


**🏢 Company**


| Method | Path                               | Description                              |
| ------ | ---------------------------------- | ---------------------------------------- |
| GET    | `/v1/company`                      | List organizations                       |
| POST   | `/v1/company`                      | Create organization                      |
| GET    | `/v1/company/:id`                  | Get organization                         |
| PUT    | `/v1/company/:id`                  | Update organization                      |
| DELETE | `/v1/company/:id`                  | Delete organization                      |
| POST   | `/v1/company/:id/switch`           | Switch organization context (admin only) |
| POST   | `/v1/company/switch`               | Exit organization context                |


**👥 Company Members**


| Method | Path                                                | Description        |
| ------ | --------------------------------------------------- | ------------------ |
| GET    | `/v1/company/:id/members`                           | List members       |
| POST   | `/v1/company/:id/members/invite`                    | Invite member      |
| PUT    | `/v1/company/:id/members/:member_id/role`           | Update member role |
| DELETE | `/v1/company/:id/members/:member_id`                | Remove member      |


**✉️ Invitations**


| Method | Path                                                   | Description               |
| ------ | ------------------------------------------------------ | ------------------------- |
| GET    | `/v1/company/:id/invitations`                          | List invitations          |
| DELETE | `/v1/company/:id/invitations/:invitation_id`           | Revoke invitation         |
| GET    | `/v1/invitations/:token`                               | Invitation details        |
| POST   | `/v1/invitations/:token/accept`                        | Accept invitation         |


**🔑 API Keys**


| Method | Path                          | Description                            |
| ------ | ----------------------------- | -------------------------------------- |
| GET    | `/v1/settings/api`            | List API keys                          |
| POST   | `/v1/settings/api`            | Create API key                         |
| PATCH  | `/v1/settings/api/:id`        | Enable/disable key (`{enabled: bool}`) |
| DELETE | `/v1/settings/api/:id`        | Revoke key                             |


**🪝 Webhooks**


| Method | Path                           | Description    |
| ------ | ------------------------------ | -------------- |
| GET    | `/v1/settings/webhooks`        | List webhooks  |
| POST   | `/v1/settings/webhooks`          | Create webhook |
| PUT    | `/v1/settings/webhooks/:id`      | Update webhook |
| DELETE | `/v1/settings/webhooks/:id`      | Delete webhook |


**⚙️ Settings**


| Method | Path                        | Description           |
| ------ | --------------------------- | --------------------- |
| GET    | `/v1/settings`              | Get settings          |
| PUT    | `/v1/settings/email`        | Update email config   |
| PUT    | `/v1/settings/sms`          | Update SMS config     |
| PUT    | `/v1/settings/storage`      | Update storage config |
| PUT    | `/v1/settings/geolocation`  | Update GeoIP config   |


**🎨 Branding & i18n**


| Method | Path                                | Description            |
| ------ | ----------------------------------- | ---------------------- |
| GET    | `/v1/settings/branding`             | Get branding settings  |
| PUT    | `/v1/settings/branding`             | Update branding        |
| GET    | `/v1/settings/i18n/locales`           | List available locales |
| PUT    | `/v1/settings/i18n/user/locale`       | Update user locale     |
| PUT    | `/v1/settings/i18n/account/locale`    | Update account locale  |


**✉️ Email Templates**


| Method | Path                                    | Description     |
| ------ | --------------------------------------- | --------------- |
| GET    | `/v1/settings/email/templates`          | List templates  |
| POST   | `/v1/settings/email/templates`          | Create template |
| PUT    | `/v1/settings/email/templates/:id`      | Update template |


**📊 Events & Stats**


| Method | Path             | Description             |
| ------ | ---------------- | ----------------------- |
| GET    | `/v1/events` | List events (audit log) |
| GET    | `/v1/stats`  | Get statistics          |


> Authoritative route list: `internal/routes/api_routes.go` and handler `RegisterRoutes` methods · OpenAPI generation: [docs/SWAGGER.md](docs/SWAGGER.md)

## Configuration

All configuration is via environment variables with the `GOSIGN_` prefix. Infrastructure settings are read at startup; application settings (SMTP, storage, branding) are managed in the database via Admin UI.


| Variable                      | Default          | Description                             |
| ----------------------------- | ---------------- | --------------------------------------- |
| `GOSIGN_JWT_SECRET`           | — (**required**) | Secret for signing JWT tokens           |
| `GOSIGN_HTTP_ADDR`            | `0.0.0.0:8088`   | HTTP server address                     |
| `GOSIGN_DEV_MODE`             | `false`          | Development mode (relaxed CORS)         |
| `GOSIGN_POSTGRES_URL`         | —                | PostgreSQL connection URL               |
| `GOSIGN_REDIS_ADDRESS`        | `localhost:6379` | Redis address                           |
| `GOSIGN_REDIS_PASSWORD`       | —                | Redis password                          |
| `GOSIGN_CORS_ALLOWED_ORIGINS` | —                | Comma-separated allowed origins         |
| `GOSIGN_APP_URL`              | —                | Public URL of the web app (emails, OAuth redirects, /embed iframe); empty = same origin |
| `GOSIGN_DATA_DIR`             | executable dir   | Data directory for `lc_*` folders       |


## Development

### Running Tests

```bash
# All tests (DB-backed tests auto-skip when the test DB is down)
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./pkg/pdf/sign/...
```

Database-backed tests need a dedicated PostgreSQL at `localhost:5453` — see [docs/TESTING.md](docs/TESTING.md).

### Building for Production

```bash
# Backend
go build -o bin/gosign-server ./cmd/server
go build -o bin/gosign-worker ./cmd/worker

# Frontend
cd web/private && bun run build
```

### Docker

```bash
cp .env.example .env   # set GOSIGN_JWT_SECRET (and domains if needed)
docker compose up -d
```

By default the stack is available at `http://app.localhost` (web app) and `http://api.localhost` (API). Override with `API_DOMAIN` / `APP_DOMAIN` / `PUBLIC_SCHEME` in `.env`.

Compose uses prebuilt images from GitHub Container Registry (`ghcr.io/shurco/gosign-*:latest`, published by the `docker` workflow on each release):

- `migrate` — `ghcr.io/shurco/gosign-migrate` — self-contained goose image with `migrations/` baked in; runs and exits (API waits for it)
- `gosign` — `ghcr.io/shurco/gosign-server` — HTTP API
- `worker` — `ghcr.io/shurco/gosign-worker` — background tasks, shares the `base` volume with the API
- `frontend` — `ghcr.io/shurco/gosign-frontend` — SvelteKit SPA served by unprivileged nginx; the API base URL is injected at container start via `GOSIGN_API_URL` (no rebuild needed per domain)
- `nginx` — edge proxy on port 80: `api.<domain>` -> backend, `app.<domain>` -> frontend
- `postgres`, `redis`, `minio` — infrastructure

To build the images locally instead of pulling from ghcr (Dockerfiles live in `docker/`):

```bash
docker build -t ghcr.io/shurco/gosign-server:latest   -f docker/Dockerfile.server .
docker build -t ghcr.io/shurco/gosign-worker:latest   -f docker/Dockerfile.worker .
docker build -t ghcr.io/shurco/gosign-migrate:latest  -f docker/Dockerfile.migrate .
docker build -t ghcr.io/shurco/gosign-frontend:latest -f docker/Dockerfile.frontend .
```

Re-run migrations manually:

```bash
docker compose run --rm migrate                       # up
GOOSE_COMMAND=status docker compose run --rm migrate  # status
```

## Scripts

Located in `scripts/`:


| Script                  | Description                                 |
| ----------------------- | ------------------------------------------- |
| `migration`             | Database migration management (wraps goose) |
| `migration dev up/down` | Load/unload development fixtures            |
| `clean`                 | Clean build artifacts                       |
| `key`                   | Generate cryptographic keys                 |
| `models`                | Generate data models                        |
| `tools`                 | Development tools                           |


Migration commands:

```bash
./scripts/migration up        # Apply all pending migrations
./scripts/migration up1       # Apply one migration
./scripts/migration down      # Roll back all migrations
./scripts/migration down1     # Roll back one migration
./scripts/migration status    # Show migration status
./scripts/migration create    # Create new migration file
```

## Documentation


| Document                                                   | Description                           |
| ---------------------------------------------------------- | ------------------------------------- |
| [docs/API_AUTHENTICATION.md](docs/API_AUTHENTICATION.md)   | JWT and API key authentication guide  |
| [docs/EMBEDDED_SIGNING.md](docs/EMBEDDED_SIGNING.md)       | JavaScript SDK for iframe integration |
| [docs/SWAGGER.md](docs/SWAGGER.md)                         | Swagger documentation generation      |
| [docs/TESTING.md](docs/TESTING.md)                         | Testing strategy and guidelines       |
| [docs/MULTILINGUAL.md](docs/MULTILINGUAL.md)               | i18n and signing portal languages     |
| [docs/CONDITIONAL_FIELDS.md](docs/CONDITIONAL_FIELDS.md)   | Dynamic show/hide field logic         |
| [docs/FORMULAS.md](docs/FORMULAS.md)                       | Formula engine and builder            |
| [docs/WHITE_LABEL.md](docs/WHITE_LABEL.md)                 | White-label branding and themes       |
| [docs/FRONTEND_COMPONENTS.md](docs/FRONTEND_COMPONENTS.md) | Frontend component architecture       |


## Roadmap

- [ ] GCS and Azure blob storage
- [ ] Advanced analytics dashboard
- [ ] External CA integration
- [ ] Mobile application
- [ ] eIDAS e-signature standards

## License

Licensed under the [GNU General Public License v3.0](LICENSE).

- You may use, modify, and distribute this software
- You must preserve the GPL-3.0 license when distributing
- You cannot use this software in proprietary (closed-source) applications

## Contributing

Contributions are welcome. Please open an issue or pull request on [GitHub](https://github.com/shurco/gosign).
