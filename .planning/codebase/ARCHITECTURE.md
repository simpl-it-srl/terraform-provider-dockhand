# Architecture

**Analysis Date:** 2026-03-07

## Pattern Overview

**Overall:** Terraform Provider using the Plugin Framework (hashicorp/terraform-plugin-framework v1.17.0)

**Key Characteristics:**
- Single Go package (`internal/provider`) contains all provider logic — no subdirectories
- Flat file organisation: one file per resource/data source
- `Client` struct in `client.go` is the sole API abstraction; it owns all HTTP I/O
- Resources and data sources receive `*Client` via Terraform's `Configure` callback (dependency injection)
- Session-cookie-based authentication — a session cookie is obtained once at provider `Configure` time and attached to every subsequent request

## Layers

**Entry Point / Binary Layer:**
- Purpose: Start the provider as a gRPC plugin server consumed by Terraform CLI
- Location: `main.go`
- Contains: `main()` calling `providerserver.Serve`
- Depends on: `internal/provider`
- Used by: Terraform CLI via the plugin protocol

**Provider Layer:**
- Purpose: Declare provider schema, authenticate, build the shared `*Client`, and register all resources and data sources
- Location: `internal/provider/provider.go`
- Contains: `dockhandProvider`, `dockhandProviderModel`, `Configure`, `Resources`, `DataSources`
- Depends on: `auth.go`, `client.go`
- Used by: Terraform Plugin Framework at provider init

**Authentication Layer:**
- Purpose: Perform HTTP login against `/api/auth/login` and return a session cookie string
- Location: `internal/provider/auth.go`
- Contains: `Login()` function, `loginResponse` struct
- Depends on: standard library `net/http`
- Used by: `provider.go Configure`

**HTTP Client Layer:**
- Purpose: Centralised Dockhand REST API client — all request construction, sending, and response parsing
- Location: `internal/provider/client.go`
- Contains: `Client` struct, `NewClient`, `doJSON`/`doJSONWithStatus` helpers, all API payload/response structs, one method per Dockhand API endpoint
- Depends on: standard library `net/http`, `encoding/json`
- Used by: every resource and data source file

**Resource Layer:**
- Purpose: Implement Terraform-managed resources (CRUD lifecycle) for each Dockhand entity
- Location: `internal/provider/resource_*.go`
- Contains: one resource struct per file (e.g. `stackResource`, `containerResource`, `gitStackResource`), a `*Model` struct per resource, `Create`/`Read`/`Update`/`Delete`/`ImportState` methods
- Depends on: `client.go`
- Used by: Terraform Plugin Framework via the function pointer registered in `provider.go`

**Data Source Layer:**
- Purpose: Implement read-only Terraform data sources that query Dockhand and expose data for use in configs
- Location: `internal/provider/data_source_*.go`
- Contains: one data source struct per file (e.g. `healthDataSource`, `containersDataSource`), a `*Model` struct, `Read` method
- Depends on: `client.go`
- Used by: Terraform Plugin Framework via the function pointer registered in `provider.go`

**Helpers Layer:**
- Purpose: Shared utility functions for converting between Go types and Terraform Framework types
- Location: `internal/provider/types_helpers.go`
- Contains: `stringValueOrNull`, `int64StringValueOrNull`, and `stringSliceToListValue` (defined inline in resources)
- Depends on: `github.com/hashicorp/terraform-plugin-framework/types`
- Used by: resource and data source files

## Data Flow

**Resource Create:**

1. Terraform CLI reads HCL config and calls the provider plugin via gRPC
2. `provider.go Configure` authenticates via `auth.go Login` and constructs `*Client`
3. Terraform calls `resource.Create` on the registered resource struct
4. Resource reads the plan from `req.Plan.Get(ctx, &model)`
5. Resource calls the appropriate `Client` method (e.g. `c.client.CreateStack(...)`)
6. `Client` method calls `doJSONWithStatus` which builds the HTTP request, attaches the `Cookie` header, sends it, and decodes the JSON response
7. Resource maps API response fields back onto the model and calls `resp.State.Set(ctx, &model)`

**Resource Read / Drift Detection:**

1. Terraform calls `resource.Read` during plan/apply
2. Resource reads current state from `req.State.Get(ctx, &state)`
3. Resource calls a `Client.Get*` or `Client.List*` method
4. If the API returns 404 / resource not found, the resource calls `resp.State.RemoveResource(ctx)` to signal drift
5. Otherwise the state is refreshed with current API values

**Action Resources (one-shot side effects):**

- Files such as `resource_container_action.go`, `resource_stack_action.go`, `resource_git_stack_deploy_action.go` etc. model imperative API calls as Terraform resources
- `Create` executes the action; `Read` is a no-op (actions have no persistent state to drift)
- Re-triggering is achieved by changing the `trigger` attribute, which forces replacement

**State Management:**
- All state lives in Terraform's state file; the provider itself is stateless between operations
- The `*Client` is constructed fresh on every `terraform apply` / `terraform plan`

## Key Abstractions

**Client:**
- Purpose: Single struct encapsulating all Dockhand REST API interactions
- Location: `internal/provider/client.go`
- Pattern: Each API resource domain has a group of methods: `List*`, `Get*`, `Create*`, `Update*`, `Delete*`. All route through `doJSONWithStatus(ctx, method, path, queryParams, requestBody, &responseOut)`

**Resource Model Structs:**
- Purpose: Represent the Terraform state/plan for a resource using `tfsdk` struct tags
- Examples: `stackResourceModel`, `containerResourceModel`, `gitStackModel`
- Pattern: Each file defines `type <name>ResourceModel struct { ... }` with `types.String`, `types.Bool`, `types.Int64`, `types.List` fields tagged `tfsdk:"<attr_name>"`

**Action Resources:**
- Purpose: Wrap imperative Dockhand API calls (start/stop/deploy/scan/rename) as Terraform resources whose lifecycle is equivalent to running the action once
- Examples: `resource_container_action.go`, `resource_git_stack_deploy_action.go`, `resource_stack_scan_action.go`
- Pattern: `Create` fires the API call; `Read`/`Update` are no-ops; a `trigger` string attribute forces re-run via replace

**Data Source Model Structs:**
- Purpose: Represent queried/read-only Dockhand data in Terraform state
- Examples: `healthDataSourceModel`, `containersDataSourceModel`
- Pattern: Each file defines `type <name>DataSourceModel struct { ... }` with computed attributes

## Entry Points

**Binary Entry Point:**
- Location: `main.go`
- Triggers: Terraform CLI plugin invocation
- Responsibilities: Parse the `--debug` flag, call `providerserver.Serve` with the provider factory function and registry address `registry.terraform.io/simpl-it-srl/dockhand`

**Provider Configure:**
- Location: `internal/provider/provider.go` — `Configure` method
- Triggers: First operation in any `terraform plan` or `terraform apply`
- Responsibilities: Resolve config/env vars, call `Login`, build `*Client`, inject it as `ResourceData` and `DataSourceData`

**Resource/DataSource Constructors:**
- Pattern: Every resource/data source exposes a `New<Name>Resource()` / `New<Name>DataSource()` function returning the interface
- Registered in `provider.go Resources()` and `DataSources()` slices

## Error Handling

**Strategy:** Surface all errors through Terraform diagnostics (`resp.Diagnostics.AddError`)

**Patterns:**
- Every CRUD method starts with `if r.client == nil { resp.Diagnostics.AddError("Unconfigured client", ...) }`
- API call errors are wrapped: `resp.Diagnostics.AddError("Error creating ...", err.Error())`
- HTTP 404 responses in `Read` trigger `resp.State.RemoveResource(ctx)` (graceful drift handling)
- `doJSONWithStatus` returns `(statusCode, error)` allowing callers to distinguish 404 from other errors
- Compile-time interface assertions (e.g. `var _ resource.Resource = (*stackResource)(nil)`) ensure all required methods are implemented

## Cross-Cutting Concerns

**Logging:** Standard `log` package in `main.go` only; provider-level errors use Terraform diagnostics
**Validation:** Attribute validation is declared inline in `Schema()` methods using `Validators` or `PlanModifiers`; required vs optional is set per attribute
**Authentication:** Cookie-based session — `Login()` POSTs to `/api/auth/login`, extracts `dockhand_session` cookie, stored on `Client.sessionCookie` and sent as a `Cookie` header on every request
**Environment scoping:** Most resources accept an optional `env` attribute (Dockhand environment ID); `Client.defaultEnv` is used as fallback when `env` is not set on a resource

---

*Architecture analysis: 2026-03-07*
