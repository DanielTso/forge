# forge

[![CI](https://github.com/DanielTso/forge/actions/workflows/ci.yml/badge.svg)](https://github.com/DanielTso/forge/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/DanielTso/forge)](https://github.com/DanielTso/forge/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/DanielTso/forge.svg)](https://pkg.go.dev/github.com/DanielTso/forge)

`forge` scaffolds new projects using industry-standard structure — README, LICENSE, `.gitignore`,
`.editorconfig`, CHANGELOG, CONTRIBUTING, `docs/`, CI workflows, a Makefile and starter code with tests —
so every project starts from the same solid baseline. Go, Python, Node and Rust.

**📖 [Read the wiki](https://github.com/DanielTso/forge/wiki)**: [Quick Start](https://github.com/DanielTso/forge/wiki/Quick-Start) · [Project Types](https://github.com/DanielTso/forge/wiki/Project-Types) · [Using forge with AI](https://github.com/DanielTso/forge/wiki/Using-forge-with-AI) · [Troubleshooting](https://github.com/DanielTso/forge/wiki/Troubleshooting-and-FAQ)

## Quick look

```console
$ forge new todo-app -m github.com/you/todo-app
Created go-cli project "todo-app" in ./todo-app
  ...
Initialized a git repository (branch: main).

$ cd todo-app && tree
.
├── .github/              # CI workflow, issue + PR templates
├── cmd/todo-app/main.go  # entry point
├── docs/                 # README + architecture notes
├── internal/app/         # your code, with a passing test
├── .editorconfig  .gitignore  .golangci.yml
├── CHANGELOG.md  CONTRIBUTING.md  LICENSE  README.md
├── Makefile              # make help | test | run | lint | check
└── go.mod

$ make test
ok  	github.com/you/todo-app/internal/app	coverage: 100.0% of statements

$ forge add dockerfile    # add pieces later; never overwrites existing files
Added dockerfile to this go-cli project:
  .dockerignore
  Dockerfile
```

Every generated project is tested in CI with its own toolchain (golangci-lint; ruff, mypy, pytest;
`node --test`; cargo fmt, clippy, test), and every generated Dockerfile is built and run.

## Install

```sh
go install github.com/DanielTso/forge@latest
```

No Go toolchain? Download a prebuilt binary for Linux, macOS or Windows (amd64 / arm64) from the
[latest release](https://github.com/DanielTso/forge/releases/latest), unpack it, and put `forge` on your `PATH`.
Each release includes `checksums.txt` for verification.

### Shell completion

```sh
forge completion bash > ~/.local/share/bash-completion/completions/forge   # bash
forge completion zsh  > "${fpath[1]}/_forge"                               # zsh
forge completion fish > ~/.config/fish/completions/forge.fish              # fish
```

## Usage

```sh
forge list                                   # project types and licenses
forge new                                    # interactive: asks for each option
forge new my-tool                            # Go CLI in ./my-tool (default type)
forge new my-api -t go-service -m github.com/me/my-api
forge new my-app -t python                   # also: node, rust, go-lib, generic
forge new my-lib -t go-lib -l Apache-2.0 --dry-run   # preview only
mkdir notes && cd notes && forge init -t generic     # scaffold in the current directory
forge add                                    # list components
forge add ci dockerfile                      # add pieces to an existing project
```

### Interactive mode

A bare `forge new` or `forge init` in a terminal prompts for the project name, type, description,
Go module (Go types), author and license, offering sensible defaults (Enter accepts them).

- `-i, --interactive` prompts for every value you did **not** give as a flag (works with piped input too).
- `-y, --yes` never prompts; flags and defaults only (use in scripts).
- Passing any value flag (`-t`, `-m`, `-d`, `-a`, `-l`, ...) or a project name skips bare-invocation prompting.

### Flags for `new` / `init`

| Flag | Meaning |
| --- | --- |
| `-t, --type` | `go-cli` (default), `go-service`, `go-lib`, `python`, `node`, `rust`, `generic` |
| `-m, --module` | Go module path (default: project name) |
| `-d, --description` | one-line description |
| `-a, --author` | copyright holder (default: `git config user.name`) |
| `-l, --license` | `MIT` (default), `Apache-2.0`, `BSD-3-Clause`, `none` |
| `--go-version` | `go` directive for `go.mod` (default: toolchain forge was built with) |
| `-o, --output` | parent directory (`new` only) |
| `--name` | project name (`init` only; default: directory name) |
| `-i`, `-y` | interactive / never prompt (see above) |
| `--no-git` | skip `git init` |
| `--force` | overwrite existing files |
| `--dry-run` | print the plan, write nothing |

Existing files are **never** overwritten unless `--force` is given; on any conflict nothing is written.
`git init` is skipped when the target is already inside a repository.

### `forge add`

Adds individual components to the project in the current directory. The language is detected from
`go.mod` (a `cmd/` directory makes it a CLI, otherwise a library), `pyproject.toml`, `package.json`
or `Cargo.toml`, so content is tailored (override with `--type`). The project name is read from the
manifest (or `cmd/<name>`), then falls back to the directory name (`--name` overrides).

| Component | Adds |
| --- | --- |
| `ci` | GitHub Actions workflow for the language |
| `dockerfile` | multi-stage / slim `Dockerfile` + `.dockerignore` (not for libraries) |
| `license` | `LICENSE` (`--license`, `--author`) |
| `gitignore`, `editorconfig` | language-aware `.gitignore`, `.editorconfig` |
| `changelog`, `contributing` | `CHANGELOG.md`, `CONTRIBUTING.md` |
| `github` | issue and PR templates |
| `dependabot` | weekly dependency + GitHub Actions updates |
| `docs` | `docs/` index and architecture page |
| `makefile` | Makefile with `help/test/check/...` targets |
| `lint` | `.golangci.yml`, `rustfmt.toml` or `eslint.config.js` |

All components are validated before anything is written, and `--dry-run` / `--force` behave as above.

## Project types

| Type | Generates |
| --- | --- |
| `go-cli` | `cmd/<name>/`, `internal/app/`, tests, Makefile, golangci config, CI |
| `go-service` | HTTP server with `/healthz`, graceful shutdown, slog, Dockerfile, CI |
| `go-lib` | root package, godoc example, `examples/`, CI |
| `python` | `src/` layout, `pyproject.toml` (hatchling), pytest, ruff, mypy, Makefile, CI |
| `node` | ESM package, `bin/` entry point, `node:test`, ESLint, Makefile, CI |
| `rust` | binary + library crate, integration test, rustfmt, clippy, Makefile, CI |
| `generic` | only the language-agnostic files below |

All types include: `README.md`, `LICENSE`, `CHANGELOG.md`, `CONTRIBUTING.md`, `.gitignore`, `.editorconfig`,
`docs/` (index + architecture), and GitHub issue/PR templates.

## Development

```sh
make check    # fmt, vet, lint (golangci-lint), test
make build
make install  # from a checkout (embeds version info)
make snapshot # build release archives locally with GoReleaser; nothing is published
```

Documentation sources:

- `scripts/wiki/gen-wiki.sh <dir>` regenerates the [wiki](https://github.com/DanielTso/forge/wiki) pages
  from real `forge` output; copy them into a clone of `forge.wiki.git` and push.
- `docs/forge-guide.html` is the source of the beginner's visual guide.

Releases: add the changes to `CHANGELOG.md`, then push a `vX.Y.Z` tag. The Release workflow builds
and publishes archives for 6 OS/arch targets.

Templates live in `internal/scaffold/templates/<set>/`, embedded into the binary with `go:embed`.

- Files ending in `.tmpl` are rendered with `text/template` using `<%` `%>` delimiters
  (so `${{ }}` and Go braces need no escaping); the suffix is stripped. Other files are copied verbatim.
- `__name__` and `__pkg__` in paths are replaced by the project name / Go package name.
- `__name__`, `__pkg__` (Go package) and `__snake__` (Python package / Rust crate) in paths are replaced.
- Files that start with a `#!` line are written executable.
- A project type is a list of template sets applied in order; later sets override earlier files.
- Optional pieces live in `templates/extras/<name>/<lang>/`; `forge add` renders them on demand.
- To add a type, create a template set and register it in `internal/scaffold/types.go`.
- To add a component for `forge add`, register it in `internal/scaffold/components.go`
  (either file patterns matched against a rendered project, or an extras set).

`make test` also generates real projects and runs their toolchains — `go` + `golangci-lint`, `node`,
`ruff` / `mypy` / `pytest`, `cargo fmt` / `clippy` / `test` — skipping any tool that is not installed.
`make test-docker` builds and runs every generated Dockerfile (needs Docker; opt-in because it pulls images).
