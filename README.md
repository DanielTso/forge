# forge

`forge` scaffolds new projects using industry-standard structure — README, LICENSE, `.gitignore`,
`.editorconfig`, CHANGELOG, CONTRIBUTING, `docs/`, CI workflows, a Makefile and starter code with tests —
so every project starts from the same solid baseline.

## Install

```sh
go install github.com/DanielTso/forge@latest   # once the repo is published
make install                                    # from a checkout (embeds version info)
```

Tagged releases (`vX.Y.Z`) publish Linux/macOS/Windows archives (amd64 + arm64) via GoReleaser;
`make snapshot` builds them locally without publishing.

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
```

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

`make test` also generates real projects and runs their toolchains (`go`, `node`, `python3`, `cargo`);
a language whose tools are not installed is skipped.
