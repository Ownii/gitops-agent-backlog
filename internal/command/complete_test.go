package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
	"github.com/Ownii/gitops-agent-backlog/internal/testutil"
)

func TestCompleteFlowsSummaryAndSetsToVerify(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	if err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")

	// Agent does work: writes summary + a source file, commits in the worktree.
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("done\n"), 0o644)
	os.MkdirAll(filepath.Join(wt, ".gab"), 0o755)
	os.WriteFile(filepath.Join(wt, ".gab", "SUMMARY.md"), []byte("## Summary\nno deviations\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "implement login")

	// complete is invoked from the worktree
	if err := Complete(wt, "T1"); err != nil {
		t.Fatal(err)
	}

	tdir, _, _ := TicketDirByID(r, "T1")
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusToVerify {
		t.Fatalf("status = %q, want to-verify", m.Status)
	}
	if _, err := os.Stat(filepath.Join(tdir, "summary.md")); err != nil {
		t.Fatalf("summary not flowed back: %v", err)
	}
}

func TestCompleteRejectsDirtyWorktree(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	os.WriteFile(filepath.Join(wt, "dirty.txt"), []byte("x"), 0o644) // uncommitted
	if err := Complete(wt, "T1"); err == nil {
		t.Fatal("expected error for dirty worktree")
	}
}

func TestCompleteWithoutSummaryStillSetsToVerify(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	if err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")

	// Agent does work: writes source file WITHOUT .gab/SUMMARY.md, commits in the worktree.
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("done\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "implement login without summary")

	// complete is invoked from the worktree
	if err := Complete(wt, "T1"); err != nil {
		t.Fatal(err)
	}

	// Assert: status is to-verify and summary.md was NOT created in main
	tdir, _, _ := TicketDirByID(r, "T1")
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusToVerify {
		t.Fatalf("status = %q, want to-verify", m.Status)
	}
	if _, err := os.Stat(filepath.Join(tdir, "summary.md")); err == nil {
		t.Fatal("summary.md should NOT exist when no SUMMARY.md was written in worktree")
	} else if !os.IsNotExist(err) {
		t.Fatalf("unexpected error checking summary.md: %v", err)
	}
}

func TestCompleteRejectsNonInProgress(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	// Do NOT call Start, so status remains "planned"

	// Try to complete a ticket that is not in-progress
	if err := Complete(dir, "T1"); err == nil {
		t.Fatal("expected error for non-in-progress ticket")
	}
}
