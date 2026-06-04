package gsuite_mcp

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallWritesHarnessAndCLIBinaries(t *testing.T) {
	fixture := newInstallFixture(t)

	output := fixture.runInstall(t)
	if !strings.Contains(output, "harness binary: "+fixture.harnessBinary) {
		t.Fatalf("install output did not report harness binary %q:\n%s", fixture.harnessBinary, output)
	}
	if !strings.Contains(output, "cli binary:     "+fixture.cliBinary) {
		t.Fatalf("install output did not report cli binary %q:\n%s", fixture.cliBinary, output)
	}

	assertBinaryReportsHead(t, fixture.harnessBinary, fixture.gitHead)
	assertBinaryReportsHead(t, fixture.cliBinary, fixture.gitHead)

	output = fixture.runInstall(t)
	if !strings.Contains(output, "nothing to do") {
		t.Fatalf("second install should no-op when both binaries match HEAD, got:\n%s", output)
	}
}

func TestInstallRefreshesStaleHarnessBinaryDespiteMatchingStamp(t *testing.T) {
	fixture := newInstallFixture(t)
	fixture.runInstall(t)

	staleScript := "#!/usr/bin/env bash\nprintf 'gsuite-mcp 0.0.0 (stale)\\n'\n"
	if err := os.WriteFile(fixture.harnessBinary, []byte(staleScript), 0755); err != nil {
		t.Fatalf("write stale harness binary: %v", err)
	}

	output := fixture.runInstall(t)
	if !strings.Contains(output, "Building gsuite-mcp") {
		t.Fatalf("expected stale harness binary to force rebuild, got:\n%s", output)
	}

	assertBinaryReportsHead(t, fixture.harnessBinary, fixture.gitHead)
	assertBinaryReportsHead(t, fixture.cliBinary, fixture.gitHead)
}

type installFixture struct {
	dir           string
	installPrefix string
	xdgDataHome   string
	harnessBinary string
	cliBinary     string
	gitHead       string
}

func newInstallFixture(t *testing.T) installFixture {
	t.Helper()

	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	installPrefix := filepath.Join(tempDir, "prefix")
	xdgDataHome := filepath.Join(tempDir, "data")

	mkdirAll(t, filepath.Join(repoDir, "cmd", "gsuite-mcp"))

	installScript, err := os.ReadFile("install.sh")
	if err != nil {
		t.Fatalf("read install.sh: %v", err)
	}
	writeFile(t, filepath.Join(repoDir, "install.sh"), installScript, 0755)
	writeFile(t, filepath.Join(repoDir, "go.mod"), []byte("module example.com/gsuite-install-fixture\n\ngo 1.23\n"), 0644)
	writeFile(t, filepath.Join(repoDir, "cmd", "gsuite-mcp", "main.go"), []byte(`package main

import (
	"fmt"
	"os"
)

var GitCommit string

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		if GitCommit != "" {
			fmt.Printf("gsuite-mcp 0.0.0 (%s)\n", GitCommit)
			return
		}
		fmt.Println("gsuite-mcp 0.0.0")
	}
}
`), 0644)

	runInDir(t, repoDir, "git", "init")
	runInDir(t, repoDir, "git", "config", "user.email", "test@example.com")
	runInDir(t, repoDir, "git", "config", "user.name", "Install Test")
	runInDir(t, repoDir, "git", "add", ".")
	runInDir(t, repoDir, "git", "commit", "-m", "initial")
	gitHead := strings.TrimSpace(runInDir(t, repoDir, "git", "rev-parse", "--short", "HEAD"))

	return installFixture{
		dir:           repoDir,
		installPrefix: installPrefix,
		xdgDataHome:   xdgDataHome,
		harnessBinary: filepath.Join(installPrefix, "libexec", "gsuite-mcp"),
		cliBinary:     filepath.Join(installPrefix, "bin", "gsuite-mcp"),
		gitHead:       gitHead,
	}
}

func (f installFixture) runInstall(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("./install.sh")
	cmd.Dir = f.dir
	cmd.Env = append(os.Environ(),
		"INSTALL_PREFIX="+f.installPrefix,
		"XDG_DATA_HOME="+f.xdgDataHome,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run install.sh: %v\n%s", err, output)
	}
	return string(output)
}

func assertBinaryReportsHead(t *testing.T, binary, gitHead string) {
	t.Helper()

	output := runInDir(t, ".", binary, "version")
	want := "(" + gitHead + ")"
	if !strings.Contains(output, want) {
		t.Fatalf("%s version output = %q, want commit %s", binary, output, want)
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path string, contents []byte, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, contents, mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runInDir(t *testing.T, dir, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
	return string(output)
}
