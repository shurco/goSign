# Authentication Test Story

## Feature: User Authentication and Authorization

As a **user**, I want to **register, login, and manage my account securely**, so that **I can access the application with proper authentication**.

## Test Coverage Summary

### Authentication Flow (auth_comprehensive_test.go)

#### Sign Up Tests
- ✅ Successful user registration with valid data
- ✅ Reject weak passwords (< 8 characters)
- ✅ Reject invalid email formats
- ✅ Reject missing required fields (email, password, first_name, last_name)
- ✅ Password complexity requirements enforcement

#### Sign In Tests  
- ✅ Reject invalid credentials (non-existent user)
- ✅ Reject empty password
- ✅ Reject malformed JSON requests
- ✅ Email enumeration prevention (same response for existing/non-existing)

#### Password Management Tests
- ✅ Request password reset with valid email
- ✅ Reject password reset with invalid email format
- ✅ Reject password reset with invalid token
- ✅ Reject weak passwords on reset
- ✅ Password hashing produces unique salted hashes
- ✅ Password verification works correctly

#### Email Verification Tests
- ✅ Reject email verification with invalid token
- ✅ Reject email verification without token parameter

#### Token Management (auth_test.go)
- ✅ Create and validate refresh tokens
- ✅ Refresh token expiration handling
- ✅ Reject invalid token formats
- ✅ Reject malformed JWT tokens
- ✅ Access token creation and validation
- ✅ Token rotation generates unique tokens
- ✅ RefreshToken endpoint returns new tokens
- ✅ RefreshToken endpoint rejects invalid tokens

### Security Features

#### Password Security
- ✅ Bcrypt hashing with proper cost
- ✅ Salt uniqueness (same password = different hashes)
- ✅ Password verification accuracy
- ✅ Weak password rejection

#### Email Enumeration Prevention
- ✅ Consistent responses for forgot password
- ✅ Same status codes for existing/non-existing emails

#### Role System
- ✅ User role (1) has basic permissions
- ✅ Moderator role (2) for future use
- ✅ Admin role (3) has highest privileges
- ✅ Roles are properly ordered (1 < 2 < 3)

## Test Metrics

### Coverage Goals
- **Unit Tests**: 80%+ coverage for auth handlers
- **Integration Tests**: Key flows (signup → verify → signin)
- **Security Tests**: All authentication vulnerabilities covered

### Performance Benchmarks
- Password hashing: < 100ms per operation
- Password verification: < 50ms per operation
- Token validation: < 10ms per operation

## Test Data

### Valid Test Users
```go
Admin:  admin@gosign.local / admin123 (role 3)
User 1: user1@gosign.local / user123  (role 1)
User 2: user2@gosign.local / user234  (role 1)
```

### Invalid Scenarios Tested
- Empty credentials
- Malformed JSON
- Invalid email formats
- Weak passwords
- Non-existent users
- Expired tokens
- Invalid tokens

## Future Test Enhancements

### Planned Tests
1. **2FA Flow**
   - Enable 2FA
   - Verify TOTP codes
   - Disable 2FA
   - 2FA during signin

2. **OAuth Integration**
   - Google OAuth flow
   - GitHub OAuth flow
   - OAuth callback handling

3. **Session Management**
   - Multiple active sessions
   - Session termination
   - Session timeout

4. **Rate Limiting**
   - Login attempt limits
   - Password reset limits
   - Token refresh limits

## Running Tests

### Unit Tests
```bash
# Run all auth tests
go test ./internal/handlers/public/... -v

# Run specific test
go test ./internal/handlers/public/ -run TestSignUpFlow -v

# Run with coverage
go test ./internal/handlers/public/... -cover -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Benchmark Tests
```bash
# Run benchmarks
go test ./internal/handlers/public/ -bench=. -benchmem

# Specific benchmark
go test ./internal/handlers/public/ -bench=BenchmarkPasswordHashing
```

### Integration Tests
```bash
# Run with database
./scripts/migration dev up
go test ./internal/handlers/public/... -tags=integration -v
```

## Test Maintenance

### When to Update Tests
- ✅ Adding new auth endpoints
- ✅ Modifying authentication logic
- ✅ Changing password requirements
- ✅ Updating token expiration
- ✅ Adding new security features

### Test Review Checklist
- [ ] All happy paths covered
- [ ] All error paths covered
- [ ] Edge cases identified and tested
- [ ] Security vulnerabilities tested
- [ ] Performance benchmarks passing
- [ ] Integration tests passing

## Security Considerations

### Tested Attack Vectors
- ✅ SQL Injection (via parameterized queries)
- ✅ Password enumeration (timing attacks prevented)
- ✅ Weak password acceptance
- ✅ Token forgery
- ✅ Session hijacking (via token rotation)

### Not Yet Tested (TODO)
- ⏳ Brute force attack mitigation
- ⏳ Account lockout after failed attempts
- ⏳ CSRF token validation
- ⏳ XSS prevention in responses

## Dependencies

### External Services
- PostgreSQL (user storage)
- Redis (token blacklist, sessions)

### Test Libraries
- `testing` (standard library)
- `testify/assert` (assertions)
- `testify/require` (requirements)
- `fiber` (HTTP testing)

## Notes

- Tests use mock data when possible to avoid database dependencies
- Integration tests require running database and Redis
- Benchmarks should be run on production-like hardware
- Security tests should be reviewed by security team

