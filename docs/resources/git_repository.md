# dockhand_git_repository

Manages a Dockhand Git repository integration via `/api/git/repositories`.

## Example Usage

```terraform
resource "dockhand_git_repository" "stacks" {
  name         = "stacks"
  url          = "https://github.com/example/your-repo.git"
  branch       = "main"
  compose_path = "myapp_stack/docker-compose.yaml"

  credential_id = dockhand_git_credential.github.id

  auto_update          = false
  auto_update_schedule = "daily"
  auto_update_cron     = "0 3 * * *"
  webhook_enabled      = false
}
```

## Schema

### Read-only

- `id` (String) Repository ID.
- `webhook_secret` (String, Sensitive)
- `last_sync` (String)
- `last_commit` (String)
- `sync_status` (String)
- `sync_error` (String)
- `created_at` (String)
- `updated_at` (String)

### Required

- `name` (String)
- `url` (String)

### Optional/Computed

- `branch` (String)
- `compose_path` (String) Path to the Docker Compose file within the repository (e.g. `myapp_stack/docker-compose.yaml`). Computed by Dockhand if not set.
- `credential_id` (String)
- `environment_id` (String)
- `auto_update` (Boolean)
- `auto_update_schedule` (String)
- `auto_update_cron` (String)
- `webhook_enabled` (Boolean)
