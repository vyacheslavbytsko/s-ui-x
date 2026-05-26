# CONTRIBUTING snippet: audit pipeline

Этот фрагмент дублирует краткий раздел из `CONTRIBUTING.md` и нужен для ссылок из audit docs.

## Запуск audit-pipeline локально

Минимальный PR gate:

```sh
make audit:build
make audit:vet
make audit:test-go
make audit:fe-lint
make audit:fe-build
make audit:test-fe
```

Soft checks:

```sh
make audit:test-go-race
make audit:lint-go
make audit:gosec
make audit:vuln
go test -tags=chaos ./tests/chaos/... -count=1 -timeout 30m
go test ./... -bench=. -benchmem -run=^$ -benchtime=2s
```

Frontend prerequisites:
- `make audit:fe-install` или `npm ci` в `frontend/`.
- Отдельного `typecheck` script нет; `npm run build` запускает `vue-tsc --noEmit`.
- Playwright требует `npx playwright install chromium`.

`test-db` fixtures:
- Полные importxui fixture-тесты требуют `test-db/x-ui.db` и `test-db/s-ui.db`.
- Если фикстур нет, такие тесты должны оставаться skipped, а не чиниться в CI-фазе.
