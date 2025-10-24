# goSign

✍️ **Sign documents without stress**

A modern web application for digital signing and verification of PDF documents with support for X.509 certificates, visual signatures, and certificate revocation lists.

## Overview

goSign is a full-stack application that provides a secure and user-friendly platform for digitally signing PDF documents. It combines a powerful Go backend with a modern Vue.js frontend to deliver a seamless document signing experience.

## Features

- 🔐 **Digital Signatures**: Sign PDF documents using X.509 certificates with PKCS7/CMS standards
- ✅ **Document Verification**: Verify signed documents and validate certificate chains
- 🎨 **Visual Signatures**: Add visible signature fields with customizable appearance
- 📜 **Certificate Management**: Generate, manage, and revoke certificates with CRL support
- 🔄 **Trust Certificate Updates**: Automatic trust certificate updates every 12 hours via cron
- 📁 **Template Support**: Create and manage signature templates for reusable configurations
- 👤 **User Authentication**: Secure JWT-based authentication system
- 📊 **Document Management**: Upload, organize, and track signed documents
- 🌐 **Multi-interface**: Public UI, Admin UI, and REST API
- 🔍 **Health Monitoring**: Built-in health check endpoints

## Tech Stack

### Backend
- **Language**: Go 1.22+
- **Web Framework**: Fiber v2
- **Database**: PostgreSQL (pgx v5)
- **Cache**: Redis
- **Authentication**: JWT (golang-jwt/jwt)
- **PDF Processing**: 
  - digitorus/pdf - PDF signing
  - pdfcpu - PDF manipulation
  - signintech/gopdf - PDF generation
- **Task Scheduling**: robfig/cron for periodic tasks
- **Logging**: zerolog

### Frontend
- **Framework**: Vue 3 with TypeScript
- **State Management**: Pinia
- **Routing**: Vue Router
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **Package Manager**: Bun
- **UI Components**: 
  - signature_pad - Digital signature capture
  - nprogress - Progress indicators

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
│   ├── handlers/            # HTTP request handlers
│   │   ├── public/          # Public API endpoints
│   │   └── private/         # Protected endpoints
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data models
│   ├── queries/             # Database queries
│   ├── routes/              # Route definitions
│   └── trust/               # Trust certificate management
├── pkg/                      # Public libraries
│   ├── pdf/                 # PDF operations
│   │   ├── sign/           # PDF signing
│   │   ├── verify/         # PDF verification
│   │   └── revocation/     # Certificate revocation
│   ├── security/            # Security utilities
│   │   ├── cert/           # Certificate management
│   │   └── password/       # Password handling
│   ├── storage/             # Storage backends
│   │   ├── postgres/       # PostgreSQL
│   │   └── redis/          # Redis
│   └── utils/               # Utility functions
├── web/                      # Frontend application
│   ├── src/
│   │   ├── components/      # Vue components
│   │   ├── layouts/         # Page layouts
│   │   ├── pages/           # Application pages
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
- `POST /api/sign` - Sign a PDF document
- `POST /api/verify` - Verify a signed document
- `GET /api/health` - Health check
- `POST /api/auth/login` - User authentication

#### Protected Endpoints (require JWT)
- `POST /api/_/edit` - Edit document templates
- `GET /api/_/templates` - List signature templates
- `POST /api/_/upload` - Upload documents

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

## Support

For issues and questions:
- GitHub Issues: [https://github.com/shurco/gosign/issues](https://github.com/shurco/gosign/issues)
- Documentation: [Add documentation link]

## Roadmap

- [ ] Multi-language support
- [ ] Batch signing operations
- [ ] Advanced certificate templates
- [ ] Integration with external CA services
- [ ] Mobile application support
- [ ] PDF form filling automation

---

**Made with ❤️ for secure document signing**

