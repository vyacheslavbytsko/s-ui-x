# Frontend accessibility and headers — Phase 6

Дата прогона: 2026-05-24.

Артефакты:
- Axe JSON: `tests/baseline/phase6/a11y/axe-results.json`.
- Security headers JSON: `tests/baseline/phase6/security-headers/headers.json`.
- Playwright JUnit: `tests/baseline/phase6/playwright.junit.xml`.

## Axe baseline

`@axe-core/playwright` прогнан по login, dashboard, migrate-xui, settings, audit на локальной Vite dev-инстанции с API-only test panel и SQLite DB в `tests/baseline/phase6/e2e-db`.

| Страница | Violations | Критичные темы |
|---|---:|---|
| login | 6 | `button-name`, `label`, `color-contrast`, landmarks/H1 |
| dashboard | 5 | `aria-required-children`, `button-name`, `image-alt`, landmarks/H1 |
| migrate-xui | 5 | `aria-required-children`, `button-name`, `color-contrast`, landmarks/H1 |
| settings | 5 | `aria-required-children`, `button-name`, `color-contrast`, landmarks/H1 |
| audit | 6 | `aria-required-children`, `aria-tooltip-name`, `button-name`, `empty-table-header`, landmarks/H1 |

Baseline не делает axe violations blocking gate: это документированный backlog для отдельного accessibility-fix диалога. Playwright test сохраняет полный JSON и остаётся green, чтобы Phase 6 e2e smoke не превращался в unrelated production-fix.

## Security headers

Проверка `frontend/tests/e2e/security-headers.spec.ts` обращается к backend test panel напрямую (`http://127.0.0.1:2095/app/`), а не к Vite, чтобы проверять `middleware.AdminSecurityHeaders()`.

Green:
- `Content-Security-Policy` содержит `frame-ancestors 'none'`.
- `X-Frame-Options: DENY`.
- `X-Content-Type-Options: nosniff`.
- `Referrer-Policy: strict-origin-when-cross-origin`.

Skipped by condition:
- `Strict-Transport-Security` проверяется только для HTTPS URL. Phase 6 test panel работает по HTTP, поэтому HSTS отсутствие не считается finding.
