package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DanielTso/forge/internal/scaffold"
)

// generateFlags are shared by the "new" and "init" commands.
type generateFlags struct {
	projectType string
	module      string
	description string
	author      string
	license     string
	goVersion   string
	noGit       bool
	yes         bool
	interactive bool
	force       bool
	dryRun      bool
}

func (g *generateFlags) register(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVarP(&g.projectType, "type", "t", "go-cli", "project type ("+strings.Join(scaffold.TypeIDs(), ", ")+")")
	f.StringVarP(&g.module, "module", "m", "", "Go module path (default: the project name)")
	f.StringVarP(&g.description, "description", "d", "", "one-line project description")
	f.StringVarP(&g.author, "author", "a", "", "author / copyright holder (default: git user.name)")
	f.StringVarP(&g.license, "license", "l", "MIT", "license ("+strings.Join(scaffold.Licenses(), ", ")+", none)")
	f.StringVar(&g.goVersion, "go-version", "", "Go version for go.mod (default: version forge was built with)")
	f.BoolVar(&g.noGit, "no-git", false, "do not run git init")
	f.BoolVarP(&g.force, "force", "f", false, "overwrite files that already exist")
	f.BoolVarP(&g.interactive, "interactive", "i", false, "prompt for each option not given as a flag")
	f.BoolVarP(&g.yes, "yes", "y", false, "never prompt; use flags and defaults")
	cmd.MarkFlagsMutuallyExclusive("interactive", "yes")
	f.BoolVar(&g.dryRun, "dry-run", false, "print what would be created without writing anything")

	_ = cmd.RegisterFlagCompletionFunc("type", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return scaffold.TypeIDs(), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("license", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return append(scaffold.Licenses(), "none"), cobra.ShellCompDirectiveNoFileComp
	})
}

// generate renders the project and writes it into dir.
func generate(cmd *cobra.Command, name, dir string, g *generateFlags) error {
	out := cmd.OutOrStdout()

	opts := scaffold.Options{
		Name:        name,
		Module:      g.module,
		Description: g.description,
		Author:      g.author,
		License:     g.license,
		Type:        g.projectType,
		GoVersion:   g.goVersion,
	}
	if opts.Author == "" {
		opts.Author = gitConfig("user.name")
	}
	if err := opts.Normalize(); err != nil {
		return err
	}

	files, err := scaffold.Render(opts)
	if err != nil {
		return err
	}

	if g.dryRun {
		conflicts, err := scaffold.Conflicts(dir, files)
		if err != nil {
			return err
		}
		existing := make(map[string]bool, len(conflicts))
		for _, c := range conflicts {
			existing[c] = true
		}
		fmt.Fprintf(out, "Would create %s project %q in %s:\n", opts.Type, opts.Name, dir)
		for _, f := range files {
			note := ""
			if existing[f.Path] {
				note = "  (exists: would need --force)"
			}
			fmt.Fprintf(out, "  %s%s\n", f.Path, note)
		}
		return nil
	}

	if err := scaffold.Write(dir, files, g.force); err != nil {
		return err
	}

	fmt.Fprintf(out, "Created %s project %q in %s\n", opts.Type, opts.Name, dir)
	printFiles(out, files)

	if !g.noGit {
		switch initGit(dir) {
		case gitInitialized:
			fmt.Fprintln(out, "\nInitialized a git repository (branch: main).")
		case gitSkippedInRepo:
			fmt.Fprintln(out, "\nSkipped git init: directory is already inside a git repository.")
		case gitUnavailable:
			fmt.Fprintln(out, "\nSkipped git init: git not found or failed.")
		}
	}

	printNextSteps(out, opts, dir)
	return nil
}

func printFiles(w io.Writer, files []scaffold.File) {
	for _, f := range files {
		fmt.Fprintf(w, "  %s\n", f.Path)
	}
}

func printNextSteps(w io.Writer, opts scaffold.Options, dir string) {
	fmt.Fprintln(w, "\nNext steps:")
	if rel, err := filepath.Rel(mustGetwd(), dir); err == nil && rel != "." {
		fmt.Fprintf(w, "  cd %s\n", rel)
	}
	if pt, err := scaffold.Lookup(opts.Type); err == nil {
		for _, step := range pt.NextSteps {
			fmt.Fprintf(w, "  %s\n", step)
		}
	}
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

type gitResult int

const (
	gitInitialized gitResult = iota
	gitSkippedInRepo
	gitUnavailable
)

func initGit(dir string) gitResult {
	if _, err := exec.LookPath("git"); err != nil {
		return gitUnavailable
	}
	if err := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree").Run(); err == nil {
		return gitSkippedInRepo
	}
	if err := exec.Command("git", "-C", dir, "init", "-q", "-b", "main").Run(); err != nil {
		return gitUnavailable
	}
	return gitInitialized
}

func gitConfig(key string) string {
	out, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
