package scaffold

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// usesRe matches a workflow step's `uses: owner/repo[/path]@vN[.N...]` line. Only major-version
// tags are matched: commit SHAs, branches and local actions (./) are never touched.
var usesRe = regexp.MustCompile(`(?m)^(\s*-?\s*uses:\s*["']?)([A-Za-z0-9_.-]+/[A-Za-z0-9_./-]+)@(v(\d+)(?:\.\d+)*)(["']?)`)

// ActionPin is the version forge's templates pin a GitHub Action to.
type ActionPin struct {
	Action string // e.g. "actions/checkout"
	Ref    string // e.g. "v7"
	Major  int
}

// ActionPins returns, for every GitHub Action used by forge's workflow templates, the newest
// version they pin. These are the versions `forge upgrade` moves a project's workflows to.
func ActionPins() (map[string]ActionPin, error) {
	pins := map[string]ActionPin{}
	err := fs.WalkDir(templateFS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.Contains(p, "/.github/workflows/") {
			return nil
		}
		b, err := fs.ReadFile(templateFS, p)
		if err != nil {
			return err
		}
		for _, m := range usesRe.FindAllStringSubmatch(string(b), -1) {
			major, _ := strconv.Atoi(m[4])
			if cur, ok := pins[m[2]]; !ok || major > cur.Major {
				pins[m[2]] = ActionPin{Action: m[2], Ref: m[3], Major: major}
			}
		}
		return nil
	})
	return pins, err
}

// ActionChange is one `uses:` pin that `forge upgrade` changes.
type ActionChange struct {
	File   string // slash-separated, relative to the project root
	Line   int
	Action string
	From   string
	To     string
}

// UpgradePlan is everything `forge upgrade` would do in a project.
type UpgradePlan struct {
	Changes []ActionChange
	// Updated maps each workflow file that changes to its new content.
	Updated map[string][]byte
	// Workflows is how many workflow files were inspected.
	Workflows int
	// Suggestions name components the project lacks, e.g. "dependabot".
	Suggestions []string
}

// PlanUpgrade inspects the workflows under dir/.github/workflows and works out which action pins
// are older than forge's templates. It reads only; ApplyUpgrade writes.
func PlanUpgrade(dir string) (UpgradePlan, error) {
	plan := UpgradePlan{Updated: map[string][]byte{}}
	pins, err := ActionPins()
	if err != nil {
		return plan, err
	}

	wfDir := filepath.Join(dir, ".github", "workflows")
	entries, err := os.ReadDir(wfDir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return plan, err
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !(strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")) {
			continue
		}
		plan.Workflows++
		rel := path.Join(".github", "workflows", name)
		src, err := os.ReadFile(filepath.Join(wfDir, name))
		if err != nil {
			return plan, err
		}
		out, changes := upgradeWorkflow(rel, src, pins)
		if len(changes) > 0 {
			plan.Changes = append(plan.Changes, changes...)
			plan.Updated[rel] = out
		}
	}

	if _, err := os.Stat(filepath.Join(dir, ".github", "dependabot.yml")); errors.Is(err, fs.ErrNotExist) {
		if _, err2 := os.Stat(filepath.Join(dir, ".github", "dependabot.yaml")); errors.Is(err2, fs.ErrNotExist) {
			plan.Suggestions = append(plan.Suggestions, "dependabot")
		}
	}
	if plan.Workflows == 0 {
		plan.Suggestions = append(plan.Suggestions, "ci")
	}
	sort.Strings(plan.Suggestions)
	return plan, nil
}

// upgradeWorkflow rewrites older major-version pins in one workflow file, leaving every other
// byte as it was. A pin that is already at or above forge's version is left alone.
func upgradeWorkflow(rel string, src []byte, pins map[string]ActionPin) ([]byte, []ActionChange) {
	var changes []ActionChange
	text := string(src)
	out := usesRe.ReplaceAllStringFunc(text, func(match string) string {
		m := usesRe.FindStringSubmatch(match)
		pin, ok := pins[m[2]]
		if !ok {
			return match
		}
		major, _ := strconv.Atoi(m[4])
		if major >= pin.Major {
			return match
		}
		changes = append(changes, ActionChange{File: rel, Action: m[2], From: m[3], To: pin.Ref})
		return m[1] + m[2] + "@" + pin.Ref + m[5]
	})
	// Line numbers, for the report: find each changed pin in order in the original text.
	from := 0
	for i := range changes {
		needle := changes[i].Action + "@" + changes[i].From
		idx := strings.Index(text[from:], needle)
		if idx < 0 {
			continue
		}
		changes[i].Line = strings.Count(text[:from+idx], "\n") + 1
		from += idx + len(needle)
	}
	return []byte(out), changes
}

// ApplyUpgrade writes the plan's updated workflow files, keeping each file's permissions.
func ApplyUpgrade(dir string, plan UpgradePlan) error {
	files := make([]string, 0, len(plan.Updated))
	for f := range plan.Updated {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		target := filepath.Join(dir, filepath.FromSlash(f))
		info, err := os.Stat(target)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, plan.Updated[f], info.Mode().Perm()); err != nil {
			return fmt.Errorf("writing %s: %w", f, err)
		}
	}
	return nil
}
