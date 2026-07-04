package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

// seedPlanned creates a committed planned ticket with spec.md + plan.md.
func seedPlanned(t *testing.T, dir, id, slug string) string {
	t.Helper()
	r, _ := repo.Discover(dir)
	if err := EnsureGab(r); err != nil {
		t.Fatal(err)
	}
	tdir := filepath.Join(r.TicketsDir(), ticket.FormatFolder(10, id, slug))
	if err := os.MkdirAll(tdir, 0o755); err != nil {
		t.Fatal(err)
	}
	ticket.WriteMeta(filepath.Join(tdir, "meta.yml"), &ticket.Meta{ID: id, Title: slug, Status: ticket.StatusPlanned})
	os.WriteFile(filepath.Join(tdir, "spec.md"), []byte("## Spec\nlogin\n"), 0o644)
	os.WriteFile(filepath.Join(tdir, "plan.md"), []byte("## Plan\nstep 1\n"), 0o644)
	if _, err := gitx.Run(dir, "add", "-A"); err != nil {
		t.Fatal(err)
	}
	if _, err := gitx.Run(dir, "commit", "-m", "seed "+id); err != nil {
		t.Fatal(err)
	}
	return tdir
}

func TestStartCreatesWorktreeBriefAndStatus(t *testing.T) {
	dir := testutil.InitRepo(t)
	tdir := seedPlanned(t, dir, "T1", "login")

	if err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	// status flipped on main
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusInProgress || m.Branch != "gab/T1-login" {
		t.Fatalf("meta after start = %+v", m)
	}
	// worktree exists with a committed BRIEF.md
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	brief, err := os.ReadFile(filepath.Join(wt, ".gab", "BRIEF.md"))
	if err != nil {
		t.Fatalf("BRIEF.md missing in worktree: %v", err)
	}
	if !strings.Contains(string(brief), "login") || !strings.Contains(string(brief), "Plan") {
		t.Fatalf("brief missing content: %s", brief)
	}
	// branch is committed (brief commit present)
	if _, err := gitx.Run(wt, "rev-parse", "gab/T1-login"); err != nil {
		t.Fatalf("branch not found: %v", err)
	}

	// Assert the core invariant: brief commit touches ONLY .gab/BRIEF.md (not .gab/tickets/*)
	changedFiles, err := gitx.Run(wt, "show", "--stat", "--name-only", "--pretty=format:", "HEAD")
	if err != nil {
		t.Fatalf("failed to get changed files: %v", err)
	}
	changedFilesList := strings.Fields(string(changedFiles))
	if len(changedFilesList) != 1 || changedFilesList[0] != ".gab/BRIEF.md" {
		t.Fatalf("brief commit should touch only .gab/BRIEF.md, got: %v", changedFilesList)
	}

	// Assert BRIEF.md does NOT contain a status: line (it is statusless)
	if strings.Contains(string(brief), "status:") {
		t.Fatalf("BRIEF.md should not contain status: line, got: %s", brief)
	}
}

func TestStartRejectsNonPlanned(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	// force status back to todo
	r, _ := repo.Discover(dir)
	tdir, _, _ := TicketDirByID(r, "T1")
	ticket.WriteMeta(filepath.Join(tdir, "meta.yml"), &ticket.Meta{ID: "T1", Status: ticket.StatusTodo})
	if err := Start(dir, "T1"); err == nil {
		t.Fatal("expected error starting a non-planned ticket")
	}
}

func TestStartRejectsExistingWorktree(t *testing.T) {
	dir := testutil.InitRepo(t)
	tdir := seedPlanned(t, dir, "T1", "login")

	// First call should succeed
	if err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	// Reset status back to planned for second call
	ticket.WriteMeta(filepath.Join(tdir, "meta.yml"), &ticket.Meta{ID: "T1", Status: ticket.StatusPlanned, Title: "login"})
	if _, err := gitx.Run(dir, "add", filepath.Join(tdir, "meta.yml")); err != nil {
		t.Fatal(err)
	}
	if _, err := gitx.Run(dir, "commit", "-m", "reset T1 to planned"); err != nil {
		t.Fatal(err)
	}

	// Second call should fail because worktree/branch already exists
	if err := Start(dir, "T1"); err == nil {
		t.Fatal("expected error starting when worktree/branch already exists")
	}
}
