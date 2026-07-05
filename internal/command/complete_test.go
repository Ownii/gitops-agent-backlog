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

func TestCompleteFlowsSummaryAndSetsToVerify(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	if _, err := Start(dir, "T1"); err != nil {
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
	if _, err := Start(dir, "T1"); err != nil {
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

func TestCompleteRejectsWrongBranch(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	if _, err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	// Invoke complete from main instead of the ticket's worktree. main is clean
	// and there is no SUMMARY.md, so the old code would silently advance the
	// ticket to to-verify without any work having happened.
	if err := Complete(dir, "T1"); err == nil {
		t.Fatal("expected error completing from the wrong branch (main)")
	}

	r, _ := repo.Discover(dir)
	tdir, _, _ := TicketDirByID(r, "T1")
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusInProgress {
		t.Fatalf("status = %q, want in-progress (unchanged)", m.Status)
	}
}

func TestCompleteDoesNotCommitForeignStagedChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	if _, err := Start(dir, "T1"); err != nil {
		t.Fatal(err)
	}
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")

	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("done\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "implement login")

	// User stages unrelated work on main before completing.
	os.WriteFile(filepath.Join(dir, "unrelated.txt"), []byte("wip\n"), 0o644)
	if _, err := gitx.Run(dir, "add", "unrelated.txt"); err != nil {
		t.Fatal(err)
	}

	if err := Complete(wt, "T1"); err != nil {
		t.Fatal(err)
	}

	files, err := gitx.Run(dir, "show", "--name-only", "--pretty=format:", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(files, "unrelated.txt") {
		t.Fatalf("to-verify commit swept up foreign staged file:\n%s", files)
	}
	staged, _ := gitx.Run(dir, "diff", "--cached", "--name-only")
	if !strings.Contains(staged, "unrelated.txt") {
		t.Fatalf("foreign staged change was lost; staged=%q", staged)
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
