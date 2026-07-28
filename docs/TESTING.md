# Testing Guide

## Overview

This document describes the testing strategy and guidelines for the goSign project.

## Test Structure

### Test Types

1. **Unit Tests** - Test individual functions and methods in isolation (no external services)
2. **Database Tests** - Handler and service tests running against an isolated PostgreSQL clone per test (via [pgtestdb](https://github.com/peterldowns/pgtestdb))
3. **Benchmark Tests** - Measure performance of critical operations
4. **Frontend Tests** - Vitest unit tests in `web/private`

## Running Tests

### Backend (Go)

```bash
# Everything (DB tests are skipped automatically if the test DB is down)
go test ./...

# Single package
go test ./internal/handlers/api/ -v

# Benchmarks
go test ./internal/handlers/public/ -bench=. -benchmem
```

### Test Database

Database-backed tests connect to a dedicated PostgreSQL at `localhost:5453` (user `postgres`, password `password`, see `internal/testutil/db.go`). When it is not reachable, those tests are **skipped** with a `skip: pgtestdb unavailable` message — unit tests still run.

Start it locally with Docker:

```bash
docker run -d --name gosign-test-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5453:5432 postgres:17-alpine
```

Each test gets its own database clone with:
- all migrations from `migrations/` applied
- fixtures from `fixtures/migration/` applied (separate goose history table)

No manual migration step is needed for tests.

### Frontend (Svelte)

```bash
cd web/private
bun run test    # vitest run
bun run check   # svelte-check
bun run lint    # prettier + eslint
```

## Test Data

Fixtures live in `fixtures/migration/` and are applied automatically in Go tests. For the dev server, load them with:

```bash
./scripts/migration dev up
```

Test users (from fixtures):

```
Admin:  admin@gosign.local / admin123 (role 3)
User 1: user1@gosign.local / user123  (role 1)
User 2: user2@gosign.local / user234  (role 1)
```

## Writing Tests

### Test Structure

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

### Database Test Example

```go
func TestSomethingWithDB(t *testing.T) {
    pool := testutil.NewTestDB(t) // isolated clone, auto-skip without test DB
    db := queries.New(pool)
    // ...
}
```

### Best Practices

1. **Use table-driven tests** for multiple scenarios
2. **Test both happy and error paths**
3. **Use meaningful test names** describing the scenario
4. **Keep tests independent** (no shared state; pgtestdb gives each test its own DB)
5. **Clean up resources** in defer or t.Cleanup()
6. **Use testify/assert** for better error messages

## Security Testing

Covered security features:

- **Password Security** - bcrypt hashing, salt uniqueness, weak password rejection
- **Email Enumeration Prevention** - consistent responses for forgot password
- **Token Security** - validation, expiration handling, invalid token rejection, rotation

## Continuous Integration

GitHub Actions workflows:

- **test.yml** - golangci-lint + `go test ./...` (with the 5453 test database service) and frontend `check`/`lint`/`test`
- **analyze.yml** - CodeQL static analysis
- **release.yml** - goreleaser builds on tags

## Troubleshooting

**Database tests are skipped**
- Start the test DB: see "Test Database" above (`localhost:5453`)

**Tests timeout**
- Increase timeout: `go test -timeout 60s`
- Check the test database is healthy: `docker logs gosign-test-db`

**Import errors**
- Run: `go mod tidy`
- Check Go version against `go.mod`

## Additional Resources

- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
- [Testify Documentation](https://github.com/stretchr/testify)
- [pgtestdb](https://github.com/peterldowns/pgtestdb)
