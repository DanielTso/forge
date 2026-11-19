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
		t.Skip("go toolchain not on PATH")
	}

	for _, typ := range []string{"go-cli", "go-service", "go-lib"} {
		t.Run(typ, func(t *testing.T) {
			parent := t.TempDir()
			if out, err := run(t, "new", "sample-app", "-t", typ, "-o", parent, "--no-git", "-m", "example.com/sample-app"); err != nil {
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
			t.Skip("node not installed")
		}
		sh(t, gen(t, "node"), "node", "--test") // node:test needs no dependencies
	})
	t.Run("python", func(t *testing.T) {
		if _, err := exec.LookPath("python3"); err != nil {
			t.Skip("python3 not installed")
		}
		dir := gen(t, "python")
		sh(t, dir, "python3", "-m", "compileall", "-q", "src", "tests") // syntax check always
		if exec.Command("python3", "-m", "pytest", "--version").Run() != nil {
			t.Log("pytest not installed; ran syntax check only")
			return
		}
		c := exec.Command("python3", "-m", "pytest", "-q")
		c.Dir = dir
		c.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(dir, "src"))
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("pytest failed: %v\n%s", err, out)
		}
	})
	t.Run("rust", func(t *testing.T) {
		if _, err := exec.LookPath("cargo"); err != nil {
			t.Skip("cargo not installed")
		}
		dir := gen(t, "rust")
		sh(t, dir, "cargo", "fmt", "--check")
		sh(t, dir, "cargo", "clippy", "--all-targets", "--", "-D", "warnings")
		sh(t, dir, "cargo", "test")
	})
}
