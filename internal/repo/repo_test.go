package repo

import (
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestDiscoverFromMain(t *testing.T) {
	dir := testutil.InitRepo(t)
	r, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	// EvalSymlinks because macOS temp dirs are symlinked (/var → /private/var).
	got, _ := filepath.EvalSymlinks(r.Main)
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("Main = %q, want %q", got, want)
	}
	if filepath.Base(r.GabDir()) != ".gab" {
		t.Fatalf("GabDir = %q", r.GabDir())
	}
}

func TestDiscoverFromFeatureWorktreeFindsMain(t *testing.T) {
	dir := testutil.InitRepo(t)
	wt := filepath.Join(t.TempDir(), "feature")
	if _, err := gitx.Run(dir, "worktree", "add", "-b", "gab/T1-x", wt, "main"); err != nil {
		t.Fatal(err)
	}
	r, err := Discover(wt) // discover from inside the feature worktree
	if err != nil {
		t.Fatal(err)
	}
	got, _ := filepath.EvalSymlinks(r.Main)
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("Main = %q, want main checkout %q", got, want)
	}
}
