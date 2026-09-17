# AGENTS.md

Desktop app for Linux (Fyne GUI) + a local HTTP widget server (gin). It scrapes live chat from several streaming/donation sources and serves the merged stream as an OBS browser-source overlay. The desktop window is the operator console; the web page is what viewers/streamer see.

Module path: `github.com/PlayingPossumHiss/possum_chat` (Go 1.27).

## What the app does (runtime behavior)

1. On start it loads `./config.json` (creates it with defaults if absent), initializes the logger, starts a `gocron` scheduler, runs the gin HTTP server in a goroutine, then opens the Fyne desktop window (`ShowAndRun`, blocking).
2. The operator (streamer) opens the GUI, enters a key per source (channel name / token), and toggles each source on/off. Toggling a source calls `Scraper.Run(ctx)` / `Scraper.Stop()`.
3. Each scraper connects to its source, buffers incoming chat messages, and (where supported) tracks the current viewer count.
4. A background job (`ask_watchers_for_messages`) drains every scraper's buffer every 30 ms and pushes the messages into a single in-memory `message_queue`.
5. The widget page polls `/api/v1/messages` and renders the merged chat plus per-source online counts.

## Commands

Use the Makefile targets; they encode non-obvious flags.

- `make test` — `go test ./...`
- `make lint` — downloads golangci-lint `v2.13.2` into `./bin/` then runs `./bin/golangci-lint run --fix` (writes fixes; run before committing)
- `make cover` — `go test -timeout 240s -short -count=1 -coverpkg=./... -coverprofile=coverage.out ./...`, opens the HTML coverage report, then removes `coverage.out`
- `make build-app` — `go build -o ./possum_chat ./cmd/main.go` then tars `possum_chat` + `static/`
- `make install-deps` — installs `minimock` into `./bin/` (once per checkout)
- `make mock-generate` — regenerates mocks from interfaces (see below)

Single test: `go test ./internal/use_case/list_messages/... -run TestUseCase_ListMessages`

Entrypoint is `cmd/main.go`. Run via `go run ./cmd/main.go` or the built binary.

## Architecture

Manual DI composition root in `internal/container/` (`Container` with lazy singleton getters — `getXxx()` methods). Wire new services/use-cases there.

- `internal/entity` — shared structs & enums (message, config, scraper state, language constants)
- `internal/service` — business services (settings, message_queue, logger, language_provider, scrapers)
- `internal/use_case` — use cases
- `internal/api` — gin HTTP handlers (self API + widget)
- `internal/ui` — Fyne desktop UI
- `internal/infra/clients/<source>` — external API/WS clients (one per source)
- `internal/service/scrapers/<source>` — scrapers wrapping the clients
- `internal/utils/time` — `Clock` interface (`Now() time.Time`) injected for deterministic tests

Settings live in `internal/service/settings`; config is loaded/validated/migrated there. All settings can be changed from the UI (except logging level/path, which are edited in `config.json`).

## Data flow & message lifecycle

```
source (YouTube/Twitch/Kick/VK Play Live/Donation Alerts)
   └─ infra client  ──►  scraper (buffers messages + online count)
                              │  GetMessages() drains buffer every 30 ms
                              ▼
                     run_watch_scrapers.UseCase
                              ▼
                     message_queue.Service  (single in-memory queue)
                              │
              ┌───────────────┼────────────────┐
              ▼               ▼                ▼
      list_messages       get_online       send_test_messages
      UseCase             UseCase          UseCase (GUI test button)
              │               │
              ▼               ▼
      GET /api/v1/messages (gin)  ──►  static/messages.html + messages.js (widget)
```

- Scrapers keep an internal `messages []entity.Message` buffer. `GetMessages()` returns a clone and resets the buffer to nil (drain semantics).
- `message_queue.PushMessages()` **overwrites `CreatedAt` with `clock.Now()`** so messages never appear mid-list when several scrapers produce at once.
- `message_queue.ListMessages(forLast)` returns messages younger than `View.TimeToHideMessage` by default, or younger than `forLast` when the query param is passed. It does NOT sort.
- `list_messages.UseCase.ListMessages()` sorts ascending by `CreatedAt` before returning (this is what the API serves).
- `message_queue.CleanOldMessages()` deletes messages older than `View.TimeToDeleteMessage` (no-op when it is `0`).
- Online count: `get_online.UseCase.GetOnline()` returns `map[entity.Source]int64`, but only when `View.ShowUserCount` is true and the scraper `Status() == ScraperStateActive`.

### Message model

`entity.Message` has `ID`, `Source`, `User`, `Content []MessageContentItem`, `CreatedAt`. `MessageContentItem` is `{Type (text|image), Value}` — each source normalizes its native message (emotes/smiles become `image` items with absolute URLs) into this unified shape. `entity.Source` order: `Youtube=1, Twitch, Kick, VkPlayLive, DonationAlerts`.

## HTTP API

All served by gin in `internal/api/api.go` (release mode). Routes:

| Route | Handler | Notes |
|---|---|---|
| `GET /messages.html` | static | widget page |
| `GET /js/messages.js` | static | widget JS |
| `GET /img/*` | static | source icons |
| `GET /css/messages.css` | `cssMainStyleCss` | main style, read from `./static/css/simple_block.css` or `simple_no_bg.css` |
| `GET /css/custom_style.css` | `cssCustomStyleCss` | custom CSS from `View.CssStyle` |
| `GET /api/v1/messages` | `apiV1Messages` | messages + online counts |
| `GET /api/v1/logging_status` | `apiV1LoggingStatus` | `{error_count, warn_count}` |

`GET /api/v1/messages` query param: `for_last` — a Go duration string (`time.ParseDuration`); overrides the hide window so the panel can show older messages.

Response shape:

```json
{
  "messages": [
    {
      "id": "youtube_...",
      "source": "youtube",                 // youtube|twitch|kick|vk_play_live|donation_alerts
      "user": "some_user",
      "message_content": [
        {"type": "text",  "value": "hello "},
        {"type": "image", "value": "https://.../emoji.png"}
      ],
      "created_at": "2026-01-01T12:00:00Z" // RFC3339
    }
  ],
  "online": [
    {"count": 123, "source": "twitch"}
  ]
}
```

### Widget

- `http://127.0.0.1:8081/messages.html` — OBS overlay (default). Newest message at the bottom; when space runs out only recent ones remain. Online count block at the bottom.
- `http://127.0.0.1:8081/messages.html?for_last=1h` — show all messages from the last hour.
- `http://127.0.0.1:8081/messages.html?for_last=1h&use_scroll=true` — scrollable list for the streamer's second monitor; also shows error/warn badges from `/api/v1/logging_status`.

`static/js/messages.js` polls `/api/v1/messages` every **50 ms**, reverses the array in JS, and updates a Vue 2 app. `use_scroll=true` additionally polls `/api/v1/logging_status` every 1 s. The "newest at bottom" behavior comes from the CSS `transform: rotate(180deg)` trick plus the JS `.reverse()`.

The page loads Vue 2 from a CDN (`cdn.jsdelivr.net/npm/vue@2.7.16`), so the widget needs internet access (unless cached).

## Sources

Per-source connection key (stored in `config.json` under `connections`) and transport:

| Source | Key | Transport / how chat & online are obtained |
|---|---|---|
| YouTube | `channel_name` — channel name **without @** OR a stream/video ID | `epjane/youtube-live-chat-downloader` (scrapes `ytInitialData`, continuation tokens). Online from the watch page `ytInitialData`. |
| Twitch | `channel_name` — without @ | Chat via `go-twitch-irc/v4` anonymous client (WebSocket). Online by scraping the channel page HTML for `clientId="..."`, then a GraphQL POST (`gqlURL`) with that client id. |
| Kick | `channel_name` | REST `https://kick.com/api/v2/channels/<name>` for room id + online; chat via WebSocket Pusher subscribe `chatrooms.<id>.v2`. |
| VK Play Live | `channel_name` | REST `https://api.live.vkvideo.ru/v1/channel/<name>` (user id) and `/v1/ws/connect` (ws token); chat + online via WebSocket `wss://pubsub.live.vkvideo.ru/...` subscribing to `channel-chat:<id>` and `channel-info:<id>`. |
| Donation Alerts | `token` (secret — masked in UI) | Socket.IO to `socket.donationalerts.ru:443`, emit `add-user`, listen on `donation`. No online count. |

Online counts are polled once per minute by each scraper's `watchOnline` loop, except VK Play Live (pushed over WS) and Donation Alerts (none). Several sources add small `time.Sleep` delays between reconnect attempts "чтобы не словить бан" (to avoid being rate-limited/banned).

## Background jobs (gocron scheduler)

Registered in `internal/container/container.go`:

| Job name | Interval | Function |
|---|---|---|
| `ask_watchers_for_messages` | 30 ms | `run_watch_scrapers.Run` — drain scrapers → push to queue |
| `clean_old_messages` | 30 ms | `message_queue.CleanOldMessages` — drop expired messages |

Job errors are logged via `AfterJobRunsWithError` listener.

## Configuration

- File: `./config.json` (relative to the working directory). JSON key for logging is misspelled `"loging"` (kept intentionally).
- `internal/service/settings/datastruct.go` holds the wire `config` struct (JSON tags), `currentVersion`, `configPath`, and `defaultConfig`.
- `internal/service/settings/conversion.go` maps JSON ⇄ `entity.Config` and validates (returns `app_errors.ErrInvalidConfig` on bad values).
- `settings.Service.UpdateConfig([]entity.ConfigUpdateOption)` applies functional options, persists to disk, then updates the in-memory copy (mutex-protected). The UI edits config through this path.

Schema (current version `1.3`):

```json
{
  "connections": {
    "youtube":        {"channel_name": "..."},
    "twitch":         {"channel_name": "..."},
    "kick":           {"channel_name": "..."},
    "vk_play_live":   {"channel_name": "..."},
    "donation_alerts":{"token": "..."}
  },
  "loging": {"log_path": "./log.log", "level": "INFO"},
  "view": {
    "css_style": "",
    "main_style": "simple_block",     // simple_block | simple_no_bg
    "time_to_hide_message": "3m0s",   // widget visibility window (Go duration)
    "time_to_delete_message": "1h0m0s", // queue retention (Go duration)
    "show_user_count": true
  },
  "ui": {"lang": "en"},              // en | ru
  "port": 8081,
  "version": "1.3"
}
```

Defaults: `port 8081`, `time_to_hide_message 3m`, `time_to_delete_message 1h`, `lang en`, log level `INFO`. `show_user_count` and `main_style` are set during migration.

### Config migration

`internal/service/settings/settings.go` `upgradeConfig()` bumps `version` when adding fields. The chain so far:

- `1.0 → 1.1`: add `ui.lang` (en)
- `1.1 → 1.2`: add `view.main_style` (`simple_block`)
- `1.2 → 1.3`: add `view.show_user_count` (true)

When adding a field: add it to `entity.Config*` **and** to the `config` struct + `defaultConfig` in `datastruct.go`, plus `configFromJson`/`configToJson` in `conversion.go`, then add a migration step and bump `currentVersion`.

## UI (Fyne)

Three tabs in `internal/ui/`:

1. **Connections** (`main_window_connections.go`) — per-source toggle button + key entry (Donation Alerts key is a password field via `Source.KeyIsSecret()`), plus links to the OBS widget, the scrollable message panel, and the repo.
2. **CSS** (`main_window_css.go`) — custom CSS editor (`View.CssStyle`), main-style picker (`simple_block` / `simple_no_bg`), and a "send test message" button (injects one fake message per source via `send_test_messages.UseCase`).
3. **Settings** (`main_window_settings.go`) — time-to-hide (seconds), time-to-delete (minutes), port, language, show-online checkbox, version.

App version shown in Settings is a hardcoded const `version = "5b94611"` in `main_window_settings.go`.

## Gotchas

- **GUI requires a display.** Fyne apps cannot run headless. Unit tests are pure and run fine in CI/headless.
- **Relative paths.** The app reads `./config.json`, `./static/css/*.css`, `./static/img/favicon.ico`, and writes `./log.log` relative to the working directory — run it from the repo root / the directory containing `static/` (the `build-app` tarball preserves this layout).
- **`config.json` is gitignored** and auto-generated on first run with defaults; edits in the repo copy are not tracked. The JSON key for logging is misspelled `"loging"` (kept intentionally).
- **Config migration**: new fields go through `upgradeConfig()` (see above), not just `entity`.
- **Port & language changes require an app restart**; after a CSS/style change, refresh the browser tab.
- **Widget needs internet** for Vue 2 (loaded from jsdelivr CDN).
- **Strict lint**: `.golangci.yml` (v2) enables a large linter set, including `mnd` (magic numbers). Magic numbers need a `//nolint` comment (see `internal/service/settings/datastruct.go`). `testpackage` requires external test packages; `paralleltest`/`tparallel` require `t.Parallel()`.
- **Comments and test descriptions are in Russian** — follow that convention. The product README (`README.md`) is also in Russian.

## Tests

- Use `minimock/v3` + `testify/assert`. Mocks live in `*/mocks/` and are generated from interfaces by `make mock-generate`. When you change an interface, regenerate the mock.
- Test files must use external package (`package foo_test`), tests should call `t.Parallel()`.
- Inject `utils_time.Clock` instead of `time.Now()` for deterministic time assertions (see `list_messages_test.go`).
- The `logger` is a package-level singleton — tests that trigger logging must call `logger.Init(configStorage)` first (see `run_watch_scrapers_test.go`).
