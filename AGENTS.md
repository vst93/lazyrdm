# AGENTS.md

Guidance for coding agents working in `lazyrdm`.

## Project Snapshot

- Language: Go (`go.mod` uses Go `1.24.0`, CI uses `1.24.3`).
- App type: terminal UI Redis manager built with `gocui`.
- Entry point: `main.go`.
- Main implementation package: `service/`.
- Current state: no automated tests and no lint config are checked in.

## Repository Map

- `main.go`: app bootstrap, GUI setup, global keybindings.
- `service/`: UI components and business logic.
- `cmd/install.sh`: install script used by curl-based install flow.
- `build.sh`: local multi-platform build + zip packaging script.
- `.github/workflows/release.yml`: release CI build matrix and packaging.
- `build/`: build artifacts.

## Canonical Commands

Use these commands from repo root: `/Users/vst/Code/goProgram/lazytr`.

### Build

- Build all packages:
  - `go build ./...`
- Build runnable binary (local):
  - `go build -o lazyrdm main.go`
- Run release-oriented local build script:
  - `./build.sh`
  - Script prompts for a version and builds zips for multiple OS/ARCH targets.

### Run

- Run app directly:
  - `go run main.go`

### Test

- Run all tests:
  - `go test ./...`
- Run tests for one package:
  - `go test ./service`
- Run a single test function (when tests exist):
  - `go test ./service -run '^TestName$' -v`
- Run a single test file pattern (when tests exist):
  - `go test ./service -run 'TestKey.*' -v`

Important: the repository currently has no `*_test.go` files, so `go test ./...` reports `[no test files]`.

### Lint / Format / Vet

No dedicated lint tool is configured (`.golangci.yml` absent).

- Format code:
  - `gofmt -w .`
- Vet code:
  - `go vet ./...`

If you introduce new code, run at least:

1. `gofmt -w` on changed files
2. `go test ./...`
3. `go build ./...`

### Install / Release Helpers

- Homebrew install:
  - `brew install vst93/tap/lazyrdm`
- Homebrew uninstall:
  - `brew uninstall lazyrdm`
- Script install:
  - `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/vst93/lazyrdm/refs/heads/master/cmd/install.sh)"`

## Coding Style Rules

Follow existing patterns in this codebase rather than introducing a new architecture style.

### Imports

- Keep imports grouped by standard Go style:
  1. stdlib
  2. external dependencies
  3. local module imports (for example `lazyrdm/service`, `tinyrdm/...`)
- Use `gofmt` ordering and spacing.
- Avoid unnecessary aliases.

### Formatting

- Use `gofmt` formatting (tabs, spacing, import order).
- Keep chained component calls readable (example pattern: `LoadKeys().Layout().KeyBind()`).
- Preserve existing mixed-language comments when editing nearby code.

### Types and APIs

- Use `any` when dynamic typing is intended (repository already uses `any` heavily).
- Prefer explicit structs for component state.
- Keep receiver methods on component types returning `*Type` when chaining is expected.

### Naming Conventions

- Exported identifiers: `PascalCase`.
- Unexported identifiers: `camelCase`.
- Component structs use `LTR` prefix (for example `LTRListKeyComponent`).
- Global shared component pointers use `Global` prefix (for example `GlobalApp`, `GlobalTipComponent`).
- Keep file names aligned with feature/component purpose (`list_key.go`, `list_db.go`, etc.).

### Component Lifecycle Pattern

Most UI components in `service/` follow this shape:

1. `InitXxxComponent()` to allocate and register
2. `Layout()` to render/re-render
3. `KeyBind()` to register key handlers
4. `KeyMapTip()` for help text

When adding new UI panes, match this lifecycle unless there is a strong reason not to.

### Error Handling

- Guard errors immediately with `if err != nil`.
- Return errors up the stack where function signatures allow it.
- Use `fmt.Errorf("...: %w", err)` when adding context to returned errors.
- In UI event handlers, surface user-facing failures through `GlobalTipComponent.LayoutTemporary(...)`.
- Use `defer` for cleanup (`Close`, state restoration, resource release).

### State Management

- This project uses global mutable pointers for major UI components.
- Reuse existing globals rather than introducing parallel state containers.
- Be careful with goroutines updating shared UI state; use existing `g.Update(...)` patterns.

### Keybinding and Interaction Patterns

- Register keybindings through shared helpers like `GuiSetKeysbinding(...)`.
- Keep Vim-like movement key support (`h/j/k/l`) where similar views already use it.
- Keep confirmation flows consistent with `GuiSetKeysbindingConfirm(...)`.

### Dependencies and Boundaries

- Keep Redis/domain operations routed through `tinyrdm/backend/services` APIs.
- Keep terminal/UI rendering responsibilities inside `service/` components.
- Avoid introducing unrelated frameworks or broad refactors in focused tasks.

## Cursor and Copilot Rule Sources

- `.cursor/rules/`: not present in this repository.
- `.cursorrules`: not present in this repository.
- `.github/copilot-instructions.md`: not present in this repository.

If any of the above files are added later, treat them as higher-priority repository instructions and update this file accordingly.

## Agent Execution Checklist

Before finishing a code change:

1. Ensure edited Go files are `gofmt`-formatted.
2. Run `go test ./...`.
3. Run `go build ./...`.
4. Confirm behavior manually when touching keybindings or UI rendering.
5. Keep changes minimal and consistent with existing component/global patterns.

When adding tests in the future, include at least one command example in PR notes for running one specific test via `-run`.
