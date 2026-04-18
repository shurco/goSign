# pkg/ — AGENTS.md

Reusable libraries that can be imported by any package inside this repo (and, in principle, by external code). Keep them **dependency-light** and **framework-agnostic**. No Fiber, no `internal/*` imports.

## Subpackages

| Package             | Responsibility                                                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `appdir/`           | Resolves data/config directories (`appdir.Base()`, `LcPages()`, `LcSigned()`, …) relative to the executable.                                |
| `pdf/`              | High-level PDF manipulation: `append`, `certificate`, `extract`, `fill`, `fonts`, `page`, `render`.                                         |
| `pdf/sign/`         | PKCS#7 / PAdES signing (TSA, revocation data, DocMDP). Network-dependent tests **must** guard with `testing.Short()`.                       |
| `pdf/verify/`       | Verification and validation of signed PDFs.                                                                                                 |
| `pdf/revocation/`   | OCSP + CRL fetching/embedding helpers.                                                                                                      |
| `security/cert/`    | X.509 certificate builder, CRL issuance, key helpers. Error sentinel: `ErrDecodeCACert`.                                                    |
| `security/password/`| Bcrypt password hashing (`GeneratePassword`, `ComparePasswords`) + `NewToken` opaque identifier + `RandomString` crypto-rand ID.            |
| `storage/`          | `interface.go` defines the storage contract. Implementations: `local`, `s3` (MinIO).                                                        |
| `storage/postgres/` | `pgxpool` factory.                                                                                                                          |
| `storage/redis/`    | Redis v9 wrapper (`redis.New`, `redis.Conn`).                                                                                               |
| `notification/`     | Provider-based notifier (`Email`, `SMS`, `Webhook`) + `Worker` + Go templates.                                                              |
| `webhook/`          | Signed webhook dispatcher with retries.                                                                                                     |
| `geolocation/`      | GeoLite2 reader (`Service`, `Reload`) + `ExtractFromTarGz` / `ExtractFromGzip` helpers.                                                     |
| `logging/`          | Zerolog wrapper (`logging.Log`).                                                                                                            |
| `utils/`            | Small pure helpers: `maputil`, `date_utils`, `file_utils`, `webutil` (HTTP response envelope).                                              |
| `fixtures/`         | Embedded test documents used by other packages.                                                                                             |

## Rules specific to `pkg/`

- **Zero coupling to `internal/`.** If you need configuration, take it as a typed struct argument or an interface.
- **Return errors, never log-and-swallow.** Callers decide how to surface them.
- **API stability.** Breaking changes must update every caller in this repo in the same change (no `Deprecated` stubs).
- **Security helpers** (`security/cert`, `security/password`) must have direct unit tests; they're the blast radius for authentication and signing.
- **Network-dependent tests** (TSA, OCSP, MaxMind) must:
  - be gated behind `if testing.Short() { t.Skip(...) }`, **and**
  - have a sensible timeout (`context.WithTimeout` or `http.Client.Timeout`).

## Conventions

- Package-level `Err…` sentinels use lowercase messages (Go convention); wrap with `%w` when re-returned.
- Constructors return `(*T, error)` when initialisation can fail (e.g. opening a database file). Idempotent setup returns only `*T`.
- Avoid `panic` outside of clearly documented "programmer error" branches (see `password.RandomString`).
- Keep test fixtures inside the package they belong to (e.g. `pkg/pdf/testdata/`).

## Tests

- `go test -short -race ./pkg/...` should pass without network access or external services.
- `go test -race ./pkg/...` (full mode) may need internet (TSA, OCSP, MaxMind); run it as part of nightly CI, not on every commit.
