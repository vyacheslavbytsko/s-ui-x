# Phase 4 gosec triage

Дата: 2026-05-24.

Источник: [`tests/baseline/phase1/gosec.txt`](../../../tests/baseline/phase1/gosec.txt). В production-код не добавлялись `//nosec` / `//nolint` комментарии.

Итоговая классификация: true positive — 16, nosec — 18, mitigated_in_review — 21.

| file:line | rule | классификация | реестр | риск | рекомендуемый фикс / обоснование |
|---|---|---:|---|---:|---|
| `database/importxui/plan.go:260` | G115 | true_positive | security triage gap | P2 | Проверять `row.ID >= 0` и верхнюю границу перед `uint(row.ID)`; импортируемая SQLite БД является внешним вводом. |
| `database/importxui/plan.go:262` | G115 | true_positive | security triage gap | P2 | То же: безопасная функция конвертации x-ui ID -> destination ID. |
| `service/server.go:161` | G115 | nosec | п.24 контекстно | P3 | `runtime.NumGoroutine()` в ответе статистики практически не достигает `uint32` overflow; это не security primitive. Лучше вернуть `int`/`uint64` для чистоты. |
| `service/server.go:174` | G115 | nosec | п.24 контекстно | P3 | Аналогично `appThreads`; confidentiality-risk в п.24 отдельный от арифметики. |
| `service/user.go:344` | G115 | mitigated_in_review | п.33 смежно | P3 | Вызывается из `apiTokenScopeAllowed`, где `len(scope) <= len("observability")`; добавить регрессию на allowed scopes и при будущем reuse заменить `byte` на `int`. |
| `util/common/random.go:18` | G404 | true_positive | security triage gap | P2 | `math/rand` fallback используется для tokens/CSRF/settings secrets при ошибке `crypto/rand`; fail-closed или возврат ошибки вместо слабой энтропии. |
| `service/secret_settings.go:35` | G101 | nosec | п.25 контекстно | P3 | `StoredSecretMarker` — UI/API marker `stored`, не credential и не secret. |
| `api/csrf.go:16` | G101 | nosec | CSRF matrix | P3 | `CSRF_TOKEN` / `CSRF_EXPIRES` — имена session keys, не секретные значения. |
| `api/apiV2Handler.go:35` | G101 | nosec | п.34 | P3 | `legacyTokenHeaderSunset` — HTTP date для deprecation, не credential. |
| `database/importxui/profile_crypto.go:163` | G703 | mitigated_in_review | xui profile crypto | P2 | `XUI_PROFILE_KEY_FILE` operator-controlled; усилить проверкой absolute path/permissions и явным audit при fallback. |
| `cmd/migration/1_2.go:31` | G703 | nosec | п.14 | P3 | Local migration читает путь из локальной runtime config, не из remote request. |
| `cmd/migration/1_2.go:35` | G703 | nosec | п.14 | P3 | То же; path traversal как web threat не применим к локальной migration-команде. |
| `database/importxui/source/xuihttp/xuihttp.go:104` | G124 | nosec | xui remote | P3 | `AddCookie` формирует outbound request к x-ui; `Secure/HttpOnly/SameSite` относятся к `Set-Cookie` response. |
| `database/importxui/source_iface.go:64` | G304 | mitigated_in_review | п.38 | P3 | Admin/database-scope feature читает выбранный SQLite source; есть SQLite/integrity validation. Для hardening — root-bound file picker / allowlist. |
| `database/importxui/source/ssh/ssh.go:108` | G304 | mitigated_in_review | п.38 | P3 | Запись идёт в temp-dir `source.db`; имя фиксировано. Оставить cleanup + добавить temp-dir lifecycle tests. |
| `database/importxui/source.go:153` | G304 | mitigated_in_review | п.38 | P3 | Хешируется уже полученный local source path; связать с root-bound source validation. |
| `database/importxui/profile_crypto.go:163` | G304 | mitigated_in_review | п.38, xui profile crypto | P2 | Operator key-file path; не remote-controlled, но стоит ограничить absolute path/permissions. |
| `database/backup.go:118` | G304 | mitigated_in_review | п.11, п.38 | P3 | `dbPath` берётся из config/current DB lifecycle; не remote path. Добавить path/root invariant tests при backup. |
| `database/backup.go:338` | G304 | mitigated_in_review | п.11, п.38 | P3 | `dst` — temp restore path рядом с DB; до rename есть SQLite validation. |
| `cmd/migration/1_2.go:35` | G304 | nosec | п.14, п.38 | P3 | Local migration file read; не web/admin endpoint. |
| `cmd/decrypt_backup.go:54` | G304 | nosec | backup CLI, п.38 | P3 | CLI явно принимает `--in`; чтение произвольного файла — контракт локальной утилиты. |
| `api/import_xui.go:220` | G304 | mitigated_in_review | п.38, п.39 | P2 | `validateRollbackPath` проверяет symlink, regular file, имя `s-ui-pre-xui-import-*.db` и каталог DB; закрепить Phase 4 tests. |
| `api/import_xui.go:312` | G304 | mitigated_in_review | п.38 | P3 | Multipart upload пишется в `MkdirTemp` + фиксированное `source.db` с mode `0600`; cleanup через defer. |
| `api/import_xui.go:362` | G304 | mitigated_in_review | п.38 | P3 | Открывается temp upload path после записи; есть SQLite header/integrity validation. |
| `database/importxui/profile_crypto.go:62` | G117 | mitigated_in_review | xui profile crypto | P2 | Source может содержать `Password`, но сразу шифруется AEAD; добавить wipe raw plaintext после encrypt. |
| `service/client.go:263` | G602 | nosec | none | P3 | `index` приходит из `for index := range clients`; out-of-range не достижим без гонки/мутации slice. |
| `web/web.go:184` | G104 | nosec | shutdown path | P3 | Cleanup in error path; ошибка `Stop` не меняет исходную причину `Start`. Можно логировать. |
| `web/web.go:221` | G104 | nosec | shutdown path | P3 | `listener.Close()` при ошибке TLS load; best-effort cleanup. |
| `sub/sub.go:168` | G104 | nosec | shutdown path | P3 | Аналогично `web.Start` cleanup. |
| `sub/sub.go:207` | G104 | nosec | shutdown path | P3 | Best-effort listener cleanup при TLS load failure. |
| `sub/jsonService.go:93` | G104 | mitigated_in_review | subscription config | P3 | Ошибка `addOthers` может скрыть invalid extra config; вернуть ошибку в будущем, сейчас не secret/audit path. |
| `service/inbounds.go:92` | G104 | mitigated_in_review | inbound contract | P3 | Invalid stored JSON silently defaults `shadowtls_version`; лучше fail-visible при corrupted config. |
| `service/inbounds.go:95` | G104 | mitigated_in_review | inbound contract | P3 | Аналогично `ss_managed`; не remote exploit без доступа к DB/settings. |
| `network/auto_https_conn.go:46` | G104 | nosec | network cleanup | P3 | Redirect write failure не влияет на security state; логирование достаточно. |
| `network/auto_https_conn.go:47` | G104 | nosec | network cleanup | P3 | Close после redirect best-effort. |
| `core/tracker_conn.go:97` | G104 | nosec | connection cleanup | P3 | Force-close tracked connection; ошибка close не меняет удаление из tracker. |
| `core/tracker_conn.go:100` | G104 | nosec | connection cleanup | P3 | То же для packet conn. |
| `core/log.go:184` | G104 | true_positive | п.16 смежно | P3 | Потеря core log write errors снижает наблюдаемость; возвращать/считать ошибки writer. |
| `core/box.go:408` | G104 | mitigated_in_review | core startup | P3 | Cleanup после init error; исходная ошибка сохраняется. Логировать close error при hardening. |
| `cmd/migration/1_1.go:31` | G104 | true_positive | п.14 | P2 | `rows.Scan` error может исказить schema migration; проверять и прерывать. |
| `cmd/migration/1_1.go:48` | G104 | true_positive | п.14 | P2 | Ошибка JSON migration silently drops/changes config; fail migration. |
| `cmd/migration/1_1.go:52` | G104 | true_positive | п.14 | P2 | Аналогично links JSON migration. |
| `cmd/migration/1_2.go:46` | G104 | true_positive | п.14 | P2 | `DropTable` в migration без проверки может оставить mixed schema. |
| `cmd/migration/1_2.go:47` | G104 | true_positive | п.14 | P2 | `AutoMigrate` error должен прерывать upgrade. |
| `cmd/migration/1_2.go:83` | G104 | true_positive | п.14 | P2 | Ошибка JSON parse адресов может привести к потере endpoint данных. |
| `cmd/migration/1_2.go:128` | G104 | true_positive | п.14 | P2 | `DropTable` outbounds/endpoints без проверки. |
| `cmd/migration/1_2.go:129` | G104 | true_positive | п.14 | P2 | `AutoMigrate` outbounds/endpoints без проверки. |
| `cmd/migration/1_2.go:152` | G104 | true_positive | п.14 | P2 | Ошибка JSON options ignored при migration. |
| `app/app.go:54` | G104 | true_positive | п.19 | P2 | `GetAllSetting` инициализирует defaults; startup должен fail-visible при ошибке. |
| `app/app.go:158` | G104 | true_positive | app restart | P2 | `RestartApp` игнорирует `Start` error после `Stop`; может оставить панель частично поднятой. |
| `api/session.go:90` | G104 | true_positive | session security | P2 | `ClearSession` должен вернуть/логировать `Save` error; иначе logout может выглядеть успешным без очистки cookie/store. |
| `api/apiService.go:363` | G104 | mitigated_in_review | database export | P3 | Response write после audit/export headers; при hardening audit должен различать started vs completed. |
| `api/apiService.go:436` | G104 | mitigated_in_review | п.25 | P3 | Encrypted backup response write error не раскрывает secret, но audit completed/failed стоит уточнить. |
| `api/apiService.go:905` | G104 | mitigated_in_review | config export | P3 | Error-body write best-effort; логирование достаточно. |
| `api/apiService.go:910` | G104 | mitigated_in_review | config export | P3 | Config response write error лучше учитывать в audit/metrics, но это не auth bypass. |

## Notes

- Все `G304` привязаны к п.38 как общему xui/import/temp-file anchor, даже когда конкретный путь относится к backup или CLI; локальные CLI findings отдельно помечены `nosec`.
- `G101` findings в baseline являются false positive: marker/session-key/date, а не секреты.
- Для `tx.Where("... ?", args...)` SQL injection false positive в текущем `gosec.txt` не найден: query strings константные и параметры передаются через GORM placeholders.
