# AGENTS.md

Go desktop app (Fyne GUI) + local HTTP widget server (gin) that scrapes live chat from multiple sources (YouTube, Twitch, Kick, VK Play Live, Donation Alerts) and serves an OBS overlay.

## Commands

Use the Makefile targets; they encode non-obvious flags.

- `make test` — `go test ./...`
- `make lint` — `golangci-lint run --fix` (writes fixes; run before committing)
- `make cover` — coverage via `go test -short -coverpkg=./...`
- `make build-app` — `go build -o ./possum_chat ./cmd/main.go` then tars `possum_chat` + `static/`
- `make install-deps` — installs `minimock` into `./bin/` (once per checkout)
- `make mock-generate` — regenerates mocks from interfaces (see below)

Single test: `go test ./internal/use_case/list_messages/... -run TestUseCase_ListMessages`

Entrypoint is `cmd/main.go`. Run via `go run ./cmd/main.go` or the built binary.

## Architecture

Manual DI composition root in `internal/container/` (`Container` with lazy singleton getters — `getXxx()` methods). Wire new services/use-cases there.

- `internal/entity` — shared structs & enums (message, config, scraper state)
- `internal/service` — business services (settings, message_queue, logger, language_provider, scrapers)
- `internal/use_case` — use cases
- `internal/api` — gin HTTP handlers (self API + widget)
- `internal/ui` — Fyne desktop UI
- `internal/infra/clients/<source>` — external API/WS clients (one per source)
- `internal/service/scrapers/<source>` — scrapers wrapping the clients

Settings live in `internal/service/settings`; config is loaded/validated/migrated there.

Main goal of this service is to collect messages from different streaming services and push it in united queue `internal/service/message_queue` that is used for provide all messages in API and UI

## Gotchas

- **GUI requires a display.** Fyne apps cannot run headless. Unit tests are pure and run fine in CI/headless.
- **`config.json` is gitignored** and auto-generated on first run with defaults; edits in the repo copy are not tracked. The JSON key for logging is misspelled `"loging"` (kept intentionally).
- **Config migration**: `internal/service/settings/settings.go` `upgradeConfig()` bumps `version` (currently `1.3`) when adding fields. New default fields must be added there, not just in `entity`.
- **Strict lint**: `.golangci.yml` (v2) enables a large linter set, including `mnd` (magic numbers). Magic numbers need a `//nolint` comment (see `datastruct.go`). `testpackage` requires external test packages.
- **Comments and test descriptions are in Russian** — follow that convention.

## Tests

- Use `minimock/v3` + `testify/assert`. Mocks live in `*/mocks/` and are generated from interfaces by `make mock-generate`. When you change an interface, regenerate the mock.
- Test files must use external package (`package foo_test`), tests should call `t.Parallel()`.
