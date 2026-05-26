# Phase 5 performance baseline

Дата прогона: 2026-05-24. Окружение: Windows amd64, Go `go1.26.2`; подробности в [`tests/baseline/env.md`](../../../tests/baseline/env.md).

Все сценарии ниже in-process/`httptest`; live-нагрузка на внешние сервисы не запускалась. Значение зафиксировано как регресс-порог для последующих фиксов. Сырые логи: [`bench-backup.txt`](../../../tests/baseline/phase5/bench-backup.txt), [`bench-all.txt`](../../../tests/baseline/phase5/bench-all.txt), очищенная выжимка benchmark lines: [`bench-all.clean.txt`](../../../tests/baseline/phase5/bench-all.clean.txt).

## Командная сводка

| Команда | Статус | Артефакт |
|---|---:|---|
| `go build ./...` | green | [`build.txt`](../../../tests/baseline/phase5/build.txt) |
| `go vet ./...` | green | [`vet.txt`](../../../tests/baseline/phase5/vet.txt) |
| `go test ./... -count=1 -timeout 5m` | green | [`test.txt`](../../../tests/baseline/phase5/test.txt) |
| `go test ./... -race -count=1 -timeout 15m` | green | [`test-race.txt`](../../../tests/baseline/phase5/test-race.txt) |
| `go test ./... -bench=. -benchmem -run=^$ -benchtime=2s` | green | [`bench-all.txt`](../../../tests/baseline/phase5/bench-all.txt) |
| coverage | green, total 43.0% | [`coverage-func.txt`](../../../tests/baseline/phase5/coverage/coverage-func.txt) |

## Benchmark anchors

| Сценарий | Baseline | Alloc/op | Доп. метрика | Реестр |
|---|---:|---:|---|---|
| `GetDb`, 100k rows/table, full | 5,996,446,400 ns/op | 616,592,008 B/op | 29,151,232 bytes out | backup baseline |
| `GetDb`, 100k rows/table, exclude heavy | 111,606,533 ns/op | 9,246,633 B/op | 606,208 bytes out | backup baseline |
| `GetDb`, 1M rows/table, full | 84,816,073,700 ns/op | 5,985,488,896 B/op | 289,628,160 bytes out | backup baseline |
| `GetDb`, 1M rows/table, exclude heavy | 110,659,350 ns/op | 9,240,950 B/op | 606,208 bytes out | backup baseline |
| `StatsService.SaveStats`, 100 clients | 7,805,924 ns/op | 398,025 B/op | 3,783 allocs/op | stats pipeline |
| `StatsService.SaveStats`, 1000 clients | 77,639,290 ns/op | 4,013,931 B/op | ~9.95x time for 10x N | stats pipeline |
| `updateClientTrafficDeltas`, 1000 clients, 0% empty | 25,489,474 ns/op | 1,252,169 B/op | 12,949 allocs/op | stats pipeline |
| `updateClientTrafficDeltas`, 1000 clients, 90% empty | 2,290,403 ns/op | 142,095 B/op | 1,331 allocs/op | stats pipeline |
| `AuditWriter_Overload10000` | 150,544,238 ns/op | 875,846 B/op | 5,904 drops/op | #16 |
| `TelegramNotifier_Enqueue`, fake send | 351.9 ns/op | 2 B/op | 7,016,416 sent | reliability queue |
| `TelegramNotifier_Enqueue`, blocked sender | 11,274 ns/op | 567 B/op | 186,710 overflows | reliability queue |
| `TokenUseDebouncer_Record`, parallel 1/16/64 | 231.9 / 335.3 / 454.7 ns/op | 0 B/op | 10,000 flushed unique | token use |
| `TokenUseDebouncer_BatchFlush` | 345,551,429 ns/op | 12,204,098 B/op | 142,671 allocs/op | token use |
| `importxui Plan`, 100 / 1000 wireguard inbounds | 14,580,779 / 208,831,580 ns/op | 1.35 MB / 13.18 MB | synthetic source | importxui |
| `importxui Apply dry-run`, 100 / 1000 | 28,087,296 / 257,673,600 ns/op | 1.84 MB / 18.15 MB | synthetic source | importxui |
| `importxui Apply real`, 100 / 1000 | 136,752,783 / 344,723,883 ns/op | 2.79 MB / 19.10 MB | includes pre-import backup | importxui |
| `ipmonitor loadCacheEntry`, 1000 clients x 5 IP | 133,392 ns/op | 10,892 B/op | warmup 126.9 ms | #18 |
| `ipmonitor Allow`, known IP, 1000 clients x 5 IP | 2,092 ns/op | 464 B/op | 7 allocs/op | #18 |
| `ipmonitor Allow`, reject over limit, 1000 clients x 5 IP | 2,979 ns/op | 880 B/op | 12 allocs/op | #18 |
| `realtime Publish`, 10 / 100 / 1000 subscribers | 4,160 / 31,322 / 361,725 ns/op | 424 / 1,239 / 8,535 B/op | 0 drops | #32 |
| WS connect/disconnect, 10 / 20 clients | 16,931,580 / 35,317,561 ns/op | 629,280 / 1,255,765 B/op | capacity 5/user, 20/IP -> 429 | #32 |
| Cron XUI sync missing source | 303,697,286 ns/op | 31,090 B/op | 3 attempts, 300 ms expected backoff | #40 |
| `GET /load` via `httptest`, parallel=100 | 9,279,263 ns/op | 592,366 B/op | 1,000/1,000 HTTP 200 in load test | API perf |
| `GET /stats` via `httptest`, parallel=100 | 5,027,445 ns/op | 879,479 B/op | 1,000/1,000 HTTP 200 | API perf |
| `GET /onlines` via `httptest`, parallel=100 | 4,066 ns/op | 6,303 B/op | 1,000/1,000 HTTP 200 | API perf |
| `POST /save`, serial settings change | 829,571 ns/op | 65,004 B/op | 988 allocs/op | API perf |
| `GET /import-xui/reports` | 3,226,927 ns/op | 375,337 B/op | 100 requests -> 5x 200, 95x 429 | #36, #44 |

Notes:
- Full backup benches are single-iteration on Windows once operation time exceeds `benchtime`; `bench-all` showed higher 1M full time (128.6s) than dedicated `bench-backup` (84.8s). Treat dedicated `bench-backup` as the local regression-порог and re-confirm on Linux runner before hard gating.
- `realtime` latency distribution uses nanosecond samples around very short operations; Windows timer granularity produced `p50_ns=0` for smaller subscriber sets. Use `ns/op` and `drops` as the primary anchor.

## Reliability anchors

| Пункт | Anchor | Наблюдение |
|---|---|---|
| #16 audit writer severity priority | `TestAuditWriterOverloadSeverityPriorityAnchorIssue16Phase5` | FIFO overflow: `dropped=5904`, `lost_warn_security=5000`, `lost_info=904`, kept only `info=4096`; severity priority отсутствует. |
| Telegram notifier overflow | `TestTelegramNotifierOverflowAnchorPhase5` | capacity 256, enqueue capacity+100 under blocked sender -> `overflows=99`. |
| #18 ipmonitor allow path | `BenchmarkLoadCacheEntry`, `BenchmarkAllow`, targeted anchor | 1000 clients x 5 IP warmup 126.9 ms; allow known ~2.1 us/op; reject over limit ~3.0 us/op. |
| #32 WS reconnect/hub server side | `BenchmarkPublishToNClients`, `BenchmarkRealtimeWSConnectDisconnect`, capacity test | 1000 subscribers publish 361.7 us/op, 0 drops; max per user/IP returns 429. |
| #36 import-xui report rate-limit | `TestAPIImportXUIReportsRateLimitPhase5` | 100 same-actor requests -> 5 allowed, 95 rate-limited. |
| #40 cron sync lost network | `TestXUISyncJobLostNetworkBackoffAnchorIssue40Phase5` | Current retry shape: 3 attempts, 100 ms + 200 ms backoff, elapsed 303.4 ms. |
| #44 rollback/report server-side baseline | `BenchmarkAPI_ImportXUIReports`, rate-limit test | Server-side import report path has a latency/rate-limit anchor for Phase 6 frontend polling/rollback work; frontend e2e remains Phase 6. |

## pprof snapshots

Profiles:
- [`backup-cpu.pprof`](../../../tests/baseline/phase5/profiles/backup-cpu.pprof), top: [`backup-cpu.top.txt`](../../../tests/baseline/phase5/profiles/backup-cpu.top.txt)
- [`backup-mem.pprof`](../../../tests/baseline/phase5/profiles/backup-mem.pprof), top: [`backup-mem.top.txt`](../../../tests/baseline/phase5/profiles/backup-mem.top.txt)
- [`api-load-cpu.pprof`](../../../tests/baseline/phase5/profiles/api-load-cpu.pprof), top: [`api-load-cpu.top.txt`](../../../tests/baseline/phase5/profiles/api-load-cpu.top.txt)

Top CPU, backup:
- `runtime.cgocall`: 49.38s flat / 62.57%.
- Application cumulative: `database.GetDb` at `backup.go:95` 64.02s cum; `copyBackupTable` 64.02s cum; `gorm.FindInBatches` 48.03s cum.

Top alloc, backup:
- `sqlite3.(*SQLiteConn).query`: 846.34 MB.
- `database/sql.driverArgsConnLocked`: 830.89 MB.
- `reflect.unsafe_New`: 741.03 MB.
- Application cumulative: `database.GetDb` 5,717.62 MB; `copyBackupTable` 5,717.62 MB.

Top CPU, API `/load`:
- `runtime.cgocall`: 25.96s flat / 79.32%.
- Application cumulative: `ApiService.LoadData` 29.44s; `getData` -> `SettingService.GetFinalSubURI` 27.04s; `gorm.(*processor).Execute` 27.57s.

## Delta к Phase 4

| Метрика | Phase 4 | Phase 5 | Дельта |
|---|---:|---:|---:|
| Coverage total | 42.7% | 43.0% | +0.3 |
| `api` coverage | 57.5% | 59.5% | +2.0 |
| `service` coverage | 48.1% | 48.1% | +0.0 |
| `database` coverage | 66.5% | 66.5% | +0.0 |
| `ipmonitor` coverage | 82.4% | 82.4% | +0.0 |
| `realtime` coverage | 92.5% | 92.5% | +0.0 |

Go build/vet/test/race остались green. Phase 3 race finding (#47) не воспроизвёлся в Phase 5 full `-race`, но остаётся отдельным зарегистрированным concurrency issue для отдельного фикса.
