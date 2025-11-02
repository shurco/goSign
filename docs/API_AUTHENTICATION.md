# API Authentication & Rate Limiting

## Overview

The system supports two authentication methods:
1. **JWT tokens** - for web interface users
2. **API keys** - for external integrations and automation

Both methods are protected by rate limiting to prevent abuse.

## JWT Authentication

### User Registration

```bash
POST /auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecureP@ssw0rd123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Registration successful. Please check your email to verify your account.",
  "data": {
    "user_id": "uuid"
  }
}
```

### Getting a Token

```bash
POST /auth/signin
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password",
  "code": "123456"  // Optional: 2FA code if enabled
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "token_type": "Bearer"
  }
}
```

### Refreshing a Token

```bash
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGc..."
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "token_type": "Bearer"
  }
}
```

### Using a Token

```bash
Authorization: Bearer <jwt_token>
```

### Characteristics
- **Access Token Lifetime**: 10 minutes
- **Refresh Token Lifetime**: 7 days
- **Token Refresh**: Automatic refresh via `/auth/refresh` endpoint
- **Contains**: user_id, email, name, organization_id (if in organization context)

### Organization Context
When a user switches to an organization context, the JWT token includes an `organization_id` field. This enables multi-tenant data isolation:

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "organization_id": "org_uuid",  // Included when in org context
  "exp": 1234567890
}
```

To switch organization context:
```bash
POST /api/v1/organizations/{organization_id}/switch
Authorization: Bearer <jwt_token>
```

This updates your JWT token to include the organization_id for subsequent requests.

## API Key Authentication

### Creating an API Key

```bash
POST /api/v1/apikeys
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Production Integration",
  "expires_at": 1735689600  // Unix timestamp (optional)
}
```

**Response (save the key - it won't be shown again!):**
```json
{
  "success": true,
  "message": "API key created successfully. Save this key securely - it won't be shown again.",
  "data": {
    "key": "abc123...xyz",  // Save this value!
    "api_key": {
      "id": "key_123",
      "name": "Production Integration",
      "enabled": true,
      "expires_at": "2025-01-01T00:00:00Z",
      "created_at": "2024-12-01T00:00:00Z"
    }
  }
}
```

### Using an API Key

```bash
X-API-Key: abc123...xyz
```

### Managing API Keys

#### List Keys
```bash
GET /api/v1/apikeys
Authorization: Bearer <jwt_token>
```

#### Disable Key
```bash
PUT /api/v1/apikeys/{id}/disable
Authorization: Bearer <jwt_token>
```

#### Enable Key
```bash
PUT /api/v1/apikeys/{id}/enable
Authorization: Bearer <jwt_token>
```

#### Delete Key
```bash
DELETE /api/v1/apikeys/{id}
Authorization: Bearer <jwt_token>
```

## Rate Limiting

### Standard Limits

**API endpoints (most):**
- 100 requests / minute per user/API key
- Grouping by `user_id` or `api_key_id`
- Fallback to IP address for public endpoints

**Sensitive operations (settings, apikeys):**
- 10 requests / minute per user/API key
- Applied to settings modification operations

### Response When Limit Exceeded

```json
{
  "success": false,
  "message": "Rate limit exceeded. Please try again later."
}
```

**HTTP Status:** 429 Too Many Requests

### Usage Examples

#### With JWT Token
```bash
curl -H "Authorization: Bearer <token>" \
     https://api.example.com/api/v1/submissions
```

#### With API Key
```bash
curl -H "X-API-Key: abc123...xyz" \
     https://api.example.com/api/v1/submissions
```

## Security Best Practices

### API Keys
1. **Store securely** - use secrets manager
2. **Rotation** - regularly create new keys and delete old ones
3. **Expiration** - set expiration dates for automatic rotation
4. **Naming** - use descriptive names for identification
5. **Monitoring** - track `last_used_at` to identify unused keys

### JWT Tokens
1. **HTTPS only** - never transmit over unencrypted connections
2. **Short-lived** - current lifetime is 10 minutes
3. **Secure storage** - use httpOnly cookies in production
4. **Refresh strategy** - refresh tokens implementation planned

## Technical Details

### API Key Generation
- Length: 32 bytes (43 characters base64)
- Algorithm: crypto/rand
- Storage: SHA256 hash in database
- Format: base64 URL-safe encoding

### Rate Limiter
- Engine: Fiber built-in middleware
- Storage: In-memory (Redis planned for clusters)
- Key format:
  - `apikey:{key_id}` for API keys
  - `user:{user_id}` for JWT
  - IP address for unauthenticated users

### Authentication Flow

```
Request → Check X-API-Key header
       ├─ Found → Validate → Set auth context → Next
       └─ Not found → Check Authorization header
                    ├─ Found → Validate JWT → Set auth context → Next
                    └─ Not found → 401 Unauthorized
```

### Auth Context Structure

```go
type AuthContext struct {
    Type      AuthType  // "jwt" | "api_key"
    UserID    string
    AccountID string
    Email     string    // JWT only
    Name      string    // JWT only
}
```

## Migration Notes

### For Existing Users
- JWT authentication works as before
- API keys are a new feature, requires creation through UI/API

### For Developers
- Use `middleware.GetAuthContext(c)` to get authentication information
- Check `auth.Type` to differentiate between JWT and API keys
- Rate limiting is applied automatically to protected routes

## Password Management

### Forgot Password

```bash
POST /auth/password/forgot
Content-Type: application/json

{
  "email": "user@example.com"
}
```

**Response:** Always returns success (to prevent email enumeration)

### Reset Password

```bash
POST /auth/password/reset
Content-Type: application/json

{
  "token": "reset-token-from-email",
  "new_password": "NewSecureP@ssw0rd123"
}
```

## Email Verification

### Verify Email

```bash
GET /auth/verify-email?token=<verification-token>
```

**Response:**
```json
{
  "success": true,
  "message": "Email verified successfully"
}
```

## Two-Factor Authentication (2FA)

### Enable 2FA

```bash
POST /auth/2fa/enable
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "qr_code": "data:image/png;base64,...",
    "secret": "JBSWY3DPEHPK3PXP",
    "backup_codes": ["123456", "234567", ...]
  }
}
```

### Verify 2FA

```bash
POST /auth/2fa/verify
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "code": "123456"
}
```

### Disable 2FA

```bash
POST /auth/2fa/disable
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "password": "current-password"
}
```

## OAuth Authentication

### Google OAuth

1. **Initiate OAuth:**
```bash
GET /auth/oauth/google
```

2. **OAuth Callback:**
```bash
GET /auth/oauth/google/callback?code=<auth-code>&state=<state>
```

### GitHub OAuth

1. **Initiate OAuth:**
```bash
GET /auth/oauth/github
```

2. **OAuth Callback:**
```bash
GET /auth/oauth/github/callback?code=<auth-code>&state=<state>
```

**Note:** OAuth providers must be configured in the application settings with client ID and secret.

## Sign Out

```bash
POST /auth/signout
Authorization: Bearer <jwt_token>
```

Invalidates the refresh token and clears session.

## Roadmap

- [ ] Redis backend for rate limiting (currently in-memory)
- [ ] IP whitelist for API keys
- [ ] Webhook signatures with API keys
- [ ] Audit log for API key usage
- [ ] Dashboard with usage statistics
- [ ] Additional OAuth providers (Microsoft, Apple)

