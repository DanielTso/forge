package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/DanielTso/forge/internal/scaffold"
)

// isTerminal reports whether r is an interactive terminal. It is a variable so tests can override it.
var isTerminal = func(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

// shouldPrompt decides whether to ask questions instead of using flag values/defaults.
//
//   - --yes never prompts;
//   - --interactive always prompts (even without a terminal, e.g. piped input);
//   - otherwise a bare invocation in a terminal prompts: no arguments and none of the
//     flags that set project values (behavioural flags such as --no-git or --force don't count).
func shouldPrompt(cmd *cobra.Command, g *generateFlags, nargs int) bool {
	if g.yes {
		return false
	}
	if g.interactive {
		return true
	}
	if nargs != 0 || !isTerminal(cmd.InOrStdin()) {
		return false
	}
	for _, f := range valueFlags {
		if cmd.Flags().Changed(f) {
			return false
		}
	}
	return true
}

// valueFlags are the flags that supply project values; giving any of them opts out of bare-invocation prompting.
var valueFlags = []string{"name", "type", "module", "description", "author", "license", "go-version", "output"}

type prompter struct {
	r *bufio.Reader
	w io.Writer
}

func newPrompter(cmd *cobra.Command) *prompter {
	return &prompter{r: bufio.NewReader(cmd.InOrStdin()), w: cmd.OutOrStdout()}
}

// ask prompts for a value, applying def on empty input and re-asking until validate passes.
// At end of input it falls back to def, or fails if def is not valid.
func (p *prompter) ask(label, def string, validate func(string) error) (string, error) {
	for {
		if def != "" {
			fmt.Fprintf(p.w, "%s [%s]: ", label, def)
		} else {
			fmt.Fprintf(p.w, "%s: ", label)
		}
		line, err := p.r.ReadString('\n')
		eof := errors.Is(err, io.EOF)
		if err != nil && !eof {
			return "", err
		}
		if eof {
			fmt.Fprintln(p.w)
		}
		val := strings.TrimSpace(line)
		if val == "" {
			val = def
		}
		if validate != nil {
			if verr := validate(val); verr != nil {
				if eof {
					return "", verr
				}
				fmt.Fprintf(p.w, "  %v\n", verr)
				continue
			}
		}
		return val, nil
	}
}

// chooseType shows a numbered list of project types and accepts a number or an ID.
func (p *prompter) chooseType(def string) (string, error) {
	types := scaffold.Types()
	fmt.Fprintln(p.w, "Project type:")
	defIdx := "1"
	for i, t := range types {
		fmt.Fprintf(p.w, "  %d) %-11s %s\n", i+1, t.ID, t.Description)
		if t.ID == def {
			defIdx = strconv.Itoa(i + 1)
		}
	}
	var chosen string
	_, err := p.ask("Choose", defIdx, func(s string) error {
		if n, err := strconv.Atoi(s); err == nil {
			if n >= 1 && n <= len(types) {
				chosen = types[n-1].ID
				return nil
			}
			return fmt.Errorf("enter a number between 1 and %d, or a type name", len(types))
		}
		if _, err := scaffold.Lookup(s); err != nil {
			return err
		}
		chosen = s
		return nil
	})
	return chosen, err
}

// promptProject fills in every value the user did not already supply via flags.
// name is the project name (possibly empty); the (possibly updated) name is returned.
func promptProject(cmd *cobra.Command, g *generateFlags, name string) (string, error) {
	p := newPrompter(cmd)
	changed := cmd.Flags().Changed

	fmt.Fprintln(p.w, "Let's set up your project (press Enter to accept the default).")
	var err error

	if name == "" || cmd.Name() == "init" && !changed("name") {
		if name, err = p.ask("Project name", name, scaffold.ValidateName); err != nil {
			return "", err
		}
	}
	if !changed("type") {
		if g.projectType, err = p.chooseType(g.projectType); err != nil {
			return "", err
		}
	}
	if !changed("description") {
		if g.description, err = p.ask("Description", g.description, nil); err != nil {
			return "", err
		}
	}

	pt, err := scaffold.Lookup(g.projectType)
	if err != nil {
		return "", err
	}
	if pt.Lang == "go" && !changed("module") {
		def := g.module
		if def == "" {
			def = name
		}
		if g.module, err = p.ask("Go module path", def, nil); err != nil {
			return "", err
		}
	}
	if !changed("author") {
		def := g.author
		if def == "" {
			def = gitConfig("user.name")
		}
		if g.author, err = p.ask("Author", def, nil); err != nil {
			return "", err
		}
	}
	if !changed("license") {
		licenses := append(scaffold.Licenses(), "none")
		if g.license, err = p.ask("License ("+strings.Join(licenses, ", ")+")", g.license, func(s string) error {
			for _, l := range licenses {
				if strings.EqualFold(l, s) {
					return nil
				}
			}
			return fmt.Errorf("unknown license %q", s)
		}); err != nil {
			return "", err
		}
		for _, l := range licenses { // normalise case
			if strings.EqualFold(l, g.license) {
				g.license = l
			}
		}
	}
	fmt.Fprintln(p.w)
	return name, nil
}
