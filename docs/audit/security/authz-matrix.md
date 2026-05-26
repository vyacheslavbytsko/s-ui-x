# Phase 4 AuthZ matrix

Дата: 2026-05-24.

Важно: фактический v2 prefix в коде — `/apiv2` (`web/web.go`), хотя в плане фазы он назван `/api/v2`.

## Наблюдаемый контракт

- v2 использует token middleware `APIv2Handler.checkToken`. При отсутствующем, неверном или expired token сейчас возвращается HTTP 200 с `success:false`, а не 401/403. Это зафиксировано как Phase 4 auth-status gap.
- v1 использует browser session middleware `checkLogin` и CSRF middleware. Без session browser-flow отдаёт 307 redirect на `/login`; XHR-flow отдаёт HTTP 200 с `success:false`.
- `requireTokenScopeAny` пропускает browser session без token scope. Поэтому v1 scope в таблице означает implicit admin session.
- Дублированные `/import-xui/*` routes в v1 и v2 остаются issue 35: v1 protected by session+CSRF, v2 protected by API token; часть remote routes в v2 требует только `xui_remote`, а не `admin`.

## v2 `/apiv2`

| Метод/путь | Требуемый scope | Middleware | Без токена | Wrong scope | Expired token | Correct scope |
|---|---|---|---|---|---|---|
| GET `/apiv2/security/audit` | `admin` | token | 200 `success:false` | 403 | 200 `success:false` | 200 |
| POST `/apiv2/rotateSubSecret` | `admin` или `write` | token | 200 `success:false` | 403 | 200 `success:false` | 200 |
| POST `/apiv2/telegram/test` | `admin` | token | 200 `success:false` | 403 | 200 `success:false` | 200 |
| POST `/apiv2/telegram/backup` | `admin` или `telegram` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/telegram/backup/run` | `admin` или `telegram` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/plan` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/apply` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/rollback` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| GET `/apiv2/import-xui/reports` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/remote/plan` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/remote/apply` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| GET `/apiv2/import-xui/remote/status` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | 200 |
| GET `/apiv2/import-xui/sync/profiles` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/sync/profiles` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/sync/run` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui/sync/disable` | `xui_remote` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/save` | any valid token (gap: should be `write/admin`) | token | 200 `success:false` | N/A | 200 `success:false` | reaches handler |
| POST `/apiv2/restartApp` | any valid token (gap) | token | 200 `success:false` | N/A | 200 `success:false` | reaches handler |
| POST `/apiv2/restartSb` | any valid token (gap) | token | 200 `success:false` | N/A | 200 `success:false` | reaches handler |
| POST `/apiv2/linkConvert` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | reaches handler |
| POST `/apiv2/subConvert` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | reaches handler |
| POST `/apiv2/importdb` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| POST `/apiv2/import-xui` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| GET `/apiv2/load` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/inbounds` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/outbounds` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/endpoints` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/services` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/tls` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/clients` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/config` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/users` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/settings` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/stats` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/status` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/onlines` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/logs` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/changes` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/keypairs` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | 200 |
| GET `/apiv2/getdb` | `admin` или `database` | token | 200 `success:false` | 403 | 200 `success:false` | reaches handler |
| GET `/apiv2/checkOutbound` | any valid token | token | 200 `success:false` | N/A | 200 `success:false` | reaches handler |

## v1 `/api`

| Метод/путь | Требуемый scope | Middleware | Без session | Wrong scope | Expired token | Correct session |
|---|---|---|---|---|---|---|
| POST `/api/login` | public | CSRF-exempt | reaches handler | N/A | N/A | reaches handler |
| GET `/api/logout` | public logout helper | session-exempt, no CSRF | 200 | N/A | N/A | 200 |
| GET `/api/csrf` | admin browser session | session | 307/200 XHR false | N/A | N/A | 200 |
| GET `/api/load` and partial GETs | admin browser session | session | 307/200 XHR false | N/A | N/A | 200 |
| GET `/api/users`, `/api/settings`, `/api/stats`, `/api/status`, `/api/onlines`, `/api/logs`, `/api/changes`, `/api/keypairs`, `/api/tokens`, `/api/singbox-config`, `/api/checkOutbound`, `/api/version` | admin browser session | session | 307/200 XHR false | N/A | N/A | reaches handler |
| GET `/api/getdb` | admin browser session; handler allows no token scope | session | 307/200 XHR false | N/A | N/A | reaches handler |
| GET `/api/security/audit` | admin browser session | session | 307/200 XHR false | N/A | N/A | 200 |
| GET `/api/realtime/ws-token` | admin browser session | session | 307/200 XHR false | N/A | N/A | 200 |
| GET `/api/realtime/ws` | admin browser session + ws token | session | 307/200 XHR false | N/A | N/A | websocket auth |
| GET `/api/ip-monitor/:client` | admin browser session | session | 307/200 XHR false | N/A | N/A | reaches handler |
| GET `/api/observability/history`, `/api/observability/core-history` | admin browser session | session | 307/200 XHR false | N/A | N/A | 200 |
| POST `/api/changePass`, `/api/save`, `/api/restartApp`, `/api/restartSb`, `/api/linkConvert`, `/api/subConvert`, `/api/addToken`, `/api/deleteToken`, `/api/setTokenEnabled`, `/api/logoutAllAdmins`, `/api/checkOutbounds`, `/api/rotateSubSecret`, `/api/ip-monitor/:client/clear` | admin browser session | session + CSRF | 307/200 XHR false before CSRF | N/A | N/A | CSRF-gated |
| POST `/api/importdb`, `/api/import-xui`, `/api/import-xui/plan`, `/api/import-xui/apply`, `/api/import-xui/rollback` | admin browser session; handler database gate is bypassed for no token scope | session + CSRF | 307/200 XHR false before CSRF | N/A | N/A | CSRF-gated |
| POST `/api/import-xui/remote/plan`, `/api/import-xui/remote/apply`, `/api/import-xui/sync/profiles`, `/api/import-xui/sync/run`, `/api/import-xui/sync/disable` | admin browser session; handler allows no token scope | session + CSRF | 307/200 XHR false before CSRF | N/A | N/A | CSRF-gated |
| POST `/api/telegram/test`, `/api/telegram/backup`, `/api/telegram/backup/run` | admin browser session; handler allows no token scope | session + CSRF | 307/200 XHR false before CSRF | N/A | N/A | CSRF-gated |

## Regression anchors

- [`api/security_authz_test.go`](../../../api/security_authz_test.go) anchors scope matrix rows with `requireTokenScopeAny` and records current HTTP-status gaps as skipped XFAIL.
- [`api/security_csrf_test.go`](../../../api/security_csrf_test.go) anchors all v1 POST routes for missing/expired/rotated CSRF rejection.
