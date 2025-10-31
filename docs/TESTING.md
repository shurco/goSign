# Testing Guide

## Overview

This document describes the testing strategy and guidelines for the goSign project, with a focus on authentication testing.

## Test Structure

### Test Types

1. **Unit Tests** - Test individual functions and methods in isolation
2. **Integration Tests** - Test complete workflows with database and external services
3. **Benchmark Tests** - Measure performance of critical operations
4. **Security Tests** - Verify security features and prevent vulnerabilities

## Running Tests

### Authentication Tests

#### Quick Start

```bash
# Run all working tests (without database)
go test ./internal/handlers/public/ -run "TestPassword|TestValidation|TestRefreshToken" -v

# Run benchmarks
go test ./internal/handlers/public/ -bench=. -benchmem

# Use convenient script
./scripts/test_auth.sh help
```

#### Test Categories

**Password Security Tests** (No DB required)
```bash
go test ./internal/handlers/public/ -run TestPasswordHashing -v
```
- Password hashing with bcrypt
- Hash uniqueness (salt verification)
- Password verification accuracy

**Validation Tests** (No DB required)
```bash
go test ./internal/handlers/public/ -run TestValidation -v
```
- Email format validation
- Required fields validation
- Data structure validation

**Token Tests** (No DB required)
```bash
go test ./internal/handlers/public/ -run TestRefreshToken -v
```
- Token creation and validation
- Token expiration handling
- Invalid token rejection

**Integration Tests** (Requires DB)
```bash
# Setup database first
./scripts/migration dev up

# Run integration tests
go test ./internal/handlers/public/... -v
```
- SignUp flow
- SignIn flow
- Password reset flow
- Email verification

### Using Test Script

The project includes a convenient test runner:

```bash
# Show help
./scripts/test_auth.sh help

# Run unit tests only
./scripts/test_auth.sh unit

# Generate coverage report
./scripts/test_auth.sh coverage

# Run benchmarks
./scripts/test_auth.sh bench

# Run specific test
./scripts/test_auth.sh specific TestPasswordHashing

# Run all tests
./scripts/test_auth.sh all
```

## Test Coverage

### Current Coverage

✅ **Implemented and Passing:**
- Password hashing and verification (100%)
- Data validation helpers (100%)
- Token creation and validation (80%)
- User role constants (100%)

⏳ **Implemented, Requires Database:**
- SignUp endpoint (integration)
- SignIn endpoint (integration)
- Password reset flow (integration)
- Email verification (integration)

🔄 **Planned:**
- 2FA flow tests
- OAuth integration tests
- Rate limiting tests
- Session management tests

### Coverage Goals

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: All critical paths
- **Security Tests**: All attack vectors

## Performance Benchmarks

### Current Benchmarks

```
BenchmarkPasswordHashing      ~1.5ms/op  5KB/op  10 allocs/op
BenchmarkPasswordVerification ~1.6ms/op  5KB/op  11 allocs/op
```

### Performance Targets

- Password operations: < 100ms
- Token validation: < 10ms
- API response time: < 200ms

## Test Data

### Test Users (fixtures/migration)

```
Admin:  admin@gosign.local / admin123 (role 3)
User 1: user1@gosign.local / user123  (role 1)
User 2: user2@gosign.local / user234  (role 1)
```

### Test Accounts

Load test data:
```bash
./scripts/migration dev up
```

## Writing Tests

### Test Structure

Follow the project's test story pattern:

```go
// Test Story: Feature Description
// As a [user type], I want to [action]
// so that [benefit]

func TestFeatureName(t *testing.T) {
    t.Run("scenario description", func(t *testing.T) {
        // Given: Setup test conditions
        // When: Perform action
        // Then: Assert expected results
    })
}
```

### Example Test

```go
func TestPasswordHashing(t *testing.T) {
    t.Run("password hashing and verification", func(t *testing.T) {
        // Given: a plain text password
        plainPassword := "SecureP@ssw0rd123"

        // When: hashing the password
        hashedPassword := password.GeneratePassword(plainPassword)

        // Then: should verify correctly
        assert.True(t, password.ComparePasswords(hashedPassword, plainPassword))
    })
}
```

### Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Test both happy and error paths**
3. **Use meaningful test names** describing the scenario
4. **Keep tests independent** (no shared state)
5. **Clean up resources** in defer or t.Cleanup()
6. **Use testify/assert** for better error messages
7. **Document test stories** in separate .md files

## Security Testing

### Tested Security Features

✅ **Password Security**
- Bcrypt hashing with proper cost
- Salt uniqueness
- Weak password rejection
- Password complexity requirements

✅ **Email Enumeration Prevention**
- Consistent responses for forgot password
- Same timing for existing/non-existing emails

✅ **Token Security**
- Token validation
- Expiration handling
- Invalid token rejection
- Token rotation

### Security Test Examples

```go
func TestSecurityFeatures(t *testing.T) {
    t.Run("password complexity requirements", func(t *testing.T) {
        weakPasswords := []string{"123", "password", "abc123"}
        
        for _, pwd := range weakPasswords {
            // Test that weak passwords are rejected
            result := validatePassword(pwd)
            assert.False(t, result, "Weak password should be rejected")
        }
    })
}
```

## Integration Testing

### Prerequisites

1. **PostgreSQL** running on localhost:5432
2. **Redis** running on localhost:6379
3. **Migrations** applied: `./scripts/migration dev up`

### Running Integration Tests

```bash
# Full integration test suite
go test ./internal/handlers/public/... -tags=integration -v

# With coverage
go test ./internal/handlers/public/... -tags=integration -cover -v
```

### Integration Test Example

```go
//go:build integration
// +build integration

func TestSignUpIntegration(t *testing.T) {
    // Requires database connection
    // Tests full signup flow including database operations
}
```

## Continuous Integration

### CI Pipeline

Tests run automatically on:
- Pull requests
- Commits to main branch
- Release tags

### CI Test Stages

1. **Lint** - Code quality checks
2. **Unit Tests** - Fast tests without dependencies
3. **Integration Tests** - Tests with database
4. **Benchmarks** - Performance regression detection

## Troubleshooting

### Common Issues

**Tests timeout**
- Increase timeout: `go test -timeout 60s`
- Check database connection
- Verify Redis is running

**Database connection errors**
- Run migrations: `./scripts/migration dev up`
- Check PostgreSQL is running
- Verify connection string

**Import errors**
- Run: `go mod tidy`
- Check Go version: `go version` (requires 1.22+)

## Additional Resources

- [Test Story Documentation](../internal/handlers/public/AUTH_TEST_STORY.md)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Testify Documentation](https://github.com/stretchr/testify)

## Contributing

When adding new features:

1. Write tests first (TDD approach)
2. Ensure all tests pass
3. Add test story documentation
4. Update this guide if needed
5. Maintain 80%+ coverage

