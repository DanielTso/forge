# Changelog

All notable changes to this project are documented here.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-11-19

### Added

- `forge upgrade`: updates an existing project's GitHub Actions pins (`uses: owner/action@vN`) to the
  versions forge's templates use, changing only the version text. It never downgrades, leaves SHA and
  branch pins and unknown actions alone, supports `--dry-run`, and suggests `forge add dependabot` / `ci`
  when missing. A test guarantees freshly generated projects never need an upgrade.

## [0.2.1] - 2026-11-19

### Changed

- `forge add` / `forge init` detection: a repo with both `package.json` and `tsconfig.json` is now
  detected as `typescript` (it was `node` before 0.2.0), so `forge add` writes TypeScript-aware files there.

### Fixed

- A Python project that also has a TypeScript or Deno frontend was detected as `typescript`/`deno` in 0.2.0;
  Python is detected first again, as in 0.1.x.
- Scoped package names (`@me/tool`) in `package.json` and `deno.json` now give the project name `tool`,
  and Deno projects get their name from `deno.json`.
- Zig packages named in quoted form (`.name = .@"my-lib"`) are detected with their full name.
- README: `dockerfile` is unavailable for `go-lib`, `gleam`, `zig` and `generic`, not only libraries.

## [0.2.0] - 2026-11-19

### Added

- Four new project types: `typescript` (strict TypeScript on Node, pinned to 6.x so language servers work,
  node:test, typescript-eslint), `deno` (built-in fmt/lint/check/test), `gleam` (Erlang VM, gleeunit) and
  `zig` (0.17, executable + library module, generated package fingerprint).
- `forge add` detects these projects (`tsconfig.json`, `deno.json`, `gleam.toml`, `build.zig`); `dockerfile`
  is available for `typescript` and `deno`.
- forge's CI installs Python tools, Deno, Gleam/Erlang and Zig so every template is exercised on each push.

## [0.1.3] - 2026-11-19

### Added

- Generated projects include `.github/dependabot.yml` (language ecosystem + GitHub Actions,
  weekly); `forge add dependabot` adds it to existing projects.

### Changed

- Generated CI workflows use current action versions: checkout v7, setup-go v7, setup-node v7,
  setup-python v7, golangci-lint-action v9.

### Fixed

- Generated `package.json` had its `"license"` line mis-indented (template whitespace trimming).

## [0.1.2] - 2026-11-18

### Changed

- `forge --version` omits the commit when it is unknown (e.g. `go install` builds now print
  `forge v0.1.1` instead of `forge v0.1.1 (none)`), and marks local builds with uncommitted
  changes as `-dirty`.

## [0.1.1] - 2026-11-18

### Fixed

- `forge --version` showed `dev (none)` when installed with `go install`; it now reports the
  module version from Go's embedded build info (and the commit, for local builds).

## [0.1.0] - 2026-11-18

### Added

- `forge new`, `forge init`, `forge add` and `forge list` commands.
- Project types: `go-cli`, `go-service`, `go-lib`, `python`, `node`, `rust`, `generic`.
- Components for `forge add`: ci, dockerfile, license, gitignore, editorconfig, changelog,
  contributing, github, docs, makefile, lint.
- Interactive prompts for a bare `forge new` / `forge init`.
- MIT, Apache-2.0 and BSD-3-Clause licenses.
- Go Dockerfile accepts `VERSION` / `COMMIT` build args and stamps them into the binary;
  the go-service `make docker` target passes them.
- `make test-docker`: opt-in test that builds and runs every generated Dockerfile and checks
  the image does not run as root.
- Integration tests run each generated project's own toolchain (golangci-lint; ruff, mypy and
  pytest; node --test; cargo fmt, clippy and test).

[Unreleased]: https://github.com/DanielTso/forge/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/DanielTso/forge/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/DanielTso/forge/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/DanielTso/forge/compare/v0.1.3...v0.2.0
[0.1.3]: https://github.com/DanielTso/forge/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/DanielTso/forge/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/DanielTso/forge/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/DanielTso/forge/releases/tag/v0.1.0
