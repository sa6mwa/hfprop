# Repository Guidelines

## Project Structure & Module Organization
- Core library: `hfprop.go` (LGDC client, HF propagation helpers).
- CLI: `cmd/hfprop` (`main.go` bootstraps Cobra; `cmd/root.go` sets global flags; `fof2.go` holds the foF2 command; `functions.go` parses natural-language times).
- Tests: `hfprop_test.go` uses `httptest` fixtures; build outputs land in `bin/` via the Makefile.

## Build, Test, and Development Commands
- `go fmt ./...` — format; run first.
- `go vet ./...` — static checks.
- `golint ./...` — linting (expect style warnings).
- `modernize ./...` — codebase modernization pass (run after lint).
- `go test ./...` — unit tests; offline.
- `make build` — create `bin/hfprop` with stripped symbols.
- `go install ./cmd/hfprop@latest` — install CLI locally when you need a manual end-to-end check.

## Coding Style & API Contracts
- Go 1.19; keep code `gofmt`-clean, exported names in PascalCase, locals in mixedCaps; prefer early returns and small helpers.
- Follow the Cobra pattern: define flags in `init`, keep command logic in `Run`.
- User-facing functions must return at most two values; if more data is needed, return `(ResultStruct, error)` rather than multiple scalars.
- When adding LGDC params, extend the `HFProp` API and constants (`LgdcKey*`, `Default*`) instead of per-command hacks.

## Testing Guidelines
- Use the standard `testing` package; prefer table-driven tests for varied inputs.
- Mock LGDC with `httptest` servers and minimal fixtures; avoid real network calls in automated runs.
- Name tests `TestFunction_Scenario` and assert both parameter and value expectations.

## Branching & Workflow
- Do not create commits or PRs here; leave changes uncommitted so the maintainer can stage and commit.
- Work in short-lived branches, develop, then merge; include the command set above in your own pre-merge checklist.

## Architecture & Growth Notes
- Purpose: provide HF propagation data and forecasting (MUF and related metrics) for radio-communication use.
- Backlog: add smart forecasting that outputs JSON with hourly MUF and confidence for up to ~1 week, based on historic LGDC data from one or more ionosondes.
- As the codebase grows, modularize into `internal/` packages and keep components behind narrow interfaces to simplify testing and swaps.

## Security & Configuration
- Default LGDC base URL is public; do not embed secrets or API keys.
- Flags `--base-url` and `--ursi-code` let you target mirrors or mocks; keep defaults intact for users.
