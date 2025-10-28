# API Authentication & Rate Limiting

## Overview

The system supports two authentication methods:
1. **JWT tokens** - for web interface users
2. **API keys** - for external integrations and automation

Both methods are protected by rate limiting to prevent abuse.

## JWT Authentication

### Getting a Token

```bash
POST /auth/signin
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password"
}
```

### Using a Token

```bash
Authorization: Bearer <jwt_token>
```

### Characteristics
- Lifetime: 10 minutes
- Automatic refresh via refresh token (TODO)
- Contains: user_id, email, name

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

## Roadmap

- [ ] Refresh tokens for JWT
- [ ] Redis backend for rate limiting
- [ ] IP whitelist for API keys
- [ ] Webhook signatures with API keys
- [ ] Audit log for API key usage
- [ ] Dashboard with usage statistics

