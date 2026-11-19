package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetect(t *testing.T) {
	write := func(t *testing.T, dir, name, content string) {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name  string
		files map[string]string
		want  Detected
	}{
		{"empty", nil, Detected{Type: "generic"}},
		{"go lib", map[string]string{"go.mod": "module example.com/lib\n\ngo 1.22.1\n"},
			Detected{Type: "go-lib", Module: "example.com/lib", GoVersion: "1.22"}},
		{"go cli single cmd", map[string]string{"go.mod": "module m\n\ngo 1.24\n", "cmd/tool/main.go": "package main"},
			Detected{Type: "go-cli", Name: "tool", Module: "m", GoVersion: "1.24"}},
		{"go cli many cmds", map[string]string{"go.mod": "module m\n", "cmd/a/main.go": "x", "cmd/b/main.go": "x"},
			Detected{Type: "go-cli", Module: "m"}},
		{"python", map[string]string{"pyproject.toml": "[build-system]\nname = \"wrong\"\n\n[project]\nname = \"pyname\"\n"},
			Detected{Type: "python", Name: "pyname"}},
		{"node", map[string]string{"package.json": `{"name": "nodename", "version": "1.0.0"}`},
			Detected{Type: "node", Name: "nodename"}},
		{"node bad json", map[string]string{"package.json": `{nope`}, Detected{Type: "node"}},
		{"deno", map[string]string{"deno.json": "{}", "package.json": `{"name":"x"}`}, Detected{Type: "deno"}},
		{"deno scoped name", map[string]string{"deno.json": `{"name": "@me/denotool"}`}, Detected{Type: "deno", Name: "denotool"}},
		{"node scoped name", map[string]string{"package.json": `{"name": "@acme/widget"}`}, Detected{Type: "node", Name: "widget"}},
		{"zig quoted name", map[string]string{"build.zig": "", "build.zig.zon": ".{\n    .name = .@\"zig-tool\",\n}\n"},
			Detected{Type: "zig", Name: "zig-tool"}},
		// Precedence: Python wins over a JS/TS frontend in the same repo (as it did before 0.2.0).
		{"python with ts frontend", map[string]string{"pyproject.toml": "[project]\nname = \"webapp\"\n", "package.json": `{"name":"ui"}`, "tsconfig.json": "{}"},
			Detected{Type: "python", Name: "webapp"}},
		{"python with deno", map[string]string{"pyproject.toml": "[project]\nname = \"webapp\"\n", "deno.json": "{}"},
			Detected{Type: "python", Name: "webapp"}},
		{"node with ts is typescript", map[string]string{"package.json": `{"name":"app"}`, "tsconfig.json": "{}"},
			Detected{Type: "typescript", Name: "app"}},
		{"gleam", map[string]string{"gleam.toml": "name = \"glm\"\nversion = \"1.0.0\"\n\n[dependencies]\nname = \"nope\"\n"},
			Detected{Type: "gleam", Name: "glm"}},
		{"zig", map[string]string{"build.zig": "", "build.zig.zon": ".{\n    .name = .ziggy,\n}\n"}, Detected{Type: "zig", Name: "ziggy"}},
		{"typescript", map[string]string{"package.json": `{"name":"tsname"}`, "tsconfig.json": "{}"},
			Detected{Type: "typescript", Name: "tsname"}},
		{"java", map[string]string{"pom.xml": "<project/>"}, Detected{Type: "java"}},
		{"new bun project (no lockfile yet)", map[string]string{"package.json": `{"name":"b","packageManager":"bun@1.4.4"}`, "tsconfig.json": "{}"},
			Detected{Type: "bun", Name: "b"}},
		{"bun types dependency", map[string]string{"package.json": `{"name":"b","devDependencies":{"@types/bun":"^1"}}`},
			Detected{Type: "bun", Name: "b"}},
		{"bun", map[string]string{"package.json": `{"name":"bunny"}`, "tsconfig.json": "{}", "bun.lock": ""}, Detected{Type: "bun", Name: "bunny"}},
		{"python with bun frontend", map[string]string{"pyproject.toml": "[project]\nname = \"webapp\"\n", "bun.lock": ""},
			Detected{Type: "python", Name: "webapp"}},
		{"rust", map[string]string{"Cargo.toml": "[package]\nname = 'rustname'\n[dependencies]\nname = \"nope\"\n"},
			Detected{Type: "rust", Name: "rustname"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for n, c := range tt.files {
				write(t, dir, n, c)
			}
			got, err := Detect(dir)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Detect = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestEveryComponentRendersForSomeType(t *testing.T) {
	for _, c := range Components() {
		rendered := false
		for _, pt := range Types() {
			o := newOpts(t, pt.ID)
			files, err := RenderComponent(o, c)
			if err != nil {
				continue
			}
			rendered = true
			if len(files) == 0 {
				t.Errorf("%s/%s: no files", c.Name, pt.ID)
			}
		}
		if !rendered {
			t.Errorf("component %q cannot be rendered for any project type (dead component)", c.Name)
		}
	}
}

func TestComponentErrors(t *testing.T) {
	if _, err := LookupComponent("nope"); err == nil {
		t.Error("expected error for unknown component")
	}
	dockerfile, _ := LookupComponent("dockerfile")
	for _, typ := range []string{"go-lib", "gleam", "zig", "generic"} {
		if _, err := RenderComponent(newOpts(t, typ), dockerfile); err == nil {
			t.Errorf("dockerfile should not apply to %s", typ)
		}
	}
	ci, _ := LookupComponent("ci")
	if _, err := RenderComponent(newOpts(t, "generic"), ci); err == nil {
		t.Error("ci should be unavailable for generic")
	}
}

func TestDockerfileIsLanguageSpecific(t *testing.T) {
	want := map[string]string{
		"go-cli": "golang:", "go-service": "golang:", "python": "python:", "node": "node:", "rust": "rust:",
		"typescript": "node:", "deno": "denoland/deno:", "java": "eclipse-temurin:", "bun": "oven/bun:",
	}
	c, _ := LookupComponent("dockerfile")
	for typ, base := range want {
		files, err := RenderComponent(newOpts(t, typ), c)
		if err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		got := paths(files)
		df, ok := got["Dockerfile"]
		if !ok || !strings.Contains(string(df.Content), base) {
			t.Errorf("%s: Dockerfile should use base %q", typ, base)
		}
		if _, ok := got[".dockerignore"]; !ok {
			t.Errorf("%s: missing .dockerignore", typ)
		}
	}
}
