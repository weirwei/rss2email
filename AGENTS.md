# Repository Guidelines

## Project Structure & Module Organization
This repository is a Go CLI service that fetches RSS feeds and sends email updates.

- `main.go`: application entrypoint.
- `cmd/`: Cobra commands (`root`, `register`, `add-feed`, `sql`).
- `service/`: source-specific feed logic and scheduler wiring.
- `rss/`: feed fetching/parsing.
- `models/`: GORM models and DB access.
- `helpers/` and `conf/`: initialization for DB, email, and YAML config.
- `db/`: schema (`rss.sql`) and example SQLite DB.
- `test/`: shared test bootstrap utilities.

Keep new code in the closest existing package; avoid creating new top-level directories unless needed.

## Build, Test, and Development Commands
- `go build -o rss2email main.go`: build local binary.
- `./rss2email`: start scheduler and run configured jobs.
- `./rss2email register you@example.com ruanyifeng sspai`: add user subscriptions.
- `./rss2email add-feed <subscription_id> <feed_url> <name> [content_field] [schedule_type] [cron_spec]`: add a feed source.
- `go test ./...`: run all tests.
- `make docker-image-build`: build Docker image tagged from `VERSION`.
- `make docker-run`: run container mounting local `./db`.

## Coding Style & Naming Conventions
Use standard Go style and keep code `gofmt`-clean before committing.

- Package names: short, lowercase (`service`, `helpers`).
- Files: snake_case by feature (`add_feed.go`, `usersubscription.go`).
- Exported identifiers: `PascalCase`; internal helpers: `camelCase`.
- Keep command logic in `cmd/`, business behavior in `service/`, persistence in `models/`.

## Testing Guidelines
Tests use Go’s `testing` package (`*_test.go`, `TestXxx` functions).

- Run `go test ./...` locally.
- Some tests require reachable external feeds and initialized DB tables (notably `feed_sources`); seed schema from `db/rss.sql` when needed.
- Add table-driven tests for parser/service changes and keep assertions deterministic.

## Commit & Pull Request Guidelines
Recent history follows Conventional Commit prefixes: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`.

- Keep commit subjects imperative and scoped (example: `feat: add startup retry backoff`).
- One logical change per commit.
- PRs should include: purpose, key changes, test command(s) run, and any config/DB migration notes.
- For behavior changes, include sample CLI usage or logs to show expected output.
