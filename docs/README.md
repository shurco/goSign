# goSign Documentation

**Last Updated**: 2025-10-28 08:30 UTC

## 📚 Available Documentation

### Getting Started
- **[IMPLEMENTATION_COMPLETE.md](IMPLEMENTATION_COMPLETE.md)** - Full implementation report
  - All implemented features (23/23 tasks completed)
  - Architecture overview
  - Technical stack details
  - File structure and key files

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

## 🔗 Quick Access

### API Reference (When Running)
- **Swagger UI**: http://localhost:8088/swagger/index.html
- **OpenAPI JSON**: http://localhost:8088/swagger/doc.json
- **OpenAPI YAML**: http://localhost:8088/swagger/doc.yaml

### Main API Endpoints
```
Authentication:
  POST /auth/signin             - User login

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

### Advanced Features
- **Bulk Operations**: CSV/XLSX import for mass submission creation
- **Webhook System**: Real-time event notifications
- **Rate Limiting**: 100 req/min standard, 10 req/min for sensitive ops
- **PDF Assembly**: Dynamic field filling and audit trail generation
- **Reminders**: Scheduled notifications for pending signatures

## 📊 Documentation Coverage

| Category | Coverage | Status |
|----------|----------|--------|
| API Documentation | 100% | ✅ Complete |
| Authentication | 100% | ✅ Complete |
| Integration Guides | 100% | ✅ Complete |
| Frontend Architecture | 100% | ✅ Complete |
| Code Examples | 100% | ✅ Complete |
| Architecture | 100% | ✅ Complete |

---

**Status**: ✅ Complete  
**Total Documents**: 5  
**Version**: 2.1.0

