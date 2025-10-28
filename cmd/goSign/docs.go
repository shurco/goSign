package main

// @title goSign API
// @version 1.0
// @description API for goSign digital document signing platform
// @description
// @description Key features:
// @description - Multi-signer workflow with email notifications
// @description - 12 field types for filling (signature, initials, text, date, number, checkbox, radio, select, multi_select, file, image, cells)
// @description - Webhook integrations for events
// @description - Flexible storage (local, S3, GCS, Azure)
// @description - Audit trail for each document

// @contact.name goSign Support
// @contact.email support@gosign.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8088
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token in format: Bearer {token}

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API key for service authentication

// @tag.name Auth
// @tag.description Authentication and authorization

// @tag.name Submissions
// @tag.description Submission management (documents for signing)

// @tag.name Submitters
// @tag.description Submitter management (document signers)

// @tag.name Templates
// @tag.description Document template management

// @tag.name Webhooks
// @tag.description Webhook integration management

// @tag.name Settings
// @tag.description Application settings (SMTP, Storage, Branding)

