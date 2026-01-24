# goSign

✍️ **Sign documents without stress**

A modern, full-featured document signing platform with multi-signer workflows, email notifications, and comprehensive API. Built with Go and Vue.js, goSign provides enterprise-grade capabilities for secure digital document signing.

## Overview

goSign is a complete document signing solution that combines powerful backend services with an intuitive frontend interface. It supports multi-party signing workflows, automated notifications, embedded signing, and extensive API integration capabilities.

## ✨ Key Features

### Core Signing Features
- 🔐 **Digital Signatures**: X.509 certificates with PKCS7/CMS standards
- ✅ **Document Verification**: Full certificate chain validation
- 🎨 **Visual Signatures**: Customizable signature appearance and placement
- 📜 **Certificate Management**: Generate, manage, and revoke certificates with CRL
- 🔄 **Trust Updates**: Automatic trust certificate updates (every 12h)

### Document Workflow
- 👥 **Multi-signer Workflow**: Sequential or parallel signing with state machine
- 📧 **Email Notifications**: Automated invitations, reminders, and status updates
- 📱 **SMS Support**: Optional SMS notifications for signers
- ⏰ **Scheduled Reminders**: Configurable reminder system
- 📊 **Status Tracking**: Real-time submission and signer status

### API & Integration
- 🔑 **Dual Authentication**: JWT tokens and API keys with rate limiting
- 📚 **Swagger Documentation**: Interactive API documentation
- 🔗 **Webhook Support**: Real-time event notifications
- 🖼️ **Embedded Signing**: JavaScript SDK for iframe integration
- 📦 **Bulk Operations**: CSV/XLSX import for mass submissions

### Advanced Features
- 📁 **Template System**: Reusable document templates with 14 field types (PDF file import supported)
- 🗄️ **Flexible Storage**: Local, S3 (GCS, Azure planned)
- ⚡ **Rate Limiting**: Configurable API rate limits
- 🔍 **Event Logging**: Comprehensive audit trail
- 🎯 **Generic CRUD API**: Consistent REST API design

### Enterprise Features
- 🏢 **Organizations & Teams**: Multi-tenant organization management
- 👥 **Role-Based Access Control**: Owner, Admin, Member, Viewer roles
- 🔐 **Organization Context**: JWT tokens with organization scope
- 📋 **Team Collaboration**: Invite members, manage permissions
- 🗂️ **Organization Templates**: Templates scoped to organizations
- 📊 **Team Analytics**: Organization-level statistics and insights
- 🌐 **Multilingual (i18n)**: 7 UI languages, 14 signing portal languages, RTL support
- 📝 **Conditional Fields**: Show/hide fields based on dynamic conditions
- 📐 **Formula Engine**: Dynamic field calculations with formula builder
- 🎨 **White-Label Branding**: Custom logos, colors, fonts, signing themes
- 📧 **Email Templates**: Customizable templates with locale support

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.25+
- **Framework**: Fiber v2 (HTTP server)
- **Database**: PostgreSQL 14+ with JSONB
- **Cache**: Redis 6+
- **Authentication**: JWT + API Keys
- **Email**: SMTP/SendGrid support
- **Storage**: Local, S3 (GCS, Azure planned)
- **PDF Processing**: 
  - digitorus/pdf - PDF reading and digital signing
  - signintech/gopdf - PDF creation and manipulation
- **Task Scheduling**: Built-in Go scheduler
- **Logging**: zerolog
- **API Docs**: Swagger/OpenAPI

### Frontend
- **Framework**: Vue 3 + TypeScript (Composition API)
- **State Management**: Pinia
- **Routing**: Vue Router
- **Styling**: Tailwind CSS v4
- **Build Tool**: Vite
- **Package Manager**: Bun
- **Components**: 
  - **UI Library**: 21 reusable components (Button, Input, Modal, Table, etc.)
  - **Common Components**: FieldInput (14 field types), ResourceTable, FormModal
  - **Template Components**: Area, Document, Page, Preview
  - signature_pad for capture

## Project Structure

```
goSign/
├── cmd/                      # Command-line applications
│   ├── goSign/              # Main application
│   ├── cert/                # Certificate utilities
│   ├── pdf/                 # PDF utilities
├── internal/                 # Private application code
│   ├── config/              # Configuration management
│   ├── handlers/
│   │   ├── api/            # REST API v1 handlers
│   │   └── public/         # Public and auth endpoints
│   ├── middleware/          # JWT, rate limiting, CORS
│   ├── models/              # Data models (14 models)
│   ├── queries/             # Database repositories
│   ├── routes/              # API v1 routes
│   ├── services/            # Business logic
│   │   ├── submission/     # Multi-signer workflow
│   │   ├── apikey/         # API key management
│   │   └── reminder/       # Reminder scheduling
│   ├── trust/               # Trust certificate management
│   └── worker/              # Background tasks
├── pkg/                      # Public libraries
│   ├── pdf/
│   │   ├── sign/           # Digital signing
│   │   ├── verify/         # Signature verification
│   │   ├── fill/           # PDF form filling
│   │   └── revocation/     # CRL management
│   ├── notification/        # Email/SMS service
│   ├── webhook/             # Webhook dispatcher
│   ├── storage/             # Blob storage
│   │   ├── local/          # Local filesystem
│   │   └── s3/             # AWS S3/MinIO
│   ├── security/
│   │   ├── cert/           # Certificate operations
│   │   └── password/       # Hashing and validation
│   └── utils/               # Helper functions
├── web/                      # Frontend application
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/          # 21 reusable UI components
│   │   │   ├── common/      # Generic components (FieldInput, FormModal, ResourceTable)
│   │   │   ├── field/       # Field-specific components
│   │   │   └── template/    # Document template components
│   │   ├── composables/     # Vue composables (conditions, formulas, i18n)
│   │   ├── layouts/         # Page layouts (Main, Profile, Settings)
│   │   ├── models/          # TypeScript models
│   │   ├── pages/           # Application pages (Dashboard, Sign, Verify, Settings, etc.)
│   │   ├── stores/          # Pinia stores
│   │   └── utils/           # Frontend utilities
├── migrations/               # Database migrations
├── fixtures/                 # Test data and fixtures
└── docker/                   # Docker configuration

```

## Installation

### Prerequisites
- Go 1.25 or higher
- PostgreSQL 14+
- Redis 6+
- Bun (for frontend development)
- Node.js 18+ (alternative to Bun)
- **PDF→JPG (preview generation)**: `pdftoppm` from **poppler-utils** (required for template previews when creating templates from PDF)

  | OS | Package | Install |
  |----|---------|---------|
  | Debian / Ubuntu | `poppler-utils` | `sudo apt install poppler-utils` |
  | RHEL / Fedora / CentOS | `poppler-utils` | `sudo dnf install poppler-utils` |
  | Alpine | `poppler-utils` | `apk add poppler-utils` |
  | Arch | `poppler` | `pacman -S poppler` |
  | macOS (Homebrew) | `poppler` | `brew install poppler` |

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

3. Generate or copy configuration (creates `gosign.toml` in project root):
```bash
go run cmd/goSign/main.go gen --config
# or: cp cmd/goSign/gosign.example.toml ./gosign.toml
```

4. Edit `gosign.toml`: set Postgres URL, Redis, and SMTP

5. Run database migrations:
```bash
./scripts/migration up
```

6. (Optional) Load test data for development:
```bash
./scripts/migration dev up
```

This will create test users:
- **Admin**: `admin@gosign.local` / `admin123`
- **User 1**: `user1@gosign.local` / `user123`
- **User 2**: `user2@gosign.local` / `user234`

See `fixtures/migration/README.md` for more details about test data.

### Frontend Setup

1. Navigate to the web directory:
```bash
cd web
```

2. Install dependencies:
```bash
bun install
# or
npm install
```

3. Start development server:
```bash
bun run dev
# or
npm run dev
```

## Usage

### Starting the Application

Run the main application:
```bash
go run cmd/goSign/main.go serve
```

The application will start on `http://localhost:8088` (default) with three interfaces:
- **Public UI**: `http://localhost:8088/` - Document signing and verification
- **Admin UI**: `http://localhost:8088/_/` - Administration panel
- **API**: `http://localhost:8088/api/` - REST API endpoints

### CLI Commands

```bash
# Start the server
go run cmd/goSign/main.go serve

# Generate configuration
go run cmd/goSign/main.go gen --config

# Run certificate utilities
go run cmd/cert/main.go [options]
```

### API Endpoints

#### Public Endpoints
- `POST /verify/pdf` - Verify signed document
- `POST /sign` - Sign PDF document
- `GET /s/:slug` - Submitter signing portal
- `GET /health` - Health check

#### Authentication
- `POST /auth/signup` - User registration
- `POST /auth/signin` - User login (returns JWT + refresh token)
- `POST /auth/refresh` - Refresh access token
- `POST /auth/signout` - User logout
- `POST /auth/password/forgot` - Request password reset
- `POST /auth/password/reset` - Reset password with token
- `GET /auth/verify-email` - Verify email address
- `POST /auth/2fa/enable` - Enable two-factor authentication
- `POST /auth/2fa/verify` - Verify 2FA code
- `POST /auth/2fa/disable` - Disable 2FA
- `GET /auth/oauth/google` - Google OAuth login
- `GET /auth/oauth/github` - GitHub OAuth login

#### API v1 (Protected)

> **Note**: Below are key examples. Full API includes **42+ endpoints** across all resources.

**Submissions** (8 endpoints)
- `GET /api/v1/submissions` - List submissions
- `GET /api/v1/submissions/:id` - Get submission details
- `POST /api/v1/submissions` - Create submission
- `PUT /api/v1/submissions/:id` - Update submission
- `POST /api/v1/submissions/send` - Send to signers
- `POST /api/v1/submissions/bulk` - Bulk import from CSV

**Submitters** (6 endpoints)
- `GET /api/v1/submitters` - List submitters
- `GET /api/v1/submitters/:id` - Get submitter details
- `POST /api/v1/submitters/:id/resend` - Resend invitation
- `POST /api/v1/submitters/:id/complete` - Complete signing
- `POST /api/v1/submitters/:id/decline` - Decline signing

**Templates** (7 endpoints)
- `GET /api/v1/templates` - List templates
- `GET /api/v1/templates/:id` - Get template details
- `POST /api/v1/templates` - Create template
- `PUT /api/v1/templates/:id` - Update template
- `POST /api/v1/templates/clone` - Clone template
- `POST /api/v1/templates/from-file` - Create from PDF

**Organizations** (6 endpoints)
- `GET /api/v1/organizations` - List user's organizations
- `GET /api/v1/organizations/:id` - Get organization details
- `POST /api/v1/organizations` - Create organization
- `PUT /api/v1/organizations/:id` - Update organization
- `DELETE /api/v1/organizations/:id` - Delete organization
- `POST /api/v1/organizations/:id/switch` - Switch organization context

**Organization Members** (7 endpoints)
- `GET /api/v1/organizations/:id/members` - List members
- `POST /api/v1/organizations/:id/members` - Add member
- `PUT /api/v1/organizations/:id/members/:user_id` - Update member role
- `DELETE /api/v1/organizations/:id/members/:user_id` - Remove member

**Organization Invitations** (5 endpoints)
- `GET /api/v1/organizations/:id/invitations` - List invitations
- `POST /api/v1/organizations/:id/invitations` - Send invitation
- `POST /api/v1/invitations/:token/accept` - Accept invitation
- `DELETE /api/v1/invitations/:id` - Revoke invitation

**API Keys** (6 endpoints)
- `GET /api/v1/apikeys` - List API keys
- `POST /api/v1/apikeys` - Create API key
- `DELETE /api/v1/apikeys/:id` - Revoke key
- `POST /api/v1/apikeys/:id/enable` - Enable key
- `POST /api/v1/apikeys/:id/disable` - Disable key

**Webhooks** (5 endpoints)
- `GET /api/v1/webhooks` - List webhooks
- `POST /api/v1/webhooks` - Create webhook
- `PUT /api/v1/webhooks/:id` - Update webhook
- `DELETE /api/v1/webhooks/:id` - Delete webhook

**Settings** (4 endpoints)
- `GET /api/v1/settings` - Get settings
- `PUT /api/v1/settings/email` - Update email config
- `PUT /api/v1/settings/storage` - Update storage config
- `PUT /api/v1/settings/branding` - Update branding

**Branding, i18n, Email Templates**
- `GET /api/v1/branding`, `PUT /api/v1/branding` - White-label branding
- `GET /api/v1/i18n/locales` - Available locales
- `GET /api/v1/email-templates`, `POST /api/v1/email-templates`, `PUT /api/v1/email-templates/:id` - Email templates

**📚 Complete API Reference:**
- **Interactive Docs**: [Swagger UI](http://localhost:8088/swagger/index.html)
- **Full Endpoint List**: [docs/SWAGGER.md](docs/SWAGGER.md)
- **Authentication Guide**: [docs/API_AUTHENTICATION.md](docs/API_AUTHENTICATION.md)

## Configuration

Configuration is managed through a TOML file (`gosign.toml` in project root).

### Quick Setup

1. **Copy example configuration to project root:**
   ```bash
   cp cmd/goSign/gosign.example.toml ./gosign.toml
   ```

2. **Update required values in `gosign.toml`:**
   - Database connection (PostgreSQL)
   - Redis connection (for authentication features)
   - SMTP settings (for email notifications)

### Key Configuration Sections

- **http-addr**: Server address (default: `0.0.0.0:8088`)
- **DevMode**: Development mode flag
- **Postgres**: Database connection settings
- **Redis**: Session storage and caching
- **Settings**: Email, Storage, Webhook, Features
- **Trust**: Certificate trust sources and updates

See `cmd/goSign/gosign.example.toml` for all configuration options.

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/pdf/sign/...
```

### Building for Production

Backend:
```bash
go build -o gosign cmd/goSign/main.go
```

Frontend:
```bash
cd web
bun run build
# or
npm run build
```

### Docker Deployment

```bash
docker-compose -f docker/docker-compose.yaml up -d
```

## Key Features Details

### PDF Signing
- Supports PAdES (PDF Advanced Electronic Signatures)
- PKCS#7/CMS signature format
- Visual signature placement
- Multiple signature fields support
- Timestamp support

### Certificate Management
- X.509 certificate generation
- Certificate Revocation Lists (CRL)
- Certificate chain validation
- Trust store management
- Automatic trust certificate updates

### Security
- JWT-based authentication with refresh tokens (7 days)
- Password hashing with bcrypt
- Two-factor authentication (2FA) support
- OAuth integration (Google, GitHub)
- Email verification system
- Password reset with secure tokens
- Secure certificate storage
- Input validation with go-playground/validator

## Scripts

Utility scripts are located in the `scripts/` directory:
- `clean` - Clean build artifacts and temporary files
- `key` - Generate cryptographic keys
- `migration` - Database migration management
- `models` - Generate data models
- `tools` - Development tools

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0).

See the [LICENSE](LICENSE) file for the full license text.

**Summary:**
- ✅ You are free to use, modify, and distribute this software
- ✅ You must keep the same license when distributing
- ✅ You must include the full license text and source code
- ❌ You cannot use this software in proprietary (closed-source) applications

For more information about GPL-3.0, visit: https://www.gnu.org/licenses/gpl-3.0.html

## Contributing

Contributions are welcome! Please read the contributing guidelines before submitting pull requests.

## 💬 Support

For issues and questions:
- **GitHub Issues**: [https://github.com/shurco/gosign/issues](https://github.com/shurco/gosign/issues)
- **Documentation**: [docs/README.md](docs/README.md)
- **API Reference**: http://localhost:8088/swagger/index.html

## 🌟 What's New in v2.0

goSign v2.0 introduces enterprise document signing capabilities:

- ✅ **Multi-party Signing**: Complete workflow with sequential/parallel signing
- ✅ **Notification System**: Automated emails, SMS, and reminders
- ✅ **API Keys**: Secure service-to-service authentication
- ✅ **Rate Limiting**: Protection against abuse (100-10 req/min)
- ✅ **Embedded Signing**: JavaScript SDK for iframe integration
- ✅ **Bulk Operations**: CSV/XLSX import for mass creation
- ✅ **Webhooks**: Real-time event notifications
- ✅ **Storage Options**: S3, GCS, Azure, or local
- ✅ **Swagger Docs**: Interactive API documentation

## 🏢 What's New in v2.1

goSign v2.1 adds enterprise team collaboration features:

- ✅ **Organizations**: Multi-tenant organization management
- ✅ **Role-Based Access**: Four roles (Owner, Admin, Member, Viewer)
- ✅ **Team Invitations**: Email-based member invitations
- ✅ **Organization Context**: JWT tokens with organization scope
- ✅ **Team Templates**: Templates shared within organizations
- ✅ **Organization Isolation**: Data separation between organizations

## 🌐 What's New in v2.4

goSign v2.4 adds advanced enterprise features:

- ✅ **Multilingual (i18n)**: 7 UI and 14 signing portal languages, RTL support
- ✅ **Conditional Fields**: Show/hide fields based on conditions
- ✅ **Formula Engine**: Dynamic calculations with formula builder
- ✅ **White-Label Branding**: Custom logos, colors, fonts, signing themes
- ✅ **Email Templates**: Customizable templates with locale support

See [IMPLEMENTATION_COMPLETE.md](docs/IMPLEMENTATION_COMPLETE.md) and [docs/README.md](docs/README.md) for full details.

## 📖 Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Implementation Report](docs/IMPLEMENTATION_COMPLETE.md)** - Full feature list and architecture
- **[API Authentication](docs/API_AUTHENTICATION.md)** - JWT and API key setup
- **[Embedded Signing](docs/EMBEDDED_SIGNING.md)** - JavaScript SDK integration
- **[Frontend Components](docs/FRONTEND_COMPONENTS.md)** - Component architecture and UI library
- **[Swagger Guide](docs/SWAGGER.md)** - API documentation generation
- **[Multilingual](docs/MULTILINGUAL.md)** - i18n and signing portal languages
- **[Conditional Fields](docs/CONDITIONAL_FIELDS.md)** - Dynamic show/hide logic
- **[Formulas](docs/FORMULAS.md)** - Formula engine and builder
- **[White-Label](docs/WHITE_LABEL.md)** - Branding and themes

**Quick Links:**
- Swagger UI: http://localhost:8088/swagger/index.html
- Documentation Index: [docs/README.md](docs/README.md)

## 🗺️ Roadmap

### Completed ✅
- [x] Multi-signer workflows
- [x] Email/SMS notifications
- [x] API keys and rate limiting
- [x] Embedded signing SDK
- [x] Bulk operations
- [x] Webhook system
- [x] Swagger documentation
- [x] Organizations and role-based access
- [x] Multilingual support (i18n)
- [x] Conditional fields
- [x] Formula engine
- [x] White-label branding
- [x] Custom email templates

### Planned
- [ ] GCS and Azure blob storage
- [ ] Advanced analytics dashboard
- [ ] External CA integration
- [ ] Mobile application
- [ ] E-signature standards (eIDAS)
- [ ] Advanced PDF form automation

---

**Made with ❤️ for secure document signing**

