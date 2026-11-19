// Package scaffold renders project templates and writes them to disk.
package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"
)

// Options holds everything templates may reference.
type Options struct {
	Name        string // project name, e.g. "my-tool"
	Pkg         string // Go-identifier-safe package name derived from Name (lowercase, no separators)
	Snake       string // snake_case identifier derived from Name (Python package, Rust crate)
	Module      string // Go module path
	Description string
	Author      string
	License     string // license ID, or "none"
	Type        string // project type ID
	GoVersion   string // e.g. "1.27"
	Year        int
}

// File is a rendered file ready to be written.
type File struct {
	Path    string // slash-separated, relative to the project root
	Content []byte
	Mode    fs.FileMode
}

// The methods below are used by templates for conditional content.

// IsGo reports whether the project type is a Go project.
func (o Options) IsGo() bool { return o.Lang() == "go" }

// IsPython reports whether the project type is a Python project.
func (o Options) IsPython() bool { return o.Lang() == "python" }

// IsNode reports whether the project type is a Node.js project.
func (o Options) IsNode() bool { return o.Lang() == "node" }

// IsRust reports whether the project type is a Rust project.
func (o Options) IsRust() bool { return o.Lang() == "rust" }

// HasMakefile reports whether the project type ships a Makefile.
func (o Options) HasMakefile() bool { return o.Lang() != "" }

// NpmName is the project name as a valid (lowercase) npm package name.
func (o Options) NpmName() string { return strings.ToLower(o.Name) }

// Lang returns the programming language of the project type, or "" for language-agnostic types.
func (o Options) Lang() string {
	pt, err := Lookup(o.Type)
	if err != nil {
		return ""
	}
	return pt.Lang
}

var nameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]*$`)

// ValidateName reports whether name is usable as a project name.
func ValidateName(name string) error {
	if !nameRe.MatchString(name) {
		return fmt.Errorf("invalid project name %q: must start with a letter and contain only letters, digits, '.', '_' or '-'", name)
	}
	return nil
}

// SnakeName derives a snake_case identifier from a project name (e.g. "my-tool" -> "my_tool").
func SnakeName(name string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastUnderscore = false
		} else if !lastUnderscore && b.Len() > 0 {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.TrimRight(b.String(), "_")
	if out == "" || unicode.IsDigit(rune(out[0])) {
		out = "p" + out
	}
	return out
}

// PackageName derives a valid, lowercase Go package name from a project name.
func PackageName(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	pkg := b.String()
	if pkg == "" || unicode.IsDigit(rune(pkg[0])) {
		pkg = "p" + pkg
	}
	return pkg
}

// DefaultGoVersion returns the major.minor of the Go toolchain forge was built with.
func DefaultGoVersion() string {
	v := strings.TrimPrefix(runtime.Version(), "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 && parts[0] == "1" {
		return parts[0] + "." + parts[1]
	}
	return "1.24"
}

// Normalize validates opts and fills in derived/default values.
func (o *Options) Normalize() error {
	if err := ValidateName(o.Name); err != nil {
		return err
	}
	if _, err := Lookup(o.Type); err != nil {
		return err
	}
	if o.IsRust() && strings.Contains(o.Name, ".") {
		return fmt.Errorf("invalid project name %q for a Rust crate: '.' is not allowed", o.Name)
	}
	if o.License == "" {
		o.License = "MIT"
	}
	if o.License != "none" {
		if _, err := fs.Stat(templateFS, path.Join("templates", "licenses", o.License+".tmpl")); err != nil {
			return fmt.Errorf("unknown license %q (available: %s, none)", o.License, strings.Join(Licenses(), ", "))
		}
	}
	if o.Pkg == "" {
		o.Pkg = PackageName(o.Name)
	}
	if o.Snake == "" {
		o.Snake = SnakeName(o.Name)
	}
	if o.Module == "" {
		o.Module = o.Name
	}
	if o.Author == "" {
		o.Author = "The " + o.Name + " Authors"
	}
	if o.Description == "" {
		o.Description = "TODO: describe " + o.Name
	}
	if o.GoVersion == "" {
		o.GoVersion = DefaultGoVersion()
	}
	if o.Year == 0 {
		o.Year = time.Now().Year()
	}
	return nil
}

// Render produces every file for the project described by opts, sorted by path.
// opts must already be normalized.
func Render(opts Options) ([]File, error) {
	pt, err := Lookup(opts.Type)
	if err != nil {
		return nil, err
	}

	byPath := map[string]File{}
	// Later template sets override earlier ones, so type-specific files win over common ones.
	for _, set := range pt.Sets {
		files, err := renderSet(set, opts)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			byPath[f.Path] = f
		}
	}

	if opts.License != "none" {
		content, err := renderTemplate("license", mustRead(path.Join("templates", "licenses", opts.License+".tmpl")), opts)
		if err != nil {
			return nil, err
		}
		byPath["LICENSE"] = File{Path: "LICENSE", Content: content, Mode: 0o644}
	}

	out := make([]File, 0, len(byPath))
	for _, f := range byPath {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// RenderExtra renders an optional extra (e.g. "dockerfile") for the language of opts.
// opts must already be normalized. It returns nil if the extra has no variant for the language.
func RenderExtra(opts Options, name string) ([]File, error) {
	set := path.Join("extras", name, opts.Lang())
	if _, err := fs.Stat(templateFS, path.Join("templates", set)); err != nil {
		return nil, nil
	}
	return renderSet(set, opts)
}

func renderSet(set string, opts Options) ([]File, error) {
	root := path.Join("templates", set)
	var files []File
	err := fs.WalkDir(templateFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel := strings.TrimPrefix(p, root+"/")
		raw, err := fs.ReadFile(templateFS, p)
		if err != nil {
			return err
		}

		dest := strings.NewReplacer("__name__", opts.Name, "__pkg__", opts.Pkg, "__snake__", opts.Snake).Replace(rel)
		content := raw
		if strings.HasSuffix(dest, ".tmpl") {
			dest = strings.TrimSuffix(dest, ".tmpl")
			content, err = renderTemplate(rel, raw, opts)
			if err != nil {
				return err
			}
		}

		mode := fs.FileMode(0o644)
		if strings.HasSuffix(dest, ".sh") || bytes.HasPrefix(content, []byte("#!")) {
			mode = 0o755
		}
		files = append(files, File{Path: dest, Content: content, Mode: mode})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("rendering template set %q: %w", set, err)
	}
	return files, nil
}

func renderTemplate(name string, src []byte, opts Options) ([]byte, error) {
	// Custom delimiters so templates can contain GitHub Actions' ${{ }} and Go braces untouched.
	t, err := template.New(name).Delims("<%", "%>").Option("missingkey=error").Parse(string(src))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, opts); err != nil {
		return nil, fmt.Errorf("executing template %s: %w", name, err)
	}
	return buf.Bytes(), nil
}

func mustRead(p string) []byte {
	b, err := fs.ReadFile(templateFS, p)
	if err != nil {
		panic(err) // embedded files are verified by tests
	}
	return b
}

// Conflicts returns the paths in files that already exist under dir.
func Conflicts(dir string, files []File) ([]string, error) {
	var conflicts []string
	for _, f := range files {
		_, err := os.Lstat(filepath.Join(dir, filepath.FromSlash(f.Path)))
		switch {
		case err == nil:
			conflicts = append(conflicts, f.Path)
		case errors.Is(err, fs.ErrNotExist):
		default:
			return nil, err
		}
	}
	return conflicts, nil
}

// Write writes files beneath dir, creating directories as needed. Existing files are
// never overwritten unless force is true; otherwise Write fails before touching anything.
func Write(dir string, files []File, force bool) error {
	if !force {
		conflicts, err := Conflicts(dir, files)
		if err != nil {
			return err
		}
		if len(conflicts) > 0 {
			return &ConflictError{Paths: conflicts}
		}
	}
	for _, f := range files {
		target := filepath.Join(dir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, f.Content, f.Mode); err != nil {
			return err
		}
		// WriteFile doesn't change the mode of an existing file.
		if err := os.Chmod(target, f.Mode); err != nil {
			return err
		}
	}
	return nil
}

// ConflictError is returned by Write when files already exist.
type ConflictError struct{ Paths []string }

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%d file(s) already exist: %s (use --force to overwrite)", len(e.Paths), strings.Join(e.Paths, ", "))
}
