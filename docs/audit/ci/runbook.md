# CI audit runbook

Дата: 2026-05-24.

## Ручной запуск workflows

Через GitHub UI:
- Actions -> `Audit Go` -> Run workflow.
- Actions -> `Audit Frontend` -> Run workflow.
- Actions -> `Audit Chaos` -> Run workflow.
- Actions -> `Audit Perf` -> Run workflow.
- Actions -> `Audit Dashboard` -> Run workflow.

Через `gh`:

```sh
gh workflow run audit-go.yml
gh workflow run audit-frontend.yml
gh workflow run audit-chaos.yml
gh workflow run audit-perf.yml
gh workflow run audit.yml
```

Docker chaos opt-in:

```sh
gh workflow run audit-chaos.yml -f run_docker_chaos=true
```

Или задайте repository variable:

```text
RUN_DOCKER_CHAOS=true
```

## Локальный запуск

Go:

```sh
make audit:build
make audit:vet
make audit:test-go
make audit:test-go-race
make audit:lint-go
make audit:gosec
make audit:vuln
```

Frontend:

```sh
make audit:fe-install
make audit:fe-lint
make audit:fe-build
make audit:test-fe
```

Playwright e2e напрямую:

```sh
cd frontend
npx playwright install chromium
npx playwright test
```

Chaos:

```sh
go test -tags=chaos ./tests/chaos/... -count=1 -timeout 30m
```

Perf:

```sh
go test ./... -bench=. -benchmem -run=^$ -benchtime=2s
```

## Как читать `summary.html`

`scripts/audit/aggregate.sh` и `scripts/audit/aggregate.ps1` собирают все `tests/baseline/phase*/**/*.junit.xml` и пишут:
- `tests/baseline/phase8/summary.json`
- `tests/baseline/phase8/summary.html`
- `tests/baseline/phase8/aggregate.junit.xml`

В dashboard:
- `green` = passed testcases.
- `red` = failures + errors.
- `skipped` = skipped testcases.
- `XFAIL markers` = текстовые маркеры `XFAIL`, `expected-fail`, `fixme` в JUnit/log files.

`deltaVsBaselineMarkers` в JSON сравнивает JUnit totals с текстовыми marker-counts из `tests/baseline/SUMMARY.md`. Это не замена triage, а быстрый индикатор: стало ли больше red/skipped/XFAIL сигналов относительно описанного baseline.

## Как добавить новый check

1. Добавьте команду в `Makefile`, если check должен быть локально воспроизводимым.
2. Пишите лог и JUnit через `tests/baseline/run-command.ps1`.
3. Добавьте job в соответствующий `audit-*.yml`.
4. Решите gate policy:
   - REQUIRED только для green, стабильных и дешёвых проверок.
   - SOFT для known red, flaky, nightly, perf/security triage.
5. Добавьте artifact upload.
6. Обновите `docs/audit/ci/required-checks.md` и `tests/baseline/SUMMARY.md`.

## `act` dry-run

Локально Phase 8 проверяет только сборочный job:

```sh
act -W .github/workflows/audit-go.yml -j build
```

Если `act` не установлен, dry-run фиксируется как skipped и workflow верифицируется на push/PR.
