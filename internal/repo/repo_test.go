package repo

import (
	"os"
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

func TestWorktreePathIncludesRepoName(t *testing.T) {
	// Sibling of the repo, namespaced by repo name so two gab repos in the same
	// parent directory cannot collide on the same id+slug.
	r := &Repo{Main: filepath.Join("/home", "dev", "myrepo"), Trunk: "main"}
	got := r.WorktreePath("T1", "login")
	want := filepath.Join("/home", "dev", ".gab-worktrees", "myrepo", "T1-login")
	if got != want {
		t.Fatalf("WorktreePath = %q, want %q", got, want)
	}
}

func TestDiscoverDetectsNonMainTrunk(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-b", "master"},
		{"config", "user.email", "t@example.com"},
		{"config", "user.name", "T"},
	} {
		if _, err := gitx.Run(dir, args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := gitx.Run(dir, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := gitx.Run(dir, "commit", "-m", "init"); err != nil {
		t.Fatal(err)
	}

	r, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if r.Trunk != "master" {
		t.Fatalf("Trunk = %q, want master", r.Trunk)
	}
	got, _ := filepath.EvalSymlinks(r.Main)
	want, _ := filepath.EvalSymlinks(dir)
	if got != want {
		t.Fatalf("Main = %q, want %q", got, want)
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
