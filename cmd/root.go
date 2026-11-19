// Package cmd defines forge's command-line interface.
package cmd

import (
	"github.com/spf13/cobra"
)

// Build information, overridden at build time via -ldflags (see Makefile).
var (
	version = "dev"
	commit  = "none"
)

// NewRootCmd builds the root command and its subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "forge",
		Short: "Scaffold projects using industry-standard structure",
		Long: `forge generates new projects with a sensible, conventional layout:
README, LICENSE, .gitignore, .editorconfig, CHANGELOG, contributing guide,
docs/, CI workflows, a Makefile and starter code with tests.

Run "forge list" to see the available project types.`,
		Version:       version + " (" + commit + ")",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.SetVersionTemplate("forge {{.Version}}\n")

	root.AddCommand(newNewCmd(), newInitCmd(), newAddCmd(), newListCmd())
	return root
}

// Execute runs the CLI.
func Execute() error {
	return NewRootCmd().Execute()
}
