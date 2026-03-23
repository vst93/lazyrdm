# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview
- `lazyrdm` is a Go-based TUI for managing Redis connections that mirrors tiny-rdm configurations and uses the `gocui` library for rendering (README.md).
- Target Go version is 1.24.x (`go.mod` pins go1.24.3 via `toolchain`).
- The binary is typically distributed via Homebrew (`brew install vst93/tap/lazyrdm`) or the curl-based script in `cmd/install.sh` (README.md, install section).

## Core commands
Run all commands from the repository root.

| Task | Command | Notes |
| --- | --- | --- |
| Build module | `go build ./...` | Basic compilation check. |
| Build runnable binary | `go build -o lazyrdm main.go` | Produces the CLI for local testing. |
| Run the app | `go run main.go` | Launches the TUI directly. |
| Full release build | `./build.sh` | Prompts for a version, builds CGO-disabled binaries for darwin/linux/windows + Android/Termux, zips outputs under `build/`. |
| Format changed files | `gofmt -w <paths>` | Ensure Go formatting before committing. |
| Vet | `go vet ./...` | Lightweight static analysis. |
| All tests | `go test ./...` | Currently reports `[no test files]`, but run after adding tests. |
| Single package test | `go test ./service -run '^TestName$' -v` | Use once tests exist. |

## High-level architecture

### Entry point and runtime loop
- `main.go` creates a `gocui.Gui`, binds top-level shortcuts (Ctrl+Q to quit, Tab to cycle panes, Ctrl+W to exit to the connection list, `?` for help), and hands off all rendering to the `service` package.
- The Windows-specific `checkAndRelaunchInWT` helper relaunches inside Windows Terminal when the binary is double-clicked.
- `service.NewMainApp` initializes `GlobalApp`, registers the layout manager, populates the connection list view, and starts a goroutine-driven resize watcher (service/app.go).

### Service package layout
- UI is organized around component structs (e.g., `LTRConnectionComponent`, `LTRListDBComponent`, `LTRListKeyComponent`, `LTRKeyInfoComponent`, `LTRKeyInfoDetailComponent`, `LTRTipComponent`) declared under `service/`.
- Each component typically exposes `Init...`, `Layout`, and `KeyBind` methods and registers itself into `GlobalApp.ViewNameList`. Components share global singletons (`GlobalConnectionComponent`, `GlobalDBComponent`, etc.) to simplify cross-pane communication.
- Redis connectivity delegates to `tinyrdm/backend/services`: the connection component reads the tiny-rdm configuration, opens browser sessions via `services.Browser()`, and hands DB/key metadata to downstream panes (service/list_connection.go, service/list_db.go, service/list_key.go, service/key_info*.go).
- Layout helpers (`SetViewSafe`, `GuiSetKeysbinding`, tip rendering) centralize gocui view sizing, consistent keybinding registration, and tooltip messaging.

### Interaction flow
1. **Connection list** (`InitConnectionComponent`): loads groups/entries from the shared tiny-rdm config, supports Vim-like navigation, editing, creation, deletion, exporting, and connects via `services.Browser().OpenConnection`.
2. **Database list** (`InitDBComponent`): after a connection succeeds, shows logical Redis DBs, lets the user select one, and triggers a key list refresh.
3. **Key list & detail panes**: fetch keys via the browser service, allow filtering/pagination, display TTL/value metadata, and render value previews with optional format toggles.
4. **Auxiliary panes**: confirmation dialogs (`page_confirm`), help page (`page_help`), server info page, tip/status bar, and file selector / embedded editor utilities for editing connection definitions.
5. **Resize + focus handling**: the layout manager recalculates component bounds whenever the terminal resizes, and tabbing cycles focus following `GlobalApp.ViewNameList` ordering.

### Patterns to preserve
- Maintain the `Init → Layout → KeyBind` lifecycle when adding panes so keybindings are registered after view creation.
- Update `GlobalTipComponent` whenever a component needs to show hints or transient feedback; use `LayoutTemporary` for toasts instead of printing.
- Use the existing helpers around `services.*` to access configuration, open DBs, export/import configs, and trigger Redis commands; avoid bypassing them with direct client calls.
- Keep goroutine UI updates wrapped in `GlobalApp.Gui.Update(...)` to avoid race conditions.

## Distribution & automation
- `build.sh` is the canonical local release script; it wipes `build/`, cross-compiles for the supported OS/ARCH matrix, zips each artifact, and deletes the raw binary after packaging.
- GitHub Actions release workflow (`.github/workflows/release.yml`) mirrors the matrix build, enabling Android builds by pointing `CC` at NDK toolchains when necessary, and uploads zipped binaries + SHA256 sums on tag creation.

## Install references
- Homebrew tap: `brew install vst93/tap/lazyrdm` / `brew uninstall lazyrdm`.
- Shell installer: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)"`.

## Additional notes
- There are no `_test.go` files checked in yet; expect `go test` to log `[no test files]` until tests are added.
- No lint configuration (e.g., golangci-lint) or Cursor/Copilot rule files are present; follow the conventions documented here and in `AGENTS.md`.
- `build/` contains generated artifacts and is safe to clean when needed.
