package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGeneratedDockerfilesBuildAndRun builds the Dockerfile of every project type that has one,
// runs the image, and checks it runs as a non-root user. It needs a working Docker daemon and
// pulls base images, so it is opt-in: set FORGE_DOCKER_TESTS=1 (or run `make test-docker`).
func TestGeneratedDockerfilesBuildAndRun(t *testing.T) {
	if os.Getenv("FORGE_DOCKER_TESTS") != "1" {
		t.Skip("set FORGE_DOCKER_TESTS=1 to run Docker tests")
	}
	if err := exec.Command("docker", "info").Run(); err != nil {
		t.Skipf("docker daemon not reachable: %v", err)
	}

	cases := []struct {
		typ     string
		args    []string // container args
		wantOut string
	}{
		{"go-cli", []string{"-version"}, "dkr-go-cli 1.2.3 (abc1234)"},
		{"python", []string{"Ada"}, "Hello, Ada!"},
		{"node", []string{"Ada"}, "Hello, Ada!"},
		{"rust", []string{"Ada"}, "Hello, Ada!"},
		{"typescript", []string{"Ada"}, "Hello, Ada!"},
		{"deno", []string{"Ada"}, "Hello, Ada!"},
	}
	for _, tc := range cases {
		t.Run(tc.typ, func(t *testing.T) {
			name := "dkr-" + tc.typ
			parent := t.TempDir()
			if out, err := run(t, "new", name, "-t", tc.typ, "-o", parent, "--no-git", "-y", "-m", "example.com/"+name); err != nil {
				t.Fatalf("%v\n%s", err, out)
			}
			dir := filepath.Join(parent, name)
			t.Chdir(dir)
			if out, err := run(t, "add", "dockerfile"); err != nil {
				t.Fatalf("%v\n%s", err, out)
			}

			image := "forge-test/" + name
			build := exec.Command("docker", "build", "-q", "-t", image,
				"--build-arg", "VERSION=1.2.3", "--build-arg", "COMMIT=abc1234", ".")
			build.Dir = dir
			if out, err := build.CombinedOutput(); err != nil {
				t.Fatalf("docker build: %v\n%s", err, out)
			}
			t.Cleanup(func() { _ = exec.Command("docker", "rmi", "-f", image).Run() })

			out, err := exec.Command("docker", append([]string{"run", "--rm", image}, tc.args...)...).CombinedOutput()
			if err != nil {
				t.Fatalf("docker run: %v\n%s", err, out)
			}
			if !strings.Contains(string(out), tc.wantOut) {
				t.Errorf("output %q does not contain %q", out, tc.wantOut)
			}

			user, err := exec.Command("docker", "image", "inspect", "-f", "{{.Config.User}}", image).Output()
			if err != nil {
				t.Fatal(err)
			}
			if u := strings.TrimSpace(string(user)); u == "" || u == "root" || u == "0" {
				t.Errorf("image runs as root (User=%q)", u)
			}
		})
	}
}
