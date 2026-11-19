package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newNewCmd() *cobra.Command {
	var g generateFlags
	var output string

	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new project in a new directory",
		Long: `Create a new project named <name> in ./<name> (or under --output).

Run "forge new" with no arguments in a terminal to be prompted for each option;
pass --interactive to be prompted for whatever flags you left out, or --yes to
never be prompted.`,
		Example: `  forge new                       # interactive
  forge new my-tool
  forge new my-api --type go-service --module github.com/me/my-api
  forge new my-lib -t go-lib -l Apache-2.0 --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := ""
			if len(args) == 1 {
				name = args[0]
			}
			if shouldPrompt(cmd, &g, len(args)) {
				var err error
				if name, err = promptProject(cmd, &g, name); err != nil {
					return err
				}
			}
			if name == "" {
				return fmt.Errorf("a project name is required (e.g. \"forge new my-tool\"), or run in a terminal for interactive mode")
			}
			dir, err := filepath.Abs(filepath.Join(output, name))
			if err != nil {
				return err
			}
			return generate(cmd, name, dir, &g)
		},
	}
	g.register(cmd)
	cmd.Flags().StringVarP(&output, "output", "o", ".", "parent directory in which to create the project")
	return cmd
}
