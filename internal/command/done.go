package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Ownii/gitops-agent-backlog/internal/gitx"
	"github.com/Ownii/gitops-agent-backlog/internal/repo"
	"github.com/Ownii/gitops-agent-backlog/internal/ticket"
)

func Done(cwd, id string) error {
	r, err := repo.Discover(cwd)
	if err != nil {
		return err
	}
	tdir, folder, err := TicketDirByID(r, id)
	if err != nil {
		return err
	}
	m, err := ticket.ReadMeta(filepath.Join(tdir, "meta.yml"))
	if err != nil {
		return err
	}
	if m.Status != ticket.StatusToVerify {
		return fmt.Errorf("ticket %s is %q, must be %q for done", id, m.Status, ticket.StatusToVerify)
	}
	if m.Branch == "" {
		return fmt.Errorf("ticket %s has no branch recorded", id)
	}

	// Guard: done removes the ticket's worktree at the end. If it runs from
	// inside that worktree, it would delete the process's own cwd. Fail fast.
	wt := r.WorktreePath(folder.ID, folder.Slug)
	if isInside(cwd, wt) {
		return fmt.Errorf("run done from the main checkout (%s), not from inside the ticket's worktree (%s)", r.Main, wt)
	}

	// Precondition: main's worktree must be clean, so that a rollback
	// (git reset --hard) can safely restore it without discarding unrelated
	// work. Untracked files (e.g. an uncommitted freshly-`new`ed ticket) do
	// not block this — reset --hard leaves untracked files alone.
	dirty, err := gitx.Run(r.Main, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return err
	}
	if strings.TrimSpace(dirty) != "" {
		return fmt.Errorf("main worktree has uncommitted changes; commit or stash them before done:\n%s", dirty)
	}

	// Record the pre-merge commit so the atomic phase below can be rolled
	// back as a unit on any failure.
	startSHA, err := gitx.Run(r.Main, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	rollback := func(cause error) error {
		if _, rbErr := gitx.Run(r.Main, "reset", "-q", "--hard", startSHA); rbErr != nil {
			return fmt.Errorf("%w (rollback ALSO failed: %v — main may be left partially merged)", cause, rbErr)
		}
		// reset --hard restores tracked files but leaves untracked ones. If the
		// squash-merge already dumped the branch's .gab residue and a later step
		// unstaged it, drop it so a failed done leaves no stray files on main.
		for _, f := range []string{"BRIEF.md", "SUMMARY.md"} {
			os.Remove(filepath.Join(r.GabDir(), f))
		}
		return fmt.Errorf("done aborted, main rolled back to %s: %w", startSHA[:7], cause)
	}

	// --- Atomic phase: squash-merge + commits, rolled back together on error ---

	if _, err := gitx.Run(r.Main, "merge", "--squash", m.Branch); err != nil {
		return rollback(fmt.Errorf("squash-merge %s: %w", m.Branch, err))
	}
	// main owns .gab truth: drop any .gab changes the branch carried.
	if _, err := gitx.Run(r.Main, "reset", "-q", "--", ".gab"); err != nil {
		return rollback(err)
	}
	if _, err := gitx.Run(r.Main, "checkout", "--", ".gab"); err != nil {
		return rollback(err)
	}
	// Remove only the known branch-side .gab residue files (not the whole
	// .gab tree — an uncommitted freshly-`new`ed ticket may live there).
	for _, f := range []string{"BRIEF.md", "SUMMARY.md"} {
		if err := os.Remove(filepath.Join(r.GabDir(), f)); err != nil && !os.IsNotExist(err) {
			return rollback(fmt.Errorf("remove .gab residue %s: %w", f, err))
		}
	}
	// Commit the merged code — but only if something is staged. A branch whose
	// only changes were under .gab leaves nothing to commit here, and an empty
	// `git commit` would fail; skip it and go straight to archiving.
	staged, err := hasStagedChanges(r.Main)
	if err != nil {
		return rollback(err)
	}
	if staged {
		if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("feat: %s (%s)", m.Title, id)); err != nil {
			return rollback(fmt.Errorf("commit merged code: %w", err))
		}
	}

	// Archive the ticket folder to done/.
	if err := os.MkdirAll(r.DoneDir(), 0o755); err != nil {
		return rollback(fmt.Errorf("ensure done dir: %w", err))
	}
	dest := filepath.Join(r.DoneDir(), folder.Name)
	if _, err := gitx.Run(r.Main, "mv", tdir, dest); err != nil {
		return rollback(fmt.Errorf("archive ticket %s: %w", id, err))
	}
	if _, err := gitx.Run(r.Main, "commit", "-m", fmt.Sprintf("chore(gab): archive %s", id)); err != nil {
		return rollback(fmt.Errorf("commit archive: %w", err))
	}

	// --- Best-effort cleanup: past the point of no return. The merge is
	// committed; a cleanup failure must NOT roll back real work, so warn. ---

	if _, err := gitx.Run(r.Main, "worktree", "remove", "--force", wt); err != nil {
		fmt.Fprintf(os.Stderr, "gab: warning: could not remove worktree %s: %v\n", wt, err)
	}
	if _, err := gitx.Run(r.Main, "branch", "-D", m.Branch); err != nil {
		fmt.Fprintf(os.Stderr, "gab: warning: could not delete branch %s: %v\n", m.Branch, err)
	}
	return nil
}

// hasStagedChanges reports whether dir has anything staged in the index.
func hasStagedChanges(dir string) (bool, error) {
	out, err := gitx.Run(dir, "diff", "--cached", "--name-only")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}
