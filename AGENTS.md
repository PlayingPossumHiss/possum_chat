# AGENTS.md

Desktop app for Linux (Fyne GUI) + a local HTTP widget server (gin). It scrapes live chat from several streaming/donation sources and serves the merged stream as an OBS browser-source overlay. The desktop window is the operator console; the web page is what viewers/streamer see.

Module path: `github.com/PlayingPossumHiss/possum_chat` (Go 1.27).

## What the app does (runtime behavior)

1. On start it loads `./config.json` (creates it with defaults if absent), initializes the logger, starts a `gocron` scheduler, runs the gin HTTP server in a goroutine, then opens the Fyne desktop window (`ShowAndRun`, blocking).
2. The operator (streamer) opens the GUI, enters a key per source (channel name / token), and toggles each source on/off. Toggling a source calls `Scraper.Run(ctx)` / `Scraper.Stop()`.
3. Each scraper connects to its source, buffers incoming chat messages, and (where supported) tracks the current viewer count.
4. A background job (`ask_watchers_for_messages`) drains every scraper's buffer every 30 ms and pushes the messages into a single in-memory `message_queue`.
5. The widget page polls `/api/v1/messages` and renders the merged chat plus per-source online counts.
6. The streamer can run a chat poll (`--vote variant1;variant2`); viewers vote by number and the results render on a separate `/widget.html` page (see [Voting](#voting)).
7. The streamer can show arbitrary text on the same widget page with `--message` (see [Text](#text)).

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
- `internal/service` — business services (settings, message_queue, logger, language_provider, app_updater, scrapers, hooks, hooks/vote, hooks/text, election, screen_title)
- `internal/use_case` — use cases (list_messages, get_online, get_style, run_watch_scrapers, send_test_messages)
- `internal/api` — gin HTTP handlers (self API + widget)
- `internal/ui` — Fyne desktop UI
- `internal/infra/clients/<source>` — external API/WS clients (one per source: youtube, twitch, kick, vk_play_live, donation_alerts, github)
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
   ┌──────────────────┬───────────────────┬──────────────────┐
   ▼                  ▼                   ▼                  ▼
 list_messages      get_online         send_test_messages  hooks.Service
 UseCase            UseCase            UseCase (GUI test)      │
       │               │                              ┌───────┴────────┐
       ▼               ▼                              ▼                ▼
      GET /api/v1/messages (gin)  ──►  static/messages.html   vote hook    text hook
                                        + messages.js (widget)      │            │
                                                                    ▼            ▼
                                                        election.Service  screen_title.Service
                                                                    │            │
                                                                    └─────┬──────┘
                                                                          ▼
                                                           GET /api/v1/widget_content

(commands pages) POST /api/v1/command/{message,start_election,stop_election}
      └─►  screen_title.Service / election.Service (напрямую, минуя hooks)
```

- Scrapers keep an internal `messages []entity.Message` buffer. `GetMessages()` returns a clone and resets the buffer to nil (drain semantics).
- `message_queue.PushMessages()` **overwrites `CreatedAt` with `clock.Now()`** so messages never appear mid-list when several scrapers produce at once.
- `message_queue.ListMessages(forLast)` returns messages younger than `View.TimeToHideMessage` by default, or younger than `forLast` when the query param is passed. It does NOT sort.
- `list_messages.UseCase.ListMessages()` sorts ascending by `CreatedAt` before returning (this is what the API serves).
- `message_queue.CleanOldMessages()` deletes messages older than `View.TimeToDeleteMessage` (no-op when it is `0`).
- Online count: `get_online.UseCase.GetOnline()` returns `map[entity.Source]int64`, but only when `View.ShowUserCount` is true and the scraper `Status() == ScraperStateActive`.
- `message_queue.PushMessages()` also passes every pushed message to the registered `entity.Hook` (the `hooks.Service`), asynchronously — one goroutine per message, the call wrapped in a 1 s timeout context; errors are logged. Hooks are wired in `container.getMessageQueueService()`.
- `internal/service/hooks.Service` (`hooks/hooks.go`) is the aggregate hook: it holds a list of `entity.Hook`s and its `Handle(ctx, message)` runs **all** of them, joining errors with `errors.Join`. It is built once in `container.getHookService()` from `[]entity.Hook{voter, texter}` and is wired into `message_queue`.
- The vote hook (`internal/service/hooks/vote.Service`) implements `entity.Hook` and intercepts `--vote` commands, delegating vote state to `internal/service/election.Service` (see [Voting](#voting)).
- The text hook (`internal/service/hooks/text.Service`) implements `entity.Hook` and intercepts `--message` commands, delegating the text to `internal/service/screen_title.Service` (see [Text](#text)).

### Message model

`entity.Message` has `ID`, `Source`, `User`, `Content []MessageContentItem`, `CreatedAt`. `MessageContentItem` is `{Type (text|image), Value}` — each source normalizes its native message (emotes/smiles become `image` items with absolute URLs) into this unified shape. `entity.Source` order: `Youtube=1, Twitch, Kick, VkPlayLive, DonationAlerts, Internal`. `SourceInternal` is not a real chat source — it is no longer produced by any code path (the `do_command` use case that built it was removed); `common.IsMessageFromAdmin` still returns `true` for it.

## HTTP API

All served by gin in `internal/api/api.go` (release mode). Routes:

| Route | Handler | Notes |
|---|---|---|
| `GET /messages.html` | static | OBS widget page |
| `GET /messages_app.html` | static | mobile viewing page (scroll + error/warn badges) |
| `GET /commands_message_app.html` | static | mobile page to set the `--message` overlay text |
| `GET /commands_vote_app.html` | static | mobile page to start/stop a poll |
| `GET /js/messages.js` | static | messages JS (shared by `messages.html` and `messages_app.html`) |
| `GET /js/common.js` | static | shared XHR helper (`callApi`) for the commands pages |
| `GET /js/commands_message_app.js` | static | message-overlay page JS |
| `GET /js/commands_vote_app.js` | static | poll page JS |
| `GET /css/mobile_app.css` | static | mobile pages style (menu, disables the rotate trick) |
| `GET /widget.html` | static | poll widget page |
| `GET /js/widget.js` | static | poll widget JS |
| `GET /css/widget.css` | static | poll widget style |
| `GET /img/*` | static | source icons |
| `GET /css/messages.css` | `cssMainStyleCss` | main style, read from `./static/css/simple_block.css` or `simple_no_bg.css` |
| `GET /css/custom_style.css` | `cssCustomStyleCss` | custom CSS from `View.CssStyle` |
| `GET /api/v1/messages` | `apiV1Messages` | messages + online counts |
| `GET /api/v1/logging_status` | `apiV1LoggingStatus` | `{error_count, warn_count}` |
| `GET /api/v1/widget_content` | `apiV1WidgetContent` | current poll results + streamer text (see [Voting](#voting) and [Text](#text)) |
| `POST /api/v1/command/message` | `apiV1CommandMessage` | set the `--message` overlay text from `{"text": "..."}` (see [Commands](#commands)) |
| `POST /api/v1/command/start_election` | `apiV1CommandStartElection` | start a poll from `{"candidates": ["a","b"]}` |
| `POST /api/v1/command/stop_election` | `apiV1CommandStopElection` | clear/stop the poll |

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

`GET /api/v1/widget_content` response — `vote` is `[]` while no poll is active, `text` is `""` while no `--message` text is set:

```json
{
  "vote": [
    {"text": "variant1", "count": 3},
    {"text": "variant2", "count": 1}
  ],
  "text": "some text"
}
```

### Widget

- `http://127.0.0.1:8081/messages.html` — OBS overlay (default). Newest message at the bottom; when space runs out only recent ones remain. Online count block at the bottom. No error/warn badges.
- `http://127.0.0.1:8081/messages.html?for_last=1h` — show all messages from the last hour.
- `http://127.0.0.1:8081/messages_app.html?for_last=1h` — mobile viewing page. Scrollable list (the `rotate(180deg)` trick is disabled via `/css/mobile_app.css`), shows error/warn badges from `/api/v1/logging_status`, and has a nav menu (`Сообщения` / `Команды`).
- `http://127.0.0.1:8081/commands_message_app.html` — mobile page to set the `--message` overlay text (a text field + button that `POST`s `{"text": ...}` to `/api/v1/command/message`).
- `http://127.0.0.1:8081/commands_vote_app.html` — mobile page to start/stop a poll (add variants, `POST`s to `/api/v1/command/start_election` / `stop_election`).
- `http://127.0.0.1:8081/widget.html` — poll widget: renders the current `--vote` results (hidden while no poll is active) and the current `--message` text (bottom-right). Polls `/api/v1/widget_content` every 250 ms.

`static/js/messages.js` is shared by `messages.html` and `messages_app.html`. It polls `/api/v1/messages` every **50 ms** (reverses the array in JS, updates a Vue 2 app) and `/api/v1/logging_status` every **1 s**. The "newest at bottom" behavior comes from the CSS `transform: rotate(180deg)` trick plus the JS `.reverse()`.

Both pages load Vue 2 from a CDN (`cdn.jsdelivr.net/npm/vue@2.7.16`), so they need internet access (unless cached).

## Voting (`--vote`)

Chat-driven polls. The streamer starts/ends them from their own chat, viewers vote by number, and the results render on `/widget.html`.

- **Start**: the streamer posts `--vote variant1;variant2` — variants are split on `;`.
- **Vote**: a viewer posts a message that is a single number (`1`, `2`, …); an optional leading `#` or `№` and surrounding spaces are trimmed.
- **End**: the streamer posts `--vote` (no arguments) — clears the poll.
- Only the streamer's own messages can start/end a poll. `vote.Service.isValidElectionRequest` matches `message.User` case-insensitively (`EqualFold`) against the configured `channel_name` per source; supported for Kick, Twitch, YouTube and VK Play Live (Donation Alerts returns `false`).
- Each user votes once per poll: `election.Service` dedupes by `(User, Source)` in the `voted` map.
- State lives in `internal/service/election.Service` (`candidates []entity.VoteResult`, `voted` map), mutex-protected (`sync.RWMutex`); `ElectionResult()` returns a copy under the lock, `StartElection`/`StopElection`/`Vote` mutate under the lock.
- The vote hook (`hooks/vote.Service`) is wired as an `entity.Hook` in `container.getHookService()`; the `election.Service` is the shared storage, injected into `api.New` as the `Voter` interface for `/api/v1/widget_content`.

## Text (`--message`)

Chat-driven text overlay. The streamer shows a line of text on `/widget.html` from their own chat; it renders in the bottom-right corner.

- **Set**: the streamer posts `--message some text` — everything after `--message ` is displayed.
- **Clear**: the streamer posts `--message` (no arguments) — the text is removed.
- Only the streamer's own messages can set/clear it: `text.Service.isValidTextRequest` matches `message.User` case-insensitively (`EqualFold`) against the configured `channel_name` per source; supported for Kick, Twitch, YouTube and VK Play Live (Donation Alerts returns `false`).
- State lives in `internal/service/screen_title.Service` (`message string`), mutex-protected (`sync.RWMutex`); `hooks/text.Service` delegates via `SetMessage` and `GetMessage` returns it.
- The text hook (`hooks/text.Service`) is wired as an `entity.Hook` in `container.getHookService()`; the `screen_title.Service` is the shared storage, injected into `api.New` as the `ScreenTitle` interface for `/api/v1/widget_content`.

## Commands

The mobile pages let the streamer control the overlay from a phone instead of typing `--` in chat. Two pages and three endpoints replace the old single `/api/v1/command` + `do_command` pipeline.

- `commands_message_app.html` → `POST /api/v1/command/message` with `{"text": "..."}` — sets the `--message` overlay text (`""` clears it). Handler `apiV1CommandMessage` binds the JSON and calls `screen_title.Service.SetMessage` directly.
- `commands_vote_app.html` → `POST /api/v1/command/start_election` with `{"candidates": ["a","b"]}` and `POST /api/v1/command/stop_election` — start/clear the poll. Handlers call `election.Service.StartElection` / `StopElection` directly.
- These endpoints write to `screen_title.Service` / `election.Service` **directly**, bypassing the chat hooks (`hooks.Service`); there is no per-user admin check on this path — the server is localhost-only.
- Errors return `400` with `{"message": "..."}` (`responseError`); success returns `200` with `null`.
- Wired in `container.getSelfApi()`: `api.New(..., c.getScreenMessage(), c.getElectionStorage(), ...)` — the same singletons the chat hooks use, so chat-driven and page-driven updates share state.

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
- `internal/service/settings/datastruct.go` holds the wire `config` struct (JSON tags), `currentVersion`, `configPath`, `appVerssion` (hardcoded app version, see [App update](#app-update-self-update)), and `defaultConfig`.
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

1. **Connections** (`main_window_connections.go`) — per-source toggle button + key entry (Donation Alerts key is a password field via `Source.KeyIsSecret()`), plus links to the OBS widget (`messages.html`), the mobile message panel (`messages_app.html?for_last=1h`), the poll widget, and the repo.
2. **CSS** (`main_window_css.go`) — custom CSS editor (`View.CssStyle`), main-style picker (`simple_block` / `simple_no_bg`), and a "send test message" button (injects one fake message per source via `send_test_messages.UseCase`).
3. **Settings** (`main_window_settings.go`) — time-to-hide (seconds), time-to-delete (minutes), port, language, show-online checkbox, and app version + update button.

The version display and the check-for-updates / update flow are described in [App update](#app-update-self-update).

## App update (self-update)

- Current version is the hardcoded const `appVerssion` in `internal/service/settings/datastruct.go`. For release builds (the only path that uses the GitHub self-update) it is a semantic version like `1.3.1`; a commit hash is used only as a placeholder for intermediate/unreleased builds. `configFromJson` in `conversion.go` copies it into `entity.Config.AppVersion`, so it is **not** persisted in `config.json` and not written back by `configToJson`.
- `internal/infra/clients/github.Client.LatestVersion(ctx)` does `GET https://api.github.com/repos/PlayingPossumHiss/possum_chat/releases` (1 s timeout) and returns the newest release as `entity.SourceVersion{Version: tag_name, DownloadURL: browser_download_url}`, where `DownloadURL` is the `possum_chat.tar.gz` asset. Returns `(nil, nil)` when there are no releases or no `possum_chat.tar.gz` asset.
- `internal/service/app_updater.Service` wraps a `GithubClient` interface (`interfaces.go`):
  - `LatestVersion(ctx)` → delegates to the client.
  - `Update(ctx, version)` (mutex-protected) downloads the archive to `./possum_chat.tar.gz`, untars it into `./tmp`, then `swapRelease` deletes same-named files in the working directory and copies the new `possum_chat` binary + `static/` over them (`os.CopyFS`), then removes `./tmp`.
- The Settings tab (`main_window_settings.go`) reads `configStorage.Config().AppVersion`, calls `appUpdater.LatestVersion(ctx)` on open, and when the GitHub version differs (plain string `!=`, not semver) shows a "Latest version" label plus an "Update" button. Clicking Update runs `appUpdater.Update`, hides the label/button and pops a modal "Update is done. Restart the app".
- Wired in `internal/container/container.go`: `getAppUpdater()` → `app_updater.New(github.New())`, passed into `ui.New(...)` (interface `AppUpdater` in `internal/ui/interface.go`).
- New language constants `APP_LATEST_VERSION`, `UPDATE_APP_BUTTON`, `UPDATE_DONE` (`entity/language.go`, `language_provider/translations.go`).

## Gotchas

- **GUI requires a display.** Fyne apps cannot run headless. Unit tests are pure and run fine in CI/headless.
- **Relative paths.** The app reads `./config.json`, `./static/css/*.css`, `./static/img/favicon.ico`, and writes `./log.log` relative to the working directory — run it from the repo root / the directory containing `static/` (the `build-app` tarball preserves this layout).
- **`config.json` is gitignored** and auto-generated on first run with defaults; edits in the repo copy are not tracked. The JSON key for logging is misspelled `"loging"` (kept intentionally).
- **Config migration**: new fields go through `upgradeConfig()` (see above), not just `entity`.
- **Port & language changes require an app restart**; after a CSS/style change, refresh the browser tab.
- **Widget needs internet** for Vue 2 (loaded from jsdelivr CDN).
- **Self-update needs internet** (GitHub API + archive download) and overwrites the running `possum_chat` binary and `static/` in place; the operator must restart the app afterwards.
- **Strict lint**: `.golangci.yml` (v2) enables a large linter set, including `mnd` (magic numbers). Magic numbers need a `//nolint` comment (see `internal/service/settings/datastruct.go`). `testpackage` requires external test packages; `paralleltest`/`tparallel` require `t.Parallel()`.
- **Comments and test descriptions are in Russian** — follow that convention. The product README (`README.md`) is also in Russian.

## Tests

- Use `minimock/v3` + `testify/assert`. Mocks live in `*/mocks/` and are generated from interfaces by `make mock-generate`. When you change an interface, regenerate the mock.
- Do not hand-write stub structs implementing interfaces in tests — add the interface to the `mock-generate` target in the `Makefile` and run `make mock-generate` instead. Example: `entity.Hook` → `internal/entity/mocks/hook_mock.go`, used as `m_entity.NewHookMock(mc)`. For a hook that is called asynchronously and zero-or-more times, set it up with `.Optional().Set(...)`.
- Test files must use external package (`package foo_test`), tests should call `t.Parallel()`.
- Inject `utils_time.Clock` instead of `time.Now()` for deterministic time assertions (see `list_messages_test.go`).
- The `logger` is a package-level singleton — tests that trigger logging must call `logger.Init(configStorage)` first (see `run_watch_scrapers_test.go`).
