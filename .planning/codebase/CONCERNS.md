# CONCERNS

Technical debt, known issues, and areas of concern in the codebase.

---

## Critical / High Priority

### `client.go` is 2,159 lines
**File:** `internal/dockhand/client.go`
All API types, payloads, and HTTP methods live in a single file with no sub-packages. This makes navigation, testing, and maintenance difficult. Should be split into domain-specific files (e.g., `client_stack.go`, `client_registry.go`, etc.).

### Stack lookup requires full list scan
**File:** `internal/dockhand/client.go` — `GetStackByName()`
Fetches all stacks and searches in memory. No by-name API endpoint is used. This will degrade at scale if the Dockhand instance manages many stacks.

### Unstable stack list API shape
**File:** `internal/dockhand/client.go` — `parseStacks()`
Handles both array and `{"stacks": [...]}` envelope shapes at runtime with a silent empty-result fallback. If the API changes shape unexpectedly, failures are silent rather than surfaced as errors.

### SSH keys and TLS client certs stored in Terraform state
**Files:** `internal/dockhand/resource_git_repository.go`, provider config
Sensitive credentials (SSH private keys, TLS client certificates) are stored in Terraform state. Requires encrypted state backend (e.g., S3 with SSE, Terraform Cloud). No documentation warning exists.

---

## Medium Priority

### Stack Read does not refresh compose path
**File:** `internal/dockhand/resource_stack.go`
Drift goes undetected if the compose file changes out-of-band. The Read function doesn't re-fetch the current compose path from the API, only confirms the stack exists.

### Registry create/update uses `map[string]any`
**File:** `internal/dockhand/resource_registry.go`
The only resource without a typed payload struct. Uses untyped maps, making it easier to introduce subtle payload errors without compile-time catching.

### Container delete busy-waits 30 seconds
**File:** `internal/dockhand/resource_container.go`
Container deletion includes a 30-second `time.Sleep` plus a 5-retry loop adding ~6 seconds. This makes `terraform destroy` slow and the logic is fragile.

### `insecure = true` has no runtime warning
**File:** Provider configuration
Setting `insecure = true` silently disables TLS verification with no log warning or user-visible indication. This is a security risk in production environments.

---

## Testing Gaps

### Minimal unit test coverage
Only 3 unit test functions exist across the codebase:
- `internal/dockhand/client_test.go` — 1 test
- `internal/dockhand/resource_container_update_action_test.go` — 2 tests
- `internal/dockhand/data_source_schedules_executions_test.go` — 1 test

Most resources have **zero** unit test coverage.

### Acceptance tests require live Dockhand instance
All 13 acceptance tests require a live Dockhand instance. Many resources have no acceptance test at all:
- `resource_network`
- `resource_volume`
- `resource_image`
- `resource_git_repository`
- `resource_environment`
- `resource_schedule`

This makes CI/CD validation difficult without a dedicated test environment.

---

## Low Priority / Future Work

- No structured logging — uses raw `log.Printf` throughout
- No pagination support in list endpoints (could miss items on large instances)
- Error messages from API are sometimes swallowed or not propagated clearly to Terraform plan output
