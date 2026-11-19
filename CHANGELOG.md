# Changelog

All notable changes to this project are documented here.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Go Dockerfile accepts `VERSION` / `COMMIT` build args and stamps them into the binary;
  the go-service `make docker` target passes them.

### Added

- `make test-docker`: opt-in test that builds and runs every generated Dockerfile and checks
  the image does not run as root.
- Integration tests run ruff, mypy and pytest against generated Python projects.

## [0.1.0]

### Added

- `forge new`, `forge init`, `forge add` and `forge list` commands.
- Project types: `go-cli`, `go-service`, `go-lib`, `python`, `node`, `rust`, `generic`.
- Components for `forge add`: ci, dockerfile, license, gitignore, editorconfig, changelog,
  contributing, github, docs, makefile, lint.
- Interactive prompts for a bare `forge new` / `forge init`.
- MIT, Apache-2.0 and BSD-3-Clause licenses.
