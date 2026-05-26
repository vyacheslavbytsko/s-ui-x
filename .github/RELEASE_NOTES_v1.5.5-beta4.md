# S-UI v1.5.5-beta4

Audit hardening prerelease that closes the 48-item repository audit registry and
prepares the tree for GitHub release.

## Fixed

- Import-xui now treats `reset_required` as durable auth state through
  `users.force_password_reset` instead of returning generated passwords.
- Cron sync now honors saved `onlyNew` and persisted sync-profile import policy
  fields for settings, history, routing and admin import mode.
- Backup, realtime, API route/rollback/upload, system-info, token flushing,
  Telegram notifier, Telegram backup and WARP robustness fixes from the audit
  plan are included.
- MigrateXui UX now preserves apply errors, waits for rollback health before
  reload and hides generated admin passwords until explicit reveal.
- Endpoint save now blocks double-submit attempts and always clears loading
  state after failed save requests.

## Added

- Focused regression coverage and post-fix audit records for the closed audit
  plan.
- Final audit closure note marking the registry as complete: 48/48 items
  closed.
- Release cleanup for local audit scratch artifacts and workflow defaults for
  `v1.5.5-beta4`.

## Validation

- `go build ./...` - PASS.
- `go vet ./...` - PASS.
- `go test ./...` - PASS in the final Cluster E gate.
- `go test -race ./... -timeout 900s` - PASS in the final Cluster E gate.
- `govulncheck ./...` - PASS, no vulnerabilities found.
- `gosec ./...` remains the known baseline with exactly 55 issues.

## Install

```sh
bash <(curl -Ls https://raw.githubusercontent.com/deposist/s-ui-x/main/install.sh) v1.5.5-beta4
```

## Русский

Prerelease с закрытием полного audit hardening plan и подготовкой дерева к
GitHub release.

- Закрыт реестр аудита 48/48: import-xui contract, cron profile policy,
  backup/realtime/API hardening, token flushing, Telegram/WARP, system-info и
  MigrateXui UX.
- `reset_required` теперь хранится как `users.force_password_reset` без утечки
  временных паролей в отчёте.
- Endpoint save защищён от double-submit и сбрасывает loading state после
  ошибок сохранения.
- Добавлена финальная audit closure note и cleanup локальных scratch artifacts.
