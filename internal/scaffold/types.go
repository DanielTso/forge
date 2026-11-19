package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed all:templates
var templateFS embed.FS

// ProjectType describes a kind of project forge can generate.
type ProjectType struct {
	ID          string
	Description string
	// Lang is the programming language ("go", "python", "node", "rust"), or "" if language-agnostic.
	Lang string
	// Sets are the template directories (under templates/) applied in order;
	// later sets override files of the same path from earlier ones.
	Sets []string
	// NextSteps are shell commands suggested to the user after generation.
	NextSteps []string
}

var projectTypes = []ProjectType{
	{
		ID: "go-cli", Lang: "go",
		Description: "Go command-line application (cmd/, internal/, Makefile, CI)",
		Sets:        []string{"common", "go-common", "go-cli"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID: "go-service", Lang: "go",
		Description: "Go HTTP service with health check, graceful shutdown, Dockerfile",
		Sets:        []string{"common", "go-common", "extras/dockerfile/go", "go-service"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID: "go-lib", Lang: "go",
		Description: "Go library with examples and CI",
		Sets:        []string{"common", "go-common", "go-lib"},
		NextSteps:   []string{"make test"},
	},
	{
		ID: "python", Lang: "python",
		Description: "Python package (src layout, pyproject.toml, pytest, ruff, mypy, CI)",
		Sets:        []string{"common", "python"},
		NextSteps:   []string{"make install", "make test", "make run"},
	},
	{
		ID: "node", Lang: "node",
		Description: "Node.js CLI/library (ESM, node:test, ESLint, CI)",
		Sets:        []string{"common", "node"},
		NextSteps:   []string{"make install", "make test", "make run"},
	},
	{
		ID: "rust", Lang: "rust",
		Description: "Rust binary + library crate (cargo, rustfmt, clippy, CI)",
		Sets:        []string{"common", "rust"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID: "typescript", Lang: "typescript",
		Description: "TypeScript on Node (strict tsconfig, node:test, typescript-eslint, CI)",
		Sets:        []string{"common", "typescript"},
		NextSteps:   []string{"make install", "make test", "make run"},
	},
	{
		ID: "deno", Lang: "deno",
		Description: "Deno + TypeScript (built-in fmt, lint, check and test; no config)",
		Sets:        []string{"common", "deno"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID: "gleam", Lang: "gleam",
		Description: "Gleam on the Erlang VM (gleeunit tests, formatter, CI)",
		Sets:        []string{"common", "gleam"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID: "zig", Lang: "zig",
		Description: "Zig 0.17 executable + library module (zig build, tests, fmt, CI)",
		Sets:        []string{"common", "zig"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID: "bun", Lang: "bun",
		Description: "Bun + TypeScript (bun:test, built-in type-check, single-binary build)",
		Sets:        []string{"common", "bun"},
		NextSteps:   []string{"make install", "make test", "make run"},
	},
	{
		ID: "java", Lang: "java",
		Description: "Java 25 with Maven (JUnit 6, runnable jar, CI)",
		Sets:        []string{"common", "java"},
		NextSteps:   []string{"make test", "make run"},
	},
	{
		ID:          "generic",
		Description: "Language-agnostic project: README, LICENSE, docs/, contributing guides",
		Sets:        []string{"common"},
		NextSteps:   []string{"edit README.md and docs/"},
	},
}

// Types returns all known project types.
func Types() []ProjectType {
	out := make([]ProjectType, len(projectTypes))
	copy(out, projectTypes)
	return out
}

// TypeIDs returns the IDs of all project types.
func TypeIDs() []string {
	ids := make([]string, 0, len(projectTypes))
	for _, t := range projectTypes {
		ids = append(ids, t.ID)
	}
	return ids
}

// Lookup finds a project type by ID.
func Lookup(id string) (ProjectType, error) {
	for _, t := range projectTypes {
		if t.ID == id {
			return t, nil
		}
	}
	return ProjectType{}, fmt.Errorf("unknown project type %q (available: %s)", id, strings.Join(TypeIDs(), ", "))
}

// Licenses returns the IDs of the bundled licenses.
func Licenses() []string {
	entries, err := fs.ReadDir(templateFS, path.Join("templates", "licenses"))
	if err != nil {
		return nil
	}
	var ids []string
	for _, e := range entries {
		ids = append(ids, strings.TrimSuffix(e.Name(), ".tmpl"))
	}
	sort.Strings(ids)
	return ids
}
