# Repository Guidelines

## Project Structure & Module Organization
- Root module: `github.com/MeKo-Christian/GoRapidOCR` (Go 1.22).
- Core package: `paddleocr.go` and `utils.go` in the repo root, package name `GoRapidOCR`.
- Tests: `paddleocr_test.go` alongside sources.
- Docs & configs: `README.md`, `.trunk/` (linters, formatting), `.gitignore`, `LICENSE`.

## Build, Test, and Development Commands
- Build package: `go build ./...` — compiles all packages.
- Run tests: `go test ./...` — runs unit tests. For coverage: `go test ./... -race -cover`.
- Lint & format (via Trunk): `trunk check -a` and `trunk fmt`.
- Direct tools (if installed): `gofmt -s -w .`, `golangci-lint run`, `markdownlint .`.

## Coding Style & Naming Conventions
- Formatting: Go code must be `gofmt`-clean; run `trunk fmt` before pushing.
- Style: Follow standard Go conventions. Exported identifiers use CamelCase (e.g., `NewPpocr`, `OcrFileAndParse`).
- Files: Keep package files in root for now; name tests `*_test.go`.
- Comments: Provide GoDoc comments for all exported types, funcs, and constants.

## Testing Guidelines
- Framework: Standard `testing` package.
- Naming: `TestXxx` functions; table-driven tests preferred.
- External binary: Some flows require a PaddleOCR-json executable. If unavailable, focus on pure unit tests (e.g., `OcrArgs.CmdString`) or run subsets: `go test -run CmdString`.
- CI expectation: All tests green; avoid fragile, environment-dependent assertions.

## Commit & Pull Request Guidelines
- Commits: Use concise, descriptive messages. Conventional prefixes welcomed (e.g., `feat:`, `fix:`, `chore:`, `docs:`). Emojis are acceptable but optional.
- PRs must include:
  - Clear description and rationale; link issues if applicable.
  - Tests for new behavior or bug fixes.
  - Passing `trunk check -a` and `go test ./...`.
  - Updated `README.md` or comments when APIs change.

## Security & Configuration Tips
- The library launches an external OCR binary. Always pass a trusted `exePath` and validate file existence.
- Cross‑platform paths: prefer absolute paths and use correct separators (see examples in `README.md`).
- Long‑running processes are auto‑restarted to mitigate leaks; always call `Close()` when done.
