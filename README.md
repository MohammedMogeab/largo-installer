LarGo Installer (TUI)

Overview
- A cross-platform installer for the LarGo CLI featuring .
- Ensures bin dir, updates PATH safely, installs LarGo, and verifies.

Install/Run (from source)
- Prereqs: Go 1.22+
- Fetch deps: `go mod tidy`
- Run: `go run ./cmd/largo-installer`
- Flags:
  - `--largo-version <ver>`: Target LarGo version (default `latest`).
  - `--module <path>`: Go module path (default `github.com/MohammedMogeab/largo/cmd/largo`).
  - `--no-color`: Disable colors.
  - `--version`: Print installer version.

Build
- Dev build: `go build -o bin/largo-installer ./cmd/largo-installer`
- Release build: `goreleaser release --clean --snapshot` (requires GoReleaser)
- Embed version:
  `go build -ldflags "-s -w -X github.com/MohammedMogeab/largo-installer/internal/buildinfo.Version=$(git describe --tags --always)" ./cmd/largo-installer`

What it does
1) Checks Go is installed and prints version.
2) Determines install dir (`GOBIN`, `GOPATH/bin`, or `$HOME/go/bin`).
3) Ensures PATH contains that dir:
   - Windows: PowerShell (no truncation), fallback to registry.
   - Unix: adds an export line to a shell RC file.
4) Installs LarGo via `go install <module>@<version>`.
5) Verifies `largo version` via PATH or direct path.

Notes
- Open a new terminal after PATH changes to refresh the environment.
- The UI stays open after completion; press Enter or `q` to exit.

Project layout
- `cmd/largo-installer`: CLI entry.
- `internal/ui`: Bubble Tea UI.
- `internal/install`: install logic with per-OS PATH handling.
- `internal/buildinfo`: version variable set at build time.

Roadmap
- Add binary download path from GitHub Releases (no Go toolchain required).
- Add `--bin`, `--no-path`, `--use-go`, and `--verbose` flags.
- Add Windows environment broadcast for PATH change.

License
- MIT
contribute
--  we are happy with contributing with us ,
