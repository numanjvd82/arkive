# Arkive Architecture Notes

## Current Structure (Review)
- Entry: `main.go` wires config, database pool, migrations, and router.
- Routing: `core/router` builds the Gin engine and wires handlers/services.
- Handlers (HTTP): `core/handlers` is thin; each handler calls a single service method.
- Services (business logic): `core/services` performs validation, orchestration, and transactions.
- Repositories (SQL): `core/repositories` owns all SQL; repos accept a `database.PgExecutor` (pool/tx).
- Database: `core/database` defines the PgExecutor/PgPool interfaces and connection setup.
- Web UI: `core/web` provides layout, pages, and components (gomponents).
- Static assets: `core/web/static` is embedded via `core/web/assets.go` and served via `StaticFS`.
- Shared packages: `pkg/` hosts cross-cutting utilities (e.g., `tokens`, `cookies`, `storage`).
- Migrations: `migrations/` has up/down SQL files and a simple runner.

## Conventions Observed
- SQL is uppercase and kept in repos only.
- Services start a transaction per request and call repo methods with the tx.
- Web pages use reusable components (inputs, cards, buttons, icons).
- Handlers parse form inputs into structs before calling services.
- Validation errors are returned as a `validation.Errors` map and rendered inline in forms.
- Validation error messages live in `core/services*/errors.go` and are referenced from services. That way we have full control over wording and can reuse messages across services.
- Prefer modern JS: use `const`/`let` instead of `var`.
- Typography uses tokens in `core/web/static/globals.css` (Inter for UI/body, Plus Jakarta Sans for headings, JetBrains Mono for metadata). Prefer the CSS variables over hardcoded sizes.

## Potential Improvements
- **Modularization**: Consider splitting large services into smaller, focused modules.
- **Error Handling**: Standardize error responses across handlers for consistency.
- **Testing**: Increase unit test coverage for services and repositories.
- **Documentation**: Add more inline comments and README files for complex packages.
- **Configuration Management**: Explore using a configuration library for better environment handling.


## Don't Forget
- Keep dependencies minimal to avoid bloat.
- Regularly review and refactor code to maintain clarity and efficiency.
- Ensure security best practices are followed, especially in authentication and data handling.
- Stay updated with Gin and other library changes to leverage new features.
- Engage in code reviews to maintain high code quality and share knowledge among the team.
- Monitor performance and optimize database queries as needed.
- Don't run destructive commands like `git reset --hard` without backups or confirmation.

## Recent Updates
- Added `pkg/validation` for reusable validation helpers and error maps.
- Auth signup checks uniqueness via repo lookups (`GetUserByEmail`, `GetUserByBrandName`) before insert.
- Added AdBlock detection modal using shared Dialog/Button components for authenticated pages.
- Review `TODO.md` and suggest next steps when that implementation is being discussed.
