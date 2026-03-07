# Testing

**Analysis Date:** 2026-03-07

## Framework

- **Unit tests:** Standard Go `testing` package
- **Acceptance tests (TF framework):** `github.com/hashicorp/terraform-plugin-testing` — uses `resource.Test` with `TestCase` and `TestStep`
- **Acceptance tests (raw client):** Standard `testing` package calling `internal/provider` client methods directly

## Test Commands

```bash
# Run all tests (unit only in CI — acceptance tests skip without env vars)
go test ./...
make test

# Run with verbose output
go test -v ./internal/provider/...

# Run acceptance tests (requires live Dockhand instance)
DOCKHAND_TEST_ENDPOINT=http://dockhand.local \
DOCKHAND_TEST_USERNAME=admin \
DOCKHAND_TEST_PASSWORD=secret \
go test -v -run TestAcc ./internal/provider/...
```

## Test File Inventory

| File | Type | Tests |
|------|------|-------|
| `client_test.go` | Unit | `TestNewClientAllowsEmptySessionCookie` |
| `resource_container_update_action_test.go` | Unit | `TestBuildContainerUpdatePayload` (2 sub-tests) |
| `data_source_schedules_executions_test.go` | Unit | `TestMatchesScheduleExecutionFilters` |
| `acc_test.go` | Shared helpers | `testAccEnv`, `testAccDefaultEnv`, `testAccLoginSessionCookie` |
| `resource_user_acc_test.go` | Raw-client acc | `TestAccUserResource` — CRUD lifecycle via `*Client` |
| `resource_user_tf_acc_test.go` | TF acc | `TestAccUserResourceTerraform` — full TF plan/apply lifecycle |
| `resource_container_actions_tf_acc_test.go` | TF acc | `TestAccContainerRenameActionTerraform` |
| `resource_schedule_run_action_tf_acc_test.go` | TF acc | `TestAccScheduleRunActionTerraform` |
| `resource_container_file_git_stack_deploy_tf_acc_test.go` | TF acc | `TestAccContainerFileDirectoryResourceTerraform` + git stack deploy |
| `resource_new_surfaces_tf_acc_test.go` | TF acc | Multiple tests for newer resources |

## Unit Test Structure

Unit tests live alongside the code they test (`_test.go` suffix, same package `provider`).

**Example — pure function test:**
```go
func TestMatchesScheduleExecutionFilters(t *testing.T) {
    item := scheduleExecutionItemResponse{...}
    if !matchesScheduleExecutionFilters(item, "system_cleanup", "2", "success", "2") {
        t.Fatalf("expected item to match all filters")
    }
}
```

**Example — table-style sub-tests:**
```go
func TestBuildContainerUpdatePayload(t *testing.T) {
    t.Run("typed fields only", func(t *testing.T) { ... })
    t.Run("payload json overrides typed", func(t *testing.T) { ... })
}
```

## Acceptance Test Structure

### Shared Setup (`acc_test.go`)

All acceptance tests call `testAccEnv(t)` which reads required env vars and calls `t.Skip()` if they are absent — so the tests are silently skipped in CI unless a live instance is configured.

**Required env vars:**
```
DOCKHAND_TEST_ENDPOINT     URL of the Dockhand instance
DOCKHAND_TEST_USERNAME     Admin username
DOCKHAND_TEST_PASSWORD     Admin password
```

**Optional env vars:**
```
DOCKHAND_TEST_DEFAULT_ENV         Dockhand environment ID (default: "1")
DOCKHAND_TEST_SCHEDULE_TYPE       For schedule tests
DOCKHAND_TEST_SCHEDULE_ID         For schedule tests
DOCKHAND_TEST_FILE_CONTAINER_ID   For container file tests
```

### TF Framework Acceptance Tests (`*_tf_acc_test.go`)

Use `resource.Test(t, resource.TestCase{...})` with:
- `ProtoV6ProviderFactories` wired to `New("test")()`
- Provider configured via `t.Setenv` for env-var-based config
- `Steps []resource.TestStep` each containing an HCL `Config` and `Check` functions

**Pattern:**
```go
func TestAccContainerRenameActionTerraform(t *testing.T) {
    endpoint, username, password := testAccEnv(t)
    t.Setenv("DOCKHAND_ENDPOINT", endpoint)
    // ...
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
            "dockhand": providerserver.NewProtocol6WithError(New("test")()),
        },
        Steps: []resource.TestStep{
            {Config: fmt.Sprintf(`...HCL...`), Check: resource.ComposeTestCheckFunc(...)},
        },
    })
}
```

### Raw Client Acceptance Tests (`*_acc_test.go`)

Bypass Terraform entirely — call `NewClient` and provider API methods directly. Used to test CRUD lifecycle at the HTTP client level.

```go
func TestAccUserResource(t *testing.T) {
    endpoint, username, password := testAccEnv(t)
    sessionCookie := testAccLoginSessionCookie(t, endpoint, username, password)
    client, _ := NewClient(endpoint, sessionCookie, "1", true)
    // Create → Read → Update → Delete
}
```

## CI Integration

From `.github/workflows/go-ci.yml`:

```yaml
- name: Go test
  run: go test ./...
```

- No `TF_ACC` or Dockhand env vars set in CI — all acceptance tests auto-skip
- Only unit tests run in CI
- gofmt and `go mod tidy` are also enforced

## Coverage Gaps

- Only 3 unit test functions (4 test cases total) across the entire codebase
- Most resources (`resource_stack.go`, `resource_container.go`, `resource_network.go`, etc.) have **zero** unit tests
- No acceptance tests for: network, volume, image, git_repository, environment, schedule resources
- No mock HTTP server — unit tests for client behaviour are not possible without one
- No code coverage measurement or minimum threshold enforced

---

*Testing analysis: 2026-03-07*
