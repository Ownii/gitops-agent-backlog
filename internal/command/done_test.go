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

func TestDoneMergesArchivesAndCleansUp(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("feature\n"), 0o644)
	os.MkdirAll(filepath.Join(wt, ".gab"), 0o755)
	os.WriteFile(filepath.Join(wt, ".gab", "SUMMARY.md"), []byte("ok\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "impl")
	Complete(wt, "T1")

	if err := Done(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	// code merged into main
	if _, err := os.Stat(filepath.Join(dir, "app.txt")); err != nil {
		t.Fatalf("feature file not merged: %v", err)
	}
	// ticket archived, not active
	if _, _, err := TicketDirByID(r, "T1"); err == nil {
		t.Fatal("ticket should no longer be active")
	}
	if _, err := os.Stat(filepath.Join(r.DoneDir(), "010-T1-login")); err != nil {
		t.Fatalf("ticket not archived to done/: %v", err)
	}
	// main's .gab was not polluted by the branch's BRIEF.md or SUMMARY.md
	if _, err := os.Stat(filepath.Join(dir, ".gab", "BRIEF.md")); !os.IsNotExist(err) {
		t.Fatalf("BRIEF.md leaked into main .gab")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gab", "SUMMARY.md")); !os.IsNotExist(err) {
		t.Fatalf("SUMMARY.md leaked into main .gab")
	}
	// archived ticket's status should still be to-verify (not reverted by .gab discard)
	archivedMeta, err := ticket.ReadMeta(filepath.Join(r.DoneDir(), "010-T1-login", "meta.yml"))
	if err != nil {
		t.Fatalf("failed to read archived ticket meta: %v", err)
	}
	if archivedMeta.Status != ticket.StatusToVerify {
		t.Fatalf("archived ticket status should be %q, got %q", ticket.StatusToVerify, archivedMeta.Status)
	}
	// worktree + branch removed
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("worktree not removed")
	}
	if _, err := gitx.Run(dir, "rev-parse", "--verify", "gab/T1-login"); err == nil {
		t.Fatal("branch should be deleted")
	}
}

// TestDoneKeepsUncommittedTicket ensures Done's narrowed .gab residue removal
// does not sweep an uncommitted freshly-`new`ed ticket folder off main.
func TestDoneKeepsUncommittedTicket(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("feature\n"), 0o644)
	os.MkdirAll(filepath.Join(wt, ".gab"), 0o755)
	os.WriteFile(filepath.Join(wt, ".gab", "SUMMARY.md"), []byte("ok\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "impl")
	Complete(wt, "T1")

	// A second ticket, created but never committed (New does not commit).
	otherDir, err := New(dir, "other")
	if err != nil {
		t.Fatal(err)
	}

	if err := Done(dir, "T1"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(otherDir); err != nil {
		t.Fatalf("uncommitted ticket folder was swept by done: %v", err)
	}
}

// TestDoneRejectsDirtyMain ensures Done refuses to run when main has
// uncommitted tracked changes, so its rollback can't clobber unrelated work.
func TestDoneRejectsDirtyMain(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	os.WriteFile(filepath.Join(wt, "app.txt"), []byte("feature\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "impl")
	Complete(wt, "T1")

	// Dirty a tracked file on main.
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# changed\n"), 0o644)

	if err := Done(dir, "T1"); err == nil {
		t.Fatal("expected Done to reject a dirty main worktree")
	}
	// Ticket must remain active and unchanged.
	if _, _, err := TicketDirByID(r, "T1"); err != nil {
		t.Fatalf("ticket should still be active after rejected done: %v", err)
	}
}

// TestDoneRollsBackOnMergeConflict ensures a squash-merge conflict leaves main
// exactly as it was (no partial merge) and the ticket still active.
func TestDoneRollsBackOnMergeConflict(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	// Branch side adds conflict.txt.
	os.WriteFile(filepath.Join(wt, "conflict.txt"), []byte("branch\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "branch change")
	Complete(wt, "T1")

	// Main side adds the same file with different content → add/add conflict.
	os.WriteFile(filepath.Join(dir, "conflict.txt"), []byte("main\n"), 0o644)
	gitx.Run(dir, "add", "-A")
	gitx.Run(dir, "commit", "-m", "main change")

	startSHA, _ := gitx.Run(dir, "rev-parse", "HEAD")

	if err := Done(dir, "T1"); err == nil {
		t.Fatal("expected Done to fail on a squash-merge conflict")
	}

	// Main rolled back: HEAD unchanged, working tree clean.
	nowSHA, _ := gitx.Run(dir, "rev-parse", "HEAD")
	if nowSHA != startSHA {
		t.Fatalf("main HEAD moved after failed done: %s != %s", nowSHA, startSHA)
	}
	if st, _ := gitx.Run(dir, "status", "--porcelain"); st != "" {
		t.Fatalf("main worktree not clean after rollback:\n%s", st)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "conflict.txt")); string(got) != "main\n" {
		t.Fatalf("conflict.txt content clobbered: %q", got)
	}
	// Ticket still active and to-verify.
	tdir, _, err := TicketDirByID(r, "T1")
	if err != nil {
		t.Fatalf("ticket should still be active after rollback: %v", err)
	}
	m, _ := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if m.Status != ticket.StatusToVerify {
		t.Fatalf("ticket status changed after rollback: %q", m.Status)
	}
}

// TestDoneWithOnlyGabChanges ensures a branch whose only changes are under
// .gab (no code diff) still completes: the empty code commit is skipped, and
// the ticket is archived.
func TestDoneWithOnlyGabChanges(t *testing.T) {
	dir := testutil.InitRepo(t)
	seedPlanned(t, dir, "T1", "login")
	Start(dir, "T1")
	r, _ := repo.Discover(dir)
	wt := r.WorktreePath("T1", "login")
	// Only a .gab-scoped change on the branch — no code files.
	os.MkdirAll(filepath.Join(wt, ".gab"), 0o755)
	os.WriteFile(filepath.Join(wt, ".gab", "SUMMARY.md"), []byte("notes only\n"), 0o644)
	gitx.Run(wt, "add", "-A")
	gitx.Run(wt, "commit", "-m", "notes")
	Complete(wt, "T1")

	if err := Done(dir, "T1"); err != nil {
		t.Fatalf("done should succeed with only .gab changes: %v", err)
	}
	if _, err := os.Stat(filepath.Join(r.DoneDir(), "010-T1-login")); err != nil {
		t.Fatalf("ticket not archived: %v", err)
	}
	if _, _, err := TicketDirByID(r, "T1"); err == nil {
		t.Fatal("ticket should no longer be active")
	}
}
