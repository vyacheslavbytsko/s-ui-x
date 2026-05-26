# Makefile audit targets

Дата инвентаризации: 2026-05-24.

`Makefile` использует wrapper:

```make
PS ?= powershell -NoProfile -ExecutionPolicy Bypass
RUN = $(PS) -File tests/baseline/run-command.ps1 -ContinueOnError
```

Wrapper пишет:
- `tests/baseline/<phase>/<name>.txt`
- `tests/baseline/<phase>/<name>.junit.xml`

На Linux runner workflow переопределяет `PS` на `pwsh -NoProfile -ExecutionPolicy Bypass`, чтобы сохранить тот же wrapper и JUnit формат.

| Target | Команды | Phase/name артефактов | Ожидаемый baseline |
|---|---|---|---|
| `audit` | build, vet, test-go, race, cover, gosec, govulncheck, lint-go, frontend typecheck/lint/build/test/e2e | mix `phase0`, `phase1` | aggregate target для полного локального baseline |
| `audit:lint-go` | `staticcheck ./...`; `golangci-lint run` | `phase1/staticcheck.*`, `phase1/golangci-lint.*` | red baseline, soft CI |
| `audit:vet` | `go vet ./...` | `phase0/go-vet.*` | green, required CI |
| `audit:build` | `go build ./...` | `phase0/go-build.*` | green, required CI |
| `audit:test-go` | `go test ./...` | `phase0/go-test.*` | green, required CI |
| `audit:test-go-race` | `go test ./... -race -count=1` | `phase0/go-test-race.*` | red/soft из-за issue 47 |
| `audit:cover` | `go test ./... -coverprofile tests/baseline/phase0/coverage.out` | `phase0/go-cover.*`, `phase0/coverage.out` | green, informational |
| `audit:gosec` | `gosec ./...` | `phase1/gosec.*` | red classified baseline, soft CI |
| `audit:vuln` | `govulncheck ./...` | `phase1/govulncheck.*` | red, triage отдельным диалогом, soft CI |
| `audit:fe-install` | `npm ci` в `frontend/` | `phase0/npm-ci.*` | green после Phase 6 |
| `audit:fe-typecheck` | skipped: отдельного script нет | `phase0/npm-run-typecheck.*` | skipped; typecheck входит в `npm run build` |
| `audit:fe-lint` | `npm run lint` в `frontend/` | `phase0/npm-run-lint.*` | green, required CI |
| `audit:fe-build` | `npm run build` в `frontend/` | `phase0/npm-run-build.*` | green, required CI; включает `vue-tsc --noEmit` |
| `audit:test-fe` | `npm run test` в `frontend/` | `phase0/npm-run-test.*` | green, required CI |
| `audit:e2e` | skipped TODO в Makefile | `phase0/e2e.*` | заменено в CI прямым Playwright job через `frontend/playwright.config.ts` |

## Phase 8 local artifact convention

Для локального Phase 8 прогона команды из задания дополнительно сохраняются в `tests/baseline/phase8/*.txt` и `*.junit.xml`, чтобы не перетирать исторические phase0/phase1/phase6 baseline pegs.
