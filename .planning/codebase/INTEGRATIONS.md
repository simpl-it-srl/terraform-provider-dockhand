# External Integrations

**Analysis Date:** 2026-03-07

## APIs & External Services

**Dockhand API (primary target):**
- Dockhand - The sole external service this provider manages. All resources and data sources map to Dockhand REST API endpoints.
  - SDK/Client: Custom HTTP client — `internal/provider/client.go` (no third-party SDK; raw `net/http`)
  - Auth: Session cookie (`dockhand_session`) obtained via `POST /api/auth/login`
  - Base URL: configured via `DOCKHAND_ENDPOINT` env var or provider `endpoint` attribute
  - TLS: Enforces TLS 1.2 minimum; optional `insecure` flag to skip verification (dev only)
  - Timeout: 30-second HTTP timeout on auth requests

**Terraform Registry:**
- Provider published at `registry.terraform.io/simpl-it-srl/dockhand`
  - Manifest: `terraform-registry-manifest.json` declares protocol version `6.0`
  - Release artifacts signed with GPG for registry verification

## Data Storage

**Databases:**
- None — this is a stateless Terraform provider. All state is managed by Terraform's own state backend (configured by the end-user, not this provider).

**File Storage:**
- Local filesystem only — build artifacts written to `dist-local/` (gitignored); no persistent file storage in provider runtime.

**Caching:**
- None at runtime. Go module/build cache written to `.cache/` during local builds (gitignored).

## Authentication & Identity

**Auth Provider:**
- Dockhand session-cookie auth
  - Implementation: `internal/provider/auth.go` — `Login()` function POSTs credentials to `/api/auth/login`, extracts `dockhand_session` cookie, passes it as a `Cookie` header on all subsequent API requests
  - MFA: Optional MFA token support via `mfaToken` field in login payload
  - Auth providers: Configurable via `auth_provider` field (defaults to `"local"`); Dockhand supports pluggable auth providers (data source `dockhand_auth_providers` lists available ones)
  - Bootstrap mode: `allow_unauthenticated = true` bypasses login for first-install flows

## Monitoring & Observability

**Error Tracking:**
- None — no external error tracking service integrated.

**Logs:**
- Standard Go `log` package used only in `main.go` for fatal startup errors.
- Terraform plugin framework's structured logging via `github.com/hashicorp/terraform-plugin-log` is available as a transitive dependency but not directly invoked in provider code.
- API errors are surfaced as Terraform diagnostics (provider-framework pattern) rather than logs.

## CI/CD & Deployment

**Hosting:**
- Terraform Registry: `registry.terraform.io/simpl-it-srl/dockhand` (public)
- Private distribution also supported via filesystem mirror (`scripts/build-mirror.sh`, `scripts/build-packages.sh`)

**CI Pipeline:**
- GitHub Actions
  - `.github/workflows/go-ci.yml` — runs on pull_request and push to `main`; steps: `gofmt` check, `go mod tidy` verification, `go test ./...`, `go build ./...`
  - `.github/workflows/release-artifacts.yml` — runs on `v*` tag push or manual dispatch; steps: cross-platform build via `scripts/build-packages.sh`, GPG signing of SHA256SUMS, GitHub Release publish via `softprops/action-gh-release@v2`
  - Uses `crazy-max/ghaction-import-gpg@v6` for GPG key import during release

## Environment Configuration

**Required env vars (provider runtime):**
- `DOCKHAND_ENDPOINT` — Dockhand API base URL (required; no default)
- `DOCKHAND_USERNAME` + `DOCKHAND_PASSWORD` — Login credentials (required unless `DOCKHAND_ALLOW_UNAUTHENTICATED=true`)

**Optional env vars (provider runtime):**
- `DOCKHAND_MFA_TOKEN` — MFA one-time token
- `DOCKHAND_AUTH_PROVIDER` — Auth provider ID (default: `local`)
- `DOCKHAND_DEFAULT_ENV` — Default Dockhand environment ID sent as `?env=` query param

**Required env vars (acceptance tests):**
- `DOCKHAND_TEST_ENDPOINT` — Live Dockhand instance
- `DOCKHAND_TEST_USERNAME` / `DOCKHAND_TEST_PASSWORD` — Test credentials
- `DOCKHAND_TEST_DEFAULT_ENV` — Target environment (defaults to `"1"`)

**Secrets location:**
- GitHub Actions secrets: `GPG_PRIVATE_KEY` (release signing only)
- All runtime credentials passed via environment variables or Terraform provider config block (marked `Sensitive: true` for password and MFA token)

## Webhooks & Callbacks

**Incoming:**
- None — the provider does not expose any HTTP endpoints.

**Outgoing:**
- Dockhand git repository webhooks: managed as a Terraform resource (`dockhand_git_repository`) with `webhook_enabled` and `webhook_secret` attributes. The webhook URL is configured in Dockhand and called by the user's Git host (e.g. GitHub/GitLab); the provider only manages the registration of the webhook, not the HTTP handler.
- Git stack webhook action: `dockhand_git_stack_webhook_action` resource triggers webhook-based redeploy actions on Dockhand git stacks.

---

*Integration audit: 2026-03-07*
