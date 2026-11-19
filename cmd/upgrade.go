package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DanielTso/forge/internal/scaffold"
)

func newUpgradeCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Update an existing project's GitHub Actions (and safe tool versions) to forge's current ones",
		Long: `Update the GitHub Actions pinned in the current project's .github/workflows/ to the
versions forge's templates use today (for example actions/checkout@v4 -> @v7).

Only the version on each "uses:" line changes; the rest of every workflow stays as you
wrote it. A pin that is already as new or newer is left alone, and pins to a commit SHA
or a branch are never touched. Tool versions set for an action are updated when the
update is compatible (gleam-version within 1.x); otherwise (Zig, Erlang/OTP, Java)
upgrade only reports the newer version, since moving to it can need code changes.
It also points out missing pieces you can add with
"forge add" (such as Dependabot, which keeps these versions current from then on).`,
		Example: `  forge upgrade --dry-run   # show what would change
  forge upgrade`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			plan, err := scaffold.PlanUpgrade(dir)
			if err != nil {
				return err
			}

			switch {
			case plan.Workflows == 0:
				fmt.Fprintln(out, "No GitHub Actions workflows found in .github/workflows/.")
			case len(plan.Changes) == 0:
				fmt.Fprintf(out, "Everything in %d workflow file(s) is up to date.\n", plan.Workflows)
			default:
				verb := "Updated"
				if dryRun {
					verb = "Would update"
				}
				fmt.Fprintf(out, "%s %d version(s):\n", verb, len(plan.Changes))
				for _, c := range plan.Changes {
					fmt.Fprintf(out, "  %s:%d  %s  %s -> %s\n", c.File, c.Line, c.Action, c.From, c.To)
				}
				if !dryRun {
					if err := scaffold.ApplyUpgrade(dir, plan); err != nil {
						return err
					}
				}
			}

			if len(plan.Advisories) > 0 {
				fmt.Fprintln(out, "\nNewer versions available (not changed):")
				for _, a := range plan.Advisories {
					fmt.Fprintf(out, "  %s\n", a)
				}
			}
			if len(plan.Suggestions) > 0 {
				fmt.Fprintf(out, "\nSuggested: forge add %s\n", strings.Join(plan.Suggestions, " "))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would change without writing anything")
	return cmd
}
