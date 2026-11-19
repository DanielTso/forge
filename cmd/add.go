package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/DanielTso/forge/internal/scaffold"
)

func newAddCmd() *cobra.Command {
	var (
		projectType string
		name        string
		module      string
		author      string
		license     string
		force       bool
		dryRun      bool
	)

	cmd := &cobra.Command{
		Use:   "add [component...]",
		Short: "Add standard pieces (CI, Dockerfile, ...) to an existing project",
		Long: `Add individual components to the project in the current directory. The project
type is detected from its files (go.mod, pyproject.toml, package.json, Cargo.toml)
unless --type is given, so each component is tailored to the language.

Existing files are never overwritten unless --force is passed; if any would be,
nothing is written. Run "forge add" with no arguments to list the components.`,
		Example: `  forge add ci
  forge add dockerfile docs --dry-run
  forge add license --license Apache-2.0 --author "Jane Doe"`,
		ValidArgs: scaffold.ComponentNames(),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if len(args) == 0 {
				return printComponents(cmd)
			}

			// Resolve all components first so a typo doesn't leave a half-applied add.
			comps := make([]scaffold.Component, 0, len(args))
			for _, a := range args {
				c, err := scaffold.LookupComponent(a)
				if err != nil {
					return err
				}
				comps = append(comps, c)
			}

			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			det, err := scaffold.Detect(dir)
			if err != nil {
				return err
			}

			opts := scaffold.Options{
				Type:      firstNonEmpty(projectType, det.Type),
				Name:      firstNonEmpty(name, det.Name, filepath.Base(dir)),
				Module:    firstNonEmpty(module, det.Module),
				GoVersion: det.GoVersion,
				Author:    firstNonEmpty(author, gitConfig("user.name")),
				License:   license,
			}
			if err := opts.Normalize(); err != nil {
				return fmt.Errorf("%w (use --name to set the project name)", err)
			}

			var files []scaffold.File
			seen := map[string]bool{}
			for _, c := range comps {
				fs, err := scaffold.RenderComponent(opts, c)
				if err != nil {
					return err
				}
				for _, f := range fs {
					if !seen[f.Path] {
						seen[f.Path] = true
						files = append(files, f)
					}
				}
			}

			if dryRun {
				conflicts, err := scaffold.Conflicts(dir, files)
				if err != nil {
					return err
				}
				existing := map[string]bool{}
				for _, c := range conflicts {
					existing[c] = true
				}
				fmt.Fprintf(out, "Would add %s to this %s project:\n", strings.Join(args, ", "), opts.Type)
				for _, f := range files {
					note := ""
					if existing[f.Path] {
						note = "  (exists: would need --force)"
					}
					fmt.Fprintf(out, "  %s%s\n", f.Path, note)
				}
				return nil
			}

			if err := scaffold.Write(dir, files, force); err != nil {
				return err
			}
			fmt.Fprintf(out, "Added %s to this %s project:\n", strings.Join(args, ", "), opts.Type)
			printFiles(out, files)
			return nil
		},
	}

	f := cmd.Flags()
	f.StringVarP(&projectType, "type", "t", "", "project type (default: detected from the directory)")
	f.StringVar(&name, "name", "", "project name (default: detected, else the directory name)")
	f.StringVarP(&module, "module", "m", "", "Go module path (default: read from go.mod)")
	f.StringVarP(&author, "author", "a", "", "author / copyright holder (default: git user.name)")
	f.StringVarP(&license, "license", "l", "MIT", "license for the license component")
	f.BoolVarP(&force, "force", "f", false, "overwrite files that already exist")
	f.BoolVar(&dryRun, "dry-run", false, "print what would be added without writing anything")
	_ = cmd.RegisterFlagCompletionFunc("type", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return scaffold.TypeIDs(), cobra.ShellCompDirectiveNoFileComp
	})
	return cmd
}

func printComponents(cmd *cobra.Command) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "COMPONENT\tDESCRIPTION")
	for _, c := range scaffold.Components() {
		fmt.Fprintf(w, "%s\t%s\n", c.Name, c.Description)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "\nUsage: forge add <component>...")
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
