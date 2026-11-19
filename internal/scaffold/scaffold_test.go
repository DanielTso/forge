package scaffold

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newOpts(t *testing.T, typ string) Options {
	t.Helper()
	o := Options{Name: "my-proj", Module: "example.com/me/my-proj", Type: typ, Author: "Jane Doe", Year: 2026}
	if err := o.Normalize(); err != nil {
		t.Fatalf("Normalize: %v", err)
	}
	return o
}

func paths(files []File) map[string]File {
	m := make(map[string]File, len(files))
	for _, f := range files {
		m[f.Path] = f
	}
	return m
}

func TestValidateName(t *testing.T) {
	good := []string{"a", "my-tool", "my_tool", "tool2", "a.b", "MyTool"}
	bad := []string{"", "1abc", "-x", "a b", "a/b", "../x", ".", "..", "a\\b"}
	for _, n := range good {
		if err := ValidateName(n); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", n, err)
		}
	}
	for _, n := range bad {
		if err := ValidateName(n); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", n)
		}
	}
}

func TestPackageName(t *testing.T) {
	tests := map[string]string{
		"my-tool":  "mytool",
		"My_Tool":  "mytool",
		"a.b-c":    "abc",
		"tool2":    "tool2",
		"---":      "p",
		"9lives":   "p9lives",
		"Foo-Bar9": "foobar9",
	}
	for in, want := range tests {
		if got := PackageName(in); got != want {
			t.Errorf("PackageName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSnakeName(t *testing.T) {
	tests := map[string]string{
		"my-tool":   "my_tool",
		"My.Tool":   "my_tool",
		"a--b":      "a_b",
		"tool2":     "tool2",
		"foo_bar":   "foo_bar",
		"trailing-": "trailing",
	}
	for in, want := range tests {
		if got := SnakeName(in); got != want {
			t.Errorf("SnakeName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRustRejectsDotsInName(t *testing.T) {
	if err := (&Options{Name: "a.b", Type: "rust"}).Normalize(); err == nil {
		t.Error("expected error for '.' in rust crate name")
	}
	if err := (&Options{Name: "a.b", Type: "python"}).Normalize(); err != nil {
		t.Errorf("python should accept dots: %v", err)
	}
}

func TestLanguageSpecificGitignore(t *testing.T) {
	want := map[string]string{"python": ".venv/", "node": "node_modules/", "rust": "/target/", "go-cli": "/bin/"}
	for typ, entry := range want {
		files, _ := Render(newOpts(t, typ))
		gi := string(paths(files)[".gitignore"].Content)
		if !strings.Contains(gi, entry) {
			t.Errorf("%s .gitignore missing %q", typ, entry)
		}
		for other, otherEntry := range want {
			if other != typ && otherEntry != entry && strings.Contains(gi, otherEntry) {
				t.Errorf("%s .gitignore wrongly contains %s entry %q", typ, other, otherEntry)
			}
		}
	}
}

func TestGitignoreIgnoresWindowsZoneIdentifierFiles(t *testing.T) {
	// WSL creates "<file>:Zone.Identifier" next to downloaded files; the pattern must contain the colon.
	files, _ := Render(newOpts(t, "generic"))
	if !strings.Contains(string(paths(files)[".gitignore"].Content), "*:Zone.Identifier") {
		t.Error(".gitignore should ignore *:Zone.Identifier")
	}
}

func TestDependabotMatchesLanguage(t *testing.T) {
	want := map[string]string{
		"go-cli": "gomod", "go-service": "gomod", "go-lib": "gomod",
		"python": "pip", "node": "npm", "rust": "cargo", "generic": "",
	}
	ecosystems := []string{"gomod", "pip", "npm", "cargo"}
	for typ, eco := range want {
		files, err := Render(newOpts(t, typ))
		if err != nil {
			t.Fatal(err)
		}
		cfg := string(paths(files)[".github/dependabot.yml"].Content)
		// Indentation is structure in YAML: each update entry must be a 2-space list item.
		if !strings.Contains(cfg, "\nupdates:\n  - package-ecosystem: ") ||
			!strings.Contains(cfg, "\n  - package-ecosystem: github-actions\n    directory: /\n") {
			t.Errorf("%s: dependabot.yml is mis-indented:\n%s", typ, cfg)
		}
		if !strings.Contains(cfg, "package-ecosystem: github-actions") {
			t.Errorf("%s: dependabot.yml should always update github-actions", typ)
		}
		for _, e := range ecosystems {
			has := strings.Contains(cfg, "package-ecosystem: "+e+"\n")
			if has != (e == eco) {
				t.Errorf("%s: ecosystem %s present=%v, want %v\n%s", typ, e, has, e == eco, cfg)
			}
		}
	}
}

func TestPackageJSONIsValidAndIndented(t *testing.T) {
	for _, lic := range []string{"MIT", "none"} {
		o := newOpts(t, "node")
		o.License = lic
		files, _ := Render(o)
		raw := paths(files)["package.json"].Content
		var pkg map[string]any
		if err := json.Unmarshal(raw, &pkg); err != nil {
			t.Fatalf("license %s: invalid JSON: %v\n%s", lic, err, raw)
		}
		if _, has := pkg["license"]; has != (lic != "none") {
			t.Errorf("license %s: license field present=%v", lic, has)
		}
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(line, "\"") {
				t.Errorf("license %s: top-level key not indented: %q", lic, line)
			}
		}
	}
}

func TestShebangFilesAreExecutable(t *testing.T) {
	files, _ := Render(newOpts(t, "node"))
	if m := paths(files)["bin/my-proj.js"].Mode; m != 0o755 {
		t.Errorf("bin script mode = %v, want 0755", m)
	}
}

func TestNormalizeErrors(t *testing.T) {
	if err := (&Options{Name: "x", Type: "nope"}).Normalize(); err == nil {
		t.Error("expected error for unknown type")
	}
	if err := (&Options{Name: "x", Type: "generic", License: "WTFPL"}).Normalize(); err == nil {
		t.Error("expected error for unknown license")
	}
	if err := (&Options{Name: "bad name", Type: "generic"}).Normalize(); err == nil {
		t.Error("expected error for invalid name")
	}
}

func TestRenderAllTypes(t *testing.T) {
	want := map[string][]string{
		"generic":    {"README.md", "LICENSE", ".gitignore", ".editorconfig", "CHANGELOG.md", "CONTRIBUTING.md", "docs/README.md", "docs/architecture.md"},
		"go-cli":     {"go.mod", "Makefile", "cmd/my-proj/main.go", "internal/app/app.go", "internal/app/app_test.go", ".github/workflows/ci.yml", ".golangci.yml"},
		"go-service": {"go.mod", "Makefile", "Dockerfile", ".dockerignore", "cmd/my-proj/main.go", "internal/server/server.go", "internal/server/server_test.go"},
		"go-lib":     {"go.mod", "myproj.go", "myproj_test.go", "example_test.go", "examples/basic/main.go"},
		"python":     {"pyproject.toml", "Makefile", "src/my_proj/__init__.py", "src/my_proj/cli.py", "src/my_proj/__main__.py", "tests/test_cli.py", ".github/workflows/ci.yml"},
		"node":       {"package.json", "Makefile", "eslint.config.js", "src/index.js", "bin/my-proj.js", "test/index.test.js", ".github/workflows/ci.yml"},
		"rust":       {"Cargo.toml", "Makefile", "rustfmt.toml", "src/lib.rs", "src/main.rs", "tests/cli.rs", ".github/workflows/ci.yml"},
	}
	for _, pt := range Types() {
		t.Run(pt.ID, func(t *testing.T) {
			files, err := Render(newOpts(t, pt.ID))
			if err != nil {
				t.Fatal(err)
			}
			got := paths(files)
			for _, p := range append(want[pt.ID], want["generic"]...) {
				if _, ok := got[p]; !ok {
					t.Errorf("missing file %s", p)
				}
			}
			for _, f := range files {
				if strings.HasSuffix(f.Path, ".tmpl") {
					t.Errorf("%s still has .tmpl suffix", f.Path)
				}
				if strings.Contains(f.Path, "__name__") || strings.Contains(f.Path, "__pkg__") || strings.Contains(f.Path, "__snake__") {
					t.Errorf("%s has unreplaced placeholder", f.Path)
				}
				if strings.Contains(string(f.Content), "<%") || strings.Contains(string(f.Content), "%>") {
					t.Errorf("%s has unrendered template delimiters", f.Path)
				}
				if len(f.Content) == 0 {
					t.Errorf("%s is empty", f.Path)
				}
			}
		})
	}
}

func TestTypeSpecificOverridesCommon(t *testing.T) {
	files, _ := Render(newOpts(t, "go-cli"))
	arch := string(paths(files)["docs/architecture.md"].Content)
	if !strings.Contains(arch, "cmd/my-proj/") {
		t.Errorf("go-cli architecture.md should be the type-specific one, got:\n%s", arch)
	}
	files, _ = Render(newOpts(t, "generic"))
	if strings.Contains(string(paths(files)["docs/architecture.md"].Content), "cmd/") {
		t.Error("generic architecture.md should not mention cmd/")
	}
}

func TestGoSpecificContent(t *testing.T) {
	gi := func(typ string) string {
		files, _ := Render(newOpts(t, typ))
		return string(paths(files)[".gitignore"].Content)
	}
	if !strings.Contains(gi("go-cli"), "/bin/") {
		t.Error("go .gitignore should ignore /bin/")
	}
	if strings.Contains(gi("generic"), "/bin/") {
		t.Error("generic .gitignore should not contain Go entries")
	}
	files, _ := Render(newOpts(t, "go-cli"))
	if mod := string(paths(files)["go.mod"].Content); !strings.HasPrefix(mod, "module example.com/me/my-proj\n") {
		t.Errorf("unexpected go.mod:\n%s", mod)
	}
}

func TestLicenses(t *testing.T) {
	for _, l := range Licenses() {
		o := newOpts(t, "generic")
		o.License = l
		files, err := Render(o)
		if err != nil {
			t.Fatalf("%s: %v", l, err)
		}
		lic := string(paths(files)["LICENSE"].Content)
		if !strings.Contains(lic, "Jane Doe") || !strings.Contains(lic, "2026") {
			t.Errorf("%s license missing author/year", l)
		}
	}
	o := newOpts(t, "generic")
	o.License = "none"
	files, _ := Render(o)
	if _, ok := paths(files)["LICENSE"]; ok {
		t.Error("license none should not create LICENSE")
	}
	if strings.Contains(string(paths(files)["README.md"].Content), "## License") {
		t.Error("README should omit license section for none")
	}
}

func TestWriteAndConflicts(t *testing.T) {
	dir := t.TempDir()
	files, err := Render(newOpts(t, "go-cli"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Write(dir, files, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "my-proj", "main.go")); err != nil {
		t.Error(err)
	}

	// Second write must refuse and leave existing content alone.
	marker := filepath.Join(dir, "README.md")
	if err := os.WriteFile(marker, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = Write(dir, files, false)
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("want ConflictError, got %v", err)
	}
	if len(ce.Paths) != len(files) {
		t.Errorf("want %d conflicts, got %d", len(files), len(ce.Paths))
	}
	if b, _ := os.ReadFile(marker); string(b) != "mine" {
		t.Error("README.md was modified despite conflict")
	}

	// --force overwrites.
	if err := Write(dir, files, true); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(marker); string(b) == "mine" {
		t.Error("README.md not overwritten with force")
	}
}

func TestWriteIsAtomicOnConflict(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	files, _ := Render(newOpts(t, "go-cli"))
	if err := Write(dir, files, false); err == nil {
		t.Fatal("expected conflict")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("nothing else should have been written, found %d entries", len(entries))
	}
}
