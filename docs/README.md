# goSign Documentation

**Last Updated**: 2026-01-26

## 📚 Available Documentation

### Getting Started
- **[IMPLEMENTATION_COMPLETE.md](IMPLEMENTATION_COMPLETE.md)** - Full implementation report
  - All implemented features (23/23 tasks completed)
  - Architecture overview
  - Technical stack details
  - File structure and key files

### Testing
- **[TESTING.md](TESTING.md)** - Complete testing guide
  - Unit and integration testing strategies
  - Test coverage and benchmarks
  - Authentication test documentation
  - Performance testing guidelines
  
- **[TEST_QUICKSTART.md](TEST_QUICKSTART.md)** - Quick testing reference
  - Fast commands for running tests
  - Test script usage
  - Coverage and benchmark commands


### API Documentation
- **[API_AUTHENTICATION.md](API_AUTHENTICATION.md)** - Complete authentication guide
  - JWT Bearer tokens
  - API Keys management and security
  - Rate limiting configuration
  - Code examples for Go and JavaScript
  
- **[SWAGGER.md](SWAGGER.md)** - Swagger/OpenAPI documentation
  - How to generate API documentation
  - Swagger UI usage
  - Annotation examples
  - Available API endpoints

### Integration Guides
- **[EMBEDDED_SIGNING.md](EMBEDDED_SIGNING.md)** - Embedded signing integration
  - Backend API endpoints for embedding
  - Frontend JavaScript SDK
  - Usage examples and security
  - Event handling and callbacks

### Frontend Architecture
- **[FRONTEND_COMPONENTS.md](FRONTEND_COMPONENTS.md)** - Component architecture guide
  - UI component library (21 primitives)
  - Common components (FieldInput, ResourceTable, FormModal)
  - Design principles (KISS, DRY, Composition)
  - Usage examples and best practices

### Enterprise Features
- **[ENTERPRISE_IMPROVEMENTS_SUMMARY.md](ENTERPRISE_IMPROVEMENTS_SUMMARY.md)** - Complete enterprise features overview
  - Multilingual support (i18n) implementation
  - Conditional fields system
  - Formula engine
  - White-label branding
  - Email templates customization

- **[MULTILINGUAL.md](MULTILINGUAL.md)** - Multilingual support guide
  - 7 UI languages and 14 signing portal languages
  - Automatic locale detection
  - RTL support for Arabic and Hebrew
  - Field-level translations

- **[CONDITIONAL_FIELDS.md](CONDITIONAL_FIELDS.md)** - Conditional fields documentation
  - Show/hide fields based on conditions
  - Condition builder UI
  - Backend validation and evaluation

- **[FORMULAS.md](FORMULAS.md)** - Formula engine documentation
  - Dynamic field calculations
  - Formula builder interface
  - Supported functions and operators

- **[WHITE_LABEL.md](WHITE_LABEL.md)** - White-label branding guide
  - Custom logos, colors, and fonts
  - Signing page themes
  - Email template customization
  - Custom CSS support

## 🔗 Quick Access

### API Reference (When Running)
- **Swagger UI**: http://localhost:8088/swagger/index.html
- **OpenAPI JSON**: http://localhost:8088/swagger/doc.json
- **OpenAPI YAML**: http://localhost:8088/swagger/doc.yaml

### Main API Endpoints
```
Authentication:
  POST /auth/signin             - User login

Organizations (admin only):
  GET    /api/v1/organizations           - List organizations
  POST   /api/v1/organizations           - Create organization
  POST   /api/v1/organizations/:id/switch - Switch context (administrators only)
  POST   /api/v1/organizations/switch    - Exit organization (administrators only)
  UI:    /admin/organizations            - Manage organizations (admin only)

Members:
  GET    /api/v1/organizations/:id/members - List members
  POST   /api/v1/organizations/:id/members - Add member

Invitations:
  POST   /api/v1/organizations/:id/invitations - Send invitation
  POST   /api/v1/invitations/:token/accept - Accept invitation

Submissions:
  GET    /api/v1/submissions    - List submissions
  POST   /api/v1/submissions    - Create submission
  POST   /api/v1/submissions/bulk - Bulk create from CSV

Templates:
  GET    /api/v1/templates      - List templates
  POST   /api/v1/templates      - Create template

API Keys:
  GET    /api/v1/apikeys        - List API keys
  POST   /api/v1/apikeys        - Create API key

Branding:
  GET    /api/v1/branding       - Get branding settings
  PUT    /api/v1/branding       - Update branding settings

i18n:
  GET    /api/v1/i18n/locales   - List available locales
  PUT    /api/v1/account/locale - Update account locale

Email Templates (per-organization, scoped by JWT):
  GET    /api/v1/email-templates         - List email templates (current org or account)
  POST   /api/v1/email-templates         - Create email template
  PUT    /api/v1/email-templates/:id     - Update email template
  UI:    /settings/email/templates       - Configure templates (Settings, after General)

Public:
  GET    /s/:slug               - Submitter signing portal
  POST   /verify/pdf            - Verify PDF signature
```

## 🚀 Quick Start Guides

### For Developers
1. Read [IMPLEMENTATION_COMPLETE.md](IMPLEMENTATION_COMPLETE.md) for architecture overview
2. Check [API_AUTHENTICATION.md](API_AUTHENTICATION.md) for authentication setup
3. Use [SWAGGER.md](SWAGGER.md) to generate and view API documentation
4. Explore Swagger UI for interactive API testing

### For Integration Partners
1. Generate API keys through Settings UI or API
2. Review [EMBEDDED_SIGNING.md](EMBEDDED_SIGNING.md) for iframe integration
3. Implement webhook handlers for real-time event notifications
4. Test with Swagger UI before production deployment

### For Testing
1. Start application: `go run cmd/goSign/main.go serve`
2. Open Swagger UI: http://localhost:8088/swagger/index.html
3. Authorize with JWT token or API key
4. Test endpoints interactively

## 📋 Documentation Standards

All documentation follows these standards:
- **Language**: English for all code, comments, and documentation
- **Format**: Markdown with proper headers and code blocks
- **Examples**: Working code samples in multiple languages (Go, JavaScript, curl)
- **Timestamps**: YYYY-MM-DD HH:MM UTC format
- **Code Quality**: All examples tested and follow project standards

## 🎯 Feature Documentation

### Core Features
- **Multi-signer Workflow**: State machine-based submission process
- **Notification System**: Unified service for email/SMS/reminders
- **Storage Abstraction**: Support for local, S3, GCS, Azure storage
- **API Authentication**: JWT tokens and API keys with rate limiting
- **Embedded Signing**: JavaScript SDK for iframe integration
- **Settings in Database**: Global settings (SMTP, SMS, storage, branding) stored in DB; Admin UI at `/_/` for configuration
- **Field Model**: Field preferences (format, align, price, currency, date format, signature format), structured validation (pattern, min, max, message), readonly, title, description; number field type with default_value; Areas with cell_count/option_id
- **Submitter Signing UX**: In-document field overlays with labels and filled values (or field-type icon when empty), expandable field form drawer, progress dots navigation, prev/next navigation, draft persistence, signature ID display; number fields honor format/min/max and default value

### Advanced Features
- **Bulk Operations**: CSV/XLSX import for mass submission creation
- **Webhook System**: Real-time event notifications
- **Rate Limiting**: 100 req/min standard, 10 req/min for sensitive ops
- **PDF Assembly**: Dynamic field filling and audit trail generation
- **Reminders**: Scheduled notifications for pending signatures

### Enterprise Features
- **Organizations & Teams**: Multi-tenant organization management; UI under Admin section (`/admin/organizations`), access and switch/exit restricted to administrators
- **Organization Roles**: Owner, Admin, Member, Viewer with granular permissions
- **Team Invitations**: Email-based member invitation system with token expiration
- **Organization Context**: JWT tokens include organization_id for multi-tenant isolation; only admins can switch organization context via API
- **Organization Templates**: Templates can be shared within organizations
- **Team Collaboration**: Members can collaborate on templates and submissions
- **Email Templates in Settings**: Per-organization email templates configured under Settings → Email templates (after General); scoped by current organization or account
- **Multilingual Support**: 7 UI languages and 14 signing portal languages with RTL support
- **Conditional Fields**: Show/hide fields based on dynamic conditions
- **Formula Engine**: Dynamic field calculations with formula builder
- **White-Label Branding**: Custom logos, colors, fonts, and themes
- **Email Templates**: Customizable email templates with locale support

### Authentication Features
- **User Registration**: Sign up with email verification
- **Password Management**: Forgot/reset password with secure tokens
- **Two-Factor Authentication**: TOTP-based 2FA with QR codes
- **OAuth Integration**: Google and GitHub OAuth login
- **Token Management**: JWT access tokens (10min) + refresh tokens (7 days)
- **Email Verification**: Secure email verification flow

## 📊 Documentation Coverage

| Category | Coverage | Status |
|----------|----------|--------|
| API Documentation | 100% | ✅ Complete |
| Authentication | 100% | ✅ Complete |
| Integration Guides | 100% | ✅ Complete |
| Frontend Architecture | 100% | ✅ Complete |
| Code Examples | 100% | ✅ Complete |
| Architecture | 100% | ✅ Complete |
| Testing | 100% | ✅ Complete |
| Enterprise Features | 100% | ✅ Complete |

---

**Status**: ✅ Complete  
**Total Documents**: 13  
**Version**: 2.6.0

## 🆕 Enterprise Features

### v2.4.0 - Advanced Enterprise Features
goSign v2.4 adds comprehensive enterprise-grade improvements:

#### Multilingual Support (i18n)
- 7 UI languages (EN, RU, ES, FR, DE, IT, PT)
- 14 signing portal languages with additional languages
- Automatic locale detection and RTL support
- Field-level translations for templates

#### Conditional Fields
- Dynamic show/hide fields based on conditions
- Visual condition builder interface
- Support for complex logical expressions
- Backend validation and evaluation

#### Formula Engine
- Dynamic field calculations
- Formula builder with syntax highlighting
- Support for mathematical and logical operations
- Real-time formula evaluation

#### White-Label Branding
- Custom logos, colors, and fonts
- Multiple signing page themes (default, minimal, corporate)
- Email template customization
- Custom CSS support for advanced styling

#### Email Templates
- Customizable email templates per organization
- Locale-specific email templates
- Template variables and placeholders
- Rich HTML email support

### v2.6.0 - Admin-Only Organizations and Settings Email Templates
- **Organizations**: Moved to Administrator section (`/admin/organizations`). Only administrators can access the page and switch or exit organization context (API returns 403 for non-admins). Organization dropdown removed from sidebar.
- **Email Templates**: Moved from Admin settings to user Settings (`/settings/email/templates`). Each organization can have its own templates; scope is determined by JWT (organization_id or account). Settings tab order: General, Email templates, Webhooks, API keys, Branding.
- **Database**: `email_template` table extended with `organization_id` for per-organization templates (migration `20260125000001_email_template_organization.sql`).

### v2.3.0 - Organization Management
goSign v2.3 adds comprehensive organization and team management:

#### Organizations
- Create and manage organizations (admin UI at `/admin/organizations`)
- Switch between personal and organization contexts (administrators only)
- Organization-scoped data isolation

#### Team Management
- Invite members via email
- Manage member roles and permissions
- View organization members and their roles

#### Role-Based Access Control
- **Owner**: Full control, can delete organization
- **Admin**: Manage members and organization settings
- **Member**: Create and manage templates/submissions
- **Viewer**: Read-only access to organization data

### API Integration
All endpoints are documented in Swagger UI and follow RESTful conventions. See [API_AUTHENTICATION.md](API_AUTHENTICATION.md) for details on authentication and [ENTERPRISE_IMPROVEMENTS_SUMMARY.md](ENTERPRISE_IMPROVEMENTS_SUMMARY.md) for enterprise features.

