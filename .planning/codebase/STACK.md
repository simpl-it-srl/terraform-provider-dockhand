# Technology Stack

**Analysis Date:** 2026-03-07

## Languages

**Primary:**
- Go 1.24.0 - All provider logic (`main.go`, `internal/provider/`)

**Secondary:**
- HCL (Terraform) - Example configurations (`examples/`)
- Bash - Build and dev scripts (`scripts/`)

## Runtime

**Environment:**
- Go 1.24.0 (minimum required)

**Package Manager:**
- Go modules (`go mod`)
- Lockfile: `go.sum` (present and committed)

## Frameworks

**Core:**
- `github.com/hashicorp/terraform-plugin-framework` v1.17.0 - Terraform provider SDK (plugin protocol v6.0)
- `github.com/hashicorp/terraform-plugin-go` v0.29.0 - Low-level Terraform plugin protocol bindings

**Testing:**
- `github.com/hashicorp/terraform-plugin-testing` v1.14.0 - Acceptance test helpers for Terraform providers
- Go standard `testing` package - Unit tests

**Build/Dev:**
- `GNUmakefile` - Build, test, install, and dev workflow targets
- `scripts/build-packages.sh` - Cross-platform release packaging (darwin/linux/windows, amd64/arm64)
- `scripts/tf-dev.sh` - Local developer workflow using Terraform `dev_overrides`
- `scripts/build-mirror.sh` - Filesystem mirror builder for private distribution

## Key Dependencies

**Critical:**
- `github.com/hashicorp/terraform-plugin-framework` v1.17.0 - All resource/data source schema definitions and CRUD logic depend on this
- `github.com/hashicorp/terraform-plugin-go` v0.29.0 - Required by plugin-framework internals
- `github.com/hashicorp/terraform-plugin-testing` v1.14.0 - Acceptance tests require a live Dockhand instance

**Infrastructure (transitive):**
- `github.com/hashicorp/go-retryablehttp` v0.7.7 - Retry-capable HTTP (used by plugin internals, not directly by provider HTTP client)
- `github.com/hashicorp/hc-install` v0.9.2 - Terraform binary installer (used by testing framework)
- `github.com/hashicorp/terraform-exec` v0.24.0 - Running terraform CLI in tests
- `google.golang.org/grpc` v1.75.1 - gRPC transport for plugin protocol
- `google.golang.org/protobuf` v1.36.9 - Protobuf encoding for plugin protocol
- `golang.org/x/crypto` v0.45.0 - Crypto utilities (TLS, hashing)
- `github.com/ProtonMail/go-crypto` v1.1.6 - GPG crypto (used by build signing pipeline)
- `github.com/cloudflare/circl` v1.6.1 - Cryptographic primitives

## Configuration

**Environment (provider runtime):**
- `DOCKHAND_ENDPOINT` - Dockhand API base URL (required)
- `DOCKHAND_USERNAME` - Login username
- `DOCKHAND_PASSWORD` - Login password
- `DOCKHAND_MFA_TOKEN` - Optional MFA token
- `DOCKHAND_AUTH_PROVIDER` - Auth provider ID (defaults to `local`)
- `DOCKHAND_DEFAULT_ENV` - Default environment ID for API calls
- `DOCKHAND_ALLOW_UNAUTHENTICATED` - Bootstrap mode flag (`1`, `true`, `yes`)

**Environment (acceptance tests):**
- `DOCKHAND_TEST_ENDPOINT` - Live Dockhand instance URL
- `DOCKHAND_TEST_USERNAME` - Test user
- `DOCKHAND_TEST_PASSWORD` - Test password
- `DOCKHAND_TEST_DEFAULT_ENV` - Test environment ID (defaults to `"1"`)

**Build:**
- `terraform-registry-manifest.json` - Terraform Registry protocol version manifest (protocol 6.0)
- `GNUmakefile` - Main build entrypoint
- `go.mod` / `go.sum` - Module definition and lockfile

**CI Secrets (GitHub Actions):**
- `GPG_PRIVATE_KEY` - GPG key for signing release SHA256SUMS

## Platform Requirements

**Development:**
- Go 1.24.0+
- `zip` binary (for packaging)
- `terraform` or `tofu` CLI (for dev workflow via `scripts/tf-dev.sh`)
- Optional: `delve` debugger (provider supports `-debug` flag)

**Production:**
- Released as static Go binaries for: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`
- Published to Terraform Registry at `registry.terraform.io/simpl-it-srl/dockhand`
- Supports private distribution via filesystem mirror (`scripts/build-mirror.sh`)
- CGO disabled (`CGO_ENABLED=0`) — no native library dependencies at runtime

---

*Stack analysis: 2026-03-07*
