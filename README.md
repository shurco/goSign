# goSign

✍️ **Sign documents without stress**

A modern, full-featured document signing platform with multi-signer workflows, email notifications, and a comprehensive REST API. Built with Go and Vue 3, goSign provides enterprise-grade capabilities for secure digital document signing.

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
- 📚 Swagger/OpenAPI interactive documentation
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
- 🔑 JWT access tokens (10 min) + refresh tokens (7 days)
- 🧾 Two-factor authentication (TOTP with QR codes)
- 🌐 OAuth integration: Google and GitHub
- ✅ Email verification and password reset
- 🔒 bcrypt password hashing
- 🚦 Rate limiting: 100 req/min standard, 10 req/min for sensitive endpoints

## 🛠️ Tech Stack

### ⚙️ Backend
- **Language**: Go 1.26+
- **Framework**: Fiber v2
- **Database**: PostgreSQL 14+ with JSONB
- **Cache**: Redis 6+
- **Migrations**: goose
- **Authentication**: JWT + API keys
- **Email**: SMTP (go-mail)
- **Storage**: Local filesystem, S3 (MinIO-compatible)
- **PDF**: digitorus/pdf (signing/verification), signintech/gopdf (creation)
- **Formula engine**: expr-lang/expr
- **Logging**: zerolog
- **API docs**: Swagger/OpenAPI

### 🖥️ Frontend
- **Framework**: Vue 3 + TypeScript (Composition API)
- **State management**: Pinia 3
- **Routing**: Vue Router 5
- **Styling**: Tailwind CSS v4
- **Build tool**: Vite
- **Package manager**: Bun
- **i18n**: vue-i18n

## 🗺️ Project Structure

```
goSign/
├── cmd/
│   ├── goSign/              # Main application (server entrypoint)
│   ├── cert/                # Certificate utilities
│   └── pdf-cert/            # PDF certificate utilities
├── internal/
│   ├── config/              # Configuration (env vars)
│   ├── handlers/
│   │   ├── api/             # REST API v1 handlers
│   │   └── public/          # Public and auth endpoints
│   ├── middleware/          # JWT, rate limiting, CORS
│   ├── models/              # Data models
│   ├── queries/             # Database repositories
│   ├── routes/              # Route registration
│   ├── services/            # Business logic
│   │   ├── submission/      # Multi-signer workflow state machine
│   │   ├── email/           # Email template rendering
│   │   ├── field/           # Field validation
│   │   └── formula/         # Formula evaluation
│   ├── trust/               # Trust certificate management
│   └── worker/              # Background task scheduler
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
├── web/                     # Frontend application (Vue 3)
│   └── src/
│       ├── components/
│       │   ├── ui/          # Reusable UI primitives (Button, Input, Modal, etc.)
│       │   ├── common/      # Generic components (FieldInput, FormModal, ResourceTable)
│       │   ├── field/       # Field-specific components (ConditionBuilder, FormulaBuilder)
│       │   ├── template/    # Document template components
│       │   ├── organization/# Organization management components
│       │   ├── signing/     # Signing portal components
│       │   └── themes/      # White-label theme components
│       ├── composables/     # Vue composables (useConditions, useFormulas, useTheme, useCurrentUser)
│       ├── i18n/            # Translations (7 languages)
│       ├── layouts/         # Page layouts
│       ├── models/          # TypeScript interfaces
│       ├── pages/           # Application pages
│       └── stores/          # Pinia stores
├── migrations/              # SQL migrations (goose)
├── fixtures/                # Test/development data
├── docker/                  # Docker configuration
│   └── core/                # Docker Compose for infrastructure
└── scripts/                 # Utility scripts
```

## 🚀 Installation

### ✅ Prerequisites

- Go 1.26+
- PostgreSQL 14+
- Redis 6+
- Bun (or Node.js 18+ as alternative)
- `pdftoppm` from **poppler-utils** — required for PDF preview generation when creating templates from PDF files

| OS | Package | Install |
|----|---------|---------|
| Debian / Ubuntu | `poppler-utils` | `sudo apt install poppler-utils` |
| RHEL / Fedora | `poppler-utils` | `sudo dnf install poppler-utils` |
| Alpine | `poppler-utils` | `apk add poppler-utils` |
| Arch | `poppler` | `pacman -S poppler` |
| macOS | `poppler` | `brew install poppler` |

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

3. Configure environment variables (see `cmd/goSign/.env.example`):
```bash
cp cmd/goSign/.env.example cmd/goSign/.env
# Edit: GOSIGN_POSTGRES_URL, GOSIGN_REDIS_ADDRESS, GOSIGN_REDIS_PASSWORD
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
cd web
bun install
bun run dev
```

## 🧭 Usage

### ▶️ Starting the Application

```bash
go run cmd/goSign/main.go serve
```

The server starts on `http://localhost:8088` by default:

| Interface | URL |
|-----------|-----|
| Public signing/verification | `http://localhost:8088/` |
| Admin panel | `http://localhost:8088/_/` |
| REST API | `http://localhost:8088/api/v1/` |
| Swagger UI | `http://localhost:8088/swagger/index.html` |

### 🔗 API Endpoints

#### Authentication (`/auth`)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/auth/signup` | Register new user |
| POST | `/auth/signin` | Login (returns JWT + refresh token) |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/signout` | Logout |
| GET | `/auth/verify-email` | Verify email address |
| POST | `/auth/password/forgot` | Request password reset |
| POST | `/auth/password/reset` | Reset password |
| POST | `/auth/2fa/enable` | Enable 2FA |
| POST | `/auth/2fa/verify` | Verify 2FA code |
| POST | `/auth/2fa/disable` | Disable 2FA |
| GET | `/auth/oauth/google` | Google OAuth |
| GET | `/auth/oauth/github` | GitHub OAuth |

#### Public

| Method | Path | Description |
|--------|------|-------------|
| POST | `/verify/pdf` | Verify signed document |
| POST | `/sign/` | Sign PDF document |
| GET | `/s/:slug` | Submitter signing portal |
| GET | `/health` | Health check |

#### API v1 (requires JWT or API key)

**📝 Submissions**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/submissions` | List submissions |
| POST | `/api/v1/submissions` | Create submission |
| GET | `/api/v1/submissions/:id` | Get submission |
| PUT | `/api/v1/submissions/:id` | Update submission |
| DELETE | `/api/v1/submissions/:id` | Delete submission |
| POST | `/api/v1/submissions/send` | Send to signers |
| POST | `/api/v1/submissions/bulk` | Bulk import from CSV/XLSX |

**👤 Submitters**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/submitters` | List submitters |
| GET | `/api/v1/submitters/:id` | Get submitter |
| POST | `/api/v1/submitters/:id/resend` | Resend invitation |
| POST | `/api/v1/submitters/:id/complete` | Complete signing |
| POST | `/api/v1/submitters/:id/decline` | Decline signing |

**📄 Templates**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/templates` | List templates |
| POST | `/api/v1/templates` | Create template |
| GET | `/api/v1/templates/:id` | Get template |
| PUT | `/api/v1/templates/:id` | Update template |
| DELETE | `/api/v1/templates/:id` | Delete template |
| POST | `/api/v1/templates/clone` | Clone template |
| POST | `/api/v1/templates/from-file` | Create from PDF |
| POST | `/api/v1/templates/formulas/validate` | Validate formula |
| POST | `/api/v1/templates/:id/conditions/validate` | Validate conditions |

**🔗 Signing Links** (direct signing without email)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/signing-links` | List signing links |
| POST | `/api/v1/signing-links` | Create signing link |
| GET | `/api/v1/signing-links/:submission_id` | Get signing link |
| GET | `/api/v1/signing-links/:submission_id/document` | Download completed document |

**🏢 Organizations**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/organizations` | List organizations |
| POST | `/api/v1/organizations` | Create organization |
| GET | `/api/v1/organizations/:id` | Get organization |
| PUT | `/api/v1/organizations/:id` | Update organization |
| DELETE | `/api/v1/organizations/:id` | Delete organization |
| POST | `/api/v1/organizations/:id/switch` | Switch organization context (admin only) |

**👥 Organization Members**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/organizations/:id/members` | List members |
| POST | `/api/v1/organizations/:id/members` | Add member |
| PUT | `/api/v1/organizations/:id/members/:user_id` | Update member role |
| DELETE | `/api/v1/organizations/:id/members/:user_id` | Remove member |

**✉️ Invitations**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/organizations/:id/invitations` | List invitations |
| POST | `/api/v1/organizations/:id/invitations` | Send invitation |
| POST | `/api/v1/invitations/:token/accept` | Accept invitation |
| DELETE | `/api/v1/invitations/:id` | Revoke invitation |

**🔑 API Keys**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/apikeys` | List API keys |
| POST | `/api/v1/apikeys` | Create API key |
| DELETE | `/api/v1/apikeys/:id` | Revoke key |
| POST | `/api/v1/apikeys/:id/enable` | Enable key |
| POST | `/api/v1/apikeys/:id/disable` | Disable key |

**🪝 Webhooks**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/webhooks` | List webhooks |
| POST | `/api/v1/webhooks` | Create webhook |
| PUT | `/api/v1/webhooks/:id` | Update webhook |
| DELETE | `/api/v1/webhooks/:id` | Delete webhook |

**⚙️ Settings**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/settings` | Get settings |
| PUT | `/api/v1/settings/email` | Update email config |
| PUT | `/api/v1/settings/storage` | Update storage config |
| PUT | `/api/v1/settings/branding` | Update branding |

**🎨 Branding & i18n**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/branding` | Get branding settings |
| PUT | `/api/v1/branding` | Update branding |
| POST | `/api/v1/branding/assets` | Upload branding asset |
| GET | `/api/v1/i18n/locales` | List available locales |
| PUT | `/api/v1/account/locale` | Update account locale |

**✉️ Email Templates**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/email-templates` | List templates |
| POST | `/api/v1/email-templates` | Create template |
| PUT | `/api/v1/email-templates/:id` | Update template |

**📊 Events & Stats**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/events` | List events (audit log) |
| GET | `/api/v1/stats` | Get statistics |

> Full interactive reference: [Swagger UI](http://localhost:8088/swagger/index.html) · [docs/SWAGGER.md](docs/SWAGGER.md)

## Configuration

All configuration is via environment variables with the `GOSIGN_` prefix. Infrastructure settings are read at startup; application settings (SMTP, storage, branding) are managed in the database via Admin UI.

| Variable | Default | Description |
|----------|---------|-------------|
| `GOSIGN_HTTP_ADDR` | `0.0.0.0:8088` | HTTP server address |
| `GOSIGN_DEV_MODE` | `false` | Development mode |
| `GOSIGN_POSTGRES_URL` | — | PostgreSQL connection URL |
| `GOSIGN_REDIS_ADDRESS` | `localhost:6379` | Redis address |
| `GOSIGN_REDIS_PASSWORD` | — | Redis password |

## Development

### Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./pkg/pdf/sign/...
```

### Building for Production

```bash
# Backend
go build -o gosign cmd/goSign/main.go

# Frontend
cd web && bun run build
```

### Docker

```bash
docker compose -f compose.yaml up -d
```

Production compose uses a dedicated `nginx` gateway container:
- `http://localhost/` -> frontend
- `http://localhost/api/` -> backend API
- `http://localhost/swagger/index.html` -> Swagger UI

## Scripts

Located in `scripts/`:

| Script | Description |
|--------|-------------|
| `migration` | Database migration management (wraps goose) |
| `migration dev up/down` | Load/unload development fixtures |
| `clean` | Clean build artifacts |
| `key` | Generate cryptographic keys |
| `models` | Generate data models |
| `tools` | Development tools |

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

| Document | Description |
|----------|-------------|
| [docs/API_AUTHENTICATION.md](docs/API_AUTHENTICATION.md) | JWT and API key authentication guide |
| [docs/EMBEDDED_SIGNING.md](docs/EMBEDDED_SIGNING.md) | JavaScript SDK for iframe integration |
| [docs/SWAGGER.md](docs/SWAGGER.md) | Swagger documentation generation |
| [docs/TESTING.md](docs/TESTING.md) | Testing strategy and guidelines |
| [docs/MULTILINGUAL.md](docs/MULTILINGUAL.md) | i18n and signing portal languages |
| [docs/CONDITIONAL_FIELDS.md](docs/CONDITIONAL_FIELDS.md) | Dynamic show/hide field logic |
| [docs/FORMULAS.md](docs/FORMULAS.md) | Formula engine and builder |
| [docs/WHITE_LABEL.md](docs/WHITE_LABEL.md) | White-label branding and themes |
| [docs/FRONTEND_COMPONENTS.md](docs/FRONTEND_COMPONENTS.md) | Frontend component architecture |

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
