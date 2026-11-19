package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGeneratedGoProjectsBuildAndPass proves the Go templates produce code that is
// gofmt-clean, vets, compiles and passes its own tests.
func TestGeneratedGoProjectsBuildAndPass(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow integration test in -short mode")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		missingTool(t, "go toolchain not on PATH")
	}

	for _, typ := range []string{"go-cli", "go-service", "go-lib"} {
		t.Run(typ, func(t *testing.T) {
			parent := t.TempDir()
			if out, err := run(t, "new", "sample-app", "-t", typ, "-o", parent, "--no-git", "-m", "sample-app"); err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			proj := filepath.Join(parent, "sample-app")

			if out, _ := exec.Command("gofmt", "-l", proj).CombinedOutput(); len(out) > 0 {
				t.Errorf("generated code is not gofmt-clean:\n%s", out)
			}
			for _, args := range [][]string{{"vet", "./..."}, {"test", "./..."}} {
				c := exec.Command(goBin, args...)
				c.Dir = proj
				if out, err := c.CombinedOutput(); err != nil {
					t.Fatalf("go %v failed: %v\n%s", args, err, out)
				}
			}

			// The generated CI runs golangci-lint, so the generated code must pass it.
			if lint, err := exec.LookPath("golangci-lint"); err == nil {
				c := exec.Command(lint, "run", "./...")
				c.Dir = proj
				if out, err := c.CombinedOutput(); err != nil {
					t.Fatalf("golangci-lint failed on generated %s project: %v\n%s", typ, err, out)
				}
			} else {
				t.Log("golangci-lint not installed; skipped lint check")
			}
		})
	}
}

// TestGeneratedNonGoProjectsWork runs each language's own test runner against freshly generated
// projects. Tools that are not installed cause that language's subtest to be skipped.
func TestGeneratedNonGoProjectsWork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow integration test in -short mode")
	}

	gen := func(t *testing.T, typ string) string {
		t.Helper()
		parent := t.TempDir()
		if out, err := run(t, "new", "sample-app", "-t", typ, "-o", parent, "--no-git"); err != nil {
			t.Fatalf("%v\n%s", err, out)
		}
		return filepath.Join(parent, "sample-app")
	}
	sh := func(t *testing.T, dir, name string, args ...string) {
		t.Helper()
		c := exec.Command(name, args...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
		}
	}

	t.Run("node", func(t *testing.T) {
		if _, err := exec.LookPath("node"); err != nil {
			missingTool(t, "node not installed")
		}
		sh(t, gen(t, "node"), "node", "--test") // node:test needs no dependencies
	})
	t.Run("python", func(t *testing.T) {
		if _, err := exec.LookPath("python3"); err != nil {
			missingTool(t, "python3 not installed")
		}
		dir := gen(t, "python")
		sh(t, dir, "python3", "-m", "compileall", "-q", "src", "tests") // syntax check always

		// Mirror the generated CI: ruff, mypy, pytest — each only if installed.
		ran := 0
		for _, step := range [][]string{
			{"ruff", "check", "."},
			{"ruff", "format", "--check", "."},
			{"mypy"},
			{"pytest", "-q"},
		} {
			if _, err := exec.LookPath(step[0]); err != nil {
				t.Logf("%s not installed; skipped", step[0])
				continue
			}
			c := exec.Command(step[0], step[1:]...)
			c.Dir = dir
			c.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(dir, "src"))
			if out, err := c.CombinedOutput(); err != nil {
				t.Fatalf("%v failed: %v\n%s", step, err, out)
			}
			ran++
		}
		if ran == 0 {
			t.Log("no Python tools installed; ran syntax check only")
		}
	})
	t.Run("typescript", func(t *testing.T) {
		if _, err := exec.LookPath("npm"); err != nil {
			missingTool(t, "npm not installed")
		}
		dir := gen(t, "typescript")
		sh(t, dir, "npm", "install", "--no-audit", "--no-fund") // needs network: typescript, eslint
		sh(t, dir, "npm", "run", "lint")
		sh(t, dir, "npm", "run", "typecheck")
		sh(t, dir, "npm", "test")
	})
	t.Run("deno", func(t *testing.T) {
		if _, err := exec.LookPath("deno"); err != nil {
			missingTool(t, "deno not installed")
		}
		sh(t, gen(t, "deno"), "deno", "task", "check") // fmt --check, lint, check, test
	})
	t.Run("gleam", func(t *testing.T) {
		if _, err := exec.LookPath("gleam"); err != nil {
			missingTool(t, "gleam not installed")
		}
		if _, err := exec.LookPath("erl"); err != nil {
			missingTool(t, "erlang (erl) not installed")
		}
		dir := gen(t, "gleam")
		sh(t, dir, "gleam", "format", "--check", "src", "test")
		sh(t, dir, "gleam", "test") // downloads gleam_stdlib and gleeunit from Hex
	})
	t.Run("zig", func(t *testing.T) {
		if _, err := exec.LookPath("zig"); err != nil {
			missingTool(t, "zig not installed")
		}
		dir := gen(t, "zig")
		sh(t, dir, "zig", "fmt", "--check", ".")
		sh(t, dir, "zig", "build", "test") // also checks the build.zig.zon fingerprint
	})
	t.Run("rust", func(t *testing.T) {
		if _, err := exec.LookPath("cargo"); err != nil {
			missingTool(t, "cargo not installed")
		}
		dir := gen(t, "rust")
		sh(t, dir, "cargo", "fmt", "--check")
		sh(t, dir, "cargo", "clippy", "--all-targets", "--", "-D", "warnings")
		sh(t, dir, "cargo", "test")
	})
}

// missingTool skips a test whose toolchain is not installed, or fails it when
// FORGE_REQUIRE_TOOLCHAINS=1 (set in CI), so a green CI run proves every template was exercised.
func missingTool(t *testing.T, why string) {
	t.Helper()
	if os.Getenv("FORGE_REQUIRE_TOOLCHAINS") == "1" {
		t.Fatalf("%s (FORGE_REQUIRE_TOOLCHAINS=1)", why)
	}
	t.Skip(why)
}
