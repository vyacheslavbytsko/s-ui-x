# Govulncheck triage 2026-05-24

Источник baseline: [`tests/baseline/phase1/govulncheck.txt`](../../../tests/baseline/phase1/govulncheck.txt).
Дополнительный verbose-снимок для uncalled inventory: [`tests/baseline/phaseV/govulncheck-verbose-before.txt`](../../../tests/baseline/phaseV/govulncheck-verbose-before.txt).
Исходный module graph: [`tests/baseline/phaseV/go-mod-list-before.txt`](../../../tests/baseline/phaseV/go-mod-list-before.txt).

## Scope

- Разрешённые изменения: `go.mod`, `go.sum`, govulncheck/audit artifacts и audit docs.
- Production-код, frontend и npm dependencies не меняются.
- Мажорные апгрейды не выполняются.
- Severity в таблицах: `govulncheck` baseline не отдаёт CVSS/severity, поэтому severity указана как `n/a`; приоритет triage задаётся полем `called`.

## Called vulnerabilities

| GO-ID / CVE | модуль | called/uncalled | severity | текущая | целевая | риск апгрейда | привязка к реестру |
|---|---|---:|---|---|---|---|---|
| GO-2026-5026 | `golang.org/x/net` | called | n/a | v0.51.0 | v0.55.0 | minor bump, direct dep; проверить IDNA/HTTP paths build/test/race | security dependency hygiene, п. 30/SSRF hostname validation |
| GO-2026-5020 | `golang.org/x/crypto` | called | n/a | v0.48.0 | v0.52.0 | minor bump, direct dep; проверить SSH import source/core SSH paths | п. 1/2 importxui source-ssh surface |
| GO-2026-5019 | `golang.org/x/crypto` | called | n/a | v0.48.0 | v0.52.0 | same as above | п. 1/2 importxui source-ssh surface |
| GO-2026-5018 | `golang.org/x/crypto` | called | n/a | v0.48.0 | v0.52.0 | same as above | п. 1/2 importxui source-ssh surface |
| GO-2026-5017 | `golang.org/x/crypto` | called | n/a | v0.48.0 | v0.52.0 | same as above | п. 1/2 importxui source-ssh surface |
| GO-2026-5013 | `golang.org/x/crypto` | called | n/a | v0.48.0 | v0.52.0 | same as above | п. 1/2 importxui source-ssh surface |
| GO-2026-4986 | Go stdlib `net/mail` | called | n/a | go1.26.2 | go1.26.3 | patch toolchain/runtime alignment; no production code change | CI/release toolchain hygiene |
| GO-2026-4982 | Go stdlib `html/template` | called | n/a | go1.26.2 | go1.26.3 | patch toolchain/runtime alignment; no production code change | CI/release toolchain hygiene |
| GO-2026-4980 | Go stdlib `html/template` | called | n/a | go1.26.2 | go1.26.3 | patch toolchain/runtime alignment; no production code change | CI/release toolchain hygiene |
| GO-2026-4977 | Go stdlib `net/mail` | called | n/a | go1.26.2 | go1.26.3 | patch toolchain/runtime alignment; no production code change | CI/release toolchain hygiene |
| GO-2026-4971 | Go stdlib `net` | called | n/a | go1.26.2 | go1.26.3 | patch toolchain/runtime alignment; no production code change | CI/release toolchain hygiene |
| GO-2026-4918 | `golang.org/x/net` + Go stdlib `net/http` | called | n/a | x/net v0.51.0, go1.26.2 | x/net v0.53.0+, go1.26.3 | covered by x/net v0.55.0 + Go patch runtime | security dependency hygiene |

## Uncalled vulnerabilities

### Imported packages, no vulnerable symbol call

| GO-ID / CVE | модуль | called/uncalled | severity | текущая | целевая | риск апгрейда | привязка к реестру |
|---|---|---:|---|---|---|---|---|
| GO-2026-5024 | `golang.org/x/sys` | uncalled-package | n/a | v0.41.0 | v0.44.0 | indirect; фиксить только если подтянется безопасно через direct deps | dependency hygiene |
| GO-2026-5023 | `golang.org/x/crypto` | uncalled-package | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-5016 | `golang.org/x/crypto` | uncalled-package | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-5015 | `golang.org/x/crypto` | uncalled-package | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-5014 | `golang.org/x/crypto` | uncalled-package | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-4981 | Go stdlib `net` | uncalled-package | n/a | go1.26.2 | go1.26.3 | covered only by Go patch runtime/toolchain | CI/release toolchain hygiene |
| GO-2026-4976 | Go stdlib `net/http/httputil` | uncalled-package | n/a | go1.26.2 | go1.26.3 | covered only by Go patch runtime/toolchain | CI/release toolchain hygiene |

### Required modules, no vulnerable package/symbol call

| GO-ID / CVE | модуль | called/uncalled | severity | текущая | целевая | риск апгрейда | привязка к реестру |
|---|---|---:|---|---|---|---|---|
| GO-2026-5033 | `golang.org/x/crypto` | uncalled-module | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-5030 | `golang.org/x/net` | uncalled-module | n/a | v0.51.0 | v0.55.0 | covered by called x/net bump | dependency hygiene |
| GO-2026-5029 | `golang.org/x/net` | uncalled-module | n/a | v0.51.0 | v0.55.0 | covered by called x/net bump | dependency hygiene |
| GO-2026-5028 | `golang.org/x/net` | uncalled-module | n/a | v0.51.0 | v0.55.0 | covered by called x/net bump | dependency hygiene |
| GO-2026-5027 | `golang.org/x/net` | uncalled-module | n/a | v0.51.0 | v0.55.0 | covered by called x/net bump | dependency hygiene |
| GO-2026-5025 | `golang.org/x/net` | uncalled-module | n/a | v0.51.0 | v0.55.0 | covered by called x/net bump | dependency hygiene |
| GO-2026-5021 | `golang.org/x/crypto` | uncalled-module | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-5006 | `golang.org/x/crypto` | uncalled-module | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |
| GO-2026-5005 | `golang.org/x/crypto` | uncalled-module | n/a | v0.48.0 | v0.52.0 | covered by called x/crypto bump | dependency hygiene |

## Planned upgrade order

1. `go mod edit -go=1.26.3` for stdlib patch alignment, then `go mod tidy`, `go build ./...`, `go test ./... -count=1 -timeout 5m`, `go test ./... -race -count=1 -timeout 10m`.
2. `go get golang.org/x/crypto@v0.52.0`, then the same sanity commands.
3. `go get golang.org/x/net@v0.55.0`, then the same sanity commands.
4. Final `go mod tidy`, `govulncheck ./...`, audit pipeline sanity and delta documentation.

## Upgrade results

| Шаг | Изменение | Sanity result | Примечание |
|---|---|---|---|
| Stdlib/toolchain | добавлен `toolchain go1.26.3` при сохранении `go 1.25.7` | `go mod tidy`, `go build ./...`, `go test ./... -count=1 -timeout 5m`, `go test ./... -race -count=1 -timeout 10m` green | Язык модуля не повышался с 1.25.x до 1.26.x; в module context `go version` стал `go1.26.3 windows/amd64`. |
| `golang.org/x/crypto` | v0.48.0 -> v0.52.0 | `go mod tidy`, `go build ./...`, `go test ./... -count=1 -timeout 5m`, `go test ./... -race -count=1 -timeout 10m` green | Закрывает called GO-2026-5020, GO-2026-5019, GO-2026-5018, GO-2026-5017, GO-2026-5013 и uncalled crypto findings. |
| `golang.org/x/net` | v0.51.0 -> v0.55.0 | `go mod tidy`, `go build ./...`, `go test ./... -count=1 -timeout 5m`, `go test ./... -race -count=1 -timeout 10m` green | Закрывает GO-2026-5026 и module half of GO-2026-4918. |
| Transitive `x/*` deps | `x/term` v0.40.0 -> v0.43.0; `x/sys` v0.41.0 -> v0.45.0; `x/text` v0.34.0 -> v0.37.0; `x/mod` v0.33.0 -> v0.35.0; `x/sync` v0.19.0 -> v0.20.0; `x/tools` v0.42.0 -> v0.44.0 | pulled by `x/crypto`/`x/net`, final tests green | `x/sys` crosses the uncalled GO-2026-5024 fix level v0.44.0. |

## Govulncheck delta

| Группа | Baseline | After | Delta |
|---|---:|---:|---:|
| called / Symbol Results | 12 | 0 | -12 |
| imported package, no called symbol | 7 | 0 | -7 |
| required module only | 9 | 0 | -9 |

Closed called GO-IDs:
GO-2026-5026, GO-2026-5020, GO-2026-5019, GO-2026-5018, GO-2026-5017, GO-2026-5013, GO-2026-4986, GO-2026-4982, GO-2026-4980, GO-2026-4977, GO-2026-4971, GO-2026-4918.

Residual called vulnerabilities: none.

Artifacts:
- [`tests/baseline/phase1/govulncheck-after.txt`](../../../tests/baseline/phase1/govulncheck-after.txt): `No vulnerabilities found.`
- [`tests/baseline/phaseV/govulncheck-verbose-after.txt`](../../../tests/baseline/phaseV/govulncheck-verbose-after.txt): verbose scan also reports `No vulnerabilities found.`

## Command status after upgrades

| Command | Status | Artifact |
|---|---:|---|
| `go build ./...` | green | [`tests/baseline/phaseV/audit-build.txt`](../../../tests/baseline/phaseV/audit-build.txt) |
| `go vet ./...` | green | [`tests/baseline/phaseV/audit-vet.txt`](../../../tests/baseline/phaseV/audit-vet.txt) |
| `go test ./... -count=1 -timeout 5m` | green | [`tests/baseline/phaseV/audit-test-go.txt`](../../../tests/baseline/phaseV/audit-test-go.txt) |
| `go test ./... -race -count=1` | green in this run; known p. 47 did not reproduce | [`tests/baseline/phaseV/audit-test-go-race.txt`](../../../tests/baseline/phaseV/audit-test-go-race.txt) |
| `gosec ./...` | red baseline, not worse: 55 -> 55 | [`tests/baseline/phaseV/gosec-delta.txt`](../../../tests/baseline/phaseV/gosec-delta.txt) |
| `govulncheck ./...` | green | [`tests/baseline/phaseV/audit-vuln.txt`](../../../tests/baseline/phaseV/audit-vuln.txt) |
| `make audit:*` | skipped locally: `make` unavailable on Windows host | [`tests/baseline/phaseV/make-unavailable.txt`](../../../tests/baseline/phaseV/make-unavailable.txt) |

No production-code regressions were introduced by the dependency bumps. No new registry item is needed.
