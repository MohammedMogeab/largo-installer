# Contributing to LarGo Installer

Thanks for your interest in improving the LarGo installer! This guide covers how to set up your environment, propose changes, and get your work merged smoothly.

## Quick Start

- Fork the repo and create a feature branch from `main`.
- Keep PRs focused and small; split unrelated changes into separate PRs.
- Write clear commit messages in imperative mood, e.g. "add path broadcast on Windows".
- Make sure the project builds and tests pass locally before opening a PR.

## Development Environment

- Go 1.22+ recommended (CI uses 1.22.x)
- Install dependencies: `go mod tidy`
- Build: `go build ./...`
- Run installer: `go run ./cmd/largo-installer`
- With flags:
  - `--largo-version vX.Y.Z` to target a specific version
  - `--module <path>` to override the module path
  - `--no-color` to disable colors
  - `--version` to print installer version

## Project Layout

- `cmd/largo-installer`: CLI entrypoint (flag parsing, start UI)
- `internal/ui`: Bubble Tea UI (layout, steps, logs)
- `internal/install`: Installer logic + per‑OS PATH handling
- `internal/buildinfo`: Build‑time `Version` variable

## Coding Guidelines

- Prefer simple, explicit code over cleverness.
- Avoid global state; pass `io.Writer` for logs where possible.
- Keep OS‑specific code behind build tags (`//go:build windows`, `//go:build !windows`).
- When adding dependencies, justify them in the PR and keep the footprint minimal.
- Error messages should be actionable and user‑friendly.

## Testing & Checks

Run these before pushing:

```
go fmt ./...
go vet ./...
go build ./...
go test ./...
```

If you add new helpers (e.g., PATH manipulation), consider adding unit tests. Try to avoid adding tests that depend on network access or modify real user environment.

## Commit & PR Tips

- One logical change per commit when possible.
- Reference issues with `Fixes #123` or `Refs #123` when appropriate.
- Describe the motivation, the approach, and any trade‑offs in the PR description.
- Add screenshots or terminal captures for UI changes.

## Release Process (maintainers)

- Tags trigger GoReleaser; artifacts are built for major OS/arch combinations.
- To test locally: `goreleaser release --clean --snapshot`.
- Embed version via ldflags during builds: `-X .../internal/buildinfo.Version=<tag>`.

## Reporting Bugs & Security Issues

- Bugs: open a GitHub issue with OS/terminal, Go version, installer version, steps to reproduce, and logs.
- Security: please do not open a public issue. Email the maintainer (see Code of Conduct) with details to coordinate a responsible disclosure.

Thanks again for contributing!
