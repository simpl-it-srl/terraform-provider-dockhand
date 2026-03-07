# Coding Conventions

**Analysis Date:** 2026-03-07

## Naming Patterns

**Files:**
- Resources: `resource_<noun>.go` (e.g., `resource_container.go`, `resource_git_stack.go`)
- Action resources: `resource_<noun>_<verb>_action.go` (e.g., `resource_container_rename_action.go`)
- Data sources: `data_source_<noun>.go` (e.g., `data_source_containers.go`)
- Test files (unit): `<filename>_test.go` (e.g., `resource_container_test.go`)
- Test files (acceptance via Terraform framework): `<filename>_tf_acc_test.go`
- Test files (acceptance via raw client): `<filename>_acc_test.go`

**Types/Structs:**
- Resource struct: `<noun>Resource` (e.g., `containerResource`, `stackResource`)
- Data source struct: `<noun>DataSource` (e.g., `containersDataSource`)
- Terraform model: `<noun>ResourceModel` or `<noun>DataSourceModel` (e.g., `containerResourceModel`, `containersDataSourceModel`)
- API payload/request: `<noun>Payload` (e.g., `containerPayload`, `stackPayload`)
- API response: `<noun>Response` (e.g., `containerResponse`, `stackResponse`)
- Sub-model types: `<parent><field>Model` (e.g., `containerPortModel`, `containersDataSourceContainerModel`)

**Functions:**
- Constructor functions: `New<TypeName>Resource()` or `New<TypeName>DataSource()` — always return interface types
- Helper/flatten functions: camelCase descriptive verbs (e.g., `flattenEnvVars`, `flattenStringMap`, `parseContainerUpdatePayload`)
- Apply-to-state helpers: `apply<Noun>ToState` (e.g., `applyContainerRuntimeToState`)
- Format helpers: `format<Noun>` (e.g., `formatStackID`)

**Variables:**
- Local variables: camelCase (`plan`, `state`, `env`, `id`, `payload`)
- Package-level interface assertions: `var _ <Interface> = (*<Type>)(nil)`

**JSON struct tags:**
- API payload fields use camelCase JSON keys (e.g., `json:"restartPolicy"`, `json:"nanoCpus"`)
- Optional fields use `omitempty` (e.g., `json:"command,omitempty"`)
- Terraform SDK model fields use `tfsdk:"<snake_case>"` struct tags

## Code Style

**Formatting:**
- Standard `gofmt` — enforced in CI via `gofmt -l .` check that exits non-zero if any file is unformatted
- No additional style tools (no golangci-lint, no staticcheck in CI)

**Linting:**
- No dedicated lint step in CI beyond `gofmt`
- `go mod tidy` verification is enforced in CI

## Import Organization

**Order (standard `gofmt` grouping):**
1. Standard library packages
2. Third-party packages (github.com/hashicorp/...)
3. Internal packages (same module — used only in `main.go`)

**Example from `resource_container.go`:**
```go
import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)
```

**Path Aliases:** None used.

## Interface Compliance

Use compile-time interface assertions at package level. Resources declare all implemented interfaces explicitly:

```go
var (
    _ resource.Resource                = (*containerResource)(nil)
    _ resource.ResourceWithConfigure   = (*containerResource)(nil)
    _ resource.ResourceWithImportState = (*containerResource)(nil)
)
```

In `provider.go`, only the top-level provider interface is asserted:
```go
var _ provider.Provider = (*dockhandProvider)(nil)
```

## Error Handling

**Pattern:** All errors are surfaced through the Terraform diagnostics system (`resp.Diagnostics`). Never use `panic` or `log.Fatal` in resource/data source code.

**Standard error block:**
```go
if err != nil {
    resp.Diagnostics.AddError("Short summary title", err.Error())
    return
}
```

**Client nil guard** — every CRUD method starts with:
```go
if r.client == nil {
    resp.Diagnostics.AddError("Unconfigured client", "The provider client was not configured.")
    return
}
```

**Diagnostics propagation** — use `resp.Diagnostics.Append(...)` with immediate error check:
```go
resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
if resp.Diagnostics.HasError() {
    return
}
```

**Not-found handling:** Resources call `resp.State.RemoveResource(ctx)` and return without error when a resource no longer exists (drift detection).

**Delete 404 tolerance:** Delete methods treat HTTP 404 as success, and optionally re-read to verify absence before failing:
```go
status, err := r.client.DeleteStack(ctx, ...)
if err != nil && status != 404 {
    _, found, readErr := r.client.GetStackByName(ctx, ...)
    if readErr != nil || found {
        resp.Diagnostics.AddError(...)
    }
}
```

**Auth errors:** Returned as Go `error` from `auth.go:Login()` using `fmt.Errorf(...)` with descriptive messages; caller appends to diagnostics.

## Logging

**Framework:** None beyond the standard `log` package used only in `main.go` for fatal startup errors. No structured logging in provider code itself — Terraform plugin framework handles log routing.

## Comments

**When to Comment:**
- Inline comments explain non-obvious Dockhand API behaviors or quirks
- Multi-line comments document workarounds for known Dockhand API inconsistencies

**Examples from the codebase:**
```go
// Dockhand create currently starts stacks automatically. If desired state is disabled,
// explicitly stop it after creation.

// Dockhand may briefly report containers as "marked for removal". Wait until
// the container no longer appears before allowing dependent deletes (images).

// Some Dockhand builds return 500 "Failed to remove container" when the
// container is already gone. Re-check existence before failing destroy.
```

**No JSDoc/GoDoc:** Public functions at package scope are generally uncommented except for `Login()` which has a single-line doc comment.

## Function Design

**CRUD method signatures:** Fixed by Terraform plugin framework interface — all methods take `context.Context` + typed request/response.

**Helper function size:** Small, single-purpose. Flatten/convert helpers (`flattenEnvVars`, `flattenStringMap`, `flattenContainerPorts`, `stringSliceToListValue`) are typically 5–20 lines.

**Optional field handling:** Nil-pointer pattern — helper functions return `*T` for optional values:
```go
func stringPtrFromStringValue(value types.String) *string {
    if value.IsNull() || value.IsUnknown() {
        return nil
    }
    v := strings.TrimSpace(value.ValueString())
    if v == "" {
        return nil
    }
    return &v
}
```

**Null/Unknown checks:** Always check both `IsNull()` and `IsUnknown()` before calling `.ValueString()`, `.ValueBool()`, etc.

## Module Design

**Exports:** One public `New<X>Resource()` or `New<X>DataSource()` constructor per file. All struct types, model types, and API types are unexported (lowercase).

**Barrel Files:** None. The single `package provider` package contains all code flat in `internal/provider/`.

**Shared helpers:** Live in the same package. Small cross-file helpers in `types_helpers.go` (`stringValueOrNull`, `int64StringValueOrNull`). Flatten helpers collocated in the resource file that uses them, or in `resource_container.go` for widely-shared ones.

---

*Convention analysis: 2026-03-07*
