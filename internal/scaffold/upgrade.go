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
	// Advisories describe newer tool versions that upgrade reports but does not apply.
	Advisories []string
}

// PlanUpgrade inspects the workflows under dir/.github/workflows and works out which action pins
// are older than forge's templates. It reads only; ApplyUpgrade writes.
func PlanUpgrade(dir string) (UpgradePlan, error) {
	plan := UpgradePlan{Updated: map[string][]byte{}}
	pins, err := ActionPins()
	if err != nil {
		return plan, err
	}
	tools, err := ToolPins()
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
		if e.IsDir() || (!strings.HasSuffix(name, ".yml") && !strings.HasSuffix(name, ".yaml")) {
			continue
		}
		plan.Workflows++
		rel := path.Join(".github", "workflows", name)
		src, err := os.ReadFile(filepath.Join(wfDir, name))
		if err != nil {
			return plan, err
		}
		out, changes, advice := upgradeWorkflow(rel, src, pins, tools)
		plan.Advisories = append(plan.Advisories, advice...)
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

// toolPolicy says what `forge upgrade` does with a tool version set in a step's `with:` block.
type toolPolicy int

const (
	toolBump   toolPolicy = iota // update it when the major version is the same (a compatible update)
	toolAdvise                   // only report a newer version; updating may need code changes
)

// toolKeys lists, per action, the `with:` keys holding a tool version and what upgrade does with them.
var toolKeys = map[string]map[string]toolPolicy{
	"erlef/setup-beam":   {"gleam-version": toolBump, "otp-version": toolAdvise},
	"mlugg/setup-zig":    {"version": toolAdvise},
	"actions/setup-java": {"java-version": toolAdvise},
}

var (
	stepStartRe = regexp.MustCompile(`^\s*-\s`)
	withKeyRe   = regexp.MustCompile(`^(\s+)([a-z-]+):(\s*)(["']?)(\d+(?:\.\d+)*)(["']?)(.*)$`)
)

// ToolPins returns the tool versions forge's workflow templates set, keyed by action and `with:` key.
func ToolPins() (map[string]map[string]string, error) {
	pins := map[string]map[string]string{}
	err := fs.WalkDir(templateFS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.Contains(p, "/.github/workflows/") {
			return err
		}
		b, err := fs.ReadFile(templateFS, p)
		if err != nil {
			return err
		}
		action := ""
		for _, line := range strings.Split(string(b), "\n") {
			if m := usesRe.FindStringSubmatch(line); m != nil {
				action = m[2]
				continue
			}
			if stepStartRe.MatchString(line) {
				action = ""
				continue
			}
			m := withKeyRe.FindStringSubmatch(line)
			if m == nil || toolKeys[action] == nil {
				continue
			}
			if _, ok := toolKeys[action][m[2]]; !ok {
				continue
			}
			if pins[action] == nil {
				pins[action] = map[string]string{}
			}
			if cur := pins[action][m[2]]; cur == "" || compareVersions(m[5], cur) > 0 {
				pins[action][m[2]] = m[5]
			}
		}
		return nil
	})
	return pins, err
}

// compareVersions compares dotted numeric versions ("1.19.2" vs "1.20"): -1, 0 or 1.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

// upgradeWorkflow rewrites older action pins, and compatible tool versions, in one workflow file,
// leaving every other byte as it was. It also returns advice about tool versions it won't change.
func upgradeWorkflow(rel string, src []byte, pins map[string]ActionPin, tools map[string]map[string]string) ([]byte, []ActionChange, []string) {
	var (
		changes  []ActionChange
		advice   []string
		action   string
		lines    = strings.Split(string(src), "\n")
		modified bool
	)
	for i, line := range lines {
		if m := usesRe.FindStringSubmatch(line); m != nil {
			action = m[2]
			pin, ok := pins[m[2]]
			if major, _ := strconv.Atoi(m[4]); ok && major < pin.Major {
				changes = append(changes, ActionChange{File: rel, Line: i + 1, Action: m[2], From: m[3], To: pin.Ref})
				lines[i] = strings.Replace(line, m[2]+"@"+m[3], m[2]+"@"+pin.Ref, 1)
				modified = true
			}
			continue
		}
		if stepStartRe.MatchString(line) {
			action = ""
			continue
		}
		m := withKeyRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		policy, ok := toolKeys[action][m[2]]
		want := tools[action][m[2]]
		if !ok || want == "" || compareVersions(m[5], want) >= 0 {
			continue
		}
		label := action + " " + m[2]
		if policy == toolBump && strings.SplitN(m[5], ".", 2)[0] == strings.SplitN(want, ".", 2)[0] {
			changes = append(changes, ActionChange{File: rel, Line: i + 1, Action: label, From: m[5], To: want})
			lines[i] = m[1] + m[2] + ":" + m[3] + m[4] + want + m[6] + m[7]
			modified = true
			continue
		}
		advice = append(advice, fmt.Sprintf("%s:%d  %s %s -> %s is available; update it by hand (it can need code changes)", rel, i+1, label, m[5], want))
	}
	if !modified {
		return src, nil, advice
	}
	return []byte(strings.Join(lines, "\n")), changes, advice
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
