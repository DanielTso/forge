"""Builds the forge GitHub wiki pages. Run via scripts/wiki/gen-wiki.sh, which supplies the generated trees."""
import os, subprocess, sys
W, GEN = sys.argv[1], sys.argv[2]
def tree(t): return open(os.path.join(GEN, t + ".tree")).read().rstrip()
def helptext(*args):
    env = dict(os.environ, PATH=os.environ["PATH"] + ":" + os.path.expanduser("~/go/bin"))
    out = subprocess.run(["forge", *args], capture_output=True, text=True, cwd="/tmp", env=env)
    return "\n".join(l.rstrip() for l in (out.stdout + out.stderr).rstrip().splitlines())
P = {}

P["Home"] = """# forge wiki

**forge** builds the skeleton of a new software project in about a second: folder layout, starter code with a passing test, README, LICENSE, CHANGELOG, docs, a Makefile and GitHub CI. Every project starts from the same professional baseline, whether you write the code yourself or an AI writes it for you.

```console
$ forge new todo-app
Created go-cli project "todo-app" in ./todo-app
$ cd todo-app && make test
ok   github.com/you/todo-app/internal/app   coverage: 100.0% of statements
```

## Start here

| If you want to… | Read |
|---|---|
| Install forge | [Installation](Installation) |
| Make your first project in 5 minutes | [Quick Start](Quick-Start) |
| Choose between Go, Python, TypeScript, Deno, Rust, Gleam, Zig… | [Project Types](Project-Types) |
| Add CI, Docker or a license to an existing project | [Adding Components](Adding-Components) |
| Build with Claude or another AI assistant | [Using forge with AI](Using-forge-with-AI) |
| Look up a flag | [Command Reference](Command-Reference) |
| Understand what forge will and won't touch | [How forge Keeps You Safe](How-forge-Keeps-You-Safe) |
| Fix a problem | [Troubleshooting and FAQ](Troubleshooting-and-FAQ) |
| Add your own templates or contribute | [Customizing Templates](Customizing-Templates) |

## What you get, every time

- **Starter code with a passing test**, so you have a working baseline before writing anything.
- **A Makefile** with the same commands in every language: `make test`, `make run`, `make check`, `make help`.
- **GitHub CI** that runs your tests and linters on every push.
- **Dependabot**, which opens weekly pull requests to keep dependencies and CI up to date.
- **Docs and housekeeping**: README, LICENSE (MIT, Apache-2.0 or BSD-3-Clause), CHANGELOG, CONTRIBUTING, `.gitignore`, `.editorconfig`, issue and PR templates.

All of this is tested: forge's test suite generates every project type and runs that language's real tools on it (golangci-lint, ruff, mypy, pytest, `node --test`, cargo fmt, clippy and test). A separate Docker test builds and runs every generated Dockerfile.
"""

P["Installation"] = """# Installation

## Option 1: with Go (recommended)

If you have [Go](https://go.dev/dl/) 1.27 or newer:

```sh
go install github.com/DanielTso/forge@latest
forge --version
```

`go install` puts the program in `~/go/bin` (Windows: `%USERPROFILE%\\go\\bin`). If your shell says `forge: command not found`, add that folder to your `PATH`, or open a new terminal.

## Option 2: download a ready-made binary (no Go needed)

1. Open the [latest release](https://github.com/DanielTso/forge/releases/latest).
2. Download the file for your system:

   | System | File |
   |---|---|
   | Windows (most PCs) | `forge_<version>_windows_amd64.zip` |
   | Mac with Apple chip (M1–M4) | `forge_<version>_darwin_arm64.tar.gz` |
   | Mac with Intel chip | `forge_<version>_darwin_amd64.tar.gz` |
   | Linux / WSL | `forge_<version>_linux_amd64.tar.gz` |

3. Unpack it and move `forge` (or `forge.exe`) into a folder on your `PATH`, e.g. `~/.local/bin` or `/usr/local/bin`.

To check the download wasn't corrupted, compare it against `checksums.txt` from the same release:

```sh
sha256sum -c --ignore-missing checksums.txt
```

## Shell completion (optional)

Lets you press <kbd>Tab</kbd> to complete commands, project types and licenses.

```sh
forge completion bash > ~/.local/share/bash-completion/completions/forge   # bash
forge completion zsh  > "${fpath[1]}/_forge"                               # zsh
forge completion fish > ~/.config/fish/completions/forge.fish              # fish
```

Open a new terminal, then try `forge new --type <Tab>`.

## Tools for the projects you generate

forge itself needs nothing else. To *run* the projects it creates you need that language's tools:

| Project type | Needs |
|---|---|
| `go-cli`, `go-service`, `go-lib` | Go; optionally [golangci-lint](https://golangci-lint.run) for `make lint` |
| `python` | Python 3.10+ (the Makefile creates a private `.venv`) |
| `node` | Node.js 20+ |
| `typescript` | Node.js 22+ |
| `deno` | [Deno](https://deno.com) 2.x |
| `rust` | Rust via [rustup](https://rustup.rs) |
| `gleam` | [Gleam](https://gleam.run) and Erlang/OTP |
| `zig` | [Zig](https://ziglang.org) 0.17.1 |
| Any `Dockerfile` | Docker |

Next: [Quick Start](Quick-Start)
"""

P["Quick-Start"] = """# Quick Start

This walkthrough builds a small to-do app skeleton. Every block of output below was recorded from a real run.

## 1. See what you can make

```console
$ forge list
""" + helptext("list") + """
```

## 2. Preview first (optional)

`--dry-run` shows what would be created without writing anything:

```console
$ forge new todo-app --dry-run
Would create go-cli project "todo-app" in ./todo-app:
  .editorconfig
  .github/ISSUE_TEMPLATE/bug_report.md
  .github/dependabot.yml
  ...
```

## 3. Create it

Run `forge new` with no arguments and answer the questions. Press <kbd>Enter</kbd> to accept the default shown in `[brackets]`.

```console
$ forge new
Let's set up your project (press Enter to accept the default).
Project name: todo-app
Project type:
  1) go-cli      Go command-line application (cmd/, internal/, Makefile, CI)
  2) go-service  Go HTTP service with health check, graceful shutdown, Dockerfile
  ...
Choose [1]: 1
Description: A tiny to-do list app
Go module path [todo-app]: github.com/you/todo-app
Author [Your Name]:
License (Apache-2.0, BSD-3-Clause, MIT, none) [MIT]:

Created go-cli project "todo-app" in ./todo-app
  ...
Initialized a git repository (branch: main).

Next steps:
  cd todo-app
  make test
  make run
```

Prefer one line? This does the same with defaults:

```sh
forge new todo-app -d "A tiny to-do list app" -m github.com/you/todo-app
```

## 4. Look around

```text
""" + tree("go-cli").replace("my-go-cli", "todo-app") + """
```

The file you'll edit first is `internal/app/app.go`. The rest is scaffolding that keeps the project healthy; see [Project Types](Project-Types) for what each file does.

## 5. Run it

```console
$ cd todo-app
$ make test
go test -race -cover ./...
ok   github.com/you/todo-app/internal/app   1.023s   coverage: 100.0% of statements
$ make run
Hello from todo-app!
$ make help        # lists every shortcut
```

## 6. Publish it (optional)

forge already ran `git init`. To put the project on GitHub with the [GitHub CLI](https://cli.github.com):

```sh
git add -A && git commit -m "Initial scaffold from forge"
gh repo create todo-app --public --source . --push
```

The included CI workflow runs your tests on GitHub straight away.

Next: [Adding Components](Adding-Components) · [Using forge with AI](Using-forge-with-AI)
"""

notes = {
 "go-cli": ("A command-line program in Go.", "`internal/app/app.go` is your code; `cmd/<name>/main.go` only wires it up (flags, exit codes, version info). Smallest Docker image: about 3.6 MB.", "make test · make run · make build · make lint · make check"),
 "go-service": ("An HTTP API / web service in Go.", "`internal/server/server.go` holds the routes; `GET /healthz` returns `{\"status\":\"ok\"}`. The server logs JSON, reads `ADDR` (default `:8080`) and shuts down cleanly on Ctrl-C or SIGTERM. Includes a Dockerfile that builds a non-root image of about 9 MB.", "make test · make run · make docker · make check"),
 "go-lib": ("A reusable Go package for other programs to import.", "The package lives at the project root; `example_test.go` doubles as documentation on pkg.go.dev, and `examples/basic` is a runnable demo.", "make test · make lint · make check"),
 "python": ("A Python package with a command-line entry point.", "Code lives in `src/<name>/` (the *src layout*, which makes tests run against the installed package). Tests use pytest, style uses ruff, types use mypy (strict). Run `make install` first; it creates a private `.venv`.", "make install · make test · make run · make check"),
 "node": ("A Node.js command-line tool or library (ES modules).", "`src/index.js` is the library, `bin/<name>.js` the command. Tests use Node's built-in runner, so `npm test` works with no downloads; ESLint checks style after `npm install`.", "make install · make test · make lint · make run"),
 "rust": ("A Rust program with its logic in a library crate.", "`src/lib.rs` holds the logic (easy to unit-test); `src/main.rs` is a thin binary; `tests/cli.rs` runs the real compiled program. `make check` runs rustfmt, clippy (warnings are errors) and the tests.", "make test · make run · make check"),
 "typescript": ("TypeScript on Node.js, for CLIs, APIs and libraries.", "`src/` holds the code and `test/` the tests; `tsc` compiles both into `dist/` and Node's built-in test runner runs them. Strict compiler settings catch mistakes early, and ESLint uses typescript-eslint. TypeScript is pinned to 6.x because TypeScript 7 removed `tsserver`, which code-intelligence tools (including Claude's language server) still need. Run `make install` first.", "make install · make test · make check · make run"),
 "deno": ("TypeScript with Deno, the all-in-one runtime.", "No config to maintain: Deno has its own formatter, linter, type checker and test runner (`make check` runs all four). `main.ts` is the entry point and `src/` holds the code, with `*_test.ts` files beside it. `make build` compiles a single standalone binary.", "make run · make test · make check · make build"),
 "gleam": ("Gleam, a friendly type-safe language that runs on the Erlang VM.", "`src/<name>.gleam` holds the code, and tests in `test/` use gleeunit (public functions ending in `_test`). The formatter is part of the compiler, and dependencies come from Hex. Needs Gleam and Erlang/OTP installed.", "make run · make test · make check"),
 "zig": ("Zig, a fast systems language and a modern alternative to C.", "`src/root.zig` is the library (with its tests inline) and `src/main.zig` is the program. `build.zig` defines `zig build`, `zig build run` and `zig build test`. `build.zig.zon` holds the package name and a generated fingerprint that Zig requires. Targets Zig 0.17.1; Zig is pre-1.0, so expect template updates as it changes.", "make run · make test · make check · make build"),
 "generic": ("Any language, or docs-only projects.", "Only the shared, language-neutral files. Use it for documentation repos, or languages forge has no template for yet.", "(no Makefile)"),
}
pt = ["# Project Types\n\nPick the type with `-t` / `--type` (default `go-cli`), or choose from the menu in `forge new`.\n\n| Type | Best for |\n|---|---|\n"]
for t,(best,_,_) in notes.items(): pt.append(f"| [`{t}`](#{t}) | {best} |\n")
pt.append("""
## Files every type gets

| File | What it's for |
|---|---|
| `README.md` | The project's front page, with how-to-run instructions for its language |
| `LICENSE` | MIT, Apache-2.0 or BSD-3-Clause (or none), with your name and the year |
| `CHANGELOG.md` | What changed in each version ([Keep a Changelog](https://keepachangelog.com) format) |
| `CONTRIBUTING.md` | How others can help |
| `.gitignore` | Files git should never save: OS junk, secrets (`.env`), and the language's build output |
| `.editorconfig` | Consistent spacing and line endings in any editor |
| `docs/README.md`, `docs/architecture.md` | Documentation index and a map of the code layout |
| `.github/workflows/ci.yml` | GitHub runs your tests and linters on every push and pull request |
| `.github/dependabot.yml` | Weekly pull requests that update dependencies and CI actions |
| `.github/ISSUE_TEMPLATE/`, `pull_request_template.md` | Structured bug reports and PR checklists |

""")
for t,(best,detail,cmds) in notes.items():
    pt.append(f"## {t}\n\n{best} {detail}\n\n```sh\nforge new my-{t} -t {t}\n```\n\n```text\n{tree(t)}\n```\n\n**Common commands:** {cmds}\n\n")
pt.append("To add a Dockerfile to `go-cli`, `python`, `node`, `typescript`, `deno` or `rust` projects: `forge add dockerfile`. See [Adding Components](Adding-Components).\n")
P["Project-Types"] = "".join(pt)

P["Adding-Components"] = """# Adding Components

`forge add` drops individual pieces into a project that **already exists**, whether forge made it or not. Run it from the project folder.

```console
$ forge add
""" + helptext("add") + """
```

## How it tailors the files

forge looks at the folder to work out the language:

| Found | Treated as |
|---|---|
| `go.mod` and a `cmd/` folder | `go-cli` (the project name comes from `cmd/<name>` if there's exactly one) |
| `go.mod`, no `cmd/` | `go-lib` |
| `pyproject.toml` | `python` (name read from `[project]`) |
| `package.json` | `node` (name read from `"name"`) |
| `Cargo.toml` | `rust` (name read from `[package]`) |
| `deno.json` | `deno` |
| `gleam.toml` | `gleam` (name read from the top-level `name`) |
| `build.zig` | `zig` (name read from `build.zig.zon`) |
| `package.json` and `tsconfig.json` | `typescript` (name read from `"name"`) |
| none of these | `generic` |

So `forge add ci` gives a Python project a workflow that runs ruff, mypy and pytest, and a Rust project one that runs rustfmt, clippy and cargo test. Override detection with `--type`, and the name with `--name`.

## Examples

```console
$ forge add dockerfile ci --dry-run          # preview
Would add dockerfile, ci to this python project:
  .dockerignore
  Dockerfile
  .github/workflows/ci.yml  (exists: would need --force)

$ forge add dockerfile
Added dockerfile to this python project:
  .dockerignore
  Dockerfile

$ forge add license --license Apache-2.0 --author "Jane Doe"
$ forge add dependabot github                # keep deps current + issue/PR templates
```

## Rules

- **Nothing is overwritten** unless you pass `--force`. If even one file already exists, forge writes nothing and lists the conflicts.
- **All-or-nothing:** a typo in any component name (`forge add docs nope`) aborts before anything is written.
- **Some pieces don't fit some types:** `dockerfile` is refused for libraries, `gleam`, `zig` and `generic` projects; `ci` needs a language.
"""

P["Interactive-Mode"] = """# Interactive Mode

forge can ask you questions instead of making you remember flags.

## When it asks

| You run | What happens |
|---|---|
| `forge new` or `forge init` in a terminal, with no project name and no value flags | Asks for **every** option |
| any `new` / `init` command with `-i` / `--interactive` | Asks only for what you **didn't** give as a flag |
| any command with `-y` / `--yes` | Never asks; uses flags and defaults (use this in scripts) |
| input piped from a file or another program | Never asks unless you pass `-i` |

Behaviour-only flags like `--no-git`, `--force` and `--dry-run` don't stop it asking; flags that set values (`-t`, `-m`, `-d`, `-a`, `-l`, `--name`, `-o`) do.

## The questions

1. **Project name**: letters, digits, `.`, `_`, `-`; must start with a letter. (`forge init` suggests the folder name.)
2. **Project type**: type the number or the name.
3. **Description**: one line; goes into the README, package metadata and code comments.
4. **Go module path** (Go types only), e.g. `github.com/you/my-app`.
5. **Author**: defaults to your `git config user.name`.
6. **License**: `MIT`, `Apache-2.0`, `BSD-3-Clause` or `none` (not case-sensitive).

Wrong answers are explained and asked again:

```console
Project name: 1bad
  invalid project name "1bad": must start with a letter and contain only letters, digits, '.', '_' or '-'
Project name: good-name
Choose [1]: 99
  enter a number between 1 and 11, or a type name
```

## Example: mix flags and questions

```sh
forge new my-api -t go-service -i
```

Asks only for the description, module path, author and license.
"""

P["Command-Reference"] = "# Command Reference\n\nOutput of each command's `--help`, from forge v0.2.0.\n\n" + "".join(
    f"## `forge {c}`\n\n```text\n{helptext(*([c] if c else []), '--help')}\n```\n\n" for c in ["", "new", "init", "add"]
) + """## `forge list`

Prints project types and licenses. No flags.

## `forge completion`

Prints a shell completion script: `forge completion bash|zsh|fish|powershell`. See [Installation](Installation#shell-completion-optional).

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success (including `--dry-run`) |
| `1` | Any error: invalid name or type, file conflicts, unknown component, bad flags |
"""

P["Using-forge-with-AI"] = """# Using forge with AI (vibe coding)

AI assistants write better code into a good structure. Without one, an AI invents a new layout every time, skips tests, and nothing catches it when a change breaks something. Starting with forge gives the AI:

- **A known layout** to follow, described in `docs/architecture.md`.
- **A working test to copy**, so new code comes with tests in the same style.
- **One command to check everything**, `make check`, which the AI can run after each change.
- **CI on GitHub** as a safety net for anything that slips through.

## Prompts you can paste into Claude Code

Replace the words in CAPITALS.

**Start a new project**

> Use forge to create a new python project called MY-APP with the description "WHAT IT DOES". Read docs/architecture.md, then build that feature inside src/, add tests in tests/ for every function you write, and run make check until it passes.

**Fix up an existing folder**

> In this folder, run `forge add --dry-run` with ci, gitignore, license, dependabot and docs, show me what it would create, then add the ones that don't already exist.

**Add a feature safely**

> Add FEATURE. Follow the existing layout, write tests next to the existing ones, and run make check. Don't change files outside the source and test folders unless you tell me why.

**Ship it**

> Run make check. If anything fails, fix it. Then update CHANGELOG.md under "Unreleased" with what we built, commit everything, and show me the command to push to GitHub.

## Tips

- **Keep `make check` green.** Ask the AI to run it after every change; it's the fastest way to catch a broken edit.
- **Let CI be the referee.** If CI fails on GitHub, paste the failing log into the chat and ask for a fix.
- **Read the Dependabot PRs.** Each one updates a single dependency; if CI passes on the PR, it's usually safe to merge.
- **Use `--dry-run`** whenever you're unsure what a forge command will do.
"""

P["How-forge-Keeps-You-Safe"] = """# How forge Keeps You Safe

forge is designed so it's hard to lose work.

| Guarantee | Details |
|---|---|
| **Never overwrites** | If any file it wants to write already exists, `new`, `init` and `add` stop **before writing anything** and list the conflicts. Use `--force` only when you mean it. |
| **Preview everything** | `--dry-run` prints the full plan and marks files that already exist. |
| **No half-finished results** | Inputs are all checked first (name, type, license, component names), so a typo never leaves half a project behind. |
| **Careful with git** | `git init` runs only for a folder that isn't already inside a git repository, and never commits for you. Skip it with `--no-git`. |
| **Works offline** | All templates are built into the program. forge never downloads anything or sends anything anywhere. |
| **Safe names** | Project names can't contain `/` or start with `.`, so a project can't be written outside its folder. |

## What the generated projects do for safety

- `.gitignore` excludes `.env` files and secrets, plus WSL/Windows `:Zone.Identifier` files.
- Dockerfiles run as a **non-root** user, and the Go service shuts down cleanly on SIGTERM.
- CI workflows run with read-only repository permissions (`contents: read`).
- Dependabot keeps dependencies patched.
"""

P["Troubleshooting-and-FAQ"] = """# Troubleshooting and FAQ

### `forge: command not found`
Go installs programs into `~/go/bin`. Add it to your PATH (`export PATH="$PATH:$HOME/go/bin"` in `~/.bashrc`), then open a new terminal.

### `Error: N file(s) already exist: … (use --force to overwrite)`
forge refused to overwrite your files, and nothing was written. Either pick a new project name or folder, remove or rename those files, or re-run with `--force` if you really want forge's versions.

### `invalid project name "…"`
Names must start with a letter and use only letters, digits, `.`, `_`, `-`. Rust projects can't use `.`. With `forge init` or `forge add`, the folder name is used by default; pass `--name` if the folder name isn't valid (e.g. `001`).

### `component "dockerfile" does not apply to project type "go-lib"`
Libraries aren't run on their own, so they don't get a Dockerfile. If forge guessed the type wrong (e.g. a Go program without a `cmd/` folder), pass `--type go-cli`.

### `a project name is required`
You ran `forge new` without a name somewhere it can't ask questions (a script or a pipe). Give a name (`forge new my-app`) or force questions with `-i`.

### `make: command not found` (Windows)
Use WSL, or install make (e.g. `winget install GnuWin32.Make`), or run the underlying commands shown in the Makefile directly.

### Python: `No module named pytest`
Run `make install` first; it creates `.venv` and installs the dev tools there.

### `forge --version` shows no commit
Builds from `go install` include the version but not the commit. Release downloads show both, e.g. `forge 0.1.3 (b877ce6)`.

### Can I use forge on an existing project?
Yes: `forge init` fills the current folder (never overwriting), and `forge add` adds single pieces.

### Does forge need internet?
No. The generated projects' tools (npm, pip, cargo, Docker) may download dependencies when you use them.
"""

P["Customizing-Templates"] = """# Customizing Templates

For contributors: how forge's templates work and how to add a project type or component.

## Layout

```text
internal/scaffold/
├── templates/
│   ├── common/            # files every project type gets
│   ├── go-common/         # shared by all Go types (go.mod, CI, golangci config)
│   ├── go-cli/  go-service/  go-lib/  python/  node/  rust/
│   ├── extras/dockerfile/<lang>/   # optional pieces rendered on demand
│   └── licenses/          # MIT, Apache-2.0, BSD-3-Clause
├── types.go               # registry of project types
├── components.go          # registry of `forge add` components
├── detect.go              # language detection for `forge add`
└── scaffold.go            # rendering and safe writing
```

Templates are compiled into the binary with `go:embed`.

## Template rules

- Files ending in `.tmpl` are rendered with Go's `text/template` using **`<%` `%>`** delimiters (so GitHub's `${{ }}` and code braces need no escaping); the suffix is removed. Other files are copied as-is.
- In paths, `__name__` becomes the project name, `__pkg__` the Go package name and `__snake__` the Python/Rust identifier.
- Files starting with `#!` are written executable.
- Available values: `.Name .Pkg .Snake .Module .Description .Author .License .Type .GoVersion .Year`, plus `.IsGo .IsPython .IsNode .IsRust .HasMakefile .NpmName`.
- **Whitespace:** `-%>` also removes the *next line's indentation*. In YAML or indented JSON, put the dash on the left (`<%- if … %>`) instead.

## Add a project type

1. Create `internal/scaffold/templates/<your-type>/` with its files.
2. Register it in `types.go` with its `Lang`, `Sets` (e.g. `{"common", "<your-type>"}`) and `NextSteps`.
3. Add expected files to `TestRenderAllTypes`, and a toolchain run to `cmd/integration_test.go`.

## Add a component for `forge add`

Add an entry to `components.go`, either:

- **Patterns** that select files from a fully rendered project (e.g. `.github/workflows/*`), or
- **Extra**: a template set under `templates/extras/<name>/<lang>/` for pieces not generated by default.

## Testing

```sh
make check         # gofmt, vet, golangci-lint, tests (incl. generating every type and running its toolchain)
make test-docker   # builds and runs every generated Dockerfile (needs Docker)
```

Missing language tools are skipped and logged, never silently passed.

See also: [Releasing](Releasing)
"""

P["Releasing"] = """# Releasing

Releases are automated with [GoReleaser](https://goreleaser.com) and GitHub Actions.

1. Make sure `main` is green: `make check` locally, CI on GitHub.
2. In `CHANGELOG.md`, rename `## [Unreleased]` entries to `## [X.Y.Z] - YYYY-MM-DD`, keep an empty `[Unreleased]` heading, and update the compare links at the bottom.
3. Commit, then tag and push:

```sh
git commit -am "Prepare vX.Y.Z changelog"
git push origin main
git tag -a vX.Y.Z -m "forge vX.Y.Z"
git push origin vX.Y.Z
```

The **Release** workflow builds archives for Linux, macOS and Windows (amd64 and arm64), writes `checksums.txt`, and publishes the GitHub release. `go install github.com/DanielTso/forge@vX.Y.Z` works as soon as the tag is pushed.

## Checking a release

```sh
go install github.com/DanielTso/forge@vX.Y.Z && forge --version     # forge vX.Y.Z
gh release download vX.Y.Z -p 'forge_X.Y.Z_linux_amd64.tar.gz' -p checksums.txt
sha256sum -c --ignore-missing checksums.txt
```

## Dry run

`make snapshot` builds the full set of archives locally into `dist/` without publishing anything.

## Versioning

[Semantic Versioning](https://semver.org): bug fixes bump the patch (`0.1.2 → 0.1.3`); new features bump the minor; breaking changes to commands or flags bump the major (minor while below 1.0).
"""

P["_Sidebar"] = """**[Home](Home)**

**Using forge**
- [Installation](Installation)
- [Quick Start](Quick-Start)
- [Project Types](Project-Types)
- [Adding Components](Adding-Components)
- [Interactive Mode](Interactive-Mode)
- [Command Reference](Command-Reference)

**Guides**
- [Using forge with AI](Using-forge-with-AI)
- [How forge Keeps You Safe](How-forge-Keeps-You-Safe)
- [Troubleshooting and FAQ](Troubleshooting-and-FAQ)

**Contributing**
- [Customizing Templates](Customizing-Templates)
- [Releasing](Releasing)
"""
P["_Footer"] = "forge · [Repository](https://github.com/DanielTso/forge) · [Releases](https://github.com/DanielTso/forge/releases) · [Report an issue](https://github.com/DanielTso/forge/issues/new/choose)\n"

for name, body in P.items():
    with open(os.path.join(W, name + ".md"), "w") as f:
        f.write(body if body.endswith("\n") else body + "\n")
print(len(P), "pages")
