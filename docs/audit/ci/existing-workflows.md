# Existing GitHub Actions workflows

Дата инвентаризации: 2026-05-24.

## `.github/workflows/ci.yml`

Назначение: обычный CI для backend/frontend.

Триггеры:
- `pull_request`
- `push` в `main`

Runner:
- `ubuntu-24.04`

Jobs:
- `backend`: checkout, `actions/setup-go`, `go vet ./...`, `go test -race -timeout=5m ./...`.
- `frontend`: checkout, `actions/setup-node`, `npm ci`, `npm run lint`, `npm run test`, `npm run build`.

Пересечение с audit-pipeline:
- Дублирует часть `audit-go.yml`: `vet` и race-tests. В audit-pipeline race detector оставлен soft check из-за зарегистрированного пункта 47.
- Дублирует часть `audit-frontend.yml`: install/lint/test/build.
- Рекомендация на будущее: после стабилизации Phase 8 слить обычный `ci.yml` с audit required gates, чтобы не держать две конкурирующие картины статуса.

Concurrency:
- Не задана. Новые `audit-*` workflows используют собственные `concurrency.group`, поэтому с существующим CI не конкурируют за cancel policy.

## `.github/workflows/docker.yml`

Назначение: сборка и публикация Docker image в GHCR.

Триггеры:
- `workflow_dispatch` с input `tag`
- `push` тегов `v*`

Runner:
- `ubuntu-24.04`

Jobs:
- `build`: checkout с submodules, QEMU, Buildx, GHCR login через `secrets.GITHUB_TOKEN`, нормализация release tag, multi-platform Docker build/push `linux/amd64,linux/arm64`.

Пересечение с audit-pipeline:
- Не дублирует тестовый audit pipeline. Использует release/tag контур и publish permissions.

Concurrency:
- Не задана. Новые audit workflows не используют tag publish и не пишут packages.

## `.github/workflows/release.yml`

Назначение: release-сборка Linux артефактов.

Триггеры:
- `workflow_dispatch` с input `tag`
- `push` тегов `v*`

Runner:
- `ubuntu-latest`

Jobs:
- `build-frontend`: checkout, Node 25, `npm install`, `npm run build`, upload `frontend-dist`.
- `build-linux`: matrix по Linux architectures, download frontend artifact, setup Go, optional cronet-go/Bootlin toolchains, static `go build`, packaging, checksums, upload artifacts.
- `publish-linux`: download artifacts и публикация GitHub Release.

Пересечение с audit-pipeline:
- Frontend build пересекается с `audit-frontend.yml`, но release workflow делает packaging/publish и не должен заменяться audit checks.
- Linux `go build` здесь release-specific: custom tags, toolchains, static linking. Это не дубликат `audit-go.yml` build gate.

Concurrency:
- Не задана. Новые audit workflows имеют отдельные groups и не отменяют release jobs.

## `.github/workflows/windows.yml`

Назначение: release-сборка Windows артефактов.

Триггеры:
- `workflow_dispatch` с input `tag`
- `push` тегов `v*`

Runner:
- `ubuntu-latest` для frontend и Windows ARM64 cross-build
- `windows-latest` для Windows AMD64 build

Jobs:
- `build-frontend`: checkout, Node 25, `npm install`, `npm run build`, upload `frontend-dist`.
- `build-windows`: matrix `amd64`/`arm64`, setup Go, optional Chocolatey zip, `go build` с release tags, download `libcronet.dll`, package zip.
- `publish-windows`: download artifacts и публикация GitHub Release.

Пересечение с audit-pipeline:
- Frontend build пересекается с `audit-frontend.yml`.
- Windows Go build здесь release-specific и не заменяет `audit-go.yml` test/build gates.

Concurrency:
- Не задана. Новые audit workflows используют отдельные groups и не отменяют release jobs.

## Итог интеграции

Существующие workflow не удаляются и не меняются. Phase 8 добавляет отдельный audit contour:
- `audit.yml`: dashboard-only summary workflow.
- `audit-go.yml`: Go required и soft checks.
- `audit-frontend.yml`: frontend required и soft checks.
- `audit-chaos.yml`: nightly/manual chaos checks.
- `audit-perf.yml`: weekly/manual benchmark regression checks.

Branch protection рекомендуется настраивать на required checks из `docs/audit/ci/required-checks.md`, а не на весь набор soft checks.
