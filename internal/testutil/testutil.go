package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// git runs a git command in dir and fails the test on error.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// InitRepo creates a temp repo with one commit on branch main.
func InitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-b", "main")
	git(t, dir, "config", "user.email", "test@example.com")
	git(t, dir, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", "initial")
	return dir
}

// AddBareOrigin creates a bare repo and wires it as origin of dir.
func AddBareOrigin(t *testing.T, dir string) string {
	t.Helper()
	bare := t.TempDir()
	git(t, bare, "init", "--bare", "-b", "main")
	git(t, dir, "remote", "add", "origin", bare)
	return bare
}
