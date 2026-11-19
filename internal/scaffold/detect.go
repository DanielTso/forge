package scaffold

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Detected describes what forge could infer about an existing project directory.
type Detected struct {
	Type      string // project type ID
	Name      string // suggested project name ("" if unknown)
	Module    string // Go module path, if any
	GoVersion string // from go.mod, if any
}

// Detect inspects dir and infers the project type and related defaults.
// Directories with no recognised project file are reported as "generic".
func Detect(dir string) (Detected, error) {
	d := Detected{Type: "generic"}

	switch {
	case exists(dir, "go.mod"):
		mod, goVer, err := readGoMod(filepath.Join(dir, "go.mod"))
		if err != nil {
			return d, err
		}
		d.Module, d.GoVersion = mod, goVer
		d.Type = "go-lib"
		if cmds := cmdDirs(dir); len(cmds) > 0 {
			d.Type = "go-cli"
			if len(cmds) == 1 {
				d.Name = cmds[0]
			}
		}
	// Python comes before the JS/TS runtimes: a Python app with a TypeScript or Deno frontend is a
	// Python project (and was detected as one before those types existed).
	case exists(dir, "pyproject.toml"):
		d.Type = "python"
		d.Name = tomlName(filepath.Join(dir, "pyproject.toml"), "project")
	case exists(dir, "deno.json") || exists(dir, "deno.jsonc"):
		d.Type = "deno"
		// deno.jsonc may hold comments, which the JSON reader declines; the name is then left unset.
		for _, f := range []string{"deno.json", "deno.jsonc"} {
			if n := packageJSONName(filepath.Join(dir, f)); n != "" {
				d.Name = n
				break
			}
		}
	case exists(dir, "bun.lock") || exists(dir, "bun.lockb") || exists(dir, "bunfig.toml") ||
		usesBun(filepath.Join(dir, "package.json")):
		d.Type = "bun"
		d.Name = packageJSONName(filepath.Join(dir, "package.json"))
	case exists(dir, "pom.xml"):
		d.Type = "java"
	case exists(dir, "gleam.toml"):
		d.Type = "gleam"
		d.Name = tomlName(filepath.Join(dir, "gleam.toml"), "")
	case exists(dir, "build.zig"):
		d.Type = "zig"
		d.Name = zonName(filepath.Join(dir, "build.zig.zon"))
	case exists(dir, "package.json") && exists(dir, "tsconfig.json"):
		d.Type = "typescript"
		d.Name = packageJSONName(filepath.Join(dir, "package.json"))
	case exists(dir, "package.json"):
		d.Type = "node"
		d.Name = packageJSONName(filepath.Join(dir, "package.json"))
	case exists(dir, "Cargo.toml"):
		d.Type = "rust"
		d.Name = tomlName(filepath.Join(dir, "Cargo.toml"), "package")
	}
	return d, nil
}

// tomlName returns the value of `name = "..."` inside [section], or "" if not found.
// It is a deliberately tiny parser: enough for the well-known manifest layouts.
func tomlName(file, section string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }() // read-only: nothing to flush

	in := section == "" // "" means the top-level table, before any [section]
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "[") {
			in = line == "["+section+"]"
			continue
		}
		if !in {
			continue
		}
		if key, val, ok := strings.Cut(line, "="); ok && strings.TrimSpace(key) == "name" {
			return strings.Trim(strings.TrimSpace(val), `"'`)
		}
	}
	return ""
}

// zonName returns the package name from a build.zig.zon (`.name = .foo,`), or "".
func zonName(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	m := zonNameRe.FindSubmatch(b)
	if m == nil {
		return ""
	}
	if len(m[1]) > 0 { // quoted form: .name = .@"my-lib"
		return string(m[1])
	}
	return string(m[2])
}

var zonNameRe = regexp.MustCompile(`\.name\s*=\s*\.(?:@"([^"]+)"|([A-Za-z_][A-Za-z0-9_]*))`)

// usesBun reports whether a package.json is a Bun project's: `"packageManager": "bun@…"` or a
// dependency on @types/bun. A new Bun project has no bun.lock until `bun install` has run.
func usesBun(file string) bool {
	b, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	var pkg struct {
		PackageManager  string            `json:"packageManager"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(b, &pkg) != nil {
		return false
	}
	_, dep := pkg.Dependencies["@types/bun"]
	_, devDep := pkg.DevDependencies["@types/bun"]
	return strings.HasPrefix(pkg.PackageManager, "bun@") || dep || devDep
}

func packageJSONName(file string) string {
	b, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	var pkg struct {
		Name string `json:"name"`
	}
	if json.Unmarshal(b, &pkg) != nil {
		return ""
	}
	// A scoped name ("@me/tool") is not a valid project name; its last part is.
	if i := strings.LastIndex(pkg.Name, "/"); i >= 0 {
		return pkg.Name[i+1:]
	}
	return pkg.Name
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// cmdDirs lists the subdirectories of dir/cmd.
func cmdDirs(dir string) []string {
	entries, err := os.ReadDir(filepath.Join(dir, "cmd"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out
}

func readGoMod(file string) (module, goVersion string, err error) {
	f, err := os.Open(file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", "", nil
		}
		return "", "", err
	}
	defer func() { _ = f.Close() }() // read-only: nothing to flush

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "module":
			module = strings.Trim(fields[1], `"`)
		case "go":
			// "1.27.4" -> "1.27"
			parts := strings.Split(fields[1], ".")
			if len(parts) >= 2 {
				goVersion = parts[0] + "." + parts[1]
			} else {
				goVersion = fields[1]
			}
		}
	}
	return module, goVersion, sc.Err()
}
