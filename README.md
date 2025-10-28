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
- 📁 **Template System**: Reusable document templates with 14 field types
- 🗄️ **Flexible Storage**: Local, S3, GCS, or Azure Blob storage
- ⚡ **Rate Limiting**: Configurable API rate limits
- 🔍 **Event Logging**: Comprehensive audit trail
- 🎯 **Generic CRUD API**: Consistent REST API design

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.22+
- **Framework**: Fiber v2 (HTTP server)
- **Database**: PostgreSQL 14+ with JSONB
- **Cache**: Redis 6+
- **Authentication**: JWT + API Keys
- **Email**: SMTP/SendGrid support
- **Storage**: Local, S3, GCS, Azure
- **PDF Processing**: 
  - digitorus/pdf - Digital signing
  - pdfcpu - Document manipulation
  - signintech/gopdf - PDF generation
- **Task Scheduling**: robfig/cron v3
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
│   └── pdf-cert/            # PDF certificate examples
├── internal/                 # Private application code
│   ├── config/              # Configuration management
│   ├── handlers/
│   │   ├── api/            # REST API v1 handlers
│   │   ├── public/         # Public endpoints
│   │   └── private/        # Protected admin endpoints
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
│   ├── storage/             # Multi-provider storage
│   │   ├── local/          # Local filesystem
│   │   ├── s3/             # AWS S3/MinIO
│   │   ├── postgres/       # Database
│   │   └── redis/          # Cache
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
│   │   ├── composables/     # Vue composables
│   │   ├── layouts/         # Page layouts (Profile, Sidebar)
│   │   ├── models/          # TypeScript models
│   │   ├── pages/           # Application pages (9 pages)
│   │   ├── stores/          # Pinia stores
│   │   └── utils/           # Frontend utilities
├── migrations/               # Database migrations
├── fixtures/                 # Test data and fixtures
└── docker/                   # Docker configuration

```

## Installation

### Prerequisites
- Go 1.22 or higher
- PostgreSQL 14+
- Redis 6+
- Bun (for frontend development)
- Node.js 18+ (alternative to Bun)

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

3. Generate configuration file:
```bash
go run cmd/goSign/main.go gen --config
```

4. Configure database connection in the generated config file

5. Run database migrations:
```bash
./scripts/migration up
```

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

The application will start on `http://localhost:8080` (default) with three interfaces:
- **Public UI**: `http://localhost:8080/` - Document signing and verification
- **Admin UI**: `http://localhost:8080/_/` - Administration panel
- **API**: `http://localhost:8080/api/` - REST API endpoints

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
- `POST /auth/signin` - User login (returns JWT)
- `POST /auth/signout` - User logout

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

**📚 Complete API Reference:**
- **Interactive Docs**: [Swagger UI](http://localhost:8088/swagger/index.html)
- **Full Endpoint List**: [docs/SWAGGER.md](docs/SWAGGER.md)
- **Authentication Guide**: [docs/API_AUTHENTICATION.md](docs/API_AUTHENTICATION.md)

## Configuration

Configuration is managed through a TOML file. Key settings include:

- **HTTPAddr**: Server address (default: `0.0.0.0:8080`)
- **DevMode**: Development mode flag
- **Postgres**: Database connection settings
- **Redis**: Cache configuration
- **Trust**: Trust certificate sources and update settings
- **JWT**: Authentication settings

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
- JWT-based authentication
- Password hashing with bcrypt
- Secure certificate storage
- Input validation with ozzo-validation

## Scripts

Utility scripts are located in the `scripts/` directory:
- `clean` - Clean build artifacts and temporary files
- `key` - Generate cryptographic keys
- `migration` - Database migration management
- `models` - Generate data models
- `tools` - Development tools

## License

[Add license information]

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

See [IMPLEMENTATION_COMPLETE.md](docs/IMPLEMENTATION_COMPLETE.md) for full details.

## 📖 Documentation

Comprehensive documentation is available in the `docs/` directory:

- **[Implementation Report](docs/IMPLEMENTATION_COMPLETE.md)** - Full feature list and architecture
- **[API Authentication](docs/API_AUTHENTICATION.md)** - JWT and API key setup
- **[Embedded Signing](docs/EMBEDDED_SIGNING.md)** - JavaScript SDK integration
- **[Frontend Components](docs/FRONTEND_COMPONENTS.md)** - Component architecture and UI library
- **[Swagger Guide](docs/SWAGGER.md)** - API documentation generation

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

### Planned
- [ ] Multi-language support
- [ ] Advanced analytics dashboard
- [ ] External CA integration
- [ ] Mobile application
- [ ] E-signature standards (eIDAS)
- [ ] Advanced PDF form automation

---

**Made with ❤️ for secure document signing**

