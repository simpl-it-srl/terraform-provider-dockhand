# Structure

**Analysis Date:** 2026-03-07

## Directory Layout

```
terraform-provider-dockhand/
├── main.go                          # Binary entry point
├── go.mod / go.sum                  # Go module definition
├── GNUmakefile                      # Build, test, install, dev-override targets
├── terraform-registry-manifest.json # Terraform Registry metadata
├── LICENSE
├── README.md
├── AGENTS.md                        # Claude Code agent instructions
│
├── internal/provider/               # ALL provider code (single flat package)
│   ├── provider.go                  # Provider schema, Configure, Resources, DataSources
│   ├── auth.go                      # Login() — session cookie authentication
│   ├── client.go                    # HTTP client + all API types (2,159 lines)
│   ├── types_helpers.go             # Shared type conversion helpers
│   │
│   ├── resource_*.go                # One file per managed resource
│   ├── data_source_*.go             # One file per data source
│   │
│   ├── *_test.go                    # Unit tests (3 files, 4 test functions)
│   ├── *_acc_test.go                # Raw-client acceptance tests (1 file)
│   ├── *_tf_acc_test.go             # Terraform-framework acceptance tests (4 files)
│   └── acc_test.go                  # Shared acceptance test helpers
│
├── docs/                            # Provider documentation (Terraform Registry format)
│   ├── index.md                     # Provider root docs
│   ├── resources/                   # One .md per resource
│   └── data-sources/                # One .md per data source
│
├── examples/                        # HCL usage examples (referenced by docs)
│   ├── provider/
│   ├── resources/dockhand_<name>/resource.tf
│   └── data-sources/dockhand_<name>/data-source.tf
│
├── scripts/                         # Developer and release scripts
│   ├── build-mirror.sh              # Build filesystem mirror for private distribution
│   ├── build-packages.sh            # Build multi-platform release packages
│   ├── endpoint-probe.py            # Script to probe Dockhand API endpoints
│   ├── release-precheck.sh          # Pre-release validation
│   ├── release-test.sh              # Release smoke test
│   ├── tf-dev.sh                    # Terraform dev_overrides helper
│   └── tf-local.sh                  # Local Terraform runner
│
└── .github/workflows/
    ├── go-ci.yml                    # CI: gofmt, go mod tidy, go test, go build
    └── release-artifacts.yml        # Release: multi-platform builds + GitHub release
```

## Key Locations

| What | Where |
|------|-------|
| Provider schema & auth bootstrap | `internal/provider/provider.go` |
| All HTTP API calls | `internal/provider/client.go` |
| Authentication (login) | `internal/provider/auth.go` |
| Type conversion helpers | `internal/provider/types_helpers.go` |
| All resources | `internal/provider/resource_*.go` |
| All data sources | `internal/provider/data_source_*.go` |
| Unit tests | `internal/provider/*_test.go` |
| Acceptance tests (raw client) | `internal/provider/*_acc_test.go` |
| Acceptance tests (TF framework) | `internal/provider/*_tf_acc_test.go` |
| Shared acc test setup | `internal/provider/acc_test.go` |
| Documentation | `docs/` |
| HCL examples | `examples/` |

## Resource Inventory

**Managed Resources (`resource_*.go`):**

| File | Terraform Resource | Type |
|------|--------------------|------|
| `resource_auth_settings.go` | `dockhand_auth_settings` | Config |
| `resource_config_set.go` | `dockhand_config_set` | Config |
| `resource_container.go` | `dockhand_container` | CRUD |
| `resource_container_action.go` | `dockhand_container_action` | Action |
| `resource_container_check_updates_action.go` | `dockhand_container_check_updates_action` | Action |
| `resource_container_file.go` | `dockhand_container_file` | CRUD |
| `resource_container_rename_action.go` | `dockhand_container_rename_action` | Action |
| `resource_container_update_action.go` | `dockhand_container_update_action` | Action |
| `resource_environment.go` | `dockhand_environment` | CRUD |
| `resource_git_credential.go` | `dockhand_git_credential` | CRUD |
| `resource_git_repository.go` | `dockhand_git_repository` | CRUD |
| `resource_git_stack.go` | `dockhand_git_stack` | CRUD |
| `resource_git_stack_deploy_action.go` | `dockhand_git_stack_deploy_action` | Action |
| `resource_git_stack_env_file.go` | `dockhand_git_stack_env_file` | CRUD |
| `resource_git_stack_webhook_action.go` | `dockhand_git_stack_webhook_action` | Action |
| `resource_image.go` | `dockhand_image` | CRUD |
| `resource_image_push_action.go` | `dockhand_image_push_action` | Action |
| `resource_image_scan_action.go` | `dockhand_image_scan_action` | Action |
| `resource_license.go` | `dockhand_license` | Config |
| `resource_network.go` | `dockhand_network` | CRUD |
| `resource_network_connection_action.go` | `dockhand_network_connection_action` | Action |
| `resource_notification.go` | `dockhand_notification` | CRUD |
| `resource_registry.go` | `dockhand_registry` | CRUD |
| `resource_schedule.go` | `dockhand_schedule` | CRUD |
| `resource_schedule_run_action.go` | `dockhand_schedule_run_action` | Action |
| `resource_settings_general.go` | `dockhand_settings_general` | Config |
| `resource_stack.go` | `dockhand_stack` | CRUD |
| `resource_stack_action.go` | `dockhand_stack_action` | Action |
| `resource_stack_adopt_action.go` | `dockhand_stack_adopt_action` | Action |
| `resource_stack_env.go` | `dockhand_stack_env` | CRUD |
| `resource_stack_scan_action.go` | `dockhand_stack_scan_action` | Action |
| `resource_user.go` | `dockhand_user` | CRUD |
| `resource_volume.go` | `dockhand_volume` | CRUD |
| `resource_volume_clone_action.go` | `dockhand_volume_clone_action` | Action |

**Data Sources (`data_source_*.go`):**

| File | Terraform Data Source |
|------|-----------------------|
| `data_source_activity.go` | `dockhand_activity` |
| `data_source_auth_providers.go` | `dockhand_auth_providers` |
| `data_source_config_sets.go` | `dockhand_config_sets` |
| `data_source_container_inspect.go` | `dockhand_container_inspect` |
| `data_source_container_logs.go` | `dockhand_container_logs` |
| `data_source_container_pending_updates.go` | `dockhand_container_pending_updates` |
| `data_source_container_processes.go` | `dockhand_container_processes` |
| `data_source_containers.go` | `dockhand_containers` |
| `data_source_container_shells.go` | `dockhand_container_shells` |
| `data_source_container_stats.go` | `dockhand_container_stats` |
| `data_source_environments.go` | `dockhand_environments` |
| `data_source_git_credentials.go` | `dockhand_git_credentials` |
| `data_source_git_repositories.go` | `dockhand_git_repositories` |
| `data_source_hawser_status.go` | `dockhand_hawser_status` |
| `data_source_health.go` | `dockhand_health` |
| `data_source_images.go` | `dockhand_images` |
| `data_source_networks.go` | `dockhand_networks` |
| `data_source_notifications.go` | `dockhand_notifications` |
| `data_source_registries.go` | `dockhand_registries` |
| `data_source_schedules.go` | `dockhand_schedules` |
| `data_source_schedules_executions.go` | `dockhand_schedules_executions` |
| `data_source_stacks.go` | `dockhand_stacks` |
| `data_source_stack_sources.go` | `dockhand_stack_sources` |
| `data_source_users.go` | `dockhand_users` |
| `data_source_volumes.go` | `dockhand_volumes` |

## Naming Conventions

| Pattern | Example |
|---------|---------|
| Resource file | `resource_<noun>.go` |
| Action resource file | `resource_<noun>_<verb>_action.go` |
| Data source file | `data_source_<noun>.go` |
| Unit test | `<filename>_test.go` |
| TF acceptance test | `<filename>_tf_acc_test.go` |
| Raw client acceptance test | `<filename>_acc_test.go` |
| Resource struct | `<noun>Resource` |
| Data source struct | `<noun>DataSource` |
| Terraform state model | `<noun>ResourceModel` / `<noun>DataSourceModel` |
| API payload struct | `<noun>Payload` |
| API response struct | `<noun>Response` |
| Constructor function | `New<TypeName>Resource()` / `New<TypeName>DataSource()` |

---

*Structure analysis: 2026-03-07*
