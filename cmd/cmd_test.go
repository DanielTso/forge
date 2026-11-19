package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runWithInput(t, "", args...)
}

func runWithInput(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	var out bytes.Buffer
	root.SetIn(strings.NewReader(input))
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestNewCreatesProject(t *testing.T) {
	dir := t.TempDir()
	out, err := run(t, "new", "demo", "-o", dir, "--no-git", "-a", "Tester", "-m", "example.com/demo")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, p := range []string{"go.mod", "LICENSE", "cmd/demo/main.go", "docs/architecture.md"} {
		if _, err := os.Stat(filepath.Join(dir, "demo", p)); err != nil {
			t.Errorf("expected %s: %v", p, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "demo", ".git")); err == nil {
		t.Error("--no-git should not create .git")
	}
}

func TestNewDryRunWritesNothing(t *testing.T) {
	dir := t.TempDir()
	out, err := run(t, "new", "demo", "-o", dir, "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Would create") || !strings.Contains(out, "go.mod") {
		t.Errorf("unexpected output:\n%s", out)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("dry run wrote %d entries", len(entries))
	}
}

func TestNewRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"new", "bad name", "-o", dir},
		{"new", "../evil", "-o", dir},
		{"new", "ok", "-o", dir, "-t", "cobol"},
		{"new", "ok", "-o", dir, "-l", "nope"},
		{"new"},
	} {
		if _, err := run(t, args...); err == nil {
			t.Errorf("args %v: expected error", args)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("invalid input created files: %v", entries)
	}
}

func TestNewConflictRequiresForce(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, "new", "demo", "-o", dir, "--no-git"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "new", "demo", "-o", dir, "--no-git"); err == nil {
		t.Fatal("expected conflict error on second run")
	}
	if _, err := run(t, "new", "demo", "-o", dir, "--no-git", "--force"); err != nil {
		t.Fatalf("--force should succeed: %v", err)
	}
}

func TestInitUsesDirectoryName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-existing")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if _, err := run(t, "init", "-t", "generic", "--no-git"); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "# my-existing") {
		t.Errorf("README should use dir name, got %q", string(b)[:30])
	}
}

func TestList(t *testing.T) {
	out, err := run(t, "list")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"go-cli", "go-service", "go-lib", "python", "node", "rust", "typescript", "deno", "gleam", "zig", "generic", "MIT"} {
		if !strings.Contains(out, want) {
			t.Errorf("list output missing %q", want)
		}
	}
}

func TestNewAllLanguageTypes(t *testing.T) {
	want := map[string]string{
		"python":     "pyproject.toml",
		"node":       "package.json",
		"rust":       "Cargo.toml",
		"typescript": "tsconfig.json",
		"deno":       "deno.json",
		"gleam":      "gleam.toml",
		"zig":        "build.zig.zon",
	}
	for typ, manifest := range want {
		dir := t.TempDir()
		if out, err := run(t, "new", "demo-app", "-t", typ, "-o", dir, "--no-git"); err != nil {
			t.Fatalf("%s: %v\n%s", typ, err, out)
		}
		if _, err := os.Stat(filepath.Join(dir, "demo-app", manifest)); err != nil {
			t.Errorf("%s: missing %s: %v", typ, manifest, err)
		}
	}
}

func TestNewWithoutNameFailsWhenNotInteractive(t *testing.T) {
	out, err := run(t, "new", "-o", t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "project name is required") {
		t.Fatalf("want 'project name is required' error, got err=%v out=%s", err, out)
	}
}

// ---- interactive prompts ----

func TestInteractiveNewAnswersEveryPrompt(t *testing.T) {
	dir := t.TempDir()
	// name, type (3 = go-lib), description, module, author, license
	in := "my-lib\n3\nA tiny lib\nexample.com/me/my-lib\nJane Doe\nApache-2.0\n"
	out, err := runWithInput(t, in, "new", "-i", "-o", dir, "--no-git")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	proj := filepath.Join(dir, "my-lib")
	read := func(p string) string {
		b, err := os.ReadFile(filepath.Join(proj, p))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if !strings.HasPrefix(read("go.mod"), "module example.com/me/my-lib\n") {
		t.Errorf("go.mod: %s", read("go.mod"))
	}
	if _, err := os.Stat(filepath.Join(proj, "mylib.go")); err != nil {
		t.Errorf("go-lib type was not selected: %v", err)
	}
	if !strings.Contains(read("LICENSE"), "Apache License") || !strings.Contains(read("LICENSE"), "Jane Doe") {
		t.Error("license/author prompt answers not applied")
	}
	if !strings.Contains(read("README.md"), "A tiny lib") {
		t.Error("description not applied")
	}
}

func TestInteractiveAcceptsDefaultsAndTypeNames(t *testing.T) {
	dir := t.TempDir()
	// name; type by id; blank description/author/license -> defaults
	in := "defaults-app\npython\n\n\n\n"
	out, err := runWithInput(t, in, "new", "-i", "-o", dir, "--no-git")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(dir, "defaults-app", "pyproject.toml")); err != nil {
		t.Errorf("python type not selected by name: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "defaults-app", "LICENSE")); err != nil {
		t.Errorf("default MIT license missing: %v", err)
	}
}

func TestInteractiveSkipsPromptsForGivenFlags(t *testing.T) {
	dir := t.TempDir()
	// Only the description is asked (name arg + type + module + author + license come from flags).
	out, err := runWithInput(t, "From prompt\n", "new", "x-app", "-i", "-o", dir, "--no-git",
		"-t", "go-cli", "-m", "example.com/x", "-a", "A", "-l", "MIT")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "x-app", "README.md"))
	if !strings.Contains(string(b), "From prompt") {
		t.Error("description from prompt not applied")
	}
	if strings.Contains(out, "Project type:") || strings.Contains(out, "License (") {
		t.Errorf("should not prompt for values supplied by flags:\n%s", out)
	}
}

func TestInteractiveRepromptsOnInvalidInput(t *testing.T) {
	dir := t.TempDir()
	// invalid name, then valid; invalid type number, then valid; invalid license, then valid
	in := "1bad\ngood-name\n99\n4\n\n\nmit\n"
	out, err := runWithInput(t, in, "new", "-i", "-o", dir, "--no-git", "-a", "X")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "invalid project name") || !strings.Contains(out, "between 1 and") {
		t.Errorf("expected validation messages:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, "good-name", "pyproject.toml")); err != nil {
		t.Errorf("project not created from corrected answers: %v", err)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "good-name", "LICENSE"))
	if !strings.HasPrefix(string(b), "MIT License") { // "mit" normalised
		t.Errorf("license case not normalised: %q", string(b)[:20])
	}
}

func TestInteractiveEOFWithInvalidRequiredFails(t *testing.T) {
	if _, err := runWithInput(t, "", "new", "-i", "-o", t.TempDir(), "--no-git"); err == nil {
		t.Fatal("expected error: no project name and input closed")
	}
}

func TestBareInvocationPromptsOnlyOnTerminal(t *testing.T) {
	orig := isTerminal
	defer func() { isTerminal = orig }()
	isTerminal = func(io.Reader) bool { return true }

	// Bare "forge new" on a terminal prompts: name, type, description, module, author, license.
	t.Chdir(t.TempDir())
	out, err := runWithInput(t, "term-app\n1\n\n\n\n\n", "new", "--no-git")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "Project type:") {
		t.Errorf("expected prompts on a terminal:\n%s", out)
	}
	if _, err := os.Stat("term-app/go.mod"); err != nil {
		t.Errorf("project not created: %v", err)
	}

	// A bare "forge init" prompts too, offering the directory name as the default.
	dir := filepath.Join(t.TempDir(), "inited")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	out, err = runWithInput(t, "\n7\n\n\n\n", "init", "--no-git") // name default, type 7 = generic, then defaults
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if !strings.Contains(out, "Project name [inited]") {
		t.Errorf("init should default the name to the directory:\n%s", out)
	}

	// A value flag opts out of bare-invocation prompting; with no name it is an error, not a prompt.
	t.Chdir(t.TempDir())
	if _, err := runWithInput(t, "x\n", "new", "-t", "go-cli", "--no-git"); err == nil {
		t.Error("value flag should disable bare-invocation prompting")
	}

	// Without a terminal, a bare invocation must not prompt (and so errors without a name).
	isTerminal = func(io.Reader) bool { return false }
	if _, err := run(t, "new"); err == nil {
		t.Error("non-terminal bare 'new' should fail, not prompt")
	}
}

func TestYesFlagSuppressesPrompts(t *testing.T) {
	orig := isTerminal
	defer func() { isTerminal = orig }()
	isTerminal = func(io.Reader) bool { return true }

	if _, err := run(t, "new", "-y", "-o", t.TempDir()); err == nil || !strings.Contains(err.Error(), "project name is required") {
		t.Fatalf("--yes must not prompt; got %v", err)
	}
	if _, err := run(t, "new", "x", "-i", "-y"); err == nil {
		t.Error("--interactive and --yes should be mutually exclusive")
	}
}

// ---- forge add ----

func TestAddDetectsLanguageAndWrites(t *testing.T) {
	cases := []struct {
		manifest, content, wantFile string
	}{
		{"go.mod", "module example.com/x\n\ngo 1.24.0\n", ".github/workflows/ci.yml"},
		{"pyproject.toml", "[project]\nname = \"pyx\"\n", ".github/workflows/ci.yml"},
		{"package.json", `{"name":"nodex"}`, ".github/workflows/ci.yml"},
		{"Cargo.toml", "[package]\nname = \"rustx\"\n", ".github/workflows/ci.yml"},
	}
	for _, c := range cases {
		t.Run(c.manifest, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, c.manifest), []byte(c.content), 0o644); err != nil {
				t.Fatal(err)
			}
			if c.manifest == "go.mod" { // a cmd/ directory makes it a go-cli (a bare module is a library)
				if err := os.MkdirAll(filepath.Join(dir, "cmd", "sample"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			t.Chdir(dir)
			out, err := run(t, "add", "ci", "dockerfile", "gitignore", "--name", "sample")
			if err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			for _, f := range []string{c.wantFile, "Dockerfile", ".dockerignore", ".gitignore"} {
				if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
					t.Errorf("missing %s: %v", f, err)
				}
			}
		})
	}
}

func TestAddGoUsesGoModAndCmdName(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/thing\n\ngo 1.23.4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "cmd", "thingd"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if out, err := run(t, "add", "dockerfile", "--name", "thingd"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "Dockerfile"))
	if !strings.Contains(string(b), "golang:1.23 ") || !strings.Contains(string(b), "./cmd/thingd") {
		t.Errorf("Dockerfile should use go.mod version and cmd/ name:\n%s", b)
	}
}

func TestAddRefusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("LICENSE", []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "add", "license", "-a", "Z", "--name", "sample"); err == nil {
		t.Fatal("expected conflict")
	}
	if b, _ := os.ReadFile("LICENSE"); string(b) != "mine" {
		t.Error("LICENSE was modified")
	}
	if _, err := run(t, "add", "license", "-a", "Z", "--name", "sample", "--force"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile("LICENSE"); !strings.Contains(string(b), "MIT License") {
		t.Error("--force should overwrite")
	}
}

func TestAddDryRunAndErrors(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	out, err := run(t, "add", "docs", "--dry-run", "--name", "sample")
	if err != nil || !strings.Contains(out, "docs/architecture.md") {
		t.Fatalf("dry-run: %v\n%s", err, out)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Error("dry run wrote files")
	}

	if _, err := run(t, "add", "nonsense"); err == nil {
		t.Error("unknown component should fail")
	}
	// Dockerfile makes no sense for a language-agnostic project.
	if _, err := run(t, "add", "dockerfile", "--name", "sample"); err == nil {
		t.Error("dockerfile on generic should fail")
	}
	// A bad name in the list must not leave earlier components half-applied.
	if _, err := run(t, "add", "docs", "nonsense", "--name", "sample"); err == nil {
		t.Error("expected error")
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("failed add left files behind: %v", entries)
	}

	out, err = run(t, "add")
	if err != nil || !strings.Contains(out, "dockerfile") {
		t.Errorf("bare add should list components: %v\n%s", err, out)
	}
}

func TestUpgradeCommand(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, ".github", "workflows", "ci.yml")
	if err := os.MkdirAll(filepath.Dir(wf), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "jobs:\n  t:\n    steps:\n      - uses: actions/checkout@v2\n"
	if err := os.WriteFile(wf, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	out, err := run(t, "upgrade", "--dry-run")
	if err != nil || !strings.Contains(out, "Would update 1 action pin") || !strings.Contains(out, "forge add dependabot") {
		t.Fatalf("dry run: %v\n%s", err, out)
	}
	if b, _ := os.ReadFile(wf); string(b) != body {
		t.Fatal("--dry-run wrote the file")
	}

	out, err = run(t, "upgrade")
	if err != nil || !strings.Contains(out, "Updated 1 action pin") {
		t.Fatalf("upgrade: %v\n%s", err, out)
	}
	if b, _ := os.ReadFile(wf); strings.Contains(string(b), "@v2") {
		t.Errorf("pin not updated:\n%s", b)
	}

	out, _ = run(t, "upgrade")
	if !strings.Contains(out, "up to date") {
		t.Errorf("second run should report up to date:\n%s", out)
	}
}
