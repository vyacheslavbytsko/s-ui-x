# Baseline environment

Дата прогона: 2026-05-23
Рабочая директория: `C:\s-ui-x`

## Версии

- Go: `go1.26.2 windows/amd64` (`tests/baseline/phase0/go-version.txt`)
- Node.js: `v25.3.0` (`tests/baseline/phase0/node-version.txt`)
- npm: `11.6.2` (`tests/baseline/phase0/npm-version.txt`)

`go.mod` объявляет `go 1.25.7`; локальный рантайм новее объявленной версии.

## Прочитанные entrypoints

- `README.md`: описывает ручную разработческую сборку, где frontend собирается до backend, затем `go build -o sui main.go`.
- `go.mod`: модуль `github.com/deposist/s-ui-x`.
- `frontend/package.json`: frontend существует, lockfile `frontend/package-lock.json` присутствует.
- `.github/workflows/`: существующие workflow `ci.yml`, `docker.yml`, `release.yml`, `windows.yml`; новые workflow в этом baseline не добавлялись.
- `make` в локальном `PATH` не найден; `Makefile` добавлен как воспроизводимая точка входа, но локальный baseline в этом прогоне запускался напрямую через `tests/baseline/run-command.ps1`.

## npm scripts

- `dev`: `vite --host`
- `build`: `vue-tsc --noEmit && vite build`
- `preview`: `vite preview`
- `test`: `vitest run`
- `test:unit`: `vitest run`
- `lint`: `eslint .`

Отдельного `typecheck` script в `frontend/package.json` нет; typecheck выполняется внутри `npm run build`.

## test-db fixtures

Каталог `test-db/` создан с пустым `.gitkeep`. Реальные `.db` файлы не добавлялись.

Для полного локального прогона importxui-тестов нужны:

- `test-db/x-ui.db`
- `test-db/s-ui.db`

Ссылки на код, который их подгружает:

- `database/importxui/importer_test.go:211`: `setupImportTestDB(t)` копирует `x-ui.db` и `s-ui.db` во временный каталог и инициализирует main DB.
- `database/importxui/importer_test.go:227`: `fixturePath(t, name)` ищет файлы в `../../test-db/` и делает `t.Skipf`, если фикстура отсутствует.
- `database/importxui/source/file/file_test.go:20`: `TestFileSourceAcquireReturnsExistingPath` использует `test-db/x-ui.db`.
- `database/importxui/source/ssh/ssh_test.go:154`: `readSSHFixture` использует `test-db/x-ui.db`.

Если фикстур нет, соответствующие тесты должны отображаться в baseline как skipped / requires `test-db`, а не как исправляемая в этом диалоге ошибка.

## Инструменты статического анализа

- `staticcheck`: установлен через `go install honnef.co/go/tools/cmd/staticcheck@latest`; версия `staticcheck.exe 2026.1 (v0.7.0)`.
- `golangci-lint`: установлен через `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`; версия `v1.64.8`.
- `gosec`: установлен через `go install github.com/securego/gosec/v2/cmd/gosec@latest`; `gosec -version` сообщает `Version: dev`.
- `govulncheck`: установлен через `go install golang.org/x/vuln/cmd/govulncheck@latest`; scanner `govulncheck@v1.3.0`, DB updated `2026-05-22 18:28:47 +0000 UTC`.

Логи установки и версий лежат в `tests/baseline/phase0/`.

`.golangci.yml` использует формат, совместимый с локальным `golangci-lint v1.64.8`. Analyzer `nilness` подключён через `govet`, потому что в этой версии это не отдельный top-level linter. Проверка конфига: `tests/baseline/phase1/golangci-lint-config-verify.txt`.

Go baseline-команды запускались с `GOFLAGS=-mod=readonly`, чтобы прогон не менял `go.mod` / `go.sum`.

## Phase 6 frontend/e2e environment

Дата прогона: 2026-05-24.

`frontend/node_modules` восстановлен в Phase 6. Исходный Phase 0 `npm ci` падал на Windows EPERM unlink для `frontend/node_modules/@rolldown/binding-win32-x64-msvc/rolldown-binding.win32-x64-msvc.node`; повторная диагностика после закрытия локальных Node/VS Code процессов показала, что файл больше не удерживается. После этого `frontend/node_modules` удалён целиком и `npm ci` прошёл green.

Дополнительные frontend devDependencies для baseline:

- `@playwright/test`
- `playwright`
- `@axe-core/playwright`
- `axe-core`

Отдельного `typecheck` script в `frontend/package.json` по-прежнему нет; frontend typecheck покрывается командой `npm run build`, где выполняется `vue-tsc --noEmit && vite build`.

Playwright Chromium установлен локально командой `npx playwright install chromium`. E2E smoke запускался против локальной Vite dev-инстанции и test-only API helper на SQLite DB в `tests/baseline/phase6/e2e-db/`.

Полный `sui` server не использовался для Phase 6 e2e:

- `go run .` завершался на `naive outbound is disabled when built without with_naive_outbound tag`;
- `go run -tags with_naive_outbound .` на Windows завершался из-за build constraints пакета `github.com/metacubex/mihomo/transport/cro/cronet-go/all`;
- переменная `SUI_DISABLE_CORE=1` была проверена как желаемый startup guard, но production-код её не поддерживает.

Поэтому Phase 6 e2e использует `tests/e2e/panel-server/main.go`: test-only helper поднимает реальные API/session/security middleware для frontend smoke, но не стартует sing-box/core, cron и внешние сервисы. Plaintext test credentials, созданные во время e2e, удалены из артефактов командой `tests/baseline/phase6/redact-e2e-generated-secrets.txt`; проверка отсутствия таких файлов сохранена в `tests/baseline/phase6/verify-e2e-secret-redaction.txt`.

## Phase 8 CI environment

Дата прогона: 2026-05-24.

Локальные ограничения:

- `make` в локальном `PATH` не найден; Phase 8 команды запускались прямыми эквивалентами через `tests/baseline/run-command.ps1` и сохранялись в `tests/baseline/phase8/`.
- `act` в локальном `PATH` не найден; `tests/baseline/phase8/act-dryrun.txt` помечен skipped, workflow верифицируется на push/PR.
- `bash` указывает на Windows WSL launcher (`C:\Windows\system32\bash.exe`), но WSL-дистрибутивы не установлены; `aggregate.sh` рассчитан на Ubuntu CI, локально dashboard сгенерирован через `scripts/audit/aggregate.ps1`.
- YAML parse новых `audit*.yml` проверен через `node scripts/audit/validate-workflows.js`; parser взят из установленного `frontend/node_modules/yaml`.

Production-код, Go/npm зависимости и `frontend/src` в Phase 8 не менялись.
