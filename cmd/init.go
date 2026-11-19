package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var g generateFlags
	var name string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a project in the current directory",
		Long: `Scaffold a project into the current directory, using the directory name as the
project name unless --name is given. Existing files are never overwritten unless
--force is passed; if any would be, nothing is written.

Run "forge init" with no flags in a terminal to be prompted for each option.`,
		Example: `  mkdir my-tool && cd my-tool && forge init
  forge init --type generic --name acme-docs`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			if name == "" {
				name = filepath.Base(dir)
			}
			if shouldPrompt(cmd, &g, len(args)) {
				if name, err = promptProject(cmd, &g, name); err != nil {
					return err
				}
			}
			return generate(cmd, name, dir, &g)
		},
	}
	g.register(cmd)
	cmd.Flags().StringVar(&name, "name", "", "project name (default: current directory name)")
	return cmd
}
