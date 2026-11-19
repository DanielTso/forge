package scaffold

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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
	case exists(dir, "pyproject.toml"):
		d.Type = "python"
		d.Name = tomlName(filepath.Join(dir, "pyproject.toml"), "project")
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
	defer f.Close()

	in := false
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
	defer f.Close()

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
