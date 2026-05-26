# Required and soft CI checks

Дата: 2026-05-24.

## REQUIRED checks

Эти checks должны блокировать merge после включения branch protection:

| Check | Почему required |
|---|---|
| `build (ubuntu-latest)` | базовая компиляция Go на Linux |
| `build (windows-latest)` | базовая компиляция Go на Windows |
| `vet (ubuntu-latest)` | стандартная статическая проверка Go |
| `vet (windows-latest)` | Windows parity для `go vet` |
| `test-go (ubuntu-latest)` | основной Go regression suite без race |
| `test-go (windows-latest)` | Windows parity для Go tests |
| `fe-lint` | frontend lint baseline green после Phase 6 |
| `fe-build` | frontend typecheck + Vite build; `vue-tsc --noEmit` входит в `npm run build` |
| `fe-vitest` | frontend unit/regression tests |

## SOFT checks

Эти checks видны в PR, но не должны блокировать merge до отдельного triage/fix диалога:

| Check | Причина soft |
|---|---|
| `test-go-race` | известный red baseline: issue 47, `tokenUseDebouncer` vs test DB reinit |
| `lint-go` | `staticcheck` + `golangci-lint` red baseline |
| `staticcheck` | отдельный soft signal для staticcheck findings |
| `gosec` | Phase 4 classified red baseline; production fixes отдельно |
| `govulncheck` | 12 called vulnerabilities; dependency upgrades и triage отдельно |
| `fe-e2e` | Playwright содержит expected-fail/fixme anchors по п. 32, 43-46 |
| `accessibility` | axe debt зафиксирован как baseline |
| `chaos-tests` | XFAIL/skips и платформенные blockers ожидаемы |
| `docker-chaos` | требует Docker runner/privileged и включается opt-in |
| `bench-all` | perf noise, regression >20% пока warning |
| `dashboard` | informational summary |

## Registry links

Минимальная привязка к реестру:
- п. 47: `test-go-race` и chaos race anchor остаются soft до отдельного фикса.
- п. 32: `fe-e2e`/WS reconnect сценарии soft до healing reconnect фикса.
- п. 16, 18, 36, 40: perf/chaos anchors дают сигнал, но не блокируют merge.
- `govulncheck`: не мапится на конкретный фикс в Phase 8, triage вынесен отдельно.

## Branch protection recommendation

Включить required status checks только из списка REQUIRED. Не добавлять `Audit Go` workflow целиком как required, иначе known red soft jobs начнут блокировать merge.
