// Package cmd defines forge's command-line interface.
package cmd

import (
	"runtime/debug"

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
		Version:       versionString(),
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.SetVersionTemplate("forge {{.Version}}\n")

	root.AddCommand(newNewCmd(), newInitCmd(), newAddCmd(), newUpgradeCmd(), newListCmd())
	return root
}

// versionString reports the build version. Release builds stamp version/commit via -ldflags;
// `go install module@vX.Y.Z` does not, so fall back to the build info Go embeds in every binary.
func versionString() string {
	v, c := version, commit
	if info, ok := debug.ReadBuildInfo(); ok {
		v, c = fromBuildInfo(v, c, info)
	}
	if c == "none" {
		return v
	}
	return v + " (" + c + ")"
}

func fromBuildInfo(v, c string, info *debug.BuildInfo) (string, string) {
	if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	if c == "none" {
		rev, dirty := "", false
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if len(s.Value) >= 7 {
					rev = s.Value[:7]
				}
			case "vcs.modified":
				dirty = s.Value == "true"
			}
		}
		if rev != "" {
			c = rev
			if dirty {
				c += "-dirty"
			}
		}
	}
	return v, c
}

// Execute runs the CLI.
func Execute() error {
	return NewRootCmd().Execute()
}
