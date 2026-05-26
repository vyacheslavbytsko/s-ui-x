# Phase 4 CSRF matrix

Дата: 2026-05-24.

CSRF middleware подключён только в v1 browser API (`APIHandler.initRouter`). v2 `/apiv2` использует bearer/legacy API token и не использует session CSRF.

## Exceptions

| Метод/путь | Причина исключения | Наблюдаемое поведение |
|---|---|---|
| POST `/api/login` | login endpoint explicitly exempt via `csrfExemptPath` | без CSRF доходит до login handler |
| GET `/api/logout` | safe method + session-exempt suffix `logout` | без CSRF доходит до logout helper |
| GET `/api/csrf` | safe method | требует session, выпускает новый CSRF token |

## Protected POST routes

Все строки ниже требуют browser session и валидный `X-CSRF-Token`; без token и с expired token middleware возвращает 403 до вызова handler. После session generation rotation запрос отклоняется раньше на session middleware: browser-flow получает redirect, XHR-flow — `success:false`.

| Метод/путь | Middleware | Без CSRF | Expired CSRF | Rotated session |
|---|---|---:|---:|---:|
| POST `/api/changePass` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/save` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/restartApp` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/restartSb` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/linkConvert` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/subConvert` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/importdb` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/plan` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/apply` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/rollback` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/remote/plan` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/remote/apply` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/sync/profiles` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/sync/run` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/import-xui/sync/disable` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/addToken` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/deleteToken` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/setTokenEnabled` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/logoutAllAdmins` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/checkOutbounds` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/rotateSubSecret` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/telegram/test` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/telegram/backup` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/telegram/backup/run` | session + CSRF | 403 | 403 | reject before handler |
| POST `/api/ip-monitor/:client/clear` | session + CSRF | 403 | 403 | reject before handler |

Regression anchor: [`api/security_csrf_test.go`](../../../api/security_csrf_test.go).
