# Phase 8 CI design

Дата: 2026-05-24.

Цель Phase 8: упаковать уже полученный baseline Фаз 0-7 в GitHub Actions, не исправляя known red baseline и не меняя production-код.

## Phase to job matrix

| Фаза | Job | Workflow | Runner/matrix | Gate |
|---|---|---|---|---|
| 0/2/3/4/5/7 | `build` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | REQUIRED |
| 0/2/3/4/5/7 | `vet` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | REQUIRED |
| 0/2/3/4/5/7 | `test-go` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | REQUIRED |
| 3/7 | `test-go-race` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | SOFT, issue 47 |
| 1 | `lint-go` / `staticcheck` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | SOFT |
| 4 | `gosec` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | SOFT |
| 4 | `govulncheck` | `audit-go.yml` | `ubuntu-latest`, `windows-latest` | SOFT, triage отдельно |
| 6 | `fe-install` | `audit-frontend.yml` | `ubuntu-latest` | prerequisite |
| 6 | `fe-lint` | `audit-frontend.yml` | `ubuntu-latest` | REQUIRED |
| 6 | `fe-build` | `audit-frontend.yml` | `ubuntu-latest` | REQUIRED; включает `vue-tsc --noEmit` |
| 6 | `fe-vitest` | `audit-frontend.yml` | `ubuntu-latest` | REQUIRED |
| 6 | `fe-e2e` | `audit-frontend.yml` | `ubuntu-latest` | SOFT пока flaky/XFAIL |
| 6 | `accessibility` | `audit-frontend.yml` | `ubuntu-latest` | SOFT, axe debt baseline |
| 7 | `chaos-tests` | `audit-chaos.yml` | `ubuntu-latest` | SOFT |
| 7 | `docker-chaos` | `audit-chaos.yml` | `ubuntu-latest`, opt-in | SOFT, gated by `vars.RUN_DOCKER_CHAOS` или manual input |
| 5 | `bench-all` | `audit-perf.yml` | `ubuntu-latest` | SOFT |
| 8 | `dashboard` | `audit.yml` and final `audit-go.yml` job | `ubuntu-latest` | informational |

## Parallelization

Fast required Go jobs (`build`, `vet`, `test-go`) run in parallel across Linux and Windows. Known red or noisy jobs (`test-go-race`, `lint-go`, `staticcheck`, `gosec`, `govulncheck`) run in parallel too, but are marked soft so they do not block merge.

Frontend jobs run after `fe-install` proves `npm ci` is healthy. They reinstall from cache in each job instead of uploading `node_modules`, keeping artifacts small and deterministic. Playwright/e2e and accessibility are isolated soft jobs.

Chaos and perf workflows are separate schedules/manual workflows. They do not run on every PR because they are slower and contain known XFAIL/soft-warning scenarios.

## Cache plan

| Area | Mechanism | Key inputs |
|---|---|---|
| Go modules/build cache | `actions/setup-go@v5` built-in cache | `go.sum`, `go.mod` |
| npm cache | `actions/setup-node@v6` with `cache: npm` | `frontend/package-lock.json` |
| Playwright browsers | `actions/cache@v4` | OS + `frontend/package-lock.json`; path `~/.cache/ms-playwright` |
| Docker layers | Docker Buildx cache is intentionally not added to audit chaos until docker job is made non-optional |

## Critical gates

REQUIRED checks:
- `build`
- `vet`
- `test-go`
- `fe-lint`
- `fe-build`
- `fe-vitest`

SOFT checks:
- `test-go-race`, because issue 47 is a known red race anchor.
- `gosec`, because Phase 4 classified red baseline and production fixes are separate.
- `govulncheck`, because 12 called vulnerabilities require a dedicated triage dialog and no dependency upgrades happen in Phase 8.
- `staticcheck` and `golangci-lint`, because they are known red baseline.
- `chaos`, because it contains expected skips/XFAIL and platform blockers.
- `perf-bench`, because perf noise is expected and regressions are warnings until Linux baseline is stable.
- `fe-e2e` and `accessibility`, because Phase 6 contains Playwright fixme/XFAIL and axe debt.

## Artifacts and retention

Test logs and JUnit:
- upload `tests/baseline/phase*/*.txt`
- upload `tests/baseline/phase*/*.junit.xml`
- retention: 7 days

Baseline pegs and dashboards:
- upload `tests/baseline/phase8/summary.html`
- upload `tests/baseline/phase8/summary.json`
- upload benchmark comparison files
- retention: 30 days

Frontend:
- upload `tests/baseline/phase6/playwright/`
- upload `tests/baseline/phase6/playwright.junit.xml`

Coverage:
- upload `coverage.out`/coverage directories where produced by Makefile.

Security:
- upload gosec/govulncheck logs as soft artifacts. JSON output can be added later without changing gate policy.

## Matrix

Go required jobs use:
- `ubuntu-latest`
- `windows-latest`

Playwright, Docker chaos and perf use:
- `ubuntu-latest`

Reasoning:
- Linux is the primary CI runner for docker/playwright/perf.
- Windows stays in the required Go matrix because local baseline was Windows and release workflow supports Windows.
