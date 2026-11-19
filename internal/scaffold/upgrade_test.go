package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestActionPinsComeFromTemplates(t *testing.T) {
	pins, err := ActionPins()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range []string{"actions/checkout", "actions/setup-go", "actions/setup-node", "actions/setup-python", "golangci/golangci-lint-action", "denoland/setup-deno", "mlugg/setup-zig"} {
		if p, ok := pins[a]; !ok || !strings.HasPrefix(p.Ref, "v") || p.Major < 1 {
			t.Errorf("missing or bad pin for %s: %+v", a, p)
		}
	}
}

func writeWorkflow(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, ".github", "workflows", name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestUpgradeChangesOnlyOldMajorPins(t *testing.T) {
	pins, _ := ActionPins()
	checkout := pins["actions/checkout"].Ref
	dir := t.TempDir()
	body := `jobs:
  a:
    steps:
      - uses: actions/checkout@v1   # a comment
      - uses: "actions/setup-go@v2.1.0"
      - uses: actions/checkout@v999
      - uses: actions/checkout@0123456789abcdef0123456789abcdef01234567
      - uses: actions/checkout@main
      - uses: someone/unknown-action@v1
      - uses: ./local-action
      - run: echo "uses: actions/checkout@v1 in a string is not a step"
`
	p := writeWorkflow(t, dir, "ci.yml", body)

	plan, err := PlanUpgrade(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Changes) != 2 {
		t.Fatalf("want 2 changes, got %+v", plan.Changes)
	}
	if c := plan.Changes[0]; c.Action != "actions/checkout" || c.From != "v1" || c.To != checkout || c.Line != 4 {
		t.Errorf("first change: %+v", c)
	}
	if c := plan.Changes[1]; c.Action != "actions/setup-go" || c.From != "v2.1.0" || c.Line != 5 {
		t.Errorf("second change: %+v", c)
	}

	// Planning writes nothing.
	if b, _ := os.ReadFile(p); string(b) != body {
		t.Fatal("PlanUpgrade modified the file")
	}

	if err := ApplyUpgrade(dir, plan); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	want := strings.Replace(body, "actions/checkout@v1   # a comment", "actions/checkout@"+checkout+"   # a comment", 1)
	want = strings.Replace(want, `"actions/setup-go@v2.1.0"`, `"actions/setup-go@`+pins["actions/setup-go"].Ref+`"`, 1)
	if string(got) != want {
		t.Errorf("unexpected result:\n%s\nwant:\n%s", got, want)
	}
	if info, _ := os.Stat(p); info.Mode().Perm() != 0o600 {
		t.Errorf("file mode changed to %v", info.Mode().Perm())
	}

	// A second run finds nothing to do.
	again, _ := PlanUpgrade(dir)
	if len(again.Changes) != 0 {
		t.Errorf("not idempotent: %+v", again.Changes)
	}
}

func TestUpgradeSuggestions(t *testing.T) {
	dir := t.TempDir()
	plan, err := PlanUpgrade(dir)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Workflows != 0 || strings.Join(plan.Suggestions, ",") != "ci,dependabot" {
		t.Errorf("empty project: %+v", plan)
	}

	writeWorkflow(t, dir, "ci.yaml", "jobs: {}\n")
	if err := os.WriteFile(filepath.Join(dir, ".github", "dependabot.yml"), []byte("version: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, _ = PlanUpgrade(dir)
	if plan.Workflows != 1 || len(plan.Suggestions) != 0 {
		t.Errorf("complete project: %+v", plan)
	}
}

func TestFreshProjectsAreAlreadyCurrent(t *testing.T) {
	// Every project forge generates today must need no upgrade.
	for _, pt := range Types() {
		dir := t.TempDir()
		files, err := Render(newOpts(t, pt.ID))
		if err != nil {
			t.Fatal(err)
		}
		if err := Write(dir, files, false); err != nil {
			t.Fatal(err)
		}
		plan, err := PlanUpgrade(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.Changes) != 0 {
			t.Errorf("%s: freshly generated project needs upgrades: %+v", pt.ID, plan.Changes)
		}
	}
}

func TestUpgradeToolVersions(t *testing.T) {
	tools, err := ToolPins()
	if err != nil {
		t.Fatal(err)
	}
	gleam := tools["erlef/setup-beam"]["gleam-version"]
	if gleam == "" || tools["mlugg/setup-zig"]["version"] == "" || tools["actions/setup-java"]["java-version"] == "" {
		t.Fatalf("missing tool pins: %+v", tools)
	}
	dir := t.TempDir()
	body := `jobs:
  t:
    steps:
      - uses: erlef/setup-beam@v1
        with:
          otp-version: "1"
          gleam-version: "1.0.0"   # keep
      - uses: mlugg/setup-zig@v2
        with:
          version: 0.0.1
      - uses: goreleaser/goreleaser-action@v7
        with:
          version: 0.0.1
      - uses: erlef/setup-beam@v1
        with:
          gleam-version: "0.34.0"
`
	p := writeWorkflow(t, dir, "ci.yml", body)
	plan, err := PlanUpgrade(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Only gleam 1.0.0 is bumped: same major. 0.34 -> 1.x is a major jump, so it is advice.
	if len(plan.Changes) != 1 || plan.Changes[0].Line != 7 || plan.Changes[0].To != gleam {
		t.Fatalf("changes: %+v", plan.Changes)
	}
	if len(plan.Advisories) != 3 {
		t.Fatalf("want advice for otp, zig and gleam 0.34, got: %q", plan.Advisories)
	}
	for _, a := range plan.Advisories {
		if strings.Contains(a, "goreleaser") {
			t.Errorf("unrelated action's version key reported: %s", a)
		}
	}
	if err := ApplyUpgrade(dir, plan); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	if !strings.Contains(string(got), `gleam-version: "`+gleam+`"   # keep`) || !strings.Contains(string(got), "version: 0.0.1") {
		t.Errorf("unexpected file:\n%s", got)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{{"1.19.2", "1.19.2", 0}, {"1.9", "1.19", -1}, {"2", "1.99.9", 1}, {"0.17.1", "0.17", 1}}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
