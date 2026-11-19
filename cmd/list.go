package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/DanielTso/forge/internal/scaffold"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available project types and licenses",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "TYPE\tDESCRIPTION")
			for _, t := range scaffold.Types() {
				fmt.Fprintf(w, "%s\t%s\n", t.ID, t.Description)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nLicenses: %s, none\n", strings.Join(scaffold.Licenses(), ", "))
			return nil
		},
	}
}
